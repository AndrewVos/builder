package main

import (
	"testing"
)

func TestAllBuildsLoadsCommits(t *testing.T) {
	db := &PostgresDatabase{}

	ghb := &GithubBuild{Owner: "ownerrr", Repository: "repo1"}
	db.SaveGithubBuild(ghb)

	commits := []Commit{
		Commit{Sha: "csdkl22323", Message: "hellooo", Url: "something.com"},
		Commit{Sha: "324mlkm", Message: "hi there", Url: "example.com"},
	}

	build, _ := db.CreateBuild("ownerrr", "repo1", "", "", "", commits)

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
