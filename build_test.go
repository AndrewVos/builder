package main

import (
	"testing"
)

func TestBuildUrl(t *testing.T) {
	setup("")
	defer cleanup()

	build := NewBuild("", "", "", "", "", nil)
	expected := "http://localhost:1212/build_output?id=" + build.ID
	if build.URL != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v\n", expected, build.URL)
	}
}

func TestBuildUrlPort80(t *testing.T) {
	setup("")
	defer cleanup()

	oldPort := configuration.Port
	configuration.Port = "80"
	defer func() { configuration.Port = oldPort }()

	build := NewBuild("", "", "", "", "", nil)
	expected := "http://localhost/build_output?id=" + build.ID
	if build.URL != expected {
		t.Errorf("Expected:\n%v\nGot:\n%v\n", expected, build.URL)
	}
}
