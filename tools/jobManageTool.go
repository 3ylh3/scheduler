package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gorhill/cronexpr"
	"os"
	"strconv"
	"strings"
	"time"
)

type job struct {
	// 任务id
	Id int `json:"id"`
	// 任务名
	Name string `json:"name"`
	// 任务执行命令
	Cmd string `json:"cmd"`
	// cron表达式
	Cron string `json:"cron"`
	// 执行机
	ExecuteServers []string `json:"executeServers"`
	// 调度类型：1 - 全量调度，2 - 单机调度
	ScheduleType int `json:"scheduleType"`
	// 下次执行时间
	NextExecuteTime int64 `json:"nextExecuteTime"`
	// 上次执行时间
	LastExecuteTime string `json:"lastExecuteTime"`
	// 上次执行状态
	LastExecuteStatus string `json:"lastExecuteStatus"`
	// 上次执行机器ip
	LastExecuteServers []string `json:"lastExecuteServers"`
	// 状态：0 - 冻结，1 - 正常, 2 - 已调度，3 - 执行中
	Status int `json:"status"`
	// 创建/修改时间
	ModifyTime time.Time `json:"modifyTime"`
}

func main() {
	// help
	var help bool
	var jobId int
	// 任务名
	var jobName string
	// 任务执行命令
	var cmd string
	// cron表达式
	var cron string
	// 执行机
	var executeServers string
	// 调度类型：1 - 全量调度，2 - 单机调度
	var scheduleType int
	// etcd地址
	var etcdAddr string
	var add bool
	var show bool
	var update bool
	var del bool
	var freeze bool
	var unfreeze bool
	flag.BoolVar(&help, "help", false, "show help")
	flag.IntVar(&jobId, "jobId", -1, "job id")
	flag.StringVar(&jobName, "jobName", "", "job name")
	flag.StringVar(&cmd, "cmd", "", "execute commend")
	flag.StringVar(&cron, "cron", "", "cron expression")
	flag.StringVar(&executeServers, "executeServers", "", "execute servers")
	flag.IntVar(&scheduleType, "scheduleType", -1, "schedule type,default 1\n1:execute at all servers\n0:execute at one server")
	flag.StringVar(&etcdAddr, "etcdAddr", "", "etcd address")
	flag.BoolVar(&add, "add", false, "")
	flag.BoolVar(&show, "show", false, "")
	flag.BoolVar(&update, "update", false, "")
	flag.BoolVar(&del, "delete", false, "")
	flag.BoolVar(&freeze, "freeze", false, "")
	flag.BoolVar(&unfreeze, "unfreeze", false, "")
	flag.Parse()
	if help {
		// 打印使用信息
		printHelp()
		return
	}
	if !add && !show && !update && !del && !freeze && !unfreeze {
		fmt.Println("please select one option,use -help to see how to use")
		os.Exit(1)
	}
	//初始化etcd client
	client, err := initEtcdClient(etcdAddr)
	if err != nil {
		fmt.Printf("init etcd client error:%v\n", err)
		os.Exit(1)
	}
	defer client.Close()
	// 校验scheduleType
	if -1 != scheduleType && 0 != scheduleType && 1 != scheduleType {
		fmt.Println("scheduleType error,use -help to see how to use")
		os.Exit(1)
	}
	if add {
		// 校验参数
		if "" == jobName || "" == cmd || "" == cron || "" == executeServers || "" == etcdAddr {
			fmt.Println("param error,use -help to see how to use")
			os.Exit(1)
		}
		// 新增任务
		addJob(client, jobName, cmd, cron, executeServers, scheduleType)
	}
	if show {
		// 校验参数
		if "" == etcdAddr {
			fmt.Println("param error,use -help to see how to use")
			os.Exit(1)
		}
		// 查询job详情
		showJob(client, jobName, jobId)
	}
	if update {
		// 校验参数
		if -1 == jobId || "" == etcdAddr {
			fmt.Println("param error,use -help to see how to use")
			os.Exit(1)
		}
		if "" != jobName {
			fmt.Println("update job name is not allowed")
			os.Exit(1)
		}
		// 更新任务
		updateJob(client, jobId, cmd, cron, executeServers, scheduleType)
	}
	if del {
		// 校验参数
		if -1 == jobId || "" == etcdAddr {
			fmt.Println("param error,use -help to see how to use")
			os.Exit(1)
		}
		// 删除任务
		deleteJob(client, jobId)
	}
	if freeze {
		// 校验参数
		if -1 == jobId || "" == etcdAddr {
			fmt.Println("param error,use -help to see how to use")
			os.Exit(1)
		}
		// 冻结任务
		changeJobStatus(client, jobId, 0)
	}
	if unfreeze {
		// 校验参数
		if -1 == jobId || "" == etcdAddr {
			fmt.Println("param error,use -help to see how to use")
			os.Exit(1)
		}
		// 冻结任务
		changeJobStatus(client, jobId, 1)
	}
}

