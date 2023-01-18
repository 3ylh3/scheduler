package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/3ylh3/scheduler/apiserver/addjob"
	"github.com/3ylh3/scheduler/apiserver/changejobstatus"
	"github.com/3ylh3/scheduler/apiserver/deletejob"
	"github.com/3ylh3/scheduler/apiserver/job"
	"github.com/3ylh3/scheduler/apiserver/qryjobinfo"
	"github.com/3ylh3/scheduler/apiserver/updatejob"
	"github.com/3ylh3/scheduler/common"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gorhill/cronexpr"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	// help
	var help bool
	// etcd地址
	var etcdAddr string
	// 端口
	var port int
	flag.BoolVar(&help, "help", false, "show help")
	flag.StringVar(&etcdAddr, "etcdAddr", "", "etcd address")
	flag.IntVar(&port, "port", 4000, "apiserver port,default 4000")
	flag.Parse()
	if help {
		// 打印使用信息
		flag.PrintDefaults()
		return
	}
	// 校验参数
	if etcdAddr == "" {
		fmt.Println("etcd address is null,use -help to see how to use")
		os.Exit(1)
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
		os.Exit(1)
	}
	grpcServer := grpc.NewServer()
	qryjobinfo.RegisterQueryJobInfoServer(grpcServer, &QueryJobInfoServer{EtcdAddr: etcdAddr})
	addjob.RegisterAddJobServer(grpcServer, &AddJobServer{EtcdAddr: etcdAddr})
	updatejob.RegisterUpdateJobServer(grpcServer, &UpdateJobServer{EtcdAddr: etcdAddr})
	deletejob.RegisterDeleteJobServer(grpcServer, &DeleteJobServer{EtcdAddr: etcdAddr})
	changejobstatus.RegisterChangeJobStatusServer(grpcServer, &ChangeJobStatusServer{EtcdAddr: etcdAddr})
	fmt.Printf("apiserver apiserver is listing at %v\n", lis.Addr())
	err = grpcServer.Serve(lis)
	if err != nil {
		fmt.Printf("failed to serve: %v\n", err)
	}
}

type QueryJobInfoServer struct {
	qryjobinfo.UnimplementedQueryJobInfoServer
	EtcdAddr string
}

type AddJobServer struct {
	addjob.UnimplementedAddJobServer
	EtcdAddr string
}

type UpdateJobServer struct {
	updatejob.UnimplementedUpdateJobServer
	EtcdAddr string
}

type DeleteJobServer struct {
	deletejob.UnimplementedDeleteJobServer
	EtcdAddr string
}

type ChangeJobStatusServer struct {
	changejobstatus.UnimplementedChangeJobStatusServer
	EtcdAddr string
}

// 初始化etcd client
func initEtcdClient(etcdAddr string) (*clientv3.Client, error) {
	var config = clientv3.Config{
		Endpoints:   []string{etcdAddr},
		DialTimeout: 5 * time.Second,
	}
	return clientv3.New(config)
}

