package main

import (
	"errors"
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

func (p *PostgresDatabase) CreateBuild(owner string, repo string, ref string, sha string, githubURL string, commits []Commit) (*Build, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}

	githubBuild, exists := FindGithubBuild(owner, repo)
	if !exists {
		err = errors.New(fmt.Sprintf("Someone tried to create a build but we don't have an access token :/\n%v/%v", owner, repo))
		fmt.Println(err)
		return nil, err
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
		return nil, err
	}
	build_id := m[0]

	build := &Build{
		Id:         build_id,
		Owner:      owner,
		Repository: repo,
		Ref:        ref,
		Sha:        sha,
		Result:     "incomplete",
		GithubUrl:  githubURL,
		Commits:    commits,
	}

	build.Url = configuration.Host
	if configuration.Port != "80" {
		build.Url += ":" + configuration.Port
	}
	build.Url += "/build_output?id=" + strconv.Itoa(build.Id)

	database.SaveBuild(build)

	for _, commit := range commits {
		commit.BuildId = build.Id
		database.SaveCommit(&commit)
	}

	return build, nil
}
