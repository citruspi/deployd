package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	err error
)

type Config struct {
	Static []StaticProject
	Paths  PathConfig
}

type PathConfig struct {
	Deployd string
	Static  string
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
	configPath := flag.String("config", "/etc/deployd.conf", "Path to the config file")

	flag.Parse()

	if _, err = os.Stat(*configPath); os.IsNotExist(err) {
		log.Fatal(err)
	}

	data, err := ioutil.ReadFile(*configPath)

	if err != nil {
		log.Fatal(err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)

	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(config.Paths.Deployd); os.IsNotExist(err) {
		os.MkdirAll(config.Paths.Deployd, 0700)
	}

	for _, project := range config.Static {
		archivePath := fmt.Sprintf("%v/%v-%v.zip", config.Paths.Deployd, project.Name, project.Branch)

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

		unarchivedPath := fmt.Sprintf("%v/%v-%v", config.Paths.Deployd, project.Name, project.Branch)

		err = os.RemoveAll(unarchivedPath)

		if err != nil {
			log.Fatal(err)
		}

		err = unzip(archivePath, unarchivedPath)

		if err != nil {
			log.Fatal(err)
		}

		if project.GitHub {
			unarchivedPath = fmt.Sprintf("%v/%v-%v", unarchivedPath, project.Repository, project.Branch)
		}

		domainPath := fmt.Sprintf("srv/%v", project.Domain)
		projectPath := fmt.Sprintf("srv/%v/%v", project.Domain, project.Subdomain)

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
	}
}

// unzip function by http://stackoverflow.com/users/1129149/swtdrgn
// http://stackoverflow.com/questions/20357223/easy-way-to-unzip-file-with-golang/24430720#24430720

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, f.Mode())
			if err != nil {
				log.Fatal(err)
				return err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
