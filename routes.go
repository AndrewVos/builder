package main

import (
	"github.com/drone/routes"
	"net/http"
	"os"
)

func init() {
	mux := routes.New()
	mux.Get("/", homeHandler)
	mux.Get("/builds", buildsHandler)
	mux.Get("/build/:id/output", buildOutputHandler)
	mux.Get("/build/:id/output/raw", buildOutputRawHandler)
	mux.Get("/github_callback", githubLoginHandler)
	mux.Get("/logout", logoutHandler)
	mux.Get("/settings", settingsHandler)

	mux.Post("/hooks/push", pushHandler)
	mux.Post("/hooks/pull_request", pullRequestHandler)
	mux.Post("/repository", addRepositoryHandler)

	pwd, _ := os.Getwd()
	mux.Static("/assets", pwd)

	http.Handle("/", mux)
}
