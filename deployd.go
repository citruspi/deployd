package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	cachestore "github.com/citruspi/deployd/cache"
	"github.com/citruspi/deployd/configuration"
	"github.com/citruspi/deployd/deploy"
)

var (
	err error
)

func main() {

	config, err := configuration.Configure()

	if err != nil {
		log.Fatal(err)
	}

	lockPath := fmt.Sprintf("%v/deployd.pid", config.Lock)

	if _, err = os.Stat(lockPath); err == nil {
		log.Fatal("Another instance is already running.")
	} else {
		lockFile, err := os.Create(lockPath)
		lockFile.Close()

		if err != nil {
			log.Fatal(err)
		}

		pid := os.Getpid()

		err = ioutil.WriteFile(lockPath, []byte(strconv.Itoa(pid)), 0700)

		if err != nil {
			log.Fatal(err)
		}
	}

	var cache cachestore.Cache

	cache.Path = config.Cache

	err = cache.Load()

	if err != nil {
		log.Fatal(err)
	}

	for _, project := range config.Static.Projects {
		err = deploy.Static(config, cache, project)

		if err != nil {
			log.Fatal(err)
		}
	}

	err = cache.Save()

	if err != nil {
		log.Fatal(err)
	}

	err = os.RemoveAll(lockPath)

	if err != nil {
		log.Fatal(err)
	}
}
