# Run

Assuming minio is runing through the docker-compose
`go run src/main.go`

# Example

If you want to test the multipart upload

1. Create some file: 
```bash
man bash | col -b > bash.txt
```

2. Set the size you want in the `config.json`: `"slice-size": "2MB"`
3. Upload it via a curl:
```bash
 curl -F "file=@bash.txt" localhost:3000/file/
```
4. You can check a few things from the minio admin console:
 5. It has been upload in multiple shards with the proper size
 6. If you download some files, they are encrypted
5. Retrieve the full decrypted file:
```bash 
curl localhost:3000/file/bash.txt --output new-bash.txt 
```
8. Compare their hash:
```bash
cmp -s bash.txt new-bash.txt && echo "Same" || echo "Different"
```

They should be the same

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

# Possible improvement / self-criticism

* I do not have tests for some part of the code which require it (mostly `CryptoService.go`, `FileService.go` and `Config.go`)
* Cryptographic stuff is clunky
* I did not test file upload with crazy extreme value
* When upload file into multiple part, I need to generate multiple new name to upload them to Minio. I did not handle the name collision issue, so it can occur.
* Data are not persisted. The next step would be to create a SQl instance to store the relation between file and file shards.
  * Once this issue is solved, I will need to handle some case such as: upload (mode: split) and download (mode: not split) and vis versa.