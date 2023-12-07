package redis

import (
	"encoding/json"
	"fmt"
)

type Redis struct {
	Split      bool
	Ids        []string
	AesKey     string
	BucketName string
}

func (r *Redis) MarshalBinary() ([]byte, error) {
	// Convert the Redis struct to a JSON string
	jsonString, err := json.Marshal(*r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Redis object to JSON: %w", err)
	}

	return []byte(jsonString), nil
}

func (r *Redis) UnmarshalBinary(data []byte) error {
	// Decode the binary data to a JSON string
	jsonString := string(data)

	// Convert the JSON string to a Redis struct
	err := json.Unmarshal([]byte(jsonString), &r)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON to Redis object: %w", err)
	}

	return nil
}
