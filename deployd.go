package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

var (
	err error
)

type Config struct {
	Static []StaticProject
}

type StaticProject struct {
	Name       string
	Branch     string
	Domain     string
	Subdomain  string
	GitHub     string
	Bucket     string
	Owner      string
	Repository string
}

func main() {
	data, err := ioutil.ReadFile("deployd.conf")

	if err != nil {
		log.Fatal(err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)

	if err != nil {
		log.Fatal(err)
	}
}
