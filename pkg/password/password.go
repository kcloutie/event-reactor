package password

import (
	"math/rand"
	"time"
)

// TODO...maybe look into this: https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
const (
	lowerLetterBytes = "abcdefghijklmnopqrstuvwxyz"
	upperLetterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialBytes     = "!@#$%^&*()_+-=[]{}\\|;':\",.<>/?`~"
	numBytes         = "0123456789"
)

func GeneratePassword(length int, useLowerLetters, useUpperLetters, useSpecial, useNum bool, specialCharOverride string) string {

	r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
	pool := ""
	if useLowerLetters {
		pool += lowerLetterBytes
	}

	if useUpperLetters {
		pool += upperLetterBytes
	}

	if useSpecial {
		if specialCharOverride != "" {
			pool += specialCharOverride
		} else {
			pool += specialBytes
		}
	}

	if useNum {
		pool += numBytes
	}
	r2 := rand.New(rand.NewSource(time.Now().UnixNano()))
	poolBytes := []byte(pool)
	for i := 0; i < 1000; i++ {
		firstIndex := r1.Intn(len(pool))
		secondIndex := r2.Intn(len(pool))
		if secondIndex == firstIndex {
			secondIndex = rand.Intn(len(pool))
		}
		if secondIndex == firstIndex {
			continue
		}
		firstChar := pool[firstIndex]
		secondChar := pool[secondIndex]
		poolBytes[firstIndex] = secondChar
		poolBytes[secondIndex] = firstChar
	}

	pool = string(poolBytes)
	b := make([]byte, length)
	for i := range b {
		b[i] = pool[rand.Intn(len(pool))]
	}
	return string(b)
}
