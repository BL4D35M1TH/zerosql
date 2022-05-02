package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sanndy/server"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Directory string `json:"directory"`
	Database  string `json:"database"`
	Address   string `json:"address"`
}

func main() {
	var config Config
	file, err := os.ReadFile("./config.json")
	if err != nil {
		panic(fmt.Errorf("unable to read file: %w", err))
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic(fmt.Errorf("unable to parse config: %w", err))
	}
	imageServer, err := server.CreateServer(config.Directory, config.Database)
	if err != nil {
		panic(err)
	}
	panic(http.ListenAndServe(config.Address, imageServer))
}
