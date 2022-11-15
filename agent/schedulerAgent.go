package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"net"
	"os"
	"os/exec"
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
	ch := client.Watch(context.TODO(), "/agents/"+ip+"/jobs", clientv3.WithPrefix())
	for res := range ch {
		// 只处理put事件
		if 0 == res.Events[0].Type {
			job := common.Job{}
			err := json.Unmarshal(res.Events[0].Kv.Value, &job)
			if err != nil {
				fmt.Printf("%s:parse job error:%v\n", time.Now(), err)
				continue
			}
			// 执行cmd
			cmd := exec.Command("/bin/bash", "-c", job.Cmd)
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("%s:execute job error:%v,job name:%s\n", time.Now, err, job.Name)
				// TODO 更新任务上次执行状态为失败
				continue
			}

			fmt.Printf("%s:%s\n", time.Now(), string(out))

			// 更新任务状态为1
			resp, err := kv.Get(context.TODO(), "/jobs/"+job.Name)
			if err != nil {
				fmt.Printf("%s:execute job error:%v,job name:%s\n", time.Now, err, job.Name)
				continue
			}
			tmp := common.Job{}
			err = json.Unmarshal(resp.Kvs[0].Value, &tmp)
			if err != nil {
				fmt.Printf("%s:execute job error:%v,job name:%s\n", time.Now, err, job.Name)
				continue
			}
			tmp.Status = 1
			bytes, err := json.Marshal(tmp)
			if err != nil {
				fmt.Printf("%s:execute job error:%v,job name:%s\n", time.Now, err, job.Name)
				continue
			}
			jobDetail := string(bytes)
			// 更新etcd内容
			_, err = kv.Put(context.TODO(), "/jobs/"+tmp.Name, jobDetail, clientv3.WithPrevKV())
			if err != nil {
				fmt.Printf("%s:execute job error:%v,job name:%s\n", time.Now, err, job.Name)
				continue
			}

			// 删除/agents/ip/jobs节点下job内容
			_, err = kv.Delete(context.TODO(), "/agents/"+ip+"/jobs/"+job.Name)
			if err != nil {
				fmt.Printf("%s:execute job error:%v,job name:%s\n", time.Now, err, job.Name)
			}
		}
	}
}
