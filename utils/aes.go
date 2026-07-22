package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
)

// Key (customizable, must be 16, 24, or 32 bytes in length)
var key = []byte("pay451437841#xXy")

// EncryptAES Use AES to encrypt the plaintext, return the encrypted Base64 encoded string
func EncryptAES(plaintext string) (string, error) {
	// Create a cipher.Block instance
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Convert the plaintext to a byte array
	plaintextBytes := []byte(plaintext)

	// Pad the data
	paddedPlaintext := pad(plaintextBytes, block.BlockSize())

	// Initialize the IV
	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// Create a CBC mode encryptor
	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	// Append the IV to the beginning of the ciphertext
	ciphertext = append(iv, ciphertext...)

	// Return the encrypted Base64 encoded ciphertext
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES Decrypt the ciphertext, return the plaintext string
func DecryptAES(ciphertextBase64 string) (string, error) {
	// Decode the Base64 encoded ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}

	// Create a cipher.Block instance
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Extract the IV (first 16 bytes)
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Create a CBC decryptor
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove the padding characters
	plaintext = pkcs7Unpad(plaintext, block.BlockSize())

	// Return the decrypted plaintext string
	return string(plaintext), nil
}

// pad Use PKCS#7 to pad the data, ensure the length is a multiple of the block size
func pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padded := make([]byte, len(data)+padding)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}
	return padded
}

// pkcs7Unpad Remove PKCS#7 padding
func pkcs7Unpad(data []byte, blockSize int) []byte {
	padding := data[len(data)-1]
	if padding < 1 || padding > byte(blockSize) {
		return nil
	}
	for i := 0; i < int(padding); i++ {
		if data[len(data)-i-1] != padding {
			return nil
		}
	}
	return data[:len(data)-int(padding)]
}

func main() {
	// The plaintext to be encrypted
	plaintext := "Hello, AES encryption in Go!"

	// Encrypt
	encrypted, err := EncryptAES(plaintext)
	if err != nil {
		log.Fatal("Encryption failed:", err)
	}
	fmt.Println("Encrypted ciphertext (Base64 encoded):", encrypted)

	// Decrypt
	decrypted, err := DecryptAES(encrypted)
	if err != nil {
		log.Fatal("Decryption failed:", err)
	}
	fmt.Println("Decrypted plaintext:", decrypted)
}
