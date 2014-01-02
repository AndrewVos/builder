package main

type GithubBuild struct {
	AccessToken     string
	RepositoryName  string
	RepositoryOwner string
}

var githubBuilds []GithubBuild

func findGithubBuild(owner string, name string) (GithubBuild, bool) {
	for _, build := range githubBuilds {
		if build.RepositoryOwner == owner && build.RepositoryName == name {
			return build, true
		}
	}
	return GithubBuild{}, false
}

func init() {
	githubBuilds = []GithubBuild{}
}
