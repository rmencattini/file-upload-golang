package minio_client

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"mime/multipart"
)

func GetMinioClient() *minio.Client {
	endpoint := "localhost:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"

	// Initialize minio minioClient object.
	client, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		log.Fatalln(err)
	}
	return client
}

func CreateBucket(minioClient *minio.Client) {
	ctx := context.Background()
	// Make a new bucket called dev-minio.
	bucketName := "titi"
	location := "us-east-1"

	err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}
}

func UploadFile(minioClient *minio.Client, objectName string, file multipart.File) {
	ctx := context.Background()

	bucketName := "titi"

	info, err := minioClient.PutObject(ctx, bucketName, objectName, file, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})

	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Successfully uploaded %s of size %d with key: %s\n", objectName, info.Size, info.Key)
}

func GetFile(minioClient *minio.Client, key string) *minio.Object {
	ctx := context.Background()

	bucketName := "titi"

	file, err := minioClient.GetObject(ctx, bucketName, key, minio.GetObjectOptions{})

	if err != nil {
		log.Fatalln(err)
	}

	return file
}
