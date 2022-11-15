package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"net"
	"os"
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
	kv := clientv3.NewKV(client)
	_, err = kv.Put(context.TODO(), "/agents/"+ip, "1")
	if err != nil {
		fmt.Printf("init agent error:%v\n", err)
		os.Exit(1)
	}
	// watch jobs节点
	ch := client.Watch(context.TODO(), "/agents/"+ip+"/jobs", clientv3.WithPrefix())
	for res := range ch {
		fmt.Printf("%v\n", res.Events)
	}
}
