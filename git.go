package main

import (
	"io"
	"os/exec"
)

type GitTool interface {
	Retrieve(log io.Writer, url string, path string, branch string, sha string) error
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
