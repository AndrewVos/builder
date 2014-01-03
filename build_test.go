package main

import (
	"strconv"
	"testing"
)

func TestBuildUrl(t *testing.T) {
	setup("")
	defer cleanup()

	GithubBuild{Owner: "bla", Repository: "repooo"}.Save()
	build, _ := CreateBuild("bla", "repooo", "", "", "", nil)
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

	GithubBuild{Owner: "bla", Repository: "repooo"}.Save()
	build, _ := CreateBuild("bla", "repooo", "", "", "", nil)
	expected := "http://localhost/build_output?id=" + strconv.Itoa(build.Id)
	if build.Url != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v\n", expected, build.Url)
	}
}
