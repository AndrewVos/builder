package main

type Database interface {
	SaveGithubBuild(ghb *GithubBuild) error
	SaveCommit(commit *Commit) error
	SaveBuild(build *Build) error
	AllBuilds() []*Build
	CreateBuild(githubBuild *GithubBuild, build *Build) error
	FindGithubBuild(owner string, repository string) *GithubBuild
	IncompleteBuilds() []*Build
}
