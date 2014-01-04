package main

import (
	"net/http"
	"strings"
)

var githubDomain string = "https://api.github.com"
var git GitTool

func main() {
	serve()
}

func init() {
	git = Git{}
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/hooks/push", pushHandler)
	http.HandleFunc("/hooks/pull_request", pullRequestHandler)
	http.HandleFunc("/builds", buildsHandler)
	http.HandleFunc("/build_output", buildOutputHandler)
	http.HandleFunc("/build_output_raw", buildOutputRawHandler)
	http.HandleFunc("/github_callback", githubCallbackHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/add_repository", addRepositoryHandler)

	serveFile("public/scripts/jquery-2.0.3.min.js")
	serveFile("public/scripts/build.js")
	serveFile("public/scripts/build_output.js")
	serveFile("public/styles/common.css")
	serveFile("public/styles/build.css")
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

func authenticated(r *http.Request) (*http.Cookie, bool) {
	cookie, err := r.Cookie("github_access_token")
	if err != nil {
		return nil, false
	}
	return cookie, true
}
