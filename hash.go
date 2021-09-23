package main

import (
  "crypto/sha256"
  "encoding/base64"
)

func computeHash(payload string) []byte {
  sum := sha256.Sum256([]byte(payload))
  ret := sum[:]

  return ret
}

func encodeHash(hash []byte) string {
  data := base64.StdEncoding.EncodeToString(hash)

  return data
}

