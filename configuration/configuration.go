package configuration

import (
	"flag"
	"io/ioutil"
	"os"
	"os/user"

	"gopkg.in/yaml.v2"
)

var (
	err    error
	config Config
)

type Config struct {
	Cache  string
	Static StaticConfig
	Lock   string
}

type StaticConfig struct {
	Path     string
	Projects []StaticProject
}

type StaticProject struct {
	Name       string
	Branch     string
	Domain     string
	Subdomain  string
	GitHub     bool
	Bucket     string
	Owner      string
	Repository string
}

func Configure() (configuration Config, err error) {
	configPath := flag.String("config", "/etc/deployd.conf", "Path to the config file")

	flag.Parse()

	if _, err = os.Stat(*configPath); os.IsNotExist(err) {
		return config, err
	}

	data, err := ioutil.ReadFile(*configPath)

	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)

	if err != nil {
		return config, err
	}

	if config.Lock == "" {
		config.Lock = "/var/run"
	}

	if config.Cache == "" {
		usr, _ := user.Current()
		config.Cache = usr.HomeDir
	}

	if len(config.Static.Projects) > 0 {
		if config.Static.Path == "" {
			config.Static.Path = "/srv"
		}
	}

	return config, nil
}
