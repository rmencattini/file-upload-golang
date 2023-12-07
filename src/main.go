package main

import (
	"encoding/json"
	"file-upload-golang/src/infrastructure/config"
	minioservice "file-upload-golang/src/infrastructure/minio"
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
			minioClient := minioservice.GetMinioClient(appConfig)
			var byteAnswer []byte
			if appConfig.Split.Activate {
				byteAnswer = minioservice.GetFilePart(minioClient, chi.URLParam(request, "fileId"), redisClient)
			} else {
				byteAnswer = minioservice.GetFile(minioClient, chi.URLParam(request, "fileId"), redisClient)
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
