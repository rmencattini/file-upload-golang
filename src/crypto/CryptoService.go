package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"log"
	"mime/multipart"
)

type AesBlockMap map[string]cipher.Block

func (aesBlockMap AesBlockMap) Encrypt(file multipart.File, key string) ([]byte, error) {

	byteAnswer, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(byteAnswer))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(aesBlockMap[key], iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], byteAnswer)

	return ciphertext, nil
}

func (aesBlockMap AesBlockMap) Decrypt(cipherText []byte, key string) ([]byte, error) {
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(aesBlockMap[key], iv)
	stream.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}
