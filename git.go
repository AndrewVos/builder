package main

import (
	"errors"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

type GitTool interface {
	Retrieve(log io.Writer, url string, path string, branch string, sha string) error
	CreateHooks(accessToken string, owner string, repo string) error
}

type Git struct{}

func (git Git) Retrieve(log io.Writer, url string, path string, branch string, sha string) error {
	cmd := exec.Command("git", "clone", "--quiet", "--depth=50", "--branch", branch, url, path)
	cmd.Stdout = log
	cmd.Stderr = log
	err := cmd.Run()

	if err != nil {
		return err
	}

	cmd = exec.Command("git", "checkout", "--quiet", sha)
	cmd.Dir = path
	cmd.Stdout = log
	cmd.Stderr = log

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (git Git) CreateHooks(accessToken string, owner string, repo string) error {
	url := githubDomain + "/repos/" + owner + "/" + repo + "/hooks?access_token=" + accessToken

	supportedEvents := []string{"push", "pull_request"}
	for _, event := range supportedEvents {
		body := `{
      "name": "web",
      "active": true,
      "events": [ "` + event + `" ],
      "config": {
        "url": "` + configuration.Host + ":" + configuration.Port + `/hooks/` + event + `",
        "content_type": "json"
      }
    }`

		client := &http.Client{}
		request, _ := http.NewRequest("POST", url, strings.NewReader(body))
		response, err := client.Do(request)
		if err != nil {
			return err
		}

		if response.StatusCode == 401 {
			return errors.New("Access Token appears to be invalid")
		}
	}
	return nil
}
