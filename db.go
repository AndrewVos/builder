package main

import (
	"github.com/eaigner/jet"
	_ "github.com/lib/pq"
)

var connection *jet.Db

func connect() (*jet.Db, error) {
	if connection == nil {
		connection, err := jet.Open("postgres", "user=builder password="+configuration.PostgresPassword+" dbname=builder sslmode=disable")
		return connection, err
	}
	err := connection.Ping()
	if err != nil {
		return nil, err
	}
	return connection, nil
}
