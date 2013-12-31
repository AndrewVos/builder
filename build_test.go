package main

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestBuildUrl(t *testing.T) {
	setup("")
	defer cleanup()

	build := NewBuild("", "", "", "", "", nil)
	expected := "http://example.org:1212/build_output?id=" + build.ID
	if build.URL != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v\n", expected, build.URL)
	}
}

func TestBuildUrlPort80(t *testing.T) {
	setup("")
	defer cleanup()

	builderJson := `{
      "AuthToken": "lolsszz",
      "Host": "http://example.org",
      "Port": "80",
      "Repositories": [
        {"Owner": "AndrewVos", "Repository": "builder"}
      ]
    }`
	builderJson = strings.TrimSpace(builderJson)
	ioutil.WriteFile("data/builder.json", []byte(builderJson), 0700)

	build := NewBuild("", "", "", "", "", nil)
	expected := "http://example.org/build_output?id=" + build.ID
	if build.URL != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v\n", expected, build.URL)
	}
}