// 打印使用信息
func printHelp() {
	fmt.Println("Usage: jobManageTool [OPTION] [PARAMS]")
	fmt.Println("Options:")
	fmt.Printf("  -add\t\t\t\t\t\tadd job,jobName、cmd、cron、executeServers、scheduleType、etcdAddr is needed\n")
	fmt.Printf("  -show\t\t\t\t\t\tshow job details,etcdAddr is needed\n")
	fmt.Printf("  -update\t\t\t\t\tupdate job,jobId and etcdAddr is needed\n")
	fmt.Printf("  -delete\t\t\t\t\tdelete job,jobId and etcdAddr is needed\n")
	fmt.Printf("  -freeze\t\t\t\t\tfreeze job,jobId and etcdAddr is needed\n")
	fmt.Printf("  -unfreeze\t\t\t\t\tunfreeze job,jobId and etcdAddr is needed\n")
	fmt.Println("Params:")
	fmt.Printf("  -jobId\t\t\t\t\tjob id\n")
	fmt.Printf("  -jobName\t\t\t\t\tjob name\n")
	fmt.Printf("  -cmd\t\t\t\t\t\texecute commend\n")
	fmt.Printf("  -cron\t\t\t\t\t\tcron expression\n")
	fmt.Printf("  -executeServers\t\t\t\texecute servers\n")
	fmt.Printf("  -scheduleType\t\t\t\t\tschedule type,default 1\n\t\t\t\t\t\t1:execute at all servers\n\t\t\t\t\t\t0:execute at one server\n")
	fmt.Printf("  -etcdAddr\t\t\t\t\tetcd address\n")
}

// 初始化etcd client
func initEtcdClient(etcdAddr string) (*clientv3.Client, error) {
	var config = clientv3.Config{
		Endpoints:   []string{etcdAddr},
		DialTimeout: 5 * time.Second,
	}
	return clientv3.New(config)
}

