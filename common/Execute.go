package common

type Execute struct {
	TotalServersCount int      `json:"totalServersCount"`
	SuccessServers    []string `json:"successServers"`
	FailedServers     []string `json:"failedServers"`
}
