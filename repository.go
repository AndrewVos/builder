package main

type Repository struct {
	Id         int
	AccountId  int
	Owner      string
	Repository string
	Account    *Account
	Public     bool
}