// 查询job信息
func (q *QueryJobInfoServer) QueryJobInfo(ctx context.Context, req *qryjobinfo.QueryJobInfoRequest) (*qryjobinfo.QueryJobInfoResponse, error) {
	//初始化etcd client
	client, err := initEtcdClient(q.EtcdAddr)
	if err != nil {
		fmt.Printf("init etcd client error:%v\n", err)
		return &qryjobinfo.QueryJobInfoResponse{}, fmt.Errorf("init etcd client error:%v\n", err)
	}
	defer client.Close()
	var result []*job.JobInfo
	if req.JobId != "" {
		kv := clientv3.NewKV(client)
		// 获取job name
		resp, err := kv.Get(context.TODO(), "/jobIds/"+req.JobId)
		if err != nil {
			fmt.Printf("get job error:%v\n", err)
			return &qryjobinfo.QueryJobInfoResponse{}, err
		}
		if 0 == resp.Count {
			fmt.Println("not find jobs")
			return &qryjobinfo.QueryJobInfoResponse{}, fmt.Errorf("not find jobs\n")
		}
		name := string(resp.Kvs[0].Value)
		if "" != req.JobName && name != req.JobName {
			fmt.Println("job id and job name do not match")
			return &qryjobinfo.QueryJobInfoResponse{}, fmt.Errorf("job id and job name do not match\n")
		}
		// 获取任务详情
		resp, err = kv.Get(context.TODO(), "/jobs/"+name)
		if err != nil {
			fmt.Printf("get job error:%v\n", err)
			return &qryjobinfo.QueryJobInfoResponse{}, err
		}
		if 0 == resp.Count {
			// 存在id和name对应关系，但未找到job，当做不存在处理，并删除对应关系
			_, _ = kv.Delete(context.TODO(), "/jobIds/"+req.JobId)
			fmt.Println("not find job")
			return &qryjobinfo.QueryJobInfoResponse{}, fmt.Errorf("not find job\n")
		}
		job := job.JobInfo{}
		err = jsonpb.UnmarshalString(string(resp.Kvs[0].Value), &job)
		if err != nil {
			fmt.Printf("get job error:%v\n", err)
			return &qryjobinfo.QueryJobInfoResponse{}, err
		}
		result = append(result, &job)
	} else {
		kv := clientv3.NewKV(client)
		resp, err := kv.Get(context.TODO(), "/jobs/"+req.JobName, clientv3.WithPrefix())
		if err != nil {
			fmt.Printf("get jobs error:%v\n", err)
			return &qryjobinfo.QueryJobInfoResponse{}, err
		}
		if 0 == resp.Count {
			fmt.Println("not find jobs")
			return &qryjobinfo.QueryJobInfoResponse{}, fmt.Errorf("not find jobs\n")
		}
		for _, value := range resp.Kvs {
			job := job.JobInfo{}
			err := jsonpb.UnmarshalString(string(value.Value), &job)
			if err != nil {
				fmt.Printf("get job error:%v\n", err)
				return &qryjobinfo.QueryJobInfoResponse{}, err
			}
			if req.JobName == "" || req.JobName == job.Name {
				result = append(result, &job)
			}
		}
	}
	if len(result) == 0 {
		fmt.Println("not find jobs")
		return &qryjobinfo.QueryJobInfoResponse{}, fmt.Errorf("not find jobs\n")
	}
	return &qryjobinfo.QueryJobInfoResponse{Result: result}, nil
}

// 新增job
func (a *AddJobServer) AddJob(ctx context.Context, req *addjob.AddJobRequest) (*job.JobInfo, error) {
	//初始化etcd client
	client, err := initEtcdClient(a.EtcdAddr)
	if err != nil {
		fmt.Printf("init etcd client error:%v\n", err)
		return &job.JobInfo{}, fmt.Errorf("init etcd client error:%v\n", err)
	}
	defer client.Close()
	// 加分布式锁
	session, err := concurrency.NewSession(client, concurrency.WithTTL(1))
	if err != nil {
		fmt.Printf("lock error:%v\n", err)
		return &job.JobInfo{}, fmt.Errorf("lock error:%v\n", err)
	}
	defer session.Close()
	lock := concurrency.NewLocker(session, "/jobs/"+req.JobName)
	lock.Lock()
	defer lock.Unlock()
	// 校验任务名是否存在
	kv := clientv3.NewKV(client)
	resp, err := kv.Get(context.TODO(), "/jobs/"+req.JobName)
	if err != nil {
		fmt.Printf("query etcd error:%v\n", err)
		return &job.JobInfo{}, err
	}
	if 1 <= resp.Count {
		fmt.Printf("duplicate job name\n")
		return &job.JobInfo{}, fmt.Errorf("duplicate job name\n")
	}
	// 生成任务id
	id, err := generateJobId(client, session)
	if err != nil {
		fmt.Printf("generate job id error:%v\n", err)
		return &job.JobInfo{}, fmt.Errorf("generate job id error:%v\n", err)
	}
	// 解析cron表达式
	expr, err := cronexpr.Parse(req.Cron)
	if err != nil {
		fmt.Printf("parse cron expression error:%v\n", err)
		return &job.JobInfo{}, fmt.Errorf("parse cron expression error:%v\n", err)
	}
	jobInfo := common.Job{
		Id:                 id,
		Name:               req.JobName,
		Cmd:                req.Cmd,
		Cron:               req.Cron,
		NextExecuteTime:    expr.Next(time.Now()).Unix() * 1000,
		ExecuteServers:     strings.Split(req.ExecuteServers, ","),
		ScheduleType:       int(req.ScheduleType),
		LastExecuteServers: []string{},
		LastSuccessServers: []string{},
		LastFailedServers:  []string{},
		Status:             1,
		ModifyTime:         time.Now(),
	}
	bytes, err := json.Marshal(jobInfo)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		return &job.JobInfo{}, fmt.Errorf("parse job data error:%v\n", err)
	}
	jobDetail := string(bytes)
	// 存入etcd
	_, err = kv.Put(context.TODO(), "/jobs/"+req.JobName, jobDetail, clientv3.WithPrevKV())
	if err != nil {
		fmt.Printf("put into etcd error:%v\n", err)
		return &job.JobInfo{}, err
	}
	// 将id对应任务名存入etcd
	_, err = kv.Put(context.TODO(), "/jobIds/"+strconv.Itoa(id), req.JobName, clientv3.WithPrevKV())
	if err != nil {
		fmt.Printf("put into etcd error:%v\n", err)
		return &job.JobInfo{}, err
	}
	rsp := job.JobInfo{}
	err = jsonpb.UnmarshalString(jobDetail, &rsp)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		return &job.JobInfo{}, err
	}

	return &rsp, nil
}

