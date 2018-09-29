package main

import (
	"./client"
	"./server"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var App struct {
	server.Server
	client.Client
}

func init() {
	file, e := ioutil.ReadFile("./config.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	json.Unmarshal(file, &App)
}

func main() {
	app := &App
	fmt.Println(app)
	if app.Server.Target != "" {
		app.Server.Run()
	} else if app.Client.Remote != "" {
		app.Client.Run()
	}
}
