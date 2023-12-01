package main

import (
	"encoding/json"
	"file-upload-golang/src/config"
	"file-upload-golang/src/crypto"
	minioclient "file-upload-golang/src/minio-client"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
)

var appConfig = config.Config{}
var aesBlockById = crypto.AesBlockMap{}

func main() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&appConfig)
	if err != nil {
		log.Fatal(err)
	}

	minioClient := minioclient.GetMinioClient(appConfig)
	minioclient.CreateBucket(minioClient, appConfig)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/file", func(r chi.Router) {

		r.Post("/", func(writer http.ResponseWriter, request *http.Request) {
			// Get the file from the request
			file, handler, err := request.FormFile("file")
			if err != nil {
				http.Error(writer, "Unable to get file from form", http.StatusBadRequest)
				return
			}
			defer file.Close()

			// Respond to the client
			minioClient := minioclient.GetMinioClient(appConfig)
			minioclient.UploadFile(minioClient, handler.Filename, file, appConfig, aesBlockById)
			fmt.Fprintf(writer, "File %s uploaded successfully!\n", handler.Filename)
		})

		r.Get("/{fileId}", func(writer http.ResponseWriter, request *http.Request) {
			minioClient := minioclient.GetMinioClient(appConfig)
			byteAnswer := minioclient.GetFile(minioClient, chi.URLParam(request, "fileId"), appConfig, aesBlockById)
			_, err = writer.Write(byteAnswer)
			if err != nil {
				return
			}
		})

	})

	http.ListenAndServe(":3000", r)
}
