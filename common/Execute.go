package common

type Execute struct {
	TotalServers   []string `json:"totalServers"`
	SuccessServers []string `json:"successServers"`
	FailedServers  []string `json:"failedServers"`
	ExecutedTime   int64    `json:"executedTime"`
}
