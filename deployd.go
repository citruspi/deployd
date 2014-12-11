package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	cachestore "github.com/citruspi/deployd/cache"
	"github.com/citruspi/deployd/configuration"
	"github.com/citruspi/deployd/unzip"
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
		var processingPath string
		var record cachestore.CacheRecord

		for _, r := range cache.Records {
			if r.Domain == project.Domain && r.Subdomain == project.Subdomain {
				record = r
				break
			}
		}

		if record.Domain == "" {
			record.Domain = project.Domain
			record.Subdomain = project.Subdomain
		}

		processingPath = fmt.Sprintf("%v/.deployd.processing", config.Static.Path)

		if _, err = os.Stat(processingPath); os.IsNotExist(err) {
			err = os.MkdirAll(processingPath, 0700)

			if err != nil {
				log.Fatal(err)
			}
		}

		archivePath := fmt.Sprintf("%v/%v-%v.zip", processingPath, project.Name, project.Branch)

		err = os.RemoveAll(archivePath)

		if err != nil {
			log.Fatal(err)
		}

		archive, err := os.Create(archivePath)
		defer archive.Close()

		if err != nil {
			log.Fatal(err)
		}

		var archiveLocation string

		if project.GitHub {
			archiveLocation = fmt.Sprintf("https://github.com/%v/%v/archive/%v.zip", project.Owner, project.Repository, project.Branch)
		} else {
			archiveLocation = fmt.Sprintf("https://s3.amazonaws.com/%v/%v-latest.zip", project.Bucket, project.Branch)
		}

		response, err := http.Get(archiveLocation)
		defer response.Body.Close()

		if err != nil {
			log.Fatal(err)
		}

		_, err = io.Copy(archive, response.Body)

		if err != nil {
			log.Fatal(err)
		}

		archiveData, err := ioutil.ReadFile(archivePath)

		if err != nil {
			log.Fatal(err)
		}

		rawChecksum := md5.Sum(archiveData)
		checksum := fmt.Sprintf("%x", rawChecksum)

		if checksum == record.Checksum {
			fmt.Println("Duplicate")
			continue
		}

		record.Checksum = checksum
		record.LastDeployed = time.Now()

		cache.Records = append(cache.Records, record)

		unarchivedPath := fmt.Sprintf("%v/%v-%v", processingPath, project.Name, project.Branch)

		err = os.RemoveAll(unarchivedPath)

		if err != nil {
			log.Fatal(err)
		}

		err = unzip.Unzip(archivePath, unarchivedPath)

		if err != nil {
			log.Fatal(err)
		}

		if project.GitHub {
			unarchivedPath = fmt.Sprintf("%v/%v-%v", unarchivedPath, project.Repository, project.Branch)
		}

		domainPath := fmt.Sprintf("%v/%v", config.Static.Path, project.Domain)
		projectPath := fmt.Sprintf("%v/%v/%v", config.Static.Path, project.Domain, project.Subdomain)

		err = os.RemoveAll(projectPath)

		if err != nil {
			log.Fatal(err)
		}

		os.Mkdir(domainPath, 0700)

		err = os.Rename(unarchivedPath, projectPath)

		if err != nil {
			log.Fatal(err)
		}

		err = os.RemoveAll(archivePath)

		if err != nil {
			log.Fatal(err)
		}

		err = os.RemoveAll(processingPath)

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
