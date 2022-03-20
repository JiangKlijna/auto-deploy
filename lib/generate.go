package lib

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// startTime system start time
var startTime = time.Now().UnixNano()

// seed Secret.seed
var seed = strconv.FormatInt(startTime, 36) +
	strconv.FormatUint(rand.New(rand.NewSource(startTime)).Uint64(), 36) +
	runtime.Version() + runtime.GOOS + runtime.GOARCH

// GenerateSecret Get secret
// secret = sha224(seed+clientIP+userAgent).reverse()
func GenerateSecret(clientIP, userAgent string) string {
	return ReverseString(HashCalculation(sha256.New224(), seed+clientIP+userAgent))
}

// GenerateToken Get token
// token = md5(secret+md5(username+secret+password)+secret)
func GenerateToken(username, password, secret string) string {
	return HashCalculation(md5.New(), secret+HashCalculation(md5.New(), username+secret+password)+secret)
}

// GenerateCookie Get cookie
// path = sha512(((secret+token)^n).reverse()).reverse()
func GenerateCookie(secret, token string) string {
	return ReverseString(HashCalculation(sha512.New(), strings.Repeat(secret+token, 1+int(startTime%10))))
}

// GenerateAll Get secret token cookie
func GenerateAll(username, password, clientIP, userAgent string) (string, string, string) {
	secret := GenerateSecret(clientIP, userAgent)
	token := GenerateToken(username, password, secret)
	cookie := GenerateCookie(secret, token)
	return secret, token, cookie
}
