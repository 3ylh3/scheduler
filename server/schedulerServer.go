package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/3ylh3/scheduler/common"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorhill/cronexpr"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	// 配置文件
	var conf string
	flag.StringVar(&conf, "conf", "./config.yaml", "config file,default:./config.yaml")
	flag.Parse()
	if "" == conf {
		fmt.Println("config file can not be empty,use -conf to declare")
		os.Exit(1)
	}
	// 解析配置文件
	file, err := ioutil.ReadFile(conf)
	if err != nil {
		fmt.Printf("read config file error:%v\n", err)
		os.Exit(1)
	}
	c := common.Config{}
	err = yaml.Unmarshal(file, &c)
	if err != nil {
		fmt.Printf("parse config file error:%v\n", err)
		os.Exit(1)
	}
	if "" == c.Etcd.Address {
		fmt.Println("etcd address can not be empty")
		os.Exit(1)
	}
	// 初始化etcd客户端
	var config = clientv3.Config{
		Endpoints:   []string{c.Etcd.Address},
		DialTimeout: 5 * time.Second,
	}
	client, err := clientv3.New(config)
	if err != nil {
		fmt.Printf("init etcd client error:%v\n", err)
		os.Exit(1)
	}
	defer client.Close()
	kv := clientv3.NewKV(client)
	// 初始化mysql连接
	db, err := sql.Open("mysql", c.Mysql.User+":"+c.Mysql.Password+"@tcp("+
		c.Mysql.Host+":"+strconv.Itoa(c.Mysql.Port)+")/"+c.Mysql.Database)
	if nil != err {
		fmt.Printf("connect to mysql error:%v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	// 设置连接池参数
	// 最大空闲连接数
	db.SetMaxIdleConns(10)
	// 最大连接数
	db.SetMaxOpenConns(100)
	// 设置连接最大复用时间
	db.SetConnMaxLifetime(time.Hour * 1)
	// 定时每小时检查是否有异常状态的任务
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	go checkJobs(ticker, client, db)
	// 无限循环获取etcd中的job信息，若达到触发时间则挂到etcd对应的agent节点下
	for {
		resp, err := kv.Get(context.TODO(), "/jobs/", clientv3.WithPrefix())
		if err != nil {
			fmt.Printf("get jobs error:%v\n", err)
			continue
		}
		if 0 == resp.Count {
			// sleep 500毫秒，减少cpu压力
			time.Sleep(500 * time.Millisecond)
			// 暂时没有任务，继续下个循环
			continue
		}
		// 循环检查每一个job
		for _, value := range resp.Kvs {
			job := common.Job{}
			err := json.Unmarshal(value.Value, &job)
			if err != nil {
				continue
			}
			if job.Status == 0 || job.Status == 3 || job.Status == 2 {
				// 任务已冻结、异常或者正在执行中
				continue
			}
			go processJob(job.Name, client, kv, db)
		}
		// sleep 200毫秒，减少cpu压力
		time.Sleep(200 * time.Millisecond)
	}
}

// 检查是否有异常状态的任务
func checkJobs(ticker *time.Ticker, client *clientv3.Client, db *sql.DB) {
	for {
		<-ticker.C
		doCheckJobs(client, db)
	}
}

func doCheckJobs(client *clientv3.Client, db *sql.DB) {
	kv := clientv3.NewKV(client)
	resp, err := kv.Get(context.TODO(), "/jobs/", clientv3.WithPrefix())
	if err != nil {
		fmt.Printf("check jobs error:%v\n", err)
		return
	}
	if 0 == resp.Count {
		// 暂时没有任务
		return
	}
	for _, value := range resp.Kvs {
		job := common.Job{}
		err := json.Unmarshal(value.Value, &job)
		if err != nil {
			continue
		}
		if job.Status != 2 {
			// 任务未在执行中
			continue
		}
		// 检查是否有正在执行但是agent异常的任务
		checkAgent(&job, client)
		// 检查是否有执行完成但是server异常未及时更新状态的任务
		checkServer(&job, client, db)
	}
}

// 检查是否有执行完成但是server异常未及时更新状态的任务
func checkAgent(job *common.Job, client *clientv3.Client) {
	// 获取/jobs/${jobName}/execute节点数据
	kv := clientv3.NewKV(client)
	exeRsp, err := kv.Get(context.TODO(), "/jobs/"+job.Name+"/execute")
	if err != nil {
		return
	}
	if 0 == exeRsp.Count {
		// job执行完毕了
		return
	}
	execute := common.Execute{}
	err = json.Unmarshal(exeRsp.Kvs[0].Value, &execute)
	if err != nil {
		return
	}
	isHealth := true
	agentIps := execute.TotalServers
	for _, ip := range agentIps {
		// 判断agent是否存活
		agentRsp, err := kv.Get(context.TODO(), "/agents/"+ip)
		if err != nil {
			continue
		}
		if 0 == agentRsp.Count {
			// agent未存活
			isHealth = false
			break
		}
	}
	if !isHealth {
		// 任务状态异常，更新状态并删除/jobs/${jobName}/execute节点
		job.Status = 3
		job.ModifyTime = time.Now()
		bytes, err := json.Marshal(job)
		if err != nil {
			return
		}
		jobDetail := string(bytes)
		kv.Put(context.TODO(), "/jobs/"+job.Name, jobDetail, clientv3.WithPrevKV())
		kv.Delete(context.TODO(), "/jobs/"+job.Name+"/execute")
	}
}

// 检查是否有执行完成但是server异常未及时更新状态的任务
func checkServer(job *common.Job, client *clientv3.Client, db *sql.DB) {
	// 获取/jobs/${jobName}/execute节点数据
	kv := clientv3.NewKV(client)
	rsp, err := kv.Get(context.TODO(), "/jobs/"+job.Name+"/execute")
	if err != nil {
		return
	}
	// 检查所有agent上的任务执行情况
	execute := common.Execute{}
	err = json.Unmarshal(rsp.Kvs[0].Value, &execute)
	if err != nil {
		fmt.Printf("%s:watch job status error:%v,job name:%s\n", time.Now(), err, job.Name)
		return
	}
	if len(execute.TotalServers) == len(execute.SuccessServers)+len(execute.FailedServers) {
		// 所有agent均执行完成超过10分钟但是节点仍然存在，此时认为server发生了故障
		if execute.ExecutedTime-time.Now().Unix() > 600 {
			// 更新任务状态
			job.Status = 1
			if len(execute.FailedServers) == 0 {
				job.LastExecuteStatus = "success"
			} else {
				job.LastExecuteStatus = "failed"
				job.LastFailedServers = execute.FailedServers
			}
			job.LastSuccessServers = execute.SuccessServers
			job.LastExecuteServers = execute.TotalServers
			// 更新下次执行时间
			expr, err := cronexpr.Parse(job.Cron)
			if err != nil {
				return
			}
			timestamp := expr.Next(time.Now()).Unix() * 1000
			job.NextExecuteTime = timestamp
			postProcess(job, client, &execute, db)
		}
	}
}

func processJob(jobName string, client *clientv3.Client, kv clientv3.KV, db *sql.DB) {
	// 初步判断是否有锁节点，有锁节点则直接退出，避免锁节点过多
	resp, err := kv.Get(context.TODO(), "/jobs/"+jobName+"/", clientv3.WithPrefix())
	if err != nil {
		fmt.Printf("query lock entry error:%v\n", err)
		return
	}
	if resp.Count >= 1 {
		// 已经有锁节点
		return
	}
	// 加分布式锁
	session, err := concurrency.NewSession(client, concurrency.WithTTL(1))
	if err != nil {
		fmt.Printf("create session error:%v\n", err)
		return
	}
	defer session.Close()
	lock := concurrency.NewLocker(session, "/jobs/"+jobName)
	lock.Lock()
	defer lock.Unlock()
	// 加锁后获取job信息
	job := common.Job{}
	resp, err = kv.Get(context.TODO(), "/jobs/"+jobName)
	if err != nil {
		fmt.Printf("%s:get job info error:%v,job name:%s\n", time.Now(), err, jobName)
		return
	}
	err = json.Unmarshal(resp.Kvs[0].Value, &job)
	if err != nil {
		fmt.Printf("%s:get job info error:%v,job name:%s\n", time.Now(), err, jobName)
		return
	}
	// 检查是否到达触发时间（因为循环遍历任务时等待了200ms，为防止漏执行，这里判断的时候需要考虑到200ms的等待时间）
	now := time.Now().Unix() * 1000
	if job.NextExecuteTime <= now && (job.NextExecuteTime+200) >= now {
		// 到达触发时间
		if job.Status == 2 {
			// 任务已经在执行中，退出
			return
		}
		// 选择执行服务器
		var ips []string
		if job.ScheduleType == 1 {
			ips = job.ExecuteServers
		} else {
			// 生成真随机数
			randNum := 0
			result, err := rand.Int(rand.Reader, big.NewInt(int64(len(job.ExecuteServers))))
			if err == nil {
				randNum, _ = strconv.Atoi(result.String())
			}
			ips = append(ips, job.ExecuteServers[randNum])
		}
		var scheduledTime time.Time
		var executingIps []string
		// 筛选执行ip
		for _, ip := range ips {
			// 检查agent是否存在
			rsp, err := kv.Get(context.TODO(), "/agents/"+ip)
			if err != nil {
				fmt.Printf("%s:check agents error:%v,job name:%s\n", time.Now(), err, job.Name)
				return
			}
			if rsp.Count == 0 {
				fmt.Printf("%s:not find agent:%s,job name:%s\n", time.Now(), ip, job.Name)
				return
			}
			executingIps = append(executingIps, ip)
		}
		// 将执行信息挂到/jobs/${jobName}/execute节点下
		execute := common.Execute{
			TotalServers:   executingIps,
			SuccessServers: []string{},
			FailedServers:  []string{},
		}
		executeBytes, err := json.Marshal(execute)
		if err != nil {
			fmt.Printf("%s:schedule error:%v,job name:%s\n", time.Now(), err, job.Name)
			return
		}
		_, err = kv.Put(context.TODO(), "/jobs/"+job.Name+"/execute", string(executeBytes), clientv3.WithPrevKV())
		if err != nil {
			fmt.Printf("%s:schedule error:%v,job name:%s\n", time.Now(), err, job.Name)
			return
		}
		// 更新任务状态为执行中
		job.Status = 2
		// 更新上次执行时间
		job.LastExecuteTime = time.Now().String()
		bytes, err := json.Marshal(job)
		if err != nil {
			fmt.Printf("parse job data error:%v\n", err)
			return
		}
		jobDetail := string(bytes)
		_, err = kv.Put(context.TODO(), "/jobs/"+job.Name, jobDetail, clientv3.WithPrevKV())
		if err != nil {
			fmt.Printf("%s:updatejob job status error:%v,job name:%s\n", time.Now(), err, job.Name)
			return
		}
		// 监听任务执行情况
		go watchStatus(job, executingIps, client, db)
		for _, ip := range executingIps {
			// 将任务信息挂在执行ip节点下
			scheduledTime = time.Now()
			_, err = kv.Put(context.TODO(), "/agents/"+ip+"/jobs/"+job.Name, string(bytes), clientv3.WithPrevKV())
			if err != nil {
				fmt.Printf("%s:schedule error:%v,job name:%s\n", time.Now(), err, job.Name)
				return
			}
		}
		fmt.Printf("%s:%s scheduled to %s\n", scheduledTime, job.Name, ips)
	} else if now > job.NextExecuteTime {
		if job.Status == 2 || job.Status == 3 {
			// 任务正在执行
			return
		}
		// 触发时间已过
		// 更新下次执行时间
		expr, err := cronexpr.Parse(job.Cron)
		if err != nil {
			fmt.Printf("%s:parse cron expression error:%v,job name:%s\n", time.Now(), err, job.Name)
		}
		job.NextExecuteTime = expr.Next(time.Now()).Unix() * 1000
		bytes, err := json.Marshal(job)
		if err != nil {
			fmt.Printf("parse job data error:%v\n", err)
			return
		}
		jobDetail := string(bytes)
		// 更新etcd内容
		_, err = kv.Put(context.TODO(), "/jobs/"+job.Name, jobDetail, clientv3.WithPrevKV())
		if err != nil {
			fmt.Printf("%s:updatejob job status error:%v,job name:%s\n", time.Now(), err, job.Name)
			return
		}
	}
}

// 监听任务执行状态
func watchStatus(job common.Job, executingIps []string, client *clientv3.Client, db *sql.DB) {
	// watch /jobs/${jobName}/execute节点
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	execute := common.Execute{}
	ch := client.Watch(ctx, "/jobs/"+job.Name+"/execute")
	for res := range ch {
		// 处理delete事件
		if mvccpb.DELETE == res.Events[0].Type {
			return
		}
		// 处理put事件
		if mvccpb.PUT != res.Events[0].Type {
			continue
		}
		// 检查是否所有agent上的任务都执行完成
		err := json.Unmarshal(res.Events[0].Kv.Value, &execute)
		if err != nil {
			fmt.Printf("%s:watch job status error:%v,job name:%s\n", time.Now(), err, job.Name)
			return
		}
		if len(execute.TotalServers) == len(execute.SuccessServers)+len(execute.FailedServers) {
			// 所有agent均执行完成
			job.Status = 1
			if len(execute.FailedServers) == 0 {
				job.LastExecuteStatus = "success"
			} else {
				job.LastExecuteStatus = "failed"
				job.LastFailedServers = execute.FailedServers
			}
			job.LastSuccessServers = execute.SuccessServers
			job.LastExecuteServers = executingIps
			// 更新下次执行时间
			expr, err := cronexpr.Parse(job.Cron)
			if err != nil {
				fmt.Printf("%s:parse cron expression error:%v,job name:%s\n", time.Now(), err, job.Name)
				return
			}
			timestamp := expr.Next(time.Now()).Unix() * 1000
			job.NextExecuteTime = timestamp
			// 结束watch,跳出循环
			cancel()
			break
		}
	}
	// 任务执行后处理
	postProcess(&job, client, &execute, db)
	return
}

func postProcess(job *common.Job, client *clientv3.Client, execute *common.Execute, db *sql.DB) {
	// 删除/jobs/${jobName}/execute节点
	kv := clientv3.NewKV(client)
	_, err := kv.Delete(context.TODO(), "/jobs/"+job.Name+"/execute")
	if err != nil {
		fmt.Printf("%s:updatejob job status error:%v,job name:%s\n", time.Now(), err, job.Name)
		return
	}
	// 异步记录执行记录
	go saveExecutionRecord(job, execute, db)
	// 加分布式锁
	session, err := concurrency.NewSession(client, concurrency.WithTTL(1))
	if err != nil {
		fmt.Printf("create session error:%v\n", err)
		return
	}
	defer session.Close()
	lock := concurrency.NewLocker(session, "/jobs/"+job.Name)
	lock.Lock()
	defer lock.Unlock()
	bytes, err := json.Marshal(job)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		return
	}
	jobDetail := string(bytes)
	// 更新etcd内容
	_, err = kv.Put(context.TODO(), "/jobs/"+job.Name, jobDetail, clientv3.WithPrevKV())
	if err != nil {
		fmt.Printf("%s:updatejob job status error:%v,job name:%s\n", time.Now(), err, job.Name)
		return
	}
	fmt.Printf("%s:job execute in %s %s at %s,job name:%s\n", time.Now(), job.LastExecuteServers,
		job.LastExecuteStatus, time.Unix(execute.ExecutedTime, 0), job.Name)
}

// 记录执行记录
func saveExecutionRecord(job *common.Job, execute *common.Execute, db *sql.DB) {
	stmt, err := db.Prepare("insert into execution_record (job_id, job_name, execute_servers, success_servers, failed_servers, execute_status, executed_time) values (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		fmt.Printf("save execution record error:%v\n", err)
		return
	}
	tm := time.Unix(execute.ExecutedTime, 0)
	_, err = stmt.Exec(job.Id, job.Name, strings.Join(execute.TotalServers, ","), strings.Join(execute.SuccessServers, ","),
		strings.Join(execute.FailedServers, ","), job.LastExecuteStatus, tm.Format("2006-01-02 15:04:05"))
	if err != nil {
		fmt.Printf("save execution record error:%v\n", err)
		return
	}
}
