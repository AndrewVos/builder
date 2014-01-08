package main

import (
	"testing"
)

func createCleanPostgresDatabase() *PostgresDatabase {
	db, _ := connect()
	db.Query("DELETE FROM repositories").Run()
	db.Query("DELETE FROM builds").Run()
	db.Query("DELETE FROM commits").Run()
	db.Query("DELETE FROM accounts").Run()
	db.Query("DELETE FROM logins").Run()
	return &PostgresDatabase{}
}

func TestAllBuildsLoadsCommits(t *testing.T) {
	db := createCleanPostgresDatabase()

	repository := &Repository{Owner: "ownerrr", Repository: "repo1"}
	db.SaveRepository(repository)

	commits := []Commit{
		Commit{Sha: "csdkl22323", Message: "hellooo", Url: "something.com"},
		Commit{Sha: "324mlkm", Message: "hi there", Url: "example.com"},
	}

	build := &Build{Owner: "ownerrr", Repository: "repo1", Commits: commits}
	db.CreateBuild(repository, build)

	build = db.AllBuilds()[0]
	if len(build.Commits) != len(commits) {
		t.Fatalf("Expected build to load up %d commits, but had %d commits\n", len(commits), len(build.Commits))
	}
	for index, expectedCommit := range commits {
		actual := build.Commits[index]
		if actual.Sha != expectedCommit.Sha ||
			actual.Message != expectedCommit.Message ||
			actual.Url != expectedCommit.Url {
			t.Errorf("Expected commit to look like:\n%+v\nActual:\n%+v\n", expectedCommit, actual)
		}
	}
}

func TestFindRepository(t *testing.T) {
	db := createCleanPostgresDatabase()
	b1 := &Repository{Owner: "ownerrr", Repository: "repo1"}
	b2 := &Repository{Owner: "erm", Repository: "repo2"}
	db.SaveRepository(b1)
	db.SaveRepository(b2)

	repository := db.FindRepository("erm", "repo2")

	if repository == nil || repository.Owner != "erm" || repository.Repository != "repo2" {
		t.Errorf("Expected to find repository:\n%+v\nActual:\n%+v\n", b2, repository)
	}

	repository = db.FindRepository("losdsds", "sd")
	if repository != nil {
		t.Errorf("Expected not to find a repository, but found:\n%+v\n", repository)
	}
}

func TestIncompleteBuilds(t *testing.T) {
	db := createCleanPostgresDatabase()
	repository := &Repository{Owner: "ownerrr", Repository: "repo1"}
	db.SaveRepository(repository)

	build := &Build{Owner: "ownerrr", Repository: "repo1"}
	db.CreateBuild(repository, build)
	build.Complete = true
	db.SaveBuild(build)

	build = &Build{Owner: "ownerrr", Repository: "repo1"}
	db.CreateBuild(repository, build)
	build.Complete = false
	db.SaveBuild(build)

	builds := db.IncompleteBuilds()
	if len(builds) != 1 {
		t.Errorf("We should only return one build, because only one is incomplete\n%+v\n", builds)
	}
}

func TestCreateAndFindAccountById(t *testing.T) {
	db := createCleanPostgresDatabase()

	account := &Account{GithubUserId: 2455252, AccessToken: "23mf23f22n3kl2n3nkl2n3lnl2n3ln3lnl"}
	db.CreateAccount(account)

	account = db.FindAccountById(account.Id)
	if account == nil {
		t.Fatalf("Account wasn't found")
	}
	if account.Id != account.Id || account.AccessToken != "23mf23f22n3kl2n3nkl2n3lnl2n3ln3lnl" {
		t.Errorf("Expected account to be found")
	}
}

func TestCreateAndFindAccountByGithubUserId(t *testing.T) {
	db := createCleanPostgresDatabase()

	account := &Account{
		GithubUserId: 2455252,
		AccessToken:  "23mf23f22n3kl2n3nkl2n3lnl2n3ln3lnl",
	}

	db.CreateAccount(account)
	if account.Id == 0 {
		t.Fatalf("Account id should have been auto-incremented")
	}

	account = db.FindAccountByGithubUserId(account.GithubUserId)

	if account == nil {
		t.Fatalf("Account wasn't found")
	}

	if account.GithubUserId != 2455252 {
		t.Errorf("Account GithubUserId wasn't stored")
	}

	if account.AccessToken != "23mf23f22n3kl2n3nkl2n3lnl2n3ln3lnl" {
		t.Errorf("Account AccessToken wasn't stored")
	}
}

func TestCreateAccountUpdatesAccessToken(t *testing.T) {
	db := createCleanPostgresDatabase()

	account1 := &Account{GithubUserId: 2455252, AccessToken: "ZZZZZZZZZZ"}
	account2 := &Account{GithubUserId: 2455252, AccessToken: "AAAAAAAAAA"}
	db.CreateAccount(account1)
	db.CreateAccount(account2)

	foundAccount := db.FindAccountByGithubUserId(2455252)

	if foundAccount.Id != account1.Id {
		t.Errorf("Expected account to not be created again")
	}

	if foundAccount.AccessToken != "AAAAAAAAAA" {
		t.Errorf("Expected access token to get updated")
	}
}

func TestLoginExists(t *testing.T) {
	db := createCleanPostgresDatabase()

	account := &Account{GithubUserId: 2455252, AccessToken: "5T"}
	db.CreateAccount(account)

	login, _ := db.CreateLoginForAccount(account)

	if login.Token == "" {
		t.Error("Expected random token to be generated")
	}

	if db.LoginExists(account.Id, login.Token) == false {
		t.Error("Expected login to be valid")
	}

	if db.LoginExists(account.Id, "not a real token") {
		t.Error("Expected login to not be valid")
	}

	if db.LoginExists(123, "not a real token") {
		t.Error("Expected login to not be valid")
	}
}
