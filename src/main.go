package main

import (
	"file-upload-golang/src/config"
	minio_client "file-upload-golang/src/minio-client"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"net/http"
)

func main() {
	//logger := log.Logger{}

	// TODO fix the config
	_ = config.Config{}

	minioClient := minio_client.GetMinioClient()
	minio_client.CreateBucket(minioClient)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/file", func(r chi.Router) {

		r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte("list all articles id -name"))
		})

		r.Post("/", func(writer http.ResponseWriter, request *http.Request) {
			// Get the file from the request
			file, handler, err := request.FormFile("file")
			if err != nil {
				http.Error(writer, "Unable to get file from form", http.StatusBadRequest)
				return
			}
			defer file.Close()

			// Respond to the client
			minioClient := minio_client.GetMinioClient()
			minio_client.UploadFile(minioClient, handler.Filename, file)
			fmt.Fprintf(writer, "File %s uploaded successfully!\n", handler.Filename)
		})

		r.Get("/{fileId}", func(writer http.ResponseWriter, request *http.Request) {
			minioClient := minio_client.GetMinioClient()
			file := minio_client.GetFile(minioClient, chi.URLParam(request, "fileId"))
			byteAnswer, err := io.ReadAll(file)
			if err != nil {
				log.Fatal(err)
			}
			writer.Write(byteAnswer)
		})

	})

	http.ListenAndServe(":3000", r)
}
