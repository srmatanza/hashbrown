package main

import (
  "flag"
  "log"
  "time"
  "net/http"
  "hashbrown/datastore"
)

type server struct {
  router *Router
  db Datastore
  quit chan bool
}

type Datastore interface {
  Shutdown() bool
  PutHash(payload string) uint64
  GetHash(id uint64) (datastore.HashEntry, bool)
  GetStats() (uint64, time.Duration)
}

func NewServer() *server {
  s := &server{}
  s.router = &Router{}
  s.routes()
  s.quit = make(chan bool)

  s.db = datastore.NewHashStore()
  return s
}

func (s *server) shutdown() {
  s.quit<-true
  s.db.Shutdown()
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  s.router.ServeHTTP(w, r)
}

func (s *server) routes() {
  s.router.Get("^/stats$", s.handleStatsGet())
  s.router.Get("^/hash/([1-9]+[0-9]*)$", s.handleHashGet())
  s.router.Post("^/hash$", s.handleHashPost())
  s.router.Post("^/shutdown$", s.handleShutdownPost())
}

func (s *server) waitForQuit() {
  <-s.quit
}

var addr = flag.String("addr", ":8123", "hashbrown service address")

func main() {
  flag.Parse()

  s := NewServer()
  go Serve(s)
  s.waitForQuit()

  log.Printf("Closing the server gracefully")
}

// Serve will setup our basic listener. Apparently TCP keep-alives are enabled by default
// when we use this convenience method, and we probably don't want that for a basic REST
// server. It's probably necessary to spend some time analyzing the requirements of this
// service to determine what our exposure to DDoS attacks is here.
//
// To-do, spend some time getting familiar with the underlying net library.
func Serve(s *server) {
  err := http.ListenAndServe(*addr, s)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

