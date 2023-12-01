package minio_client

import (
	"bytes"
	"context"
	"crypto/aes"
	"file-upload-golang/src/config"
	"file-upload-golang/src/crypto"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
	"mime/multipart"
)

func GetMinioClient(appConfig config.Config) *minio.Client {

	// Initialize minio minioClient object.
	client, err := minio.New(appConfig.Minio.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(appConfig.Minio.AccessKeyID, appConfig.Minio.SecretAccessKey, ""),
	})
	if err != nil {
		log.Fatalln(err)
	}
	return client
}

func CreateBucket(minioClient *minio.Client, appConfig config.Config) {
	ctx := context.Background()
	// Make a new bucket called dev-minio.

	err := minioClient.MakeBucket(ctx, appConfig.Minio.BucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, appConfig.Minio.BucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", appConfig.Minio.BucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", appConfig.Minio.BucketName)
	}
}

func UploadFile(minioClient *minio.Client, objectName string, file multipart.File, appConfig config.Config, aesBlockById crypto.AesBlockMap) {
	ctx := context.Background()

	block, err := aes.NewCipher([]byte("not secured not secured!"))
	if err != nil {
		log.Fatalln(err)
	}
	aesBlockById[objectName] = block

	encryptedText, err := aesBlockById.Encrypt(file, objectName)
	if err != nil {
		log.Fatal(err)
	}

	info, err := minioClient.PutObject(ctx, appConfig.Minio.BucketName, objectName, bytes.NewReader(encryptedText), -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})

	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Successfully uploaded %s of size %d with key: %s\n", objectName, info.Size, info.Key)
}

func GetFile(minioClient *minio.Client, key string, appConfig config.Config, aesBlockById crypto.AesBlockMap) []byte {
	ctx := context.Background()

	file, err := minioClient.GetObject(ctx, appConfig.Minio.BucketName, key, minio.GetObjectOptions{})

	if err != nil {
		log.Fatalln(err)
	}

	byteAnswer, err := io.ReadAll(file)
	if err != nil {
		log.Fatalln(err)
	}

	decryptedByteAnswer, err := aesBlockById.Decrypt(byteAnswer, key)
	if err != nil {
		log.Fatal(err)
	}

	return decryptedByteAnswer
}
