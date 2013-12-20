package main

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	AuthToken    string
	Host         string
	Port         string
	Repositories []Repository
}

func CurrentConfiguration() Configuration {
	b, err := ioutil.ReadFile("data/builder.json")
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
