package main

import (
	"io"
	"io/ioutil"
	"os"
)

var fakeGit *FakeGit
var fakeDatabase *FakeDatabase

func init() {
	resetFakeGit()
	resetFakeDatabase()
}

func resetFakeDatabase() {
	fakeDatabase = &FakeDatabase{}
	database = fakeDatabase
}

func resetFakeGit() {
	fakeGit = &FakeGit{}
	git = fakeGit
}

type FakeGit struct {
	FakeRepo                  string
	UserIdToReturn            int
	AccessTokenToReturn       string
	createHooksParameters     map[string]interface{}
	IsRepositoryPrivateResult bool
	CollaboratorsToReturn     []Collaborator
}

func (g *FakeGit) Retrieve(log io.Writer, url string, path string, branch string, sha string) error {
	files, _ := ioutil.ReadDir("test-repos/" + g.FakeRepo)
	os.MkdirAll(path, 0700)
	for _, file := range files {
		b, _ := ioutil.ReadFile("test-repos/" + g.FakeRepo + "/" + file.Name())
		ioutil.WriteFile(path+"/"+file.Name(), b, 0700)
	}
	return nil
}

func (g *FakeGit) CreateHooks(accessToken string, owner string, repo string) error {
	g.createHooksParameters = map[string]interface{}{
		"accessToken": accessToken,
		"owner":       owner,
		"repository":  repo,
	}
	return nil
}

func (g *FakeGit) GetAccessToken(clientId string, clientSecret string, code string) (string, error) {
	return g.AccessTokenToReturn, nil
}

func (g *FakeGit) GetUserID(accessToken string) (int, error) {
	return g.UserIdToReturn, nil
}

func (g *FakeGit) IsRepositoryPrivate(owner string, name string) bool {
	return g.IsRepositoryPrivateResult
}

type FakeDatabase struct {
	SavedRepository         *Repository
	CreatedAccount          *Account
	LoginToReturn           *Login
	FindAccountByIdToReturn *Account
	AddedCollaborations     []map[string]int
}

func (g *FakeGit) RepositoryCollaborators(accessToken string, owner string, name string) []Collaborator {
	return g.CollaboratorsToReturn
}

func (f *FakeDatabase) AddRepositoryToAccount(account *Account, repository *Repository) error {
	f.SavedRepository = repository
	return nil
}

func (f *FakeDatabase) SaveCommit(commit *Commit) error {
	return nil
}

func (f *FakeDatabase) SaveBuild(build *Build) error {
	return nil
}

func (f *FakeDatabase) AllBuilds(account *Account) []*Build {
	return nil
}

func (f *FakeDatabase) FindPublicBuilds() []*Build {
	return nil
}

func (f *FakeDatabase) CreateBuild(githubBuild *Repository, build *Build) error {
	return nil
}

func (f *FakeDatabase) FindRepository(owner string, repository string) *Repository {
	if f.SavedRepository != nil {
		if f.SavedRepository.Owner == owner && f.SavedRepository.Repository == repository {
			return f.SavedRepository
		}
	}
	return nil
}

func (f *FakeDatabase) IncompleteBuilds() []*Build {
	return nil
}

func (f *FakeDatabase) FindAccountById(id int) *Account {
	return f.FindAccountByIdToReturn
}

func (f *FakeDatabase) CreateAccount(account *Account) error {
	f.CreatedAccount = account
	return nil
}

func (f *FakeDatabase) CreateLoginForAccount(account *Account) (*Login, error) {
	return f.LoginToReturn, nil
}

func (f *FakeDatabase) LoginExists(accountId int, token string) bool {
	return true
}

func (f *FakeDatabase) SaveCollaboration(accountId int, repositoryId int) error {
	f.AddedCollaborations = append(f.AddedCollaborations, map[string]int{
		"account_id":    accountId,
		"repository_id": repositoryId,
	})
	return nil
}
