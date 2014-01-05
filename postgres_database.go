package main

import (
	"fmt"
	"github.com/eaigner/jet"
	_ "github.com/lib/pq"
	"log"
	"strconv"
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

func (p *PostgresDatabase) AllBuilds() []*Build {
	db, err := connect()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var builds []*Build
	err = db.Query("SELECT * FROM builds").Rows(&builds)

	if err != nil {
		fmt.Println("Error getting all builds: ", err)
		return nil
	}

	if len(builds) > 0 {
		var buildIds []int
		buildsById := map[int]*Build{}
		for _, build := range builds {
			buildIds = append(buildIds, build.Id)
			buildsById[build.Id] = build
		}

		var commits []Commit
		err = db.Query("SELECT * FROM commits WHERE build_id IN ( $1 )", buildIds).Rows(&commits)
		if err != nil {
			fmt.Println("Error getting commits:", err)
			return nil
		}

		for _, commit := range commits {
			build := buildsById[commit.BuildId]
			build.Commits = append(build.Commits, commit)
		}
	}

	return builds
}

func (p *PostgresDatabase) CreateBuild(githubBuild *GithubBuild, build *Build) error {
	db, err := connect()
	if err != nil {
		return err
	}

	var m []int
	err = db.Query(`
    INSERT INTO builds (github_build_id)
      VALUES ($1)
      RETURNING (id)
    `, githubBuild.Id,
	).Rows(&m)

	if err != nil {
		fmt.Println(err)
		return err
	}
	buildId := m[0]

	build.Id = buildId
	build.Result = "incomplete"
	build.Url = configuration.Host
	if configuration.Port != "80" {
		build.Url += ":" + configuration.Port
	}
	build.Url += "/build_output?id=" + strconv.Itoa(build.Id)

	database.SaveBuild(build)

	for _, commit := range build.Commits {
		commit.BuildId = build.Id
		database.SaveCommit(&commit)
	}

	return nil
}

func (p *PostgresDatabase) FindGithubBuild(owner string, repository string) *GithubBuild {
	db, err := connect()
	if err != nil {
		log.Println(err)
		return nil
	}

	var builds []*GithubBuild
	err = db.Query(`
    SELECT * FROM github_builds
      WHERE   owner      = $1
      AND     repository = $2
    `, owner, repository).Rows(&builds)

	if err != nil {
		log.Println(err)
		return nil
	}

	if len(builds) == 0 {
		return nil
	}

	return builds[0]
}

func (p *PostgresDatabase) IncompleteBuilds() []*Build {
	db, err := connect()
	if err != nil {
		log.Println(err)
		return nil
	}

	var builds []*Build
	err = db.Query(`
    SELECT * FROM builds
      WHERE   complete = $1
    `, false).Rows(&builds)

	if err != nil {
		log.Println(err)
		return nil
	}

	return builds
}
