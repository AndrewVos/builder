package main

type Database interface {
	AddRepositoryToAccount(account *Account, repository *Repository) error
	SaveCommit(commit *Commit) error
	SaveBuild(build *Build) error
	AllBuilds(account *Account) []*Build
	FindPublicBuilds() []*Build
	CreateBuild(repository *Repository, build *Build) error
	FindRepository(owner string, name string) *Repository
	IncompleteBuilds() []*Build
	FindAccountById(id int) *Account
	CreateAccount(account *Account) error
	CreateLoginForAccount(account *Account) (*Login, error)
	LoginExists(accountId int, token string) bool
}
