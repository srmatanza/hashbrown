package main

import (
  "log"
  "encoding/json"
)

type hashEntry struct {
  id uint64
  hashLen uint16
  hash []byte
}

var inmemHashStore = make(map[uint64]hashEntry)
var hashCount uint64 = 0
var currentAvg uint64 = 0

func storeHash(hash []byte) uint64 {
  hashCount++
  he := hashEntry{hashCount, uint16(len(hash)), hash}
  inmemHashStore[he.id] = he
  log.Printf("Storing hash entry %d", he.id)

  return he.id
}

func getHash(id uint64) (hashEntry, bool) {
  if hash, ok := inmemHashStore[id]; ok == true {
    return hash, true
  }
  return hashEntry{}, false
}

func getStatsJSON() string {
  statMap := map[string]uint64 {"total": hashCount, "average": currentAvg}
  statJson, _ := json.Marshal(statMap)
  return string(statJson)
}

