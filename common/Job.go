package common

import "time"

type Job struct {
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
	// 上次执行成功机器ip
	LastSuccessServers []string `json:"lastSuccessServers"`
	// 上次执行失败机器ip
	LastFailedServers []string `json:"lastFailedServers"`
	// 状态：0 - 冻结，1 - 正常, 2 - 执行中，3 - 异常
	Status int `json:"status"`
	// 创建/修改时间
	ModifyTime time.Time `json:"modifyTime"`
}
