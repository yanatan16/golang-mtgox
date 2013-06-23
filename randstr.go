package mtgox

import (
	"crypto/sha1"
	"fmt"
	"time"
)

var nonces <-chan string
var ids <-chan string

func init() {
	nonces = nonceCreator()
	ids = idCreator()
}

func nonceCreator() <-chan string {
	nonce := make(chan string, 0)
	go func() {
		for {
			nonce <- fmt.Sprintf("%d", time.Now().Unix())
		}
	}()
	return nonce
}

func idCreator() <-chan string {
	id := make(chan string, 0)
	go func() {
		hash := sha1.New()
		base := []byte(time.Now().String())
		for {
			hash.Write(base)
			id <- fmt.Sprintf("%x", hash.Sum(nil))
		}
	}()
	return id
}
