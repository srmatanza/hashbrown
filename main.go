package main

import (
  "flag"
  "log"
)

func main() {
  addr := flag.String("addr", ":8123", "hashbrown service address")
  flag.Parse()

  s := NewServer(*addr)
  go s.Serve()
  s.WaitForQuit()

  log.Printf("Closing the server gracefully")
}
