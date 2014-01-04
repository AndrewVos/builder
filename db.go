package main

import (
	"fmt"
	"github.com/eaigner/jet"
	_ "github.com/lib/pq"
	"log"
)

var connection *jet.Db

func connect() (*jet.Db, error) {
	if connection == nil {
		c, err := jet.Open("postgres", "host=/var/run/postgresql dbname=builder sslmode=disable")
		connection = c
		return connection, err
	}
	err := connection.Ping()
	if err != nil {
		return nil, err
	}
	return connection, nil
}

type Database interface {
	SaveGithubBuild(ghb *GithubBuild) error
	SaveCommit(commit *Commit) error
	SaveBuild(build *Build) error
}

type PostgresDatabase struct {
}

func (p *PostgresDatabase) SaveGithubBuild(ghb *GithubBuild) error {
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

func (p *PostgresDatabase) SaveCommit(commit *Commit) error {
	db, err := connect()
	if err != nil {
		return err
	}

	err = db.Query(`
    INSERT INTO commits (build_id, sha, message, url)
      VALUES ($1, $2, $3, $4)
      RETURNING *
    `, commit.BuildId, commit.Sha, commit.Message, commit.Url,
	).Rows(&commit)

	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresDatabase) SaveBuild(build *Build) error {
	db, err := connect()
	if err != nil {
		return err
	}

	err = db.Query(`
    UPDATE builds
      SET
        (url, owner, repository, ref, sha, complete, success, result, github_url) = ($1, $2, $3, $4, $5, $6, $7, $8, $9)
      WHERE id = $10
	`,
		build.Url,
		build.Owner,
		build.Repository,
		build.Ref,
		build.Sha,
		build.Complete,
		build.Success,
		build.Result,
		build.GithubUrl,
		build.Id,
	).Run()

	if err != nil {
		fmt.Printf("Error saving build:\n%v\nError:\n%v\n", err)
	}
	return err
}
