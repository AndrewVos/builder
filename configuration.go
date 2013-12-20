package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Configuration struct {
	AuthToken    string
	Host         string
	Port         string
	Repositories []Repository
}

func CurrentConfiguration() Configuration {
	path := "builder.json"
	if len(os.Args) == 3 && os.Args[1] == "-c" {
		path = os.Args[2]
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var c Configuration
	err = json.Unmarshal(b, &c)
	if err != nil {
		panic(err)
	}
	return c
}

type Repository struct {
	Owner      string
	Repository string
}
