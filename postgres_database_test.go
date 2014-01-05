package main

import (
	"testing"
)

func TestAllBuildsLoadsCommits(t *testing.T) {
	db := createCleanPostgresDatabase()

	ghb := &GithubBuild{Owner: "ownerrr", Repository: "repo1"}
	db.SaveGithubBuild(ghb)

	commits := []Commit{
		Commit{Sha: "csdkl22323", Message: "hellooo", Url: "something.com"},
		Commit{Sha: "324mlkm", Message: "hi there", Url: "example.com"},
	}

	build := &Build{Owner: "ownerrr", Repository: "repo1", Commits: commits}
	db.CreateBuild(ghb, build)

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

func TestFindGithubBuild(t *testing.T) {
	db := createCleanPostgresDatabase()
	b1 := &GithubBuild{Owner: "ownerrr", Repository: "repo1"}
	b2 := &GithubBuild{Owner: "erm", Repository: "repo2"}
	db.SaveGithubBuild(b1)
	db.SaveGithubBuild(b2)

	ghb := db.FindGithubBuild("erm", "repo2")

	if ghb == nil || ghb.Owner != "erm" || ghb.Repository != "repo2" {
		t.Errorf("Expected to find github build:\n%+v\nActual:\n%+v\n", b2, ghb)
	}

	ghb = db.FindGithubBuild("losdsds", "sd")
	if ghb != nil {
		t.Errorf("Expected not to find a github build, but found:\n%+v\n", ghb)
	}
}

func createCleanPostgresDatabase() *PostgresDatabase {
	cleanDatabase()
	return &PostgresDatabase{}
}

func TestIncompleteBuilds(t *testing.T) {
	db := createCleanPostgresDatabase()
	ghb := &GithubBuild{Owner: "ownerrr", Repository: "repo1"}
	db.SaveGithubBuild(ghb)

	build := &Build{Owner: "ownerrr", Repository: "repo1"}
	db.CreateBuild(ghb, build)
	build.Complete = true
	db.SaveBuild(build)

	build = &Build{Owner: "ownerrr", Repository: "repo1"}
	db.CreateBuild(ghb, build)
	build.Complete = false
	db.SaveBuild(build)

	builds := db.IncompleteBuilds()
	if len(builds) != 1 {
		t.Errorf("We should only return one build, because only one is incomplete\n%+v\n", builds)
	}
}
