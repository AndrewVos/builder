package main

import (
	"net/http"
	"strconv"
	"strings"
)

var githubDomain string = "https://api.github.com"
var git GitTool

func main() {
	deleteIncompleteBuilds()
	serve()
}

func deleteIncompleteBuilds() {
	for _, build := range database.IncompleteBuilds() {
		if build.Complete == false {
			build.fail()
		}
	}
}

func init() {
	git = Git{}
}

func serve() {
	err := http.ListenAndServe(":"+configuration.Port, nil)
	if err != nil {
		panic(err)
	}
}

func serveFile(filename string) {
	pattern := strings.Replace(filename, "public", "", 1)
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	})
}

func currentAccount(r *http.Request) *Account {
	account_id, _ := r.Cookie("account_id")
	token, _ := r.Cookie("token")

	if account_id == nil || token == nil {
		return nil
	}

	id, _ := strconv.Atoi(account_id.Value)
	if database.LoginExists(id, token.Value) {
		account := database.FindAccountById(id)
		return account
	}
	return nil
}
