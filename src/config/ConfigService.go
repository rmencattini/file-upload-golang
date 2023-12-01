package config

import (
	"github.com/inhies/go-bytesize"
	"log"
)

type Config struct {
	Split struct {
		Activate  bool   `json:"activate"`
		SliceSize string `json:"slice-size" default:"1MB"`
	} `json:"split"`
	Minio struct {
		Endpoint        string `json:"host"`
		AccessKeyID     string `json:"id"`
		SecretAccessKey string `json:"password"`
		BucketName      string `json:"bucket-name"`
	} `json:"minio"`
}

func (config Config) GetSliceSize() bytesize.ByteSize {
	if config.Split.Activate {
		b, err := bytesize.Parse(config.Split.SliceSize)
		if err != nil {
			log.Println(err)
			log.Println("Default slice-value to 1MB")
			return bytesize.MB
		}
		return b
	}

	return bytesize.New(0)
}
