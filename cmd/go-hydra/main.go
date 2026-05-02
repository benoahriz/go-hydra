package main

import (
	"log"
	"net/http"

	"github.com/fsouza/go-dockerclient"
)

var bufLen = 1024

func clientConn() *docker.Client {
	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)
	return client
}

func main() {
	router := NewRouter()
	fs := http.FileServer(http.Dir("test"))
	router.PathPrefix("/test/").Handler(http.StripPrefix("/test/", fs))

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
