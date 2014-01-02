package main

import (
	"os"
)

type Configuration struct {
	GithubClientID     string
	GithubClientSecret string
	Host               string
	Port               string
}

var configuration Configuration

func init() {
	configuration = Configuration{
		GithubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GithubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Host:               os.Getenv("HOST"),
		Port:               os.Getenv("PORT"),
	}

	if configuration.Host == "" {
		configuration.Host = "http://localhost"
	}
	if configuration.Port == "" {
		configuration.Port = "1212"
	}
}
