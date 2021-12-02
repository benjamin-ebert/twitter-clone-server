package main

import "wtfTwitter/http"

func main() {
	//http.Run()
	server := http.NewServer()
	server.Run(server)
}