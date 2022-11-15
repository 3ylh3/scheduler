package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"math/rand"
	"os"
	"sync"
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
	// 无限循环获取etcd中的job信息，若达到触发时间则挂到etcd对应的agent节点下
	for {
		kv := clientv3.NewKV(client)
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
		var wg sync.WaitGroup
		for _, value := range resp.Kvs {
			wg.Add(1)
			go processJob(value.Value, client, kv, &wg)
		}
		wg.Wait()
	}
}

func processJob(bytes []byte, client *clientv3.Client, kv clientv3.KV, wg *sync.WaitGroup) {
	defer wg.Done()
	job := job{}
	err := json.Unmarshal(bytes, &job)
	if err != nil {
		fmt.Printf("parse job error:%v\n", err)
		return
	}
	if job.Status == 0 {
		// 任务已冻结或者正在执行
		return
	}
	// 检查是否到达触发时间
	if time.Now().Unix() >= job.NextExecuteTime {
		if job.Status == 2 || job.Status == 3 {
			// 任务正在执行
			return
		}
		// 选择执行服务器
		var ips []string
		if job.ScheduleType == 1 {
			ips = job.ExecuteServers
		} else {
			rand.Seed(time.Now().UnixNano())
			ips = append(ips, job.ExecuteServers[rand.Intn(len(job.ExecuteServers))])
		}
		// 加分布式锁
		session, err := concurrency.NewSession(client, concurrency.WithTTL(1))
		if err != nil {
			fmt.Printf("lock error:%v\n", err)
		}
		defer session.Close()
		lock := concurrency.NewLocker(session, "/jobs/"+job.Name)
		lock.Lock()
		defer lock.Unlock()
		// 将任务信息放在执行ip节点下
		for _, ip := range ips {
			// 检查agent是否存在
			rsp, err := kv.Get(context.TODO(), "/agents/"+ip, clientv3.WithPrefix())
			if err != nil {
				fmt.Printf("%s:check agents error:%v,job name:%s\n", time.Now(), err, job.Name)
				return
			}
			if rsp.Count == 0 {
				fmt.Printf("%s:not find agent:%s,job name:%s\n", time.Now(), ip, job.Name)
				return
			}
			// 存入etcd
			_, err = kv.Put(context.TODO(), "/agents/"+ip+"/jobs/"+job.Name, job.Cmd, clientv3.WithPrevKV())
			if err != nil {
				fmt.Printf("%s:schedule error:%v,job name:%s\n", time.Now(), err, job.Name)
				return
			}
		}
		// 更新任务状态为已调度
		job.Status = 2
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
		fmt.Printf("%s:%s scheduled to %s\n", time.Now(), job.Name, ips)
	}
}
