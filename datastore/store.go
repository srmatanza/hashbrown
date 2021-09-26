package datastore

import (
  "log"
  "time"
  "sync"
  "sync/atomic"
  "crypto/sha512"
)

type HashEntry struct {
  Id uint64
  HashLen uint16
  Hash []byte
}

type dbRequest struct {
  storing bool
  id uint64
  hash []byte
  tdelta time.Duration
  resp chan *HashEntry
}

type hashRequest struct {
  id uint64
  payload string
}

func computeHash(payload string) ([]byte, time.Duration) {
  start := time.Now()
  sum := sha512.Sum512([]byte(payload))
  delta := start.Sub(start)
  ret := sum[:]

  return ret, delta
}

var inmemHashStore = make(map[uint64]HashEntry)

var hashCount uint64 = 0
var availableHashCount uint64 = 0
var currentAvg time.Duration = 0

var mAvailableHashLock = &sync.Mutex{}
var mMapLock = &sync.Mutex{}

// Because we only synchronize writing to availableHashCount and currentAvg, there is a small
// possibility that data could be read while we're in this critical section. However, because
// the overall effect on the output would be minimal, it's probably not worth the performance
// overhead of synchronizing reads to these values in GetStats
func updateAvailableHashes(dt time.Duration) {
  time.Sleep(5*time.Second)
  ahc := atomic.AddUint64(&availableHashCount, 1)
  mAvailableHashLock.Lock()
  currentAvg = (currentAvg*time.Duration(ahc-1) + dt) / time.Duration(ahc)
  mAvailableHashLock.Unlock()
}

// GetStats reads willy-nilly from availableHashCount and currentAvg which are synchronized
// when they are being written to.
func GetStats() (uint64, time.Duration) {
  return availableHashCount, currentAvg
}

// hashWorker is poolable. Create as many of these as we need to compute our hashes quickly in parallel
func hashWorker(jobs <-chan hashRequest, result chan<- dbRequest, done chan<- bool) {
  for req := range jobs {
    resp := make(chan *HashEntry)
    hash, tdelta := computeHash(req.payload)
    result <- dbRequest{storing: true, id: req.id, hash: hash, tdelta: tdelta, resp: resp}
    <-resp
  }
  done<-true
}

// hashHandler will run from a single goroutine and handle updates to inmemHashStore and hashCount
func hashHandler(jobs <-chan dbRequest, done chan<- bool) {
  for req := range jobs {
    if req.storing {
      req.resp <- nil
      inmemHashStore[req.id] = HashEntry{hashCount, uint16(len(req.hash)), req.hash}
      go updateAvailableHashes(req.tdelta)
    } else {
      if req.id <= availableHashCount {
        hash, ok := inmemHashStore[req.id]
        if ok {
          req.resp <- &hash
          continue
        }
      }
      req.resp <- nil
    }
  }
  done<-true
  log.Print("HashHandler is now done")
}

var alreadyInitialized = false

var hashQueue chan hashRequest
var hashQueueDone chan bool

var dbQueue chan dbRequest
var dbQueueDone chan bool

var poolsize int

func Initialize() bool {
  if alreadyInitialized {
    log.Print("Warning; datastore.Initialize: The datastore has already been initialized")
    return false
  }
  alreadyInitialized = true

  dbQueue = make(chan dbRequest, 10)
  dbQueueDone = make(chan bool)

  go hashHandler(dbQueue, dbQueueDone)

  poolsize = 10
  hashQueue = make(chan hashRequest, 40)

  for i:=0; i<poolsize; i++ {
    go hashWorker(hashQueue, dbQueue, hashQueueDone)
  }

  return true
}

func Shutdown() bool {
  if alreadyInitialized {
    log.Print("Closing queues for datastore")
    close(dbQueue)
    <-dbQueueDone
    close(dbQueueDone)

    for i:=0; i<poolsize; i++ {
      <-hashQueueDone
    }
    alreadyInitialized = false
    return true
  }
  log.Print("Warning; datastore.Shutdown: The datastore has not been initialized")
  return false
}

// PutHash will generate a hash for the given payload in the background
// and store it in the database. This function will immediately return with the id
// that will be assigned to the hash.
func PutHash(payload string) uint64 {
  hashId := atomic.AddUint64(&hashCount, 1)
  go func() { hashQueue <- hashRequest{hashId, payload} }()
  return hashId
}

func GetHash(id uint64) (HashEntry, bool) {
  resp := make(chan *HashEntry)
  dbQueue <- dbRequest{storing: false, id: id, resp: resp}
  if hash := <-resp; hash != nil {
    return *hash, true
  }
  return HashEntry{}, false
}
