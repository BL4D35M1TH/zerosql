package main

import (
	"net/http"
	"sanndy/server"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	imageServer, err := server.CreateServer("./root", "./test.db")
	if err != nil {
		panic(err)
	}
	panic(http.ListenAndServe(":8080", imageServer))
}
