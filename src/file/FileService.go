package file

import (
	"crypto/sha256"
	"file-upload-golang/src/config"
	"io"
	"log"
	"mime/multipart"
)

type IdContent struct {
	FileId      [32]byte
	FileContent []byte
}

func GetFileIdContents(file multipart.File, appConfig config.Config) []IdContent {
	res := make([]IdContent, 0)
	for _, fileShard := range splitFiles(file, appConfig) {
		res = append(res, IdContent{
			FileId:      getFileId(fileShard),
			FileContent: fileShard,
		})
	}
	return res
}

func splitFiles(file multipart.File, appConfig config.Config) [][]byte {
	chunk := make([]byte, appConfig.GetSliceSize())
	res := make([][]byte, 0)
	for {
		_, err := file.Read(chunk)
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		res = append(res, chunk)
	}
	return res
}

func getFileId(fileContent []byte) [32]byte {
	return sha256.Sum256(fileContent)
}
