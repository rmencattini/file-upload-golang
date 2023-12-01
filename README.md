# Run

Assuming minio is runing through the docker-compose
`go run src/main.go`

# Configuration

Here is the configuration structure:
```json
{
  "split": {
    "activate": true,
    "slice-size": "2MB"
  },
  "minio": {
    "bucket-name": "custom-bucket",
    "host": "localhost:9000",
    "id": "minioadmin",
    "password": "minioadmin"
  }
}
```

:warning: The split feature is not yeat implemented.

# Possible improvement / self-criticism

* I do not have tests and some part of the code may need it (mostly `CryptoService.go`, `FileService.go` and `Config.go`)
* Cryptographic stuff is clunky
* I did not test file with crazy extreme value