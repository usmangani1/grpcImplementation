package util

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func HashPassword(password, salt string) (hashed string, saltHex string) {
	if salt == "" {
		s := make([]byte, 32)
		rand.Read(s)
		salt = hex.EncodeToString(s)
	}
	m5 := md5.New()
	m5.Write([]byte(password))
	m5.Write([]byte(salt))
	st := m5.Sum(nil)
	return hex.EncodeToString(st), salt
}

func RandString(length int) string {
	rand.Seed(time.Now().UnixNano())
	rs := make([]string, length)
	for start := 0; start < length; start++ {
		t := rand.Intn(3)
		if t == 0 {
			rs = append(rs, strconv.Itoa(rand.Intn(10)))
		} else if t == 1 {
			rs = append(rs, string(rand.Intn(26)+65))
		} else {
			rs = append(rs, string(rand.Intn(26)+97))
		}
	}
	return strings.Join(rs, "")
}
