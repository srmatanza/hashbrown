package datastore

import (
  "log"
)

type HashEntry struct {
  Id uint64
  HashLen uint16
  Hash []byte
}

var inmemHashStore = make(map[uint64]HashEntry)
var hashCount uint64 = 0
var currentAvg uint64 = 0

func PutHash(hash []byte) uint64 {
  hashCount++
  he := HashEntry{hashCount, uint16(len(hash)), hash}
  inmemHashStore[he.Id] = he
  log.Printf("Storing hash entry %d", he.Id)

  return he.Id
}

func GetHash(id uint64) (HashEntry, bool) {
  if hash, ok := inmemHashStore[id]; ok == true {
    return hash, true
  }
  return HashEntry{}, false
}

func GetStats() (uint64, uint64) {
  return hashCount, currentAvg
}

