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
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/hooks/push", pushHandler)
	http.HandleFunc("/hooks/pull_request", pullRequestHandler)
	http.HandleFunc("/builds", buildsHandler)
	http.HandleFunc("/build_output", buildOutputHandler)
	http.HandleFunc("/build_output_raw", buildOutputRawHandler)
	http.HandleFunc("/github_callback", githubLoginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/settings", settingsHandler)
	http.HandleFunc("/add_repository", addRepositoryHandler)

	serveFile("public/scripts/jquery-2.0.3.min.js")
	serveFile("public/scripts/home.js")
	serveFile("public/scripts/build_output.js")
	serveFile("public/styles/common.css")
	serveFile("public/styles/home.css")
	serveFile("public/styles/build_output.css")
	serveFile("public/styles/bootstrap.min.css")
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
