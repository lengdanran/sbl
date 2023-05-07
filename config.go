package main

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

var (
	ascii = `
	 ____  _     ____  
	/ ___|| |   | __ ) 
	\___ \| |   |  _ \ 
	 ___) | |___| |_) |
	|____/|_____|____/ 
`
)

// Config configuration details of balancer
type Config struct {
	SSLCertificateKey   string      `yaml:"ssl_certificate_key"`
	Location            []*Location `yaml:"location"`
	Schema              string      `yaml:"schema"`
	Port                int         `yaml:"port"`
	SSLCertificate      string      `yaml:"ssl_certificate"`
	HealthCheck         bool        `yaml:"tcp_health_check"`      // 是否启用健康检查
	HealthCheckInterval uint        `yaml:"health_check_interval"` // 健康检查时间间隔(s)
	MaxAllowed          uint        `yaml:"max_allowed"`           // 同时允许的最大连接数，不指定则不限制
}

// Location routing details of balancer
type Location struct {
	Pattern         string   `yaml:"pattern"`
	ProxyPass       []string `yaml:"proxy_pass"`
	ProxyPassWeight []int    `yaml:"proxy_pass_weight"`
	BalanceMode     string   `yaml:"balance_mode"` // name of balancer
}

// ReadConfig read configuration from `fileName` file
func ReadConfig(fileName string) (*Config, error) {
	in, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(in, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// test

// Print print config details
func (c *Config) Print() {
	fmt.Printf("%s\nSchema: %s\nPort: %d\nHealth Check: %v\nLocation:\n",
		ascii, c.Schema, c.Port, c.HealthCheck)
	for _, l := range c.Location {
		fmt.Printf("\tRoute: %s\n\tProxy Pass: %s\n\tProxy Pass Weight: %d\n\tMode: %s\n\n",
			l.Pattern, l.ProxyPass, l.ProxyPassWeight, l.BalanceMode)
	}
}

// Validation verify the configuration details of the balancer
func (c *Config) Validation() error {
	if c.Schema != "http" && c.Schema != "https" {
		return fmt.Errorf("the schema \"%s\" not supported", c.Schema)
	}
	if len(c.Location) == 0 {
		return errors.New("the details of location cannot be null")
	}
	if c.Schema == "https" && (len(c.SSLCertificate) == 0 || len(c.SSLCertificateKey) == 0) {
		return errors.New("the https proxy requires ssl_certificate_key and ssl_certificate")
	}
	if c.HealthCheckInterval < 1 {
		return errors.New("health_check_interval must be greater than 0")
	}
	return nil
}
