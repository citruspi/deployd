package cache

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

var (
	err error
)

type Cache struct {
	Records []CacheRecord
	Path    string
}

type CacheRecord struct {
	Domain       string
	Subdomain    string
	Checksum     string
	LastDeployed time.Time
}

func (c Cache) Load() (err error) {
	if _, err = os.Stat(c.Path); os.IsNotExist(err) {
		cacheFile, err := os.Create(c.Path)
		cacheFile.Close()

		if err != nil {
			return err
		} else {
			return nil
		}
	} else {
		cacheData, err := ioutil.ReadFile(c.Path)

		if err != nil {
			return err
		}

		err = yaml.Unmarshal(cacheData, &c.Records)

		return err
	}
}

func (c Cache) Save() (err error) {
	cacheData, err := yaml.Marshal(&c.Records)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.Path, cacheData, 0700)

	return err
}
