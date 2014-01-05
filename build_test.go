package main

import (
	"strconv"
	"testing"
)

func TestBuildUrl(t *testing.T) {
	setup("")
	defer cleanup()

	database := &PostgresDatabase{}
	githubBuild := &GithubBuild{Owner: "bla", Repository: "repooo"}
	database.SaveGithubBuild(githubBuild)
	build := &Build{
		Owner:      "bla",
		Repository: "repooo",
	}
	database.CreateBuild(githubBuild, build)
	expected := "http://localhost:1212/build_output?id=" + strconv.Itoa(build.Id)
	if build.Url != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v\n", expected, build.Url)
	}
}

func TestBuildUrlPort80(t *testing.T) {
	setup("")
	defer cleanup()

	oldPort := configuration.Port
	configuration.Port = "80"
	defer func() { configuration.Port = oldPort }()

	githubBuild := &GithubBuild{Owner: "bla", Repository: "repooo"}
	database.SaveGithubBuild(githubBuild)
	build := &Build{
		Owner:      "bla",
		Repository: "repooo",
	}
	database.CreateBuild(githubBuild, build)
	expected := "http://localhost/build_output?id=" + strconv.Itoa(build.Id)
	if build.Url != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v\n", expected, build.Url)
	}
}
