package main

import (
	"crypto/rand"
	"encoding/base64"
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

func (p *PostgresDatabase) AddRepositoryToAccount(account *Account, repository *Repository) error {
	db, err := connect()
	if err != nil {
		return err
	}

	err = db.Query(`
    INSERT INTO repositories (account_id, owner, repository)
      VALUES ($1, $2, $3)
    `, account.Id, repository.Owner, repository.Repository).Run()

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
	err = db.Query("SELECT * FROM builds ORDER BY id").Rows(&builds)

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

func (p *PostgresDatabase) CreateBuild(repository *Repository, build *Build) error {
	db, err := connect()
	if err != nil {
		return err
	}

	var m []int
	err = db.Query(`
    INSERT INTO builds (repository_id)
      VALUES ($1)
      RETURNING (id)
    `, repository.Id,
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

	p.SaveBuild(build)

	for _, commit := range build.Commits {
		commit.BuildId = build.Id
		p.SaveCommit(&commit)
	}

	return nil
}

func (p *PostgresDatabase) FindRepository(owner string, name string) *Repository {
	db, err := connect()
	if err != nil {
		log.Println(err)
		return nil
	}

	var repositories []*Repository
	err = db.Query(`
    SELECT * FROM repositories
      WHERE   owner      = $1
      AND     repository = $2
    `, owner, name).Rows(&repositories)

	if err != nil {
		log.Println(err)
		return nil
	}

	if len(repositories) == 0 {
		return nil
	}

	return repositories[0]
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

func (p *PostgresDatabase) FindAccountById(id int) *Account {
	db, err := connect()
	if err != nil {
		log.Println(err)
		return nil
	}

	var account *Account
	err = db.Query(`
    SELECT * FROM accounts
      WHERE id = $1
  `, id).Rows(&account)

	if err != nil {
		log.Println(err)
		return nil
	}

	if account == nil {
		return nil
	}

	var repositories []*Repository
	err = db.Query(`
  SELECT * FROM repositories
    WHERE account_id = $1
  `, account.Id).Rows(&repositories)
	if err != nil {
		log.Println(err)
	}

	account.Repositories = repositories

	return account
}

func (p *PostgresDatabase) CreateAccount(account *Account) error {
	db, err := connect()
	if err != nil {
		log.Println(err)
		return err
	}

	if found := p.FindAccountById(account.Id); found != nil {
		err = db.Query(`
    UPDATE accounts
      SET
        (access_token) = ($1)
      WHERE id = $2
    `, account.AccessToken, found.Id).Run()
		if err != nil {
			fmt.Println("Error updating account: ", err)
			return err
		}
		account.Id = found.Id
	} else {
		err = db.Query(`
      INSERT INTO accounts (id, access_token)
      VALUES ($1, $2)
    `, account.Id, account.AccessToken).Run()

		if err != nil {
			fmt.Println("Error creating account: ", err)
			return err
		}
	}

	return nil
}

func (p *PostgresDatabase) CreateLoginForAccount(account *Account) (*Login, error) {
	db, err := connect()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	b := make([]byte, 50)
	rand.Read(b)
	encoder := base64.URLEncoding
	token := make([]byte, encoder.EncodedLen(len(b)))
	encoder.Encode(token, b)
	t := fmt.Sprintf("%s", token)

	var id int
	err = db.Query(`
      INSERT INTO logins (account_id, token)
      VALUES ($1, $2)
      RETURNING (id)
    `, account.Id, t).Rows(&id)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &Login{
		Id:        id,
		AccountId: account.Id,
		Token:     t,
	}, nil
}

func (p *PostgresDatabase) LoginExists(accountId int, token string) bool {
	db, err := connect()
	if err != nil {
		log.Println(err)
		return false
	}

	var count int
	err = db.Query(`
      SELECT COUNT (*) FROM logins
      WHERE account_id = $1
      AND token = $1
    `, accountId, token).Rows(&count)

	if err != nil {
		fmt.Println(err)
		return false
	}

	return count == 1
}