// 新增job
func addJob(client *clientv3.Client, jobName string, cmd string, cron string, executeServers string,
	scheduleType int) {
	// 加分布式锁
	session, err := concurrency.NewSession(client, concurrency.WithTTL(1))
	if err != nil {
		fmt.Printf("lock error:%v\n", err)
	}
	defer session.Close()
	lock := concurrency.NewLocker(session, "/jobs/"+jobName)
	lock.Lock()
	defer lock.Unlock()
	// 校验任务名是否存在
	kv := clientv3.NewKV(client)
	resp, err := kv.Get(context.TODO(), "/jobs/"+jobName)
	if err != nil {
		fmt.Printf("query etcd error:%v\n", err)
		os.Exit(1)
	}
	if 1 <= resp.Count {
		fmt.Printf("duplicate job name\n")
		os.Exit(1)
	}
	// 生成任务id
	id, err := generateJobId(client)
	if err != nil {
		fmt.Printf("generate job id error:%v\n", err)
		os.Exit(1)
	}
	// 解析cron表达式
	expr, err := cronexpr.Parse(cron)
	if err != nil {
		fmt.Printf("parse cron expression error:%v\n", err)
		os.Exit(1)
	}
	job := job{
		Id:                 id,
		Name:               jobName,
		Cmd:                cmd,
		Cron:               cron,
		NextExecuteTime:    expr.Next(time.Now()).Unix(),
		ExecuteServers:     strings.Split(executeServers, ","),
		ScheduleType:       scheduleType,
		LastExecuteServers: []string{},
		Status:             1,
		ModifyTime:         time.Now(),
	}
	bytes, err := json.Marshal(job)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		os.Exit(1)
	}
	jobDetail := string(bytes)
	// 存入etcd
	_, err = kv.Put(context.TODO(), "/jobs/"+jobName, jobDetail, clientv3.WithPrevKV())
	if err != nil {
		fmt.Printf("put into etcd error:%v\n", err)
		os.Exit(1)
	}
	// 将id对应任务名存入etcd
	_, err = kv.Put(context.TODO(), "/jobIds/"+strconv.Itoa(id), jobName, clientv3.WithPrevKV())
	if err != nil {
		fmt.Printf("put into etcd error:%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("add job success,job detail:%s\n", string(bytes))
	os.Exit(0)
}

// 生成job id
func generateJobId(client *clientv3.Client) (int, error) {
	//获取id节点内容
	kv := clientv3.NewKV(client)
	resp, err := kv.Get(context.TODO(), "/id")
	if err != nil {
		return -1, err
	}
	id := 1
	if 0 != resp.Count {
		id, err = strconv.Atoi(string(resp.Kvs[0].Value))
		if err != nil {
			return -1, err
		}
	}
	//id加1并放回etcd
	_, err = kv.Put(context.TODO(), "/id", strconv.Itoa(id+1), clientv3.WithPrevKV())
	if err != nil {
		return -1, err
	}
	return id, nil
}

// 查询job
func showJob(client *clientv3.Client, jobName string, jobId int) {
	var result []job
	if -1 != jobId {
		kv := clientv3.NewKV(client)
		// 获取job name
		resp, err := kv.Get(context.TODO(), "/jobIds/"+strconv.Itoa(jobId))
		if err != nil {
			fmt.Printf("get job error:%v\n", err)
			os.Exit(1)
		}
		if 0 == resp.Count {
			fmt.Println("not find jobs")
			os.Exit(1)
		}
		name := string(resp.Kvs[0].Value)
		if "" != jobName && name != jobName {
			fmt.Println("job id and job name do not match")
			os.Exit(1)
		}
		// 获取任务详情
		resp, err = kv.Get(context.TODO(), "/jobs/"+name)
		if err != nil {
			fmt.Printf("get job error:%v\n", err)
			os.Exit(1)
		}
		if 0 == resp.Count {
			// 存在id和name对应关系，但未找到job，当做不存在处理，并删除对应关系
			_, _ = kv.Delete(context.TODO(), "/jobIds/"+strconv.Itoa(jobId))
			fmt.Println("not find jobs")
			os.Exit(1)
		}
		job := job{}
		err = json.Unmarshal(resp.Kvs[0].Value, &job)
		if err != nil {
			fmt.Printf("get job error:%v\n", err)
			os.Exit(1)
		}
		result = append(result, job)
	} else {
		kv := clientv3.NewKV(client)
		resp, err := kv.Get(context.TODO(), "/jobs/"+jobName, clientv3.WithPrefix())
		if err != nil {
			fmt.Printf("get jobs error:%v\n", err)
			os.Exit(1)
		}
		if 0 == resp.Count {
			fmt.Println("not find jobs")
			os.Exit(0)
		}
		for _, value := range resp.Kvs {
			job := job{}
			err := json.Unmarshal(value.Value, &job)
			if err != nil {
				fmt.Printf("get job error:%v\n", err)
				os.Exit(1)
			}
			if "" == jobName || jobName == job.Name {
				result = append(result, job)
			}
		}
	}
	if len(result) == 0 {
		fmt.Println("not find jobs")
		os.Exit(0)
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		os.Exit(1)
	}
	fmt.Println("jobs detail:")
	fmt.Println(string(bytes))
}

func deleteJob(client *clientv3.Client, jobId int) {
	kv := clientv3.NewKV(client)
	// 获取job name
	resp, err := kv.Get(context.TODO(), "/jobIds/"+strconv.Itoa(jobId))
	if err != nil {
		fmt.Printf("get job error:%v\n", err)
		os.Exit(1)
	}
	if 0 == resp.Count {
		fmt.Println("not find jobs")
		os.Exit(1)
	}
	name := string(resp.Kvs[0].Value)
	// 加分布式锁
	session, err := concurrency.NewSession(client, concurrency.WithTTL(1))
	if err != nil {
		fmt.Printf("lock error:%v\n", err)
	}
	defer session.Close()
	lock := concurrency.NewLocker(session, "/jobs/"+name)
	lock.Lock()
	defer lock.Unlock()
	// 删除任务
	_, err = kv.Delete(context.TODO(), "/jobs/"+name)
	if err != nil {
		fmt.Printf("delete job error:%v\n", err)
		os.Exit(1)
	}
	// 删除id name对应关系
	_, err = kv.Delete(context.TODO(), "/jobIds/"+strconv.Itoa(jobId))
	if err != nil {
		fmt.Printf("delete job error:%v\n", err)
		os.Exit(1)
	}
	fmt.Println("delete job success")
}

func updateJob(client *clientv3.Client, jobId int, cmd string, cron string,
	executeServers string, scheduleType int) {
	kv := clientv3.NewKV(client)
	// 获取job name
	resp, err := kv.Get(context.TODO(), "/jobIds/"+strconv.Itoa(jobId))
	if err != nil {
		fmt.Printf("get job error:%v\n", err)
		os.Exit(1)
	}
	if 0 == resp.Count {
		fmt.Println("not find jobs")
		os.Exit(1)
	}
	name := string(resp.Kvs[0].Value)
	// 加分布式锁
	session, err := concurrency.NewSession(client, concurrency.WithTTL(1))
	if err != nil {
		fmt.Printf("lock error:%v\n", err)
	}
	defer session.Close()
	lock := concurrency.NewLocker(session, "/jobs/"+name)
	lock.Lock()
	defer lock.Unlock()
	// 获取原job信息
	resp, err = kv.Get(context.TODO(), "/jobs/"+name)
	if err != nil {
		fmt.Printf("get job error:%v\n", err)
		os.Exit(1)
	}
	if 0 == resp.Count {
		// 存在id和name对应关系，但未找到job，当做不存在处理，并删除对应关系
		_, _ = kv.Delete(context.TODO(), "/jobIds/"+strconv.Itoa(jobId))
		fmt.Println("not find jobs")
		os.Exit(1)
	}
	job := job{}
	err = json.Unmarshal(resp.Kvs[0].Value, &job)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		os.Exit(1)
	}
	if "" != cmd {
		job.Cmd = cmd
	}
	if "" != cron {
		job.Cron = cron
		// 解析cron表达式
		expr, err := cronexpr.Parse(cron)
		if err != nil {
			fmt.Printf("parse cron expression error:%v\n", err)
			os.Exit(1)
		}
		job.NextExecuteTime = expr.Next(time.Now()).Unix()
	}
	if "" != executeServers {
		job.ExecuteServers = strings.Split(executeServers, ",")
	}
	if -1 != scheduleType {
		job.ScheduleType = scheduleType
	}
	job.ModifyTime = time.Now()
	bytes, err := json.Marshal(job)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		os.Exit(1)
	}
	jobDetail := string(bytes)
	// 更新etcd内容
	_, err = kv.Put(context.TODO(), "/jobs/"+name, jobDetail, clientv3.WithPrevKV())
	if err != nil {
		fmt.Printf("update job error:%v\n", err)
		os.Exit(1)
	}
	fmt.Println("update job success,job detail:")
	fmt.Println(jobDetail)
}

