package flywheel

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"time"
)

// Config flywheel config file
type Config struct {
	Vhosts      map[string]string `json:"vhosts"`
	Region      string            `json:"aws_region"`
	Endpoint    string            `json:"endpoint"`
	Instances   []string          `json:"instances"`
	HcInterval  Duration          `json:"healthcheck-interval"`
	IdleTimeout Duration          `json:"idle-timeout"`
	AutoScaling AutoScalingConfig `json:"autoscaling"`
}

// AutoScalingConfig list of terminate/stop AWS ASG
type AutoScalingConfig struct {
	Terminate map[string]int64 `json:"terminate"`
	Stop      []string         `json:"stop"`
}

// Duration helper type to parse duration from json
type Duration time.Duration

// UnmarshalText - unmarshal duration from JSON
func (d *Duration) UnmarshalText(b []byte) error {
	v, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

// ReadConfig - read config file from a file
func ReadConfig(filename string) (*Config, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	cfg := &Config{}
	if err = cfg.Parse(fd); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Parse the config file
func (c *Config) Parse(rd io.Reader) error {
	in, err := ioutil.ReadAll(rd)
	if err != nil {
		return err
	}
	err = json.Unmarshal(in, c)
	if err != nil {
		return fmt.Errorf("Could not decode json: %v", err)
	}

	err = c.Validate()
	if err != nil {
		return fmt.Errorf("Invalid configuration: %v", err)
	}
	return nil
}

// AwsInstances retrieve a list of AWS instance id as requested by AWK SDK
func (c *Config) AwsInstances() []*string {
	awsIds := make([]*string, len(c.Instances))
	for i := range c.Instances {
		awsIds[i] = &c.Instances[i]
	}
	return awsIds
}

// EndpointURL get endpoint URL as an URL type
func (c *Config) EndpointURL() (*url.URL, error) {
	return url.Parse(c.Endpoint)
}

// Validate config content
func (c *Config) Validate() error {
	if len(c.Instances) == 0 && len(c.AutoScaling.Stop) == 0 && len(c.AutoScaling.Terminate) == 0 {
		return fmt.Errorf("No instances or asg configured")
	}

	if len(c.Endpoint) == 0 {
		return fmt.Errorf("No endpoint configured")
	}

	if c.HcInterval <= 0 {
		c.HcInterval = Duration(30 * time.Second)
	}

	if c.IdleTimeout <= 0 {
		c.IdleTimeout = Duration(3 * time.Hour)
	}

	if c.Region == "" {
		c.Region = "ap-southeast-2"
	}

	return nil
}
