// encryption.go
// Package main handles encryption functions, including AES encryption and OTP (XOR cipher) encryption.

package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// encryptAES encrypts the plaintext using AES encryption with the provided key.
func encryptAES(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Add padding to the plaintext to make its length a multiple of the block size
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	plaintext = append(plaintext, padtext...)

	// Prepare the ciphertext slice with space for the IV
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]

	// Generate a random initialization vector (IV)
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return nil, err
	}

	// Encrypt the plaintext using AES-CBC mode
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

// decryptAES decrypts the ciphertext using AES encryption with the provided key.
func decryptAES(key, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertextData := ciphertext[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Decrypt the ciphertext using AES-CBC mode
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertextData, ciphertextData)

	// Remove padding
	paddingLength := int(ciphertextData[len(ciphertextData)-1])
	plaintext := ciphertextData[:len(ciphertextData)-paddingLength]

	return plaintext, nil
}

// encryptXOR performs OTP encryption (XOR cipher) on the message using the provided key.
func encryptXOR(message, key []byte) []byte {
	ciphertext := make([]byte, len(message))
	for i := range message {
		ciphertext[i] = message[i] ^ key[i]
	}
	return ciphertext
}
