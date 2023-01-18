package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/3ylh3/scheduler/apiserver/addjob"
	"github.com/3ylh3/scheduler/apiserver/changejobstatus"
	"github.com/3ylh3/scheduler/apiserver/deletejob"
	"github.com/3ylh3/scheduler/apiserver/qryjobinfo"
	"github.com/3ylh3/scheduler/apiserver/updatejob"
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"strconv"
	"time"
)

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
	// apiServer地址
	var apiServerAddr string
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
	flag.StringVar(&executeServers, "executeServers", "", "execute apiserver")
	flag.IntVar(&scheduleType, "scheduleType", -1, "schedule type,default 1\n1:execute at all apiserver\n0:execute at one queryJobInfoServer")
	flag.StringVar(&apiServerAddr, "apiServerAddr", "", "apiserver apiserver address")
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
	// 校验scheduleType
	if -1 != scheduleType && 1 != scheduleType && 2 != scheduleType {
		fmt.Println("scheduleType error,use -help to see how to use")
		os.Exit(1)
	}
	if show {
		// 校验参数
		if "" == apiServerAddr {
			fmt.Println("param error,use -help to see how to use")
			os.Exit(1)
		}
		// 查询job详情
		showJob(apiServerAddr, jobName, jobId)
	}
	if add {
		// 校验参数
		if "" == jobName || "" == cmd || "" == cron || "" == executeServers || "" == apiServerAddr {
			fmt.Println("param error,use -help to see how to use")
			os.Exit(1)
		}
		// 新增任务
		addJob(apiServerAddr, jobName, cmd, cron, executeServers, scheduleType)
	}
	if update {
		// 校验参数
		if -1 == jobId || "" == apiServerAddr {
			fmt.Println("param error,use -help to see how to use")
			os.Exit(1)
		}
		if "" != jobName {
			fmt.Println("updatejob job name is not allowed")
			os.Exit(1)
		}
		// 更新任务
		updateJob(apiServerAddr, jobId, cmd, cron, executeServers, scheduleType)
	}
	if del {
		// 校验参数
		if -1 == jobId || "" == apiServerAddr {
			fmt.Println("param error,use -help to see how to use")
			os.Exit(1)
		}
		// 删除任务
		deleteJob(apiServerAddr, jobId)
	}
	if freeze {
		// 校验参数
		if -1 == jobId || "" == apiServerAddr {
			fmt.Println("param error,use -help to see how to use")
			os.Exit(1)
		}
		// 冻结任务
		changeJobStatus(apiServerAddr, jobId, 0)
	}
	if unfreeze {
		// 校验参数
		if -1 == jobId || "" == apiServerAddr {
			fmt.Println("param error,use -help to see how to use")
			os.Exit(1)
		}
		// 冻结任务
		changeJobStatus(apiServerAddr, jobId, 1)
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
	fmt.Printf("  -executeServers\t\t\t\texecute apiserver\n")
	fmt.Printf("  -scheduleType\t\t\t\t\tschedule type,default 1\n\t\t\t\t\t\t1:execute at all apiserver\n\t\t\t\t\t\t0:execute at one queryJobInfoServer\n")
	fmt.Printf("  -apiServerAddr\t\t\t\t\tapiserver apiserver address\n")
}

func showJob(apiServerAddr string, jobName string, jobId int) {
	// 调用api server查询任务信息
	req := qryjobinfo.QueryJobInfoRequest{}
	if jobId == -1 {
		req.JobId = ""
	} else {
		req.JobId = strconv.Itoa(jobId)
	}
	if jobName != "" {
		req.JobName = jobName
	}
	conn, err := grpc.Dial(apiServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("connect to apiserver apiserver error:%v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	client := qryjobinfo.NewQueryJobInfoClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	rsp, err := client.QueryJobInfo(ctx, &req)
	if err != nil {
		fmt.Printf("get jobs error:%v\n", err)
		os.Exit(1)
	}
	jsonpbMarshaler := &jsonpb.Marshaler{
		EmitDefaults: true, //将字段值为空的渲染到JSON结构中
	}
	var buffer bytes.Buffer
	err = jsonpbMarshaler.Marshal(&buffer, rsp)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		os.Exit(1)
	}
	fmt.Println("jobs detail:")
	fmt.Println(buffer.String())
}

func addJob(apiServerAddr string, jobName string, cmd string, cron string, executeServers string, scheduleType int) {
	// 调用api server新增任务信息
	req := addjob.AddJobRequest{
		JobName:        jobName,
		Cmd:            cmd,
		Cron:           cron,
		ExecuteServers: executeServers,
		ScheduleType:   int32(scheduleType),
	}
	conn, err := grpc.Dial(apiServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("connect to apiserver apiserver error:%v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	client := addjob.NewAddJobClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	rsp, err := client.AddJob(ctx, &req)
	if err != nil {
		fmt.Printf("add job error:%v\n", err)
		os.Exit(1)
	}
	jsonpbMarshaler := &jsonpb.Marshaler{
		EmitDefaults: true, //将字段值为空的渲染到JSON结构中
	}
	var buffer bytes.Buffer
	err = jsonpbMarshaler.Marshal(&buffer, rsp)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("add job success,job info:%s\n", buffer.String())
}

// 更新job信息
func updateJob(apiServerAddr string, jobId int, cmd string, cron string, executeServers string, scheduleType int) {
	// 调用api server更新job信息
	req := updatejob.UpdateJobRequest{JobId: strconv.Itoa(jobId), ScheduleType: int32(scheduleType)}
	if "" != cmd {
		req.Cmd = cmd
	}
	if "" != cron {
		req.Cron = cron
	}
	if "" != executeServers {
		req.ExecuteServers = executeServers
	}
	conn, err := grpc.Dial(apiServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("connect to apiserver apiserver error:%v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	client := updatejob.NewUpdateJobClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	rsp, err := client.UpdateJob(ctx, &req)
	if err != nil {
		fmt.Printf("update job error:%v\n", err)
		os.Exit(1)
	}
	jsonpbMarshaler := &jsonpb.Marshaler{
		EmitDefaults: true, //将字段值为空的渲染到JSON结构中
	}
	var buffer bytes.Buffer
	err = jsonpbMarshaler.Marshal(&buffer, rsp)
	if err != nil {
		fmt.Printf("parse job data error:%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("update job success,job info:%s\n", buffer.String())
}

// 删除job
func deleteJob(apiServerAddr string, jobId int) {
	// 调用api server删除job
	req := deletejob.DeleteJobRequest{JobId: strconv.Itoa(jobId)}
	conn, err := grpc.Dial(apiServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("connect to apiserver apiserver error:%v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	client := deletejob.NewDeleteJobClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, err = client.DeleteJob(ctx, &req)
	if err != nil {
		fmt.Printf("delete job error:%v\n", err)
		os.Exit(1)
	}
	fmt.Println("delete job success")
}

// 更新job状态
func changeJobStatus(apiServerAddr string, jobId int, status int) {
	// 调用api server删除job
	req := changejobstatus.ChangeJobStatusRequest{JobId: strconv.Itoa(jobId), Status: int32(status)}
	conn, err := grpc.Dial(apiServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("connect to apiserver apiserver error:%v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	client := changejobstatus.NewChangeJobStatusClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, err = client.ChangeJobStatus(ctx, &req)
	if err != nil {
		fmt.Printf("change job status error:%v\n", err)
		os.Exit(1)
	}
	fmt.Println("change job status success")
}
