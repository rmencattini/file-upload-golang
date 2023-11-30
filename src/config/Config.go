package config

import (
	"github.com/inhies/go-bytesize"
	"log"
)

type Config struct {
	Split struct {
		Activate  bool   `json:"activate"`
		SliceSize string `json:"slice-size"`
	} `json:"split"`
}

func (config Config) GetSliceSize() bytesize.ByteSize {
	logger := log.Logger{}
	if config.Split.Activate {
		b, err := bytesize.Parse(config.Split.SliceSize)
		if err != nil {
			logger.Println(err)
			logger.Println("Default slice-value to 1MB")
			return bytesize.MB
		}
		return b
	}
	return bytesize.New(0)
}
