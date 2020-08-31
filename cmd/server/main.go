package main

import (
	"github.com/kelseyhightower/envconfig"
)

type Configuration struct {
	Pgdb     string // SERVER_PGDB
	Port     string // SERVER_PORT
	Endpoint string // SERVER_ENDPOINT
}

func main() {
	var c Configuration
	err := envconfig.Process("server", &c)
	if err != nil {
		panic(err)
	}
	app := NewApp(c)
	err = app.Run()
	if err != nil {
		panic(err)
	}
}