// 生成job id
func generateJobId(client *clientv3.Client, session *concurrency.Session) (int, error) {
	//加分布式锁
	lock := concurrency.NewLocker(session, "/id")
	lock.Lock()
	defer lock.Unlock()
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

// 更新job信息
func (u *UpdateJobServer) UpdateJob(ctx context.Context, req *updatejob.UpdateJobRequest) (*job.JobInfo, error) {
	//初始化etcd client
	client, err := initEtcdClient(u.EtcdAddr)
	if err != nil {
		fmt.Printf("init etcd client error:%v\n", err)
		return &job.JobInfo{}, fmt.Errorf("init etcd client error:%v\n", err)
	}
	defer client.Close()
	kv := clientv3.NewKV(client)
	// 获取job name
	resp, err := kv.Get(context.TODO(), "/jobIds/"+req.JobId)
	if err != nil {
		fmt.Printf("get job error:%v\n", err)
		return &job.JobInfo{}, err
	}
	if 0 == resp.Count {
		fmt.Println("not find job")
		return &job.JobInfo{}, fmt.Errorf("not find job\n")
	}
	name := string(resp.Kvs[0].Value)
	// 加分布式锁
	session, err := concurrency.NewSession(client, concurrency.WithTTL(1))
	if err != nil {
		fmt.Printf("lock error:%v\n", err)
		return &job.JobInfo{}, fmt.Errorf("lock error:%v\n", err)
	}
	defer session.Close()
	lock := concurrency.NewLocker(session, "/jobs/"+name)
	lock.Lock()
	defer lock.Unlock()
	// 获取原job信息
	resp, err = kv.Get(context.TODO(), "/jobs/"+name)
	if err != nil {
		fmt.Printf("get job error:%v\n", err)
		return &job.JobInfo{}, err
	}
	if 0 == resp.Count {
		// 存在id和name对应关系，但未找到job，当做不存在处理，并删除对应关系
		_, _ = kv.Delete(context.TODO(), "/jobIds/"+req.JobId)
		fmt.Println("not find job")
		return &job.JobInfo{}, fmt.Errorf("not find job\n")
	}
	jobInfo := common.Job{}
	err = json.Unmarshal(resp.Kvs[0].Value, &jobInfo)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		return &job.JobInfo{}, fmt.Errorf("parse job data error:%v\n", err)
	}
	if "" != req.Cmd {
		jobInfo.Cmd = req.Cmd
	}
	if "" != req.Cron {
		jobInfo.Cron = req.Cron
		// 解析cron表达式
		expr, err := cronexpr.Parse(req.Cron)
		if err != nil {
			fmt.Printf("parse cron expression error:%v\n", err)
			return &job.JobInfo{}, fmt.Errorf("parse cron expression error:%v\n", err)
		}
		jobInfo.NextExecuteTime = expr.Next(time.Now()).Unix() * 1000
	}
	if "" != req.ExecuteServers {
		jobInfo.ExecuteServers = strings.Split(req.ExecuteServers, ",")
	}
	if -1 != req.ScheduleType {
		jobInfo.ScheduleType = int(req.ScheduleType)
	}
	jobInfo.ModifyTime = time.Now()
	bytes, err := json.Marshal(jobInfo)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		return &job.JobInfo{}, fmt.Errorf("parse job data error:%v\n", err)
	}
	jobDetail := string(bytes)
	// 更新etcd内容
	_, err = kv.Put(context.TODO(), "/jobs/"+name, jobDetail, clientv3.WithPrevKV())
	if err != nil {
		fmt.Printf("updatejob job error:%v\n", err)
		return &job.JobInfo{}, err
	}
	rsp := job.JobInfo{}
	err = jsonpb.UnmarshalString(jobDetail, &rsp)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		return &job.JobInfo{}, err
	}
	return &rsp, nil
}

