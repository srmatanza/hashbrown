package main

import (
	"log"
	"time"
	"net/http"
	"hashbrown/datastore"
)

type server struct {
  router *Router
  db Datastore
  quit chan bool
	addr string
}

type Datastore interface {
  Shutdown() bool
  PutHash(payload string) uint64
  GetHash(id uint64) (datastore.HashEntry, bool)
  GetStats() (uint64, time.Duration)
}

func NewServer(addr string) *server {
  s := &server{addr: addr}
  s.router = &Router{}
  s.routes()
  s.quit = make(chan bool, 1)
  return s
}

func (s *server) Shutdown() {
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

func (s *server) WaitForQuit() {
  <-s.quit
}


// Serve will setup our basic listener. Apparently TCP keep-alives are enabled by default
// when we use this convenience method, and we probably don't want that for a basic REST
// server. It's probably necessary to spend some time analyzing the requirements of this
// service to determine what our exposure to DDoS attacks is here.
//
// To-do, spend some time getting familiar with the underlying net library.
func (s *server) Serve() {
  err := http.ListenAndServe(s.addr, s)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}
