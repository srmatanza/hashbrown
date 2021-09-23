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

func Serve() {
  err := http.ListenAndServe(*addr, nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