func (d *DeleteJobServer) DeleteJob(ctx context.Context, req *deletejob.DeleteJobRequest) (*emptypb.Empty, error) {
	//初始化etcd client
	client, err := initEtcdClient(d.EtcdAddr)
	if err != nil {
		fmt.Printf("init etcd client error:%v\n", err)
		return &emptypb.Empty{}, fmt.Errorf("init etcd client error:%v\n", err)
	}
	defer client.Close()
	kv := clientv3.NewKV(client)
	// 获取job name
	resp, err := kv.Get(context.TODO(), "/jobIds/"+req.JobId)
	if err != nil {
		fmt.Printf("get job error:%v\n", err)
		return &emptypb.Empty{}, err
	}
	if 0 == resp.Count {
		fmt.Println("not find job")
		return &emptypb.Empty{}, fmt.Errorf("not find job\n")
	}
	name := string(resp.Kvs[0].Value)
	// 加分布式锁
	session, err := concurrency.NewSession(client, concurrency.WithTTL(1))
	if err != nil {
		fmt.Printf("lock error:%v\n", err)
		return &emptypb.Empty{}, fmt.Errorf("lock error:%v\n", err)
	}
	defer session.Close()
	lock := concurrency.NewLocker(session, "/jobs/"+name)
	lock.Lock()
	defer lock.Unlock()
	// 删除任务
	_, err = kv.Delete(context.TODO(), "/jobs/"+name)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	// 删除id name对应关系
	_, err = kv.Delete(context.TODO(), "/jobIds/"+req.JobId)
	if err != nil {
		fmt.Printf("delete job error:%v\n", err)
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (c *ChangeJobStatusServer) ChangeJobStatus(ctx context.Context, req *changejobstatus.ChangeJobStatusRequest) (*emptypb.Empty, error) {
	//初始化etcd client
	client, err := initEtcdClient(c.EtcdAddr)
	if err != nil {
		fmt.Printf("init etcd client error:%v\n", err)
		return &emptypb.Empty{}, fmt.Errorf("init etcd client error:%v\n", err)
	}
	defer client.Close()
	kv := clientv3.NewKV(client)
	// 获取job name
	resp, err := kv.Get(context.TODO(), "/jobIds/"+req.JobId)
	if err != nil {
		fmt.Printf("get job error:%v\n", err)
		return &emptypb.Empty{}, err
	}
	if 0 == resp.Count {
		fmt.Println("not find job")
		return &emptypb.Empty{}, fmt.Errorf("not find job\n")
	}
	name := string(resp.Kvs[0].Value)
	// 加分布式锁
	session, err := concurrency.NewSession(client, concurrency.WithTTL(1))
	if err != nil {
		fmt.Printf("lock error:%v\n", err)
		return &emptypb.Empty{}, fmt.Errorf("lock error:%v\n", err)
	}
	defer session.Close()
	lock := concurrency.NewLocker(session, "/jobs/"+name)
	lock.Lock()
	defer lock.Unlock()
	// 获取job信息
	resp, err = kv.Get(context.TODO(), "/jobs/"+name)
	if err != nil {
		fmt.Printf("get job error:%v\n", err)
		return &emptypb.Empty{}, err
	}
	if 0 == resp.Count {
		// 存在id和name对应关系，但未找到job，当做不存在处理，并删除对应关系
		_, _ = kv.Delete(context.TODO(), "/jobIds/"+req.JobId)
		fmt.Println("not find job")
		return &emptypb.Empty{}, fmt.Errorf("not find job\n")
	}
	job := common.Job{}
	err = json.Unmarshal(resp.Kvs[0].Value, &job)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		return &emptypb.Empty{}, fmt.Errorf("parse job data error:%v\n", err)
	}
	if job.Status == 2 {
		fmt.Println("cannot change status while job is executing")
		return &emptypb.Empty{}, fmt.Errorf("cannot change status while job is executing\n")
	}
	job.Status = int(req.Status)
	job.ModifyTime = time.Now()
	bytes, err := json.Marshal(job)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		return &emptypb.Empty{}, fmt.Errorf("parse job data error:%v\n", err)
	}
	jobDetail := string(bytes)
	// 更新etcd内容
	_, err = kv.Put(context.TODO(), "/jobs/"+name, jobDetail, clientv3.WithPrevKV())
	if err != nil {
		fmt.Printf("change job status error:%v\n", err)
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}
