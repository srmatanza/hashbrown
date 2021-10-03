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

type HashStore struct {
  inmemHashStore map[uint64]HashEntry
  hashCount uint64
  availableHashCount uint64
  currentAvg time.Duration

  mAvailableHashLock sync.Mutex
  mMapLock sync.Mutex

  alreadyInitialized bool
  hashQueue chan hashRequest
  hashQueueDone chan bool

  dbQueue chan dbRequest
  dbQueueDone chan bool

  poolsize int
}

func computeHash(payload string) ([]byte, time.Duration) {
  start := time.Now()
  sum := sha512.Sum512([]byte(payload))
  delta := start.Sub(start)
  ret := sum[:]

  return ret, delta
}

// Because we only synchronize writing to availableHashCount and currentAvg, there is a small
// possibility that data could be read while we're in this critical section. However, because
// the overall effect on the output would be minimal, it's probably not worth the performance
// overhead of synchronizing reads to these values in GetStats
func (hs *HashStore) updateAvailableHashes(dt time.Duration) {
  time.Sleep(5*time.Second)
  ahc := atomic.AddUint64(&hs.availableHashCount, 1)
  hs.mAvailableHashLock.Lock()
  hs.currentAvg = (hs.currentAvg*time.Duration(ahc-1) + dt) / time.Duration(ahc)
  hs.mAvailableHashLock.Unlock()
}

// GetStats reads willy-nilly from availableHashCount and currentAvg which are synchronized
// when they are being written to.
func (hs *HashStore) GetStats() (uint64, time.Duration) {
  return hs.availableHashCount, hs.currentAvg
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
  log.Print("hashWorker is now done")
}

// hashHandler will run from a single goroutine and handle updates to inmemHashStore and hashCount
func (hs *HashStore) hashHandler(jobs <-chan dbRequest, done chan<- bool) {
  for req := range jobs {
    if req.storing {
      req.resp <- nil
      hs.inmemHashStore[req.id] = HashEntry{hs.hashCount, uint16(len(req.hash)), req.hash}
      go hs.updateAvailableHashes(req.tdelta)
    } else {
      if req.id <= hs.availableHashCount {
        hash, ok := hs.inmemHashStore[req.id]
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



func NewHashStore() *HashStore {
  hs := &HashStore{}
  hs.inmemHashStore = make(map[uint64]HashEntry)

  if hs.alreadyInitialized {
    log.Print("Warning; datastore.Initialize: The datastore has already been initialized")
    return nil
  }
  hs.alreadyInitialized = true

  hs.dbQueue = make(chan dbRequest, 10)
  hs.dbQueueDone = make(chan bool)

  go hs.hashHandler(hs.dbQueue, hs.dbQueueDone)

  hs.poolsize = 10
  hs.hashQueue = make(chan hashRequest, 40)
  hs.hashQueueDone = make(chan bool)

  for i:=0; i<hs.poolsize; i++ {
    go hashWorker(hs.hashQueue, hs.dbQueue, hs.hashQueueDone)
  }

  return hs
}

func (hs *HashStore) Shutdown() bool {
  if hs.alreadyInitialized {
    log.Print("Closing queues for datastore")
    close(hs.dbQueue)
    <-hs.dbQueueDone
    close(hs.dbQueueDone)

    log.Print("Closing hashQueue")
    close(hs.hashQueue)
    for i:=0; i<hs.poolsize; i++ {
      log.Print("Waiting for hashQueueDone...")
      <-hs.hashQueueDone
    }
    close(hs.hashQueueDone)
    hs.alreadyInitialized = false
    return true
  }
  log.Print("Warning; datastore.Shutdown: The datastore has not been initialized")
  return false
}

// PutHash will generate a hash for the given payload in the background
// and store it in the database. This function will immediately return with the id
// that will be assigned to the hash.
func (hs *HashStore) PutHash(payload string) uint64 {
  hashId := atomic.AddUint64(&hs.hashCount, 1)
  go func() { hs.hashQueue <- hashRequest{hashId, payload} }()
  return hashId
}

func (hs *HashStore) GetHash(id uint64) (HashEntry, bool) {
  resp := make(chan *HashEntry)
  hs.dbQueue <- dbRequest{storing: false, id: id, resp: resp}
  if hash := <-resp; hash != nil {
    return *hash, true
  }
  return HashEntry{}, false
}
