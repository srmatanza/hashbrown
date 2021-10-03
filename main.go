package main

import (
  "flag"
  "log"
  "hashbrown/datastore"
)

func main() {
  addr := flag.String("addr", ":8123", "hashbrown service address")
  flag.Parse()

  s := NewServer(*addr)
  s.db = datastore.NewHashStore()

  go s.Serve()
  s.WaitForQuit()

  log.Printf("Closing the server gracefully")
}
