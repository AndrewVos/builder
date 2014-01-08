package main

type Account struct {
	Id           int
	AccessToken  string
	Repositories []*Repository
}
