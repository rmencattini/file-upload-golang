package minio_client

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"file-upload-golang/src/config"
	"file-upload-golang/src/crypto"
	fileservice "file-upload-golang/src/file"
	redisclient "file-upload-golang/src/redis-client"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"mime/multipart"
)

func GetMinioClient(appConfig config.Config) *minio.Client {
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

func UploadFile(minioClient *minio.Client, objectName string, file multipart.File, appConfig config.Config, client *redis.Client) {
	ctx := context.Background()

	aesBlock := createAesBlock(appConfig.AesKey)
	encryptedText, err := crypto.EncryptFile(file, aesBlock)
	if err != nil {
		log.Fatal(err)
	}
	uploadEncryptedText(minioClient, objectName, ctx, encryptedText, appConfig)
	redisObject := redisclient.Redis{
		Split:      false,
		Ids:        []string{objectName},
		AesKey:     appConfig.AesKey,
		BucketName: appConfig.Minio.BucketName,
	}
	byteRedisObject, err := redisObject.MarshalBinary()
	if err != nil {
		log.Println(err)
	}
	statusCmd := client.Set(ctx, objectName, byteRedisObject, 0)
	log.Println(statusCmd)
}

func UploadFilePart(minioClient *minio.Client, objectName string, file multipart.File, appConfig config.Config, client *redis.Client) {
	ctx := context.Background()
	fileByIds := fileservice.GetFileIdContents(file, appConfig)
	redisObject := redisclient.Redis{Split: true, Ids: make([]string, 0), AesKey: appConfig.AesKey, BucketName: appConfig.Minio.BucketName}

	for i, fileById := range fileByIds {
		fileShardId := fmt.Sprintf("%s-%d-%s", objectName, i, fileById.FileId)
		uploadData(minioClient, fileShardId, fileById.FileContent, appConfig)
		redisObject.Ids = append(redisObject.Ids, fileShardId)
	}
	byteRedisObject, err := redisObject.MarshalBinary()
	if err != nil {
		log.Println(err)
	}
	statusCmd := client.Set(ctx, objectName, byteRedisObject, 0)
	log.Println(statusCmd)
}

func GetFile(minioClient *minio.Client, key string, client *redis.Client) []byte {
	redisString, err := client.Get(context.Background(), key).Result()
	if err != nil {
		log.Fatalln(err)
	}

	var redisObject redisclient.Redis
	err = json.Unmarshal([]byte(redisString), &redisObject)
	if err != nil {
		log.Fatal(err)
	}

	decryptedByteAnswer := getFile(minioClient, key, redisObject.AesKey, redisObject.BucketName)

	return decryptedByteAnswer
}

func GetFilePart(minioClient *minio.Client, key string, client *redis.Client) []byte {
	var res []byte
	redisString, err := client.Get(context.Background(), key).Result()
	if err != nil {
		log.Fatalln(err)
	}

	var redisObject redisclient.Redis
	err = json.Unmarshal([]byte(redisString), &redisObject)
	if err != nil {
		log.Fatal(err)
	}
	for _, fileShardId := range redisObject.Ids {
		res = append(res, getFile(minioClient, fileShardId, redisObject.AesKey, redisObject.BucketName)...)
	}
	return res
}

func getFile(minioClient *minio.Client, key string, aesKey string, bucketName string) []byte {
	ctx := context.Background()

	file, err := minioClient.GetObject(ctx, bucketName, key, minio.GetObjectOptions{})

	if err != nil {
		log.Fatalln(err)
	}

	byteAnswer, err := io.ReadAll(file)
	if err != nil {
		log.Fatalln(err)
	}

	aesBlock := createAesBlock(aesKey)
	decryptedByteAnswer, err := crypto.Decrypt(byteAnswer, aesBlock)
	if err != nil {
		log.Fatal(err)
	}

	return decryptedByteAnswer
}

func uploadData(minioClient *minio.Client, objectName string, data []byte, appConfig config.Config) {
	ctx := context.Background()

	aesBlock := createAesBlock(appConfig.AesKey)
	encryptedText, err := crypto.EncryptData(data, aesBlock)
	if err != nil {
		log.Fatal(err)
	}
	uploadEncryptedText(minioClient, objectName, ctx, encryptedText, appConfig)
}

func createAesBlock(aesKey string) cipher.Block {
	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		log.Fatalln(err)
	}
	return block
}

func uploadEncryptedText(minioClient *minio.Client, objectName string, ctx context.Context, encryptedText []byte, appConfig config.Config) {
	_, err := minioClient.PutObject(ctx, appConfig.Minio.BucketName, objectName, bytes.NewReader(encryptedText), -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})

	if err != nil {
		log.Fatalln(err)
	}
}
