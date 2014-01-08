package main

type Database interface {
	SaveRepository(repository *Repository) error
	SaveCommit(commit *Commit) error
	SaveBuild(build *Build) error
	AllBuilds() []*Build
	CreateBuild(repository *Repository, build *Build) error
	FindRepository(owner string, name string) *Repository
	IncompleteBuilds() []*Build
	FindAccountByGithubUserId(id int) *Account
	FindAccountById(id int) *Account
	CreateAccount(account *Account) error
	CreateLoginForAccount(account *Account) (*Login, error)
	LoginExists(accountId int, token string) bool
}
