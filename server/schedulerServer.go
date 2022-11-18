package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gorhill/cronexpr"
	"math/big"
	"os"
	"strconv"
	"tencent.com/justynyin/scheduler/common"
	"time"
)

func main() {
	// etcd地址
	var etcdAddr string
	flag.StringVar(&etcdAddr, "etcdAddr", "", "etcd address")
	flag.Parse()
	if "" == etcdAddr {
		fmt.Println("etcd address can not be empty,use -etcdAddr to declare")
		os.Exit(1)
	}
	var config = clientv3.Config{
		Endpoints:   []string{etcdAddr},
		DialTimeout: 5 * time.Second,
	}
	client, err := clientv3.New(config)
	if err != nil {
		fmt.Printf("init etcd client error:%v\n", err)
		os.Exit(1)
	}
	defer client.Close()
	kv := clientv3.NewKV(client)
	session, err := concurrency.NewSession(client, concurrency.WithTTL(1))
	if err != nil {
		fmt.Printf("create session error:%v\n", err)
	}
	defer session.Close()
	// 无限循环获取etcd中的job信息，若达到触发时间则挂到etcd对应的agent节点下
	for {
		resp, err := kv.Get(context.TODO(), "/jobs/", clientv3.WithPrefix())
		if err != nil {
			fmt.Printf("get jobs error:%v\n", err)
			continue
		}
		if 0 == resp.Count {
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
			if job.Status == 0 {
				// 任务已冻结
				continue
			}
			go processJob(job.Name, client, kv, session)
		}
		// sleep 500毫秒，减少cpu压力
		time.Sleep(500 * time.Millisecond)
	}
}

func processJob(jobName string, client *clientv3.Client, kv clientv3.KV, session *concurrency.Session) {
	// 加分布式锁
	// 生成1秒超时的context，1秒未获取锁则退出，避免造成任务执行时间过长时创建过多锁节点
	timeout, cancelFunc := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFunc()
	mutex := concurrency.NewMutex(session, "/jobs/"+jobName)
	err := mutex.Lock(timeout)
	if err != nil {
		// 加锁失败，直接返回
		return
	}
	defer mutex.Unlock(context.TODO())
	// 加锁后获取job信息
	job := common.Job{}
	resp, err := kv.Get(context.TODO(), "/jobs/"+jobName)
	if err != nil {
		fmt.Printf("%s:get job info error:%v,job name:%s\n", time.Now(), err, jobName)
		return
	}
	err = json.Unmarshal(resp.Kvs[0].Value, &job)
	if err != nil {
		fmt.Printf("%s:get job info error:%v,job name:%s\n", time.Now(), err, jobName)
		return
	}
	// 检查是否到达触发时间（因为循环遍历任务时等待了500ms，为防止漏执行，这里判断的时候需要考虑到500ms的等待时间）
	now := time.Now().Unix()
	if job.NextExecuteTime <= now && (job.NextExecuteTime+500) >= now {
		// 到达触发时间
		if job.Status == 2 {
			// 任务正在执行
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
		// 将执行机器信息挂到/jobs/${jobName}/execute节点下
		execute := common.Execute{
			TotalServersCount: len(executingIps),
			SuccessServers:    []string{},
			FailedServers:     []string{},
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
		// 更新任务状态为已调度
		job.Status = 2
		bytes, err := json.Marshal(job)
		if err != nil {
			fmt.Printf("parse job data error:%v\n", err)
			return
		}
		jobDetail := string(bytes)
		_, err = kv.Put(context.TODO(), "/jobs/"+job.Name, jobDetail, clientv3.WithPrevKV())
		if err != nil {
			fmt.Printf("%s:update job status error:%v,job name:%s\n", time.Now(), err, job.Name)
			return
		}
		// 监听任务执行情况
		go watchStatus(job, executingIps, client, session)
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
	} else if time.Now().Unix() > job.NextExecuteTime {
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
		job.NextExecuteTime = expr.Next(time.Now()).Unix()
		bytes, err := json.Marshal(job)
		if err != nil {
			fmt.Printf("parse job data error:%v\n", err)
			return
		}
		jobDetail := string(bytes)
		// 更新etcd内容
		_, err = kv.Put(context.TODO(), "/jobs/"+job.Name, jobDetail, clientv3.WithPrevKV())
		if err != nil {
			fmt.Printf("%s:update job status error:%v,job name:%s\n", time.Now(), err, job.Name)
			return
		}
	}
}

// 监听任务执行状态
func watchStatus(job common.Job, executingIps []string, client *clientv3.Client, session *concurrency.Session) {
	// watch /jobs/${jobName}/execute节点
	ch := client.Watch(context.TODO(), "/jobs/"+job.Name+"/execute")
	for res := range ch {
		// 只处理put事件
		if 0 != res.Events[0].Type {
			continue
		}
		// 检查是否所有agent上的任务都执行完成
		execute := common.Execute{}
		err := json.Unmarshal(res.Events[0].Kv.Value, &execute)
		if err != nil {
			fmt.Printf("%s:watch job status error:%v,job name:%s\n", time.Now(), err, job.Name)
			return
		}
		if execute.TotalServersCount == len(execute.SuccessServers)+len(execute.FailedServers) {
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
			job.LastExecuteTime = time.Now().String()
			// 更新下次执行时间
			expr, err := cronexpr.Parse(job.Cron)
			if err != nil {
				fmt.Printf("%s:parse cron expression error:%v,job name:%s\n", time.Now(), err, job.Name)
				return
			}
			timestamp := expr.Next(time.Now()).Unix()
			job.NextExecuteTime = timestamp
			// 跳出循环
			break
		}
	}
	// 删除/jobs/${jobName}/execute节点
	kv := clientv3.NewKV(client)
	_, err := kv.Delete(context.TODO(), "/jobs/"+job.Name+"/execute")
	if err != nil {
		fmt.Printf("%s:update job status error:%v,job name:%s\n", time.Now(), err, job.Name)
		return
	}
	// 加分布式锁
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
		fmt.Printf("%s:update job status error:%v,job name:%s\n", time.Now(), err, job.Name)
		return
	}
	fmt.Printf("%s:job execute in %s %s,job name:%s\n", job.LastExecuteTime, job.LastExecuteServers,
		job.LastExecuteStatus, job.Name)
	return
}
