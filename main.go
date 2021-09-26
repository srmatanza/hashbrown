package main

import (
  "flag"
  "log"
  "net/http"
  "hashbrown/datastore"
)

var addr = flag.String("addr", ":8123", "hashbrown service address")

func main() {
  flag.Parse()

  quit := make(chan bool)

  http.Handle("/stats", http.HandlerFunc(statsHandler))
  http.Handle("/hash/", http.HandlerFunc(getHashHandler))
  http.Handle("/hash", http.HandlerFunc(postHashHandler))
  http.Handle("/shutdown", http.HandlerFunc(func (w http.ResponseWriter, req *http.Request) {
    shutdownHandler(w, req, quit)
  }))

  datastore.Initialize()
  go Serve()
  <-quit
  log.Printf("Closing the server gracefully")
  datastore.Shutdown()
}

// Serve will setup our basic listener. Apparently TCP keep-alives are enabled by default
// when we use this convenience method, and we probably don't want that for a basic REST
// server. It's probably necessary to spend some time analyzing the requirements of this
// service to determine what our exposure to DDoS attacks is here.
//
// To-do, spend some time getting familiar with the underlying net library.
func Serve() {
  err := http.ListenAndServe(*addr, nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

