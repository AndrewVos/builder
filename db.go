package main

type Database interface {
	SaveGithubBuild(ghb *GithubBuild) error
	SaveCommit(commit *Commit) error
	SaveBuild(build *Build) error
	AllBuilds() []*Build
	CreateBuild(owner string, repo string, ref string, sha string, githubURL string, commits []Commit) (*Build, error)
	FindGithubBuild(owner string, repository string) *GithubBuild
	IncompleteBuilds() []*Build
}
