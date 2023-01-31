package common

type Config struct {
	Etcd struct {
		Address string `yaml:"address"`
	}
	Mysql struct {
		Database string `yaml:"database"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	}
	Port int `yaml:"port"`
}