func changeJobStatus(client *clientv3.Client, jobId int, status int) {
	kv := clientv3.NewKV(client)
	// 获取job name
	resp, err := kv.Get(context.TODO(), "/jobIds/"+strconv.Itoa(jobId))
	if err != nil {
		fmt.Printf("get job error:%v\n", err)
		os.Exit(1)
	}
	if 0 == resp.Count {
		fmt.Println("not find jobs")
		os.Exit(1)
	}
	name := string(resp.Kvs[0].Value)
	// 加分布式锁
	session, err := concurrency.NewSession(client, concurrency.WithTTL(1))
	if err != nil {
		fmt.Printf("lock error:%v\n", err)
	}
	defer session.Close()
	lock := concurrency.NewLocker(session, "/jobs/"+name)
	lock.Lock()
	defer lock.Unlock()
	// 获取job信息
	resp, err = kv.Get(context.TODO(), "/jobs/"+name)
	if err != nil {
		fmt.Printf("get job error:%v\n", err)
		os.Exit(1)
	}
	if 0 == resp.Count {
		// 存在id和name对应关系，但未找到job，当做不存在处理，并删除对应关系
		_, _ = kv.Delete(context.TODO(), "/jobIds/"+strconv.Itoa(jobId))
		fmt.Println("not find jobs")
		os.Exit(1)
	}
	job := job{}
	err = json.Unmarshal(resp.Kvs[0].Value, &job)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		os.Exit(1)
	}
	if job.Status == 2 {
		fmt.Printf("cannot change status while job is executing")
		os.Exit(1)
	}
	job.Status = status
	job.ModifyTime = time.Now()
	bytes, err := json.Marshal(job)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		os.Exit(1)
	}
	jobDetail := string(bytes)
	// 更新etcd内容
	_, err = kv.Put(context.TODO(), "/jobs/"+name, jobDetail, clientv3.WithPrevKV())
	if err != nil {
		fmt.Printf("change job status error:%v\n", err)
		os.Exit(1)
	}
	fmt.Println("change job status success")
}
