package main

import (
	"context"
	"encoding/json"
	"file-upload-golang/src/infrastructure/config"
	minioservice "file-upload-golang/src/infrastructure/minio"
	redisclient "file-upload-golang/src/infrastructure/redis"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

var appConfig = config.Config{}
var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
	DB:   0, // use default DB
})

func main() {
	logger := log.Logger{}
	configFile, err := os.Open("config.json")
	if err != nil {
		logger.Fatalln(err)
	}
	defer func(configFile *os.File) {
		err := configFile.Close()
		if err != nil {
			log.Println(err)
		}
	}(configFile)

	err = json.NewDecoder(configFile).Decode(&appConfig)
	if err != nil {
		log.Fatalln(err)
	}

	minioClient := minioservice.GetMinioClient(appConfig)
	minioservice.CreateBucket(minioClient, appConfig)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/file", func(r chi.Router) {

		r.Post("/", func(writer http.ResponseWriter, request *http.Request) {
			// Get the file from the request
			uploadedFile, handler, err := request.FormFile("file")
			if err != nil {
				http.Error(writer, "Unable to get file from form", http.StatusBadRequest)
				return
			}
			defer func(uploadedFile multipart.File) {
				err := uploadedFile.Close()
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
				}
			}(uploadedFile)

			// Respond to the redisClient
			minioClient := minioservice.GetMinioClient(appConfig)
			if appConfig.Split.Activate {
				minioservice.UploadFilePart(minioClient, handler.Filename, uploadedFile, appConfig, redisClient)
			} else {
				minioservice.UploadFile(minioClient, handler.Filename, uploadedFile, appConfig, redisClient)
			}
			_, err = fmt.Fprintf(writer, "File %s uploaded successfully!\n", handler.Filename)
			if err != nil {
				return
			}
		})

		r.Get("/{fileId}", func(writer http.ResponseWriter, request *http.Request) {
			objectKey := chi.URLParam(request, "fileId")
			minioClient := minioservice.GetMinioClient(appConfig)
			var byteAnswer []byte
			redisString, err := redisClient.Get(context.Background(), objectKey).Result()
			if err != nil {
				log.Println(err)
				return
			}

			var redisObject redisclient.Redis
			err = json.Unmarshal([]byte(redisString), &redisObject)
			if err != nil {
				log.Fatal(err)
			}
			if redisObject.Split {
				byteAnswer = minioservice.GetFilePart(minioClient, redisObject)
			} else {
				byteAnswer = minioservice.GetFile(minioClient, chi.URLParam(request, "fileId"), redisObject)
			}
			_, err = writer.Write(byteAnswer)
			if err != nil {
				return
			}
		})

	})

	err = http.ListenAndServe(":3000", r)
	if err != nil {
		return
	}
}
