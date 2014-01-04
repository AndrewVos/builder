package main

import (
	"log"
)

type GithubBuildPersister interface {
	Save(ghb *GithubBuild) error
}

type GithubBuildPostgresPersister struct{}

func (p *GithubBuildPostgresPersister) Save(ghb *GithubBuild) error {
	db, err := connect()
	if err != nil {
		return err
	}

	err = db.Query(`
    INSERT INTO github_builds (access_token, owner, repository)
      VALUES ($1, $2, $3)
    `, ghb.AccessToken, ghb.Owner, ghb.Repository).Run()

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

type GithubBuild struct {
	Id          int
	AccessToken string
	Owner       string
	Repository  string
}

func FindGithubBuild(owner string, repository string) (GithubBuild, bool) {
	db, err := connect()
	if err != nil {
		log.Println(err)
		return GithubBuild{}, false
	}

	var builds []GithubBuild
	err = db.Query(`
    SELECT * FROM github_builds
      WHERE   owner      = $1
      AND     repository = $2
    `, owner, repository).Rows(&builds)

	if err != nil || len(builds) == 0 {
		log.Println(err)
		return GithubBuild{}, false
	}

	return builds[0], true
}
