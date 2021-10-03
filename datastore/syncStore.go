package datastore

import (
	"log"
	"time"
	"sync"
	"sync/atomic"
)

type SyncStore struct {
	inmemHashStore map[uint64]HashEntry
	hashCount uint64
	availableHashCount uint64
	currentAvg time.Duration

	rwHashLock sync.RWMutex
	rwAvailableHashLock sync.RWMutex

	alreadyInitialized bool
}

func NewSyncStore() *SyncStore {
	hs := &SyncStore{}
	hs.inmemHashStore = make(map[uint64]HashEntry)

	if hs.alreadyInitialized {
    log.Print("Warning; The datastore has already been initialized")
    return nil
  }
	hs.alreadyInitialized = true
	return hs
}

func (hs *SyncStore) Shutdown() bool {
	if hs.alreadyInitialized {
    log.Print("Closing queues for datastore")
		hs.rwHashLock.Lock()
		hs.rwHashLock.Unlock()
		hs.rwAvailableHashLock.Lock()
		hs.rwAvailableHashLock.Unlock()
		return true
	}
	log.Print("Warning; The datastore has not been initialized")
  return false
}

func (hs *SyncStore) GetStats() (uint64, time.Duration) {
  return hs.availableHashCount, hs.currentAvg
}

func (hs *SyncStore) updateAvailableHashes(dt time.Duration) {
  time.Sleep(5*time.Second)
	hs.rwAvailableHashLock.Lock()
  hs.availableHashCount++
	ahc := hs.availableHashCount
  hs.currentAvg = (hs.currentAvg*time.Duration(ahc-1) + dt) / time.Duration(ahc)
  hs.rwAvailableHashLock.Unlock()
}

// PutHash will generate a hash for the given payload in the background
// and store it in the database. This function will immediately return with the id
// that will be assigned to the hash.
func (hs *SyncStore) PutHash(payload string) uint64 {
  hashId := atomic.AddUint64(&hs.hashCount, 1)
  go func() {
		hash, tdelta := computeHash(payload)
		go hs.updateAvailableHashes(tdelta)
		hs.rwHashLock.Lock()
		hs.inmemHashStore[hashId] = HashEntry{hashId, uint16(len(hash)), hash}
		hs.rwHashLock.Unlock()
	}()
  return hashId
}

func (hs *SyncStore) GetHash(id uint64) (HashEntry, bool) {
	if id <= hs.availableHashCount {
		hs.rwHashLock.RLock()
		hash, ok := hs.inmemHashStore[id]
		hs.rwHashLock.RUnlock()
		if ok {
			return hash, true
		}
	}
  return HashEntry{}, false
}
