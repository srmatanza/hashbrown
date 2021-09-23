package main

import (
  "crypto/sha512"
  "encoding/base64"
)

func computeHash(payload string) []byte {
  sum := sha512.Sum512([]byte(payload))
  ret := sum[:]

  return ret
}

func encodeHash(hash []byte) string {
  data := base64.StdEncoding.EncodeToString(hash)

  return data
}

