package file

import (
	"file-upload-golang/src/config"
	"github.com/google/uuid"
	"io"
	"log"
	"mime/multipart"
)

type IdContent struct {
	FileId      string
	FileContent []byte
}

func GetFileIdContents(file multipart.File, appConfig config.Config) []IdContent {
	res := make([]IdContent, 0)
	for _, fileShard := range splitFiles(file, appConfig) {
		res = append(res, IdContent{
			FileId:      getFileId(),
			FileContent: fileShard,
		})
	}
	return res
}

func splitFiles(file multipart.File, appConfig config.Config) [][]byte {
	chunk := make([]byte, appConfig.GetSliceSize())
	res := make([][]byte, 0)
	for {
		n, err := file.Read(chunk)
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		res = append(res, append([]byte(nil), chunk[:n]...))
	}
	return res
}

func getFileId() string {
	fileId, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}
	return fileId.String()
}
