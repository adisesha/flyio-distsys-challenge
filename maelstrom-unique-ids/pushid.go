package main

import (
	"math/rand"
	"sync"
	"time"
)

/**
 * See https://firebase.blog/posts/2015/02/the-2120-ways-to-ensure-unique_68
 * Fancy ID generator that creates 20-character string identifiers with the following properties:
 *
 * 1. They're based on timestamp so that they sort *after* any existing ids.
 * 2. They contain 72-bits of random data after the timestamp so that IDs won't collide with other clients' IDs.
 * 3. They sort *lexicographically* (so the timestamp is converted to characters that will sort properly).
 * 4. They're monotonically increasing.  Even if you generate more than one in the same timestamp, the
 *    latter ones will sort after the former ones.  We do this by using the previous random bits
 *    but "incrementing" them by 1 (only in the case of a timestamp collision).
 */

// Modeled after base64 web-safe chars, but ordered by ASCII.
const idChars = "-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"

func GenerateId() func() string {
	var lastPushTime int64 = 0
	lastRandChars := make([]int, 12)
	var mu sync.Mutex

	rand := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	return func() string {

		mu.Lock()
		var nowInMs = time.Now().UTC().UnixNano() / 1000000
		var duplicateTime = (nowInMs == lastPushTime)
		lastPushTime = nowInMs

		var timeStampChars []int = make([]int, 8)
		for i := 7; i >= 0; i-- {
			timeStampChars[i] = int(nowInMs % 64)
			nowInMs = nowInMs / 64
		}
		if nowInMs != 0 {
			panic("We should have converted the entire timestamp.")
		}

		var id []int = make([]int, 20)
		for i := 0; i < 8; i++ {
			id[i] = timeStampChars[i]
		}
		if !duplicateTime {
			for i := 0; i < 12; i++ {
				lastRandChars[i] = int(rand.Int31n(64))
			}
		} else {
			// If the timestamp hasn't changed since last push, use the same random number, except incremented by 1.
			i := 11
			for i >= 0 && lastRandChars[i] == 63 {
				lastRandChars[i] = 0
				i--
			}
			lastRandChars[i]++
		}
		for i := 0; i < 12; i++ {
			id[i+8] = lastRandChars[i]
		}
		mu.Unlock()
		var idStr string = ""
		for i := 0; i < len(id); i++ {
			idStr = idStr + string(idChars[id[i]])
		}
		return idStr
	}
}
