package deploy

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	cachestore "github.com/citruspi/deployd/cache"
	"github.com/citruspi/deployd/configuration"
	"github.com/citruspi/deployd/unzip"
)

var (
	err error
)

func Static(config configuration.Config, cache cachestore.Cache, name string, domain string, subdomain string, branch string, github bool, bucket string, owner string, repository string) (err error) {

	var processingPath string
	var record cachestore.CacheRecord

	for _, r := range cache.Records {
		if r.Domain == domain && r.Subdomain == subdomain {
			record = r
			break
		}
	}

	if record.Domain == "" {
		record.Domain = domain
		record.Subdomain = subdomain
	}

	processingPath = fmt.Sprintf("%v/.deployd.processing", config.Static.Path)

	if _, err = os.Stat(processingPath); os.IsNotExist(err) {
		err = os.MkdirAll(processingPath, 0700)

		if err != nil {
			log.Fatal(err)
		}
	}

	archivePath := fmt.Sprintf("%v/%v-%v.zip", processingPath, name, branch)

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

	if github {
		archiveLocation = fmt.Sprintf("https://github.com/%v/%v/archive/%v.zip", owner, repository, branch)
	} else {
		archiveLocation = fmt.Sprintf("https://s3.amazonaws.com/%v/%v-latest.zip", bucket, branch)
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
		return err
	}

	record.Checksum = checksum
	record.LastDeployed = time.Now()

	cache.Records = append(cache.Records, record)

	unarchivedPath := fmt.Sprintf("%v/%v-%v", processingPath, name, branch)

	err = os.RemoveAll(unarchivedPath)

	if err != nil {
		log.Fatal(err)
	}

	err = unzip.Unzip(archivePath, unarchivedPath)

	if err != nil {
		log.Fatal(err)
	}

	if github {
		unarchivedPath = fmt.Sprintf("%v/%v-%v", unarchivedPath, repository, branch)
	}

	domainPath := fmt.Sprintf("%v/%v", config.Static.Path, domain)
	projectPath := fmt.Sprintf("%v/%v/%v", config.Static.Path, domain, subdomain)

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

	return nil
}
