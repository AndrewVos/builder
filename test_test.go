package main

import (
	"io/ioutil"
	"os"
	"strings"
)

func setup() {
	os.Mkdir("data", 0700)
	builderJson := `{
      "AuthToken": "lolsszz",
      "Host": "http://example.org",
      "Port": "1212",
      "Repositories": [
        {"Owner": "AndrewVos", "Repository": "builder"}
      ]
    }`
	builderJson = strings.TrimSpace(builderJson)
	ioutil.WriteFile("data/builder.json", []byte(builderJson), 0700)
}

func cleanup() {
	os.RemoveAll("data")
}
