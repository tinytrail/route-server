package main

import (
	"log/slog"
	"net/http"
	"strconv"
)

func main() {
	port := 8888

	slog.Info("Starting server listening on ", "port", port)
	http.HandleFunc("/", getRoot)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!!!"))
}
