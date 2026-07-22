package google

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"strings"
	"time"
)

type Auth struct{}

func (g *Auth) GenSecret() string {
	var (
		dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		bytes      = make([]byte, 16)
	)

	_, _ = rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}

	return strings.ToUpper(string(bytes))
}

func (g *Auth) GenQrCodeUrl(user string, secret string) string {
	return fmt.Sprintf("otpauth://totp/%s?secret=%s", user, secret)
}

// To account for time errors, VerifyCode checks the current time and the time 30 seconds before and after it.
func (g *Auth) VerifyCode(secret string, code int32) bool {
	if g.getCode(secret, int64(0)) == code {
		return true
	}
	if g.getCode(secret, int64(30)) == code {
		return true
	}
	if g.getCode(secret, int64(-30)) == code {
		return true
	}

	return false
}

func (g *Auth) getCode(secret string, offset int64) int32 {
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return 0
	}
	epochSeconds := time.Now().Unix() + offset

	return int32(g.oneTimePassword(key, g.toBytes(epochSeconds/30)))
}

func (g *Auth) toBytes(value int64) []byte {
	var result []byte
	mask := int64(0xFF)
	shifts := [8]uint16{56, 48, 40, 32, 24, 16, 8, 0}
	for _, shift := range shifts {
		result = append(result, byte((value>>shift)&mask))
	}
	return result
}

func (g *Auth) toUint32(bytes []byte) uint32 {
	return (uint32(bytes[0]) << 24) + (uint32(bytes[1]) << 16) +
		(uint32(bytes[2]) << 8) + uint32(bytes[3])
}

func (g *Auth) oneTimePassword(key []byte, value []byte) uint32 {
	// sign the value using HMAC-SHA1
	hmacSha1 := hmac.New(sha1.New, key)
	hmacSha1.Write(value)
	hash := hmacSha1.Sum(nil)

	// We're going to use a subset of the generated hash.
	// Using the last nibble (half-byte) to choose the index to start from.
	// This number is always appropriate as it's maximum decimal 15, the hash will
	// have the maximum index 19 (20 bytes of SHA1) and we need 4 bytes.
	offset := hash[len(hash)-1] & 0x0F

	// get a 32-bit (4-byte) chunk from the hash starting at offset
	hashParts := hash[offset : offset+4]

	// ignore the most significant bit as per RFC 4226
	hashParts[0] = hashParts[0] & 0x7F
	number := g.toUint32(hashParts)

	// size to 6 digits
	// one million is the first number with 7 digits so the remainder
	// of the division will always return < 7 digits
	pwd := number % 1000000
	return pwd
}
