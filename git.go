package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
)

type GitTool interface {
	Retrieve(log io.Writer, url string, path string, branch string, sha string) error
	CreateHooks(accessToken string, owner string, repo string) error
	GetAccessToken(clientId string, clientSecret string, code string) (string, error)
	GetUserID(accessToken string) (int, error)
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

func (git Git) GetAccessToken(clientId string, clientSecret string, code string) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"client_id":     configuration.GithubClientID,
		"client_secret": configuration.GithubClientSecret,
		"code":          code,
	})

	client := &http.Client{}
	request, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewReader(body))

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	defer response.Body.Close()
	body, _ = ioutil.ReadAll(response.Body)
	var accessTokenResponse map[string]string
	json.Unmarshal(body, &accessTokenResponse)

	return accessTokenResponse["access_token"], nil
}

func (git Git) GetUserID(accessToken string) (int, error) {
	url := "https://api.github.com/user?access_token=" + accessToken
	response, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	b, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	fmt.Println(string(b))
	if err != nil {
		return 0, err
	}
	var m map[string]interface{}
	json.Unmarshal(b, &m)
	return m["id"].(int), nil
}
