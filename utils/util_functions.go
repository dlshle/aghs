package utils

import (
	"encoding/base64"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var randomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))

func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

func EncodeString(original string) string {
	return strings.ReplaceAll(original, "\"", "\\\"")
}

func GenerateID() string {
	return strconv.FormatInt(randomGenerator.Int63n(time.Now().Unix()), 16)
}

func Timeout(callback func(), duration time.Duration) {
	time.Sleep(duration)
	callback()
}
