package main

import (
	"errors"
	"fmt"
	"github.com/kr/pty"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

type Build struct {
	Id            int
	GithubBuildId int
	Url           string
	Owner         string
	Repository    string
	Ref           string
	Sha           string
	Complete      bool
	Success       bool
	Result        string
	GithubUrl     string
	Commits       []Commit
}

type Commit struct {
	Id      int
	BuildId int
	Sha     string
	Message string
	Url     string
}

func init() {
	for _, build := range AllBuilds() {
		if build.Complete == false {
			build.fail()
		}
	}
}

func (build *Build) ReadCommits() error {
	var commits []Commit

	db, err := connect()
	if err != nil {
		return err
	}

	err = db.Query("SELECT * FROM commits WHERE build_id = $1", build.Id).Rows(&commits)
	if err != nil {
		fmt.Println("Error getting commits:", err)
		return err
	}

	build.Commits = commits
	return nil
}

func CreateBuild(owner string, repo string, ref string, sha string, githubURL string, commits []Commit) (*Build, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}

	githubBuild, exists := FindGithubBuild(owner, repo)
	if !exists {
		return nil, errors.New(fmt.Sprintf("Someone tried to do a build but we don't have an access token :/\n%v/%v", owner, repo))
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

	build.save()

	for _, commit := range commits {
		commit.BuildId = build.Id
		database.SaveCommit(&commit)
	}

	return build, nil
}

func (build *Build) save() {
	db, err := connect()
	if err != nil {
		fmt.Println(err)
		return
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
		fmt.Println("Error saving build:", err)
		return
	}
}

func (build *Build) start() {
	err := os.MkdirAll(build.Path(), 0700)
	if err != nil {
		build.fail()
		return
	}

	output, err := os.Create(build.LogPath())
	if err != nil {
		build.fail()
		return
	}
	defer output.Close()

	err = build.checkout(output)
	if err != nil {
		build.fail()
		return
	}

	err = build.execute(output)
	if err != nil {
		build.fail()
		return
	}
	build.pass()
}

func (build *Build) checkout(output *os.File) error {
	githubBuild, found := FindGithubBuild(build.Owner, build.Repository)
	if !found {
		return errors.New("Don't have access to build this project")
	}
	url := "https://" + githubBuild.AccessToken + "@github.com/" + build.Owner + "/" + build.Repository

	err := git.Retrieve(output, url, build.SourcePath(), build.Ref, build.Sha)
	if err != nil {
		fmt.Fprintln(output, err)
		return err
	}

	return nil
}

func (build *Build) environs() []string {
	return []string{
		"BUILDER_BUILD_RESULT=" + build.Result,
		"BUILDER_BUILD_URL=" + build.Url,
		"BUILDER_BUILD_ID=" + strconv.Itoa(build.Id),
		"BUILDER_BUILD_OWNER=" + build.Owner,
		"BUILDER_BUILD_REPO=" + build.Repository,
		"BUILDER_BUILD_REF=" + build.Ref,
		"BUILDER_BUILD_SHA=" + build.Sha,
	}
}

func (build *Build) execute(output *os.File) error {
	cmd := exec.Command("bash", "./Builderfile")
	cmd.Dir = build.SourcePath()
	cmd.Stdout = output
	cmd.Stderr = output

	customEnv := build.environs()

	for _, c := range os.Environ() {
		customEnv = append(customEnv, c)
	}
	cmd.Env = customEnv

	f, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	io.Copy(output, f)

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (build *Build) pass() {
	build.Complete = true
	build.Success = true
	build.Result = "pass"
	build.save()
	build.executeHooks()
}

func (build *Build) fail() {
	build.Complete = true
	build.Success = false
	build.Result = "fail"
	build.save()
	build.executeHooks()
}

func (build *Build) executeHooks() {
	hooks, _ := ioutil.ReadDir("data/hooks")
	for _, file := range hooks {
		cmd := exec.Command("bash", "../../../data/hooks/"+file.Name())
		cmd.Dir = build.Path()

		customEnv := build.environs()

		for _, c := range os.Environ() {
			customEnv = append(customEnv, c)
		}
		cmd.Env = customEnv
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(output))
	}
}

func (b *Build) Path() string {
	return "data/builds/" + strconv.Itoa(b.Id)
}

func (b *Build) LogPath() string {
	return b.Path() + "/output.log"
}

func (b *Build) SourcePath() string {
	return b.Path() + "/source"
}

func (build *Build) ReadOutput() string {
	b, err := ioutil.ReadFile(build.LogPath())
	if err != nil {
		fmt.Println(err)
	}
	return string(b)
}

func AllBuilds() []*Build {
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

	for _, build := range builds {
		err := build.ReadCommits()
		if err != nil {
			fmt.Println("Error retrieving commits for build: ", err)
		}
	}

	return builds
}
