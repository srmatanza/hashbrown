package datastore

import (
  "log"
  "time"
)

type HashEntry struct {
  Id uint64
  HashLen uint16
  Hash []byte
}

type hashRequest struct {
  storing bool
  id uint64
  hash []byte
  tdelta time.Duration
  resp chan *HashEntry
}

var inmemHashStore = make(map[uint64]HashEntry)
var queue chan hashRequest
var statsQueue chan time.Duration

var hashCount uint64 = 0
var availableHashCount uint64 = 0
var currentAvg time.Duration = 0

// statsHandler will run from a single goroutine and handle updates to availableHashCount and currentAvg
func statsHandler(completed chan bool) {
  for dt := range statsQueue {
      availableHashCount++
      currentAvg = (currentAvg*time.Duration(availableHashCount-1) + dt) / time.Duration(availableHashCount)
      // log.Printf("statsHandler; updating the currentAvg %v, %v", currentAvg, dt)
  }
  log.Print("statsHandler is now done")
  completed <- true
}

// hashHandler will run from a single goroutine and handle updates to inmemHashStore and hashCount
func hashHandler(completed chan bool) {
  for req := range queue {
    if req.storing {
      hashCount++
      inmemHashStore[hashCount] = HashEntry{hashCount, uint16(len(req.hash)), req.hash}
      req := req
      go func() {
        req.resp <- &HashEntry{Id:hashCount}
        time.Sleep(5*time.Second)
        statsQueue <- req.tdelta
      }()
    } else {
      if req.id <= availableHashCount {
        if hash, ok := inmemHashStore[req.id]; ok == true {
          req := req
          go func() { req.resp <- &hash }()
          continue
        }
      }
      req := req
      go func() { req.resp <- nil }()
    }
  }
  log.Print("hashHandler is now done")
  completed <- true
}

var alreadyInitialized = false

var cCompleted chan bool
func Initialize() bool {
  if alreadyInitialized {
    log.Print("Warning; datastore.Initialize: The datastore has already been initialized")
    return false
  }
  alreadyInitialized = true
  cCompleted = make(chan bool)
  queue = make(chan hashRequest)
  go hashHandler(cCompleted)
  statsQueue = make(chan time.Duration)
  go statsHandler(cCompleted)
  return true
}

func Shutdown() bool {
  if alreadyInitialized {
    log.Print("Closing queues for datastore")
    close(queue)
    close(statsQueue)
    <-cCompleted
    <-cCompleted
    alreadyInitialized = false
    return true
  }
  log.Print("Warning; datastore.Shutdown: The datastore has not been initialized")
  return false
}

func PutHash(hash []byte, tdelta time.Duration) uint64 {
  resp := make(chan *HashEntry)
  queue <- hashRequest{storing: true, hash:hash, tdelta: tdelta, resp: resp}
  if entry := <-resp; entry != nil {
    // log.Printf("Storing hash entry %d", entry.Id)
    return entry.Id
  }

  return 0
}

func GetHash(id uint64) (HashEntry, bool) {
  resp := make(chan *HashEntry)
  queue <- hashRequest{storing: false, id: id, resp: resp}
  if hash := <-resp; hash != nil {
    return *hash, true
  }
  return HashEntry{}, false
}

func GetStats() (uint64, time.Duration) {
  return availableHashCount, currentAvg
}

