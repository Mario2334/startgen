package main

import (
	"net/http"
)

func SetupRoutes() {
	http.HandleFunc("/auth", AuthHandler)
}