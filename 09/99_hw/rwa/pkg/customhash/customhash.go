package customhash

import (
	"encoding/base64"
	"math/rand"

	"golang.org/x/crypto/argon2"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func Hash(plain, salt string) string {
	key := argon2.Key([]byte(plain), []byte(salt), 1, 32*1024, 4, 32)
	hashed := base64.RawStdEncoding.EncodeToString(key)

	res := make([]byte, 0, len(salt)+len(hashed))
	res = append(res, salt...)
	res = append(res, hashed...)

	return string(res)
}

func Salt(n int) string {
	v := make([]rune, n)
	for i := range n {
		v[i] = letters[rand.Intn(len(letters))]
	}

	return string(v)
}
