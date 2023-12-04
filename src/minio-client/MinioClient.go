package minio_client

import (
	"bytes"
	"context"
	"crypto/aes"
	"file-upload-golang/src/config"
	"file-upload-golang/src/crypto"
	fileservice "file-upload-golang/src/file"
	"fmt"
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

	createAesBlock(appConfig, aesBlockById, objectName)
	encryptedText, err := aesBlockById.EncryptFile(file, objectName)
	if err != nil {
		log.Fatal(err)
	}
	uploadEncryptedText(minioClient, objectName, ctx, encryptedText, appConfig)
}

func UploadFilePart(minioClient *minio.Client, objectName string, file multipart.File, appConfig config.Config, aesBlockById crypto.AesBlockMap, shardsByFile fileservice.ShardsByFile) {
	fileByIds := fileservice.GetFileIdContents(file, appConfig)
	shardsByFile[objectName] = []string{}
	for i, fileById := range fileByIds {
		fileShardId := fmt.Sprintf("%s-%d-%s", objectName, i, fileById.FileId)
		uploadData(minioClient, fileShardId, fileById.FileContent, appConfig, aesBlockById)
		shardsByFile[objectName] = append(shardsByFile[objectName], fileShardId)
	}
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

func GetFilePart(minioClient *minio.Client, key string, appConfig config.Config, aesBlockById crypto.AesBlockMap, shardsByFile fileservice.ShardsByFile) []byte {
	var res []byte
	for _, fileShardId := range shardsByFile[key] {
		res = append(res, GetFile(minioClient, fileShardId, appConfig, aesBlockById)...)
	}
	return res
}

func uploadData(minioClient *minio.Client, objectName string, data []byte, appConfig config.Config, aesBlockById crypto.AesBlockMap) {
	ctx := context.Background()

	createAesBlock(appConfig, aesBlockById, objectName)
	encryptedText, err := aesBlockById.EncryptData(data, objectName)
	if err != nil {
		log.Fatal(err)
	}
	uploadEncryptedText(minioClient, objectName, ctx, encryptedText, appConfig)
}

func createAesBlock(appConfig config.Config, aesBlockById crypto.AesBlockMap, objectName string) {
	block, err := aes.NewCipher([]byte(appConfig.AesKey))
	if err != nil {
		log.Fatalln(err)
	}
	aesBlockById[objectName] = block
}

func uploadEncryptedText(minioClient *minio.Client, objectName string, ctx context.Context, encryptedText []byte, appConfig config.Config) {
	_, err := minioClient.PutObject(ctx, appConfig.Minio.BucketName, objectName, bytes.NewReader(encryptedText), -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})

	if err != nil {
		log.Fatalln(err)
	}
}
