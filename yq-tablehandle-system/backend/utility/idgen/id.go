package idgen

import (
	"crypto/rand"
	"encoding/hex"
)

func Hex128() string {
	var b [16]byte
	if _, e := rand.Read(b[:]); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b[:])
}
