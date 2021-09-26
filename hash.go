package main

import (
  "encoding/base64"
)

func encodeHash(hash []byte) string {
  data := base64.StdEncoding.EncodeToString(hash)

  return data
}

