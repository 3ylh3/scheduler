package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/3ylh3/scheduler/common"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log"
	"net"
	"os"
	"os/exec"
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
	// 获取本机ip
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("get local ip error:%v\n", err)
		os.Exit(1)
	}
	var ip string
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				break
			}
		}
	}
	// 注册agent信息到etcd
	lease := clientv3.NewLease(client)
	leaseGrantResp, err := lease.Grant(context.TODO(), 10)
	if err != nil {
		log.Fatalf("grant error:%v\n", err)
	}
	leaseId := leaseGrantResp.ID
	kv := clientv3.NewKV(client)
	_, err = kv.Put(context.TODO(), "/agents/"+ip, "1", clientv3.WithLease(leaseId))
	if err != nil {
		fmt.Printf("init agent error:%v\n", err)
		os.Exit(1)
	}
	// 自动续租
	keepAliveChan, err := lease.KeepAlive(context.TODO(), leaseId)
	if err != nil {
		fmt.Printf("init agent error:%v\n", err)
		os.Exit(1)
	}
	go func() {
		for {
			select {
			case resp := <-keepAliveChan:
				if resp == nil {
					// 租约过期
					os.Exit(1)
				}
			}
		}
	}()
	// watch jobs节点
	session, err := concurrency.NewSession(client, concurrency.WithTTL(1))
	if err != nil {
		fmt.Printf("%s:create session error:%v\n", time.Now(), err)
		return
	}
	defer session.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := client.Watch(ctx, "/agents/"+ip+"/jobs", clientv3.WithPrefix())
	for res := range ch {
		go func(event *clientv3.Event) {
			// 只处理put事件
			if mvccpb.PUT == event.Type {
				job := common.Job{}
				err := json.Unmarshal(event.Kv.Value, &job)
				if err != nil {
					fmt.Printf("%s:parse job error:%v\n", time.Now(), err)
					return
				}
				defer func() {
					// 删除/agents/ip/jobs节点下job内容
					_, err = kv.Delete(context.TODO(), "/agents/"+ip+"/jobs/"+job.Name)
				}()
				fmt.Printf("%s:start to execute job:%s\n", time.Now(), job.Name)
				// 执行cmd
				success := true
				cmd := exec.Command("/bin/bash", "-c", job.Cmd)
				out, err := cmd.CombinedOutput()
				if err != nil {
					// 执行失败
					success = false
					fmt.Printf("%s:execute job error:%v,job name:%s\n", time.Now(), err, job.Name)
				}
				// 获取/jobs/${jobName}/execute节点内容
				// 加分布式锁
				lock := concurrency.NewLocker(session, "/jobs/"+job.Name+"/execute")
				lock.Lock()
				defer lock.Unlock()
				resp, err := kv.Get(context.TODO(), "/jobs/"+job.Name+"/execute")
				if err != nil {
					fmt.Printf("%s:updatejob execute result error:%v,job name:%s\n", time.Now(), err, job.Name)
					return
				}
				execute := common.Execute{}
				err = json.Unmarshal(resp.Kvs[0].Value, &execute)
				if err != nil {
					fmt.Printf("%s:updatejob execute result error:%v,job name:%s\n", time.Now(), err, job.Name)
					return
				}
				fmt.Printf("%s:execute %s output:\n%s", time.Now(), job.Name, string(out))

				if success {
					// 更新/jobs/${jobName}/execute节点下SuccessServers信息
					execute.SuccessServers = append(execute.SuccessServers, ip)
				} else {
					// 更新/jobs/${jobName}/execute节点下FailedServers信息
					execute.FailedServers = append(execute.FailedServers, ip)
				}
				execute.ExecutedTime = time.Now().Unix()
				executeDetail, err := json.Marshal(execute)
				if err != nil {
					fmt.Printf("%s:updatejob execute result error:%v,job name:%s\n", time.Now(), err, job.Name)
					return
				}
				_, err = kv.Put(context.TODO(), "/jobs/"+job.Name+"/execute", string(executeDetail))
				if err != nil {
					fmt.Printf("%s:updatejob execute result error:%v,job name:%s\n", time.Now(), err, job.Name)
					return
				}
				fmt.Printf("%s:job execute completed,job name:%s\n", time.Now(), job.Name)
			}
		}(res.Events[0])
	}
}
