package main

import (
	"archive/zip"
	"crypto/md5"
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
	Cache   string
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

type CacheRecord struct {
	Domain    string
	Subdomain string
	Checksum  string
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

	var cache []CacheRecord

	if _, err = os.Stat(config.Paths.Cache); os.IsNotExist(err) {
		cacheFile, err := os.Create(config.Paths.Cache)
		cacheFile.Close()

		if err != nil {
			log.Fatal(err)
		}
	} else {
		cacheData, err := ioutil.ReadFile(config.Paths.Cache)

		if err != nil {
			log.Fatal(err)
		}

		err = yaml.Unmarshal(cacheData, &cache)

		if err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(config.Paths.Deployd); os.IsNotExist(err) {
		os.MkdirAll(config.Paths.Deployd, 0700)
	}

	if _, err := os.Stat(config.Paths.Static); os.IsNotExist(err) {
		os.MkdirAll(config.Paths.Static, 0700)
	}

	for _, project := range config.Static {
		var record CacheRecord

		for _, r := range cache {
			if r.Domain == project.Domain && r.Subdomain == project.Subdomain {
				record = r
				break
			}
		}

		if record.Domain == "" {
			record.Domain = project.Domain
			record.Subdomain = project.Subdomain
		}

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

		archiveData, err := ioutil.ReadFile(archivePath)

		if err != nil {
			log.Fatal(err)
		}

		rawChecksum := md5.Sum(archiveData)
		checksum := fmt.Sprintf("%x", rawChecksum)

		if checksum == record.Checksum {
			continue
		}
		fmt.Println("Replacing")
		record.Checksum = checksum

		cache = append(cache, record)

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

		domainPath := fmt.Sprintf("%v/%v", config.Paths.Static, project.Domain)
		projectPath := fmt.Sprintf("%v/%v/%v", config.Paths.Static, project.Domain, project.Subdomain)

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

	cacheData, err := yaml.Marshal(&cache)

	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(config.Paths.Cache, cacheData, 0700)

	if err != nil {
		log.Fatal(err)
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
