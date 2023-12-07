package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"log"
	"mime/multipart"
)

func EncryptFile(file multipart.File, aesBlock cipher.Block) ([]byte, error) {

	byteAnswer, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	return EncryptData(byteAnswer, aesBlock)
}
func EncryptData(data []byte, aesBlock cipher.Block) ([]byte, error) {
	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(aesBlock, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], data)

	return ciphertext, nil
}

func Decrypt(cipherText []byte, aesBlock cipher.Block) ([]byte, error) {
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(aesBlock, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}
