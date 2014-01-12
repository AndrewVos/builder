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

	account := &Account{}
	db.CreateAccount(account)
	repository := &Repository{Owner: "ownerrr", Repository: "repo1"}
	db.AddRepositoryToAccount(account, repository)

	commits := []Commit{
		Commit{Sha: "csdkl22323", Message: "hellooo", Url: "something.com"},
		Commit{Sha: "324mlkm", Message: "hi there", Url: "example.com"},
	}

	build := &Build{Owner: "ownerrr", Repository: "repo1", Commits: commits}
	db.CreateBuild(repository, build)

	build = db.AllBuilds(account)[0]
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

func TestFindPublicBuilds(t *testing.T) {
	db := createCleanPostgresDatabase()

	account := &Account{}
	db.CreateAccount(account)

	repository1 := &Repository{Owner: "ownerrr", Repository: "repo1", Public: false}
	repository2 := &Repository{Owner: "ownerrr", Repository: "repo2", Public: true}

	db.AddRepositoryToAccount(account, repository1)
	db.AddRepositoryToAccount(account, repository2)

	db.CreateBuild(repository1, &Build{Owner: "ownerrr", Repository: "repo1"})

	commits := []Commit{
		Commit{Sha: "dssdsd", Message: "hellooo", Url: "something.com"},
	}

	db.CreateBuild(repository2, &Build{Owner: "ownerrr", Repository: "repo2", Commits: commits})

	builds := db.FindPublicBuilds()

	if len(builds) != 1 {
		t.Fatalf("Should have only found one build")
	}
	build := builds[0]

	if build.Owner != "ownerrr" || build.Repository != "repo2" {
		t.Errorf("This wasn't the build on the public repo")
	}
	if len(build.Commits) != len(commits) {
		t.Fatalf("Expected %d commits, but has %d commits\n", len(commits), len(build.Commits))
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

func TestAllBuildsOnlyLoadsBuildsForAccount(t *testing.T) {
	db := createCleanPostgresDatabase()

	account := &Account{Id: 1111}
	db.CreateAccount(account)
	repository := &Repository{Owner: "ownerrr", Repository: "repo1"}
	db.AddRepositoryToAccount(account, repository)
	db.CreateBuild(repository, &Build{Owner: "ownerrr", Repository: "repo1"})

	otherAccount := &Account{Id: 2323}
	db.CreateAccount(otherAccount)
	repository = &Repository{Owner: "something", Repository: "else"}
	db.AddRepositoryToAccount(otherAccount, repository)
	db.CreateBuild(repository, &Build{Owner: "something", Repository: "else"})

	allBuilds := db.AllBuilds(account)
	if len(allBuilds) > 1 {
		t.Errorf("We should only return builds that this account owns, we returned %d", len(allBuilds))
	}
}

func TestFindRepository(t *testing.T) {
	db := createCleanPostgresDatabase()

	account := &Account{Id: 1267}
	db.CreateAccount(account)

	b1 := &Repository{Owner: "ownerrr", Repository: "repo1"}
	b2 := &Repository{Owner: "erm", Repository: "repo2"}
	db.AddRepositoryToAccount(account, b1)
	db.AddRepositoryToAccount(account, b2)

	repository := db.FindRepository("erm", "repo2")

	if repository == nil || repository.Owner != "erm" || repository.Repository != "repo2" {
		t.Errorf("Expected to find repository:\n%+v\nActual:\n%+v\n", b2, repository)
	}
	if repository.Account == nil || repository.Account.Id != account.Id {
		t.Errorf("Repository.Account should be populated")
	}

	repository = db.FindRepository("losdsds", "sd")
	if repository != nil {
		t.Errorf("Expected not to find a repository, but found:\n%+v\n", repository)
	}
}

func TestIncompleteBuilds(t *testing.T) {
	db := createCleanPostgresDatabase()
	account := &Account{}
	db.CreateAccount(account)
	repository := &Repository{Owner: "ownerrr", Repository: "repo1"}
	db.AddRepositoryToAccount(account, repository)

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

	account := &Account{Id: 2455252, AccessToken: "23mf23f22n3kl2n3nkl2n3lnl2n3ln3lnl"}
	db.CreateAccount(account)

	found := db.FindAccountById(2455252)
	if found == nil {
		t.Fatalf("Account wasn't found")
	}
	if found.Id != account.Id {
		t.Errorf("Account id is wrong")
	}
	if found.AccessToken != "23mf23f22n3kl2n3nkl2n3lnl2n3ln3lnl" {
		t.Errorf("Account AccessToken wasn't stored")
	}
}

func TestAddRepositoryToAccount(t *testing.T) {
	db := createCleanPostgresDatabase()

	account := &Account{Id: 595, AccessToken: "23mf23f22n3kl2n3nkl2n3lnl2n3ln3lnl"}
	db.CreateAccount(account)

	repository := &Repository{Owner: "eer", Repository: "somename", Public: true}
	db.AddRepositoryToAccount(account, repository)

	if repository.Id == 0 {
		t.Errorf("Repository id should have been updated")
	}

	foundAccount := db.FindAccountById(595)
	if len(foundAccount.Repositories) == 0 {
		t.Fatalf("Repository wasn't added to account")
	}
	if foundAccount.Repositories[0].Account.Id != account.Id {
		t.Errorf("Repository should have an account")
	}
	if foundAccount.Repositories[0].Owner != "eer" {
		t.Errorf("Owner was %v", foundAccount.Repositories[0].Owner)
	}
	if foundAccount.Repositories[0].Repository != "somename" {
		t.Errorf("Name was %v", foundAccount.Repositories[0].Repository)
	}
	if foundAccount.Repositories[0].Public != true {
		t.Errorf("Public was %v", foundAccount.Repositories[0].Public)
	}
}

func TestCreateAccountUpdatesAccessToken(t *testing.T) {
	db := createCleanPostgresDatabase()

	account1 := &Account{Id: 2455252, AccessToken: "ZZZZZZZZZZ"}
	account2 := &Account{Id: 2455252, AccessToken: "AAAAAAAAAA"}
	db.CreateAccount(account1)
	db.CreateAccount(account2)

	foundAccount := db.FindAccountById(2455252)

	if foundAccount.Id != account1.Id {
		t.Errorf("Expected account to not be created again")
	}

	if foundAccount.AccessToken != "AAAAAAAAAA" {
		t.Errorf("Expected access token to get updated")
	}
}

func TestLoginExists(t *testing.T) {
	db := createCleanPostgresDatabase()

	account := &Account{Id: 2455252, AccessToken: "5T"}
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
