package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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
	GitHub     bool
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

	for _, project := range config.Static {
		/*
			path = fmt.Sprintf("/tmp/%v-%v", project.Name, project.Branch)
			dir, err := os.Stat(path)

			if path.IsDir() {
				os.Remove(path)
			}

			err = os.MkDir(path, 0700)

			if err != nil {
				log.Fatal(err)
			}
		*/

		archivePath := fmt.Sprintf("/tmp/%v-%v.zip", project.Name, project.Branch)

		if _, err := os.Stat(archivePath); err == nil {
			err := os.Remove(archivePath)

			if err != nil {
				log.Fatal(err)
			}
		}

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
	}
}
