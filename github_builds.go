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

func addGithubBuild(accessToken string, owner string, repo string) error {
	err := createHooks(accessToken, owner, repo)
	if err != nil {
		return err
	}

	githubBuilds = append(githubBuilds, GithubBuild{
		AccessToken:     accessToken,
		RepositoryOwner: owner,
		RepositoryName:  repo,
	})
	return nil
}

func init() {
	githubBuilds = []GithubBuild{}
}
