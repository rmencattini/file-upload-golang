package main

import (
	"encoding/json"
	"file-upload-golang/src/config"
	"file-upload-golang/src/crypto"
	"file-upload-golang/src/file"
	minioclient "file-upload-golang/src/minio-client"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

var appConfig = config.Config{}
var aesBlockById = crypto.AesBlockMap{}
var fileShardsByFile = file.ShardsByFile{}

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

	minioClient := minioclient.GetMinioClient(appConfig)
	minioclient.CreateBucket(minioClient, appConfig)

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

			// Respond to the client
			minioClient := minioclient.GetMinioClient(appConfig)
			if appConfig.Split.Activate {
				minioclient.UploadFilePart(minioClient, handler.Filename, uploadedFile, appConfig, aesBlockById, fileShardsByFile)
			} else {
				minioclient.UploadFile(minioClient, handler.Filename, uploadedFile, appConfig, aesBlockById)
			}
			_, err = fmt.Fprintf(writer, "File %s uploaded successfully!\n", handler.Filename)
			if err != nil {
				return
			}
		})

		r.Get("/{fileId}", func(writer http.ResponseWriter, request *http.Request) {
			minioClient := minioclient.GetMinioClient(appConfig)
			var byteAnswer []byte
			if appConfig.Split.Activate {
				byteAnswer = minioclient.GetFilePart(minioClient, chi.URLParam(request, "fileId"), appConfig, aesBlockById, fileShardsByFile)
			} else {
				byteAnswer = minioclient.GetFile(minioClient, chi.URLParam(request, "fileId"), appConfig, aesBlockById)
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
