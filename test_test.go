package main

import (
	"io"
	"io/ioutil"
	"os"
)

var fakeGit *FakeGit
var fakeDatabase *FakeDatabase

func init() {
	fakeGit = &FakeGit{}
	git = fakeGit

	fakeDatabase = &FakeDatabase{}
	database = fakeDatabase
}

type FakeGit struct {
	FakeRepo              string
	createHooksParameters map[string]interface{}
}

func (g *FakeGit) Retrieve(log io.Writer, url string, path string, branch string, sha string) error {
	files, _ := ioutil.ReadDir("test-repos/" + g.FakeRepo)
	os.MkdirAll(path, 0700)
	for _, file := range files {
		b, _ := ioutil.ReadFile("test-repos/" + g.FakeRepo + "/" + file.Name())
		ioutil.WriteFile(path+"/"+file.Name(), b, 0700)
	}
	return nil
}

func (g *FakeGit) CreateHooks(accessToken string, owner string, repo string) error {
	g.createHooksParameters = map[string]interface{}{
		"accessToken": accessToken,
		"owner":       owner,
		"repository":  repo,
	}
	return nil
}

type FakeDatabase struct {
	GithubBuild *GithubBuild
}

func (f *FakeDatabase) SaveGithubBuild(ghb *GithubBuild) error {
	f.GithubBuild = ghb
	return nil
}

func (f *FakeDatabase) SaveCommit(commit *Commit) error {
	return nil
}

func (f *FakeDatabase) SaveBuild(build *Build) error {
	return nil
}

func (f *FakeDatabase) AllBuilds() []*Build {
	return nil
}

func (f *FakeDatabase) CreateBuild(githubBuild *GithubBuild, build *Build) error {
	return nil
}

func (f *FakeDatabase) FindGithubBuild(owner string, repository string) *GithubBuild {
	if f.GithubBuild != nil {
		if f.GithubBuild.Owner == owner && f.GithubBuild.Repository == repository {
			return f.GithubBuild
		}
	}
	return nil
}

func (f *FakeDatabase) IncompleteBuilds() []*Build {
	return nil
}
