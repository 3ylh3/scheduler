package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gorhill/cronexpr"
	"math/rand"
	"os"
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
		for _, value := range resp.Kvs {
			go processJob(value.Value, client, kv)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func processJob(bytes []byte, client *clientv3.Client, kv clientv3.KV) {
	job := common.Job{}
	err := json.Unmarshal(bytes, &job)
	if err != nil {
		return
	}
	if job.Status == 0 {
		// 任务已冻结或者正在执行
		return
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
	// 检查是否到达触发时间
	if time.Now().Unix() == job.NextExecuteTime {
		// 到达触发时间
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
		// 将任务信息放在执行ip节点下
		var scheduledTime time.Time
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
			scheduledTime = time.Now()
			_, err = kv.Put(context.TODO(), "/agents/"+ip+"/jobs/"+job.Name, string(bytes), clientv3.WithPrevKV())
			if err != nil {
				fmt.Printf("%s:schedule error:%v,job name:%s\n", time.Now(), err, job.Name)
				return
			}
		}
		// 更新任务状态为已调度
		job.Status = 2
		// 更新下次执行时间
		expr, err := cronexpr.Parse(job.Cron)
		if err != nil {
			fmt.Printf("%s:parse cron expression error:%v,job name:%s\n", time.Now(), err, job.Name)
		}
		timestamp := expr.Next(time.Unix(job.NextExecuteTime, 0)).Unix()
		if timestamp < time.Now().Unix() {
			timestamp = expr.Next(time.Now()).Unix()
		}
		job.NextExecuteTime = timestamp
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
