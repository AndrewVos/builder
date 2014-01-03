package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

type GithubBuild struct {
	AccessToken     string
	RepositoryName  string
	RepositoryOwner string
}

var githubBuilds []GithubBuild

var connection *sql.DB

func connect() (*sql.DB, error) {
	if connection == nil {
		connection, err := sql.Open("postgres", "user=builder password="+configuration.PostgresPassword+" dbname=builder sslmode=disable")
		return connection, err
	}
	err := connection.Ping()
	if err != nil {
		return nil, err
	}
	return connection, nil
}

func findGithubBuild(owner string, repository string) (GithubBuild, bool) {
	db, err := connect()
	if err != nil {
		log.Println(err)
		return GithubBuild{}, false
	}

	rows, err := db.Query(`
    SELECT (access_token) FROM github_builds
      WHERE   owner      = $1
      AND     repository = $2
  `, owner, repository)

	if err != nil {
		return GithubBuild{}, false
	}

	rows.Next()
	var accessToken string
	err = rows.Scan(&accessToken)
	if err != nil {
		return GithubBuild{}, false
	}
	return GithubBuild{AccessToken: accessToken, RepositoryOwner: owner, RepositoryName: repository}, true
}

func (ghb GithubBuild) Save() error {
	db, err := connect()
	if err != nil {
		return err
	}

	_, err = db.Query(`INSERT INTO github_builds(access_token, owner, repository)
    VALUES($1, $2, $3)`, ghb.AccessToken, ghb.RepositoryOwner, ghb.RepositoryName)

	if err != nil {
		return err
	}

	return nil
}

func addGithubBuild(accessToken string, owner string, repo string) error {
	err := createHooks(accessToken, owner, repo)

	ghb := GithubBuild{
		AccessToken:     accessToken,
		RepositoryOwner: owner,
		RepositoryName:  repo,
	}

	err = ghb.Save()

	if err != nil {
		return err
	}

	return nil
}
