# Requirement

The solution can:
1. Accept files of arbitrary size, encrypt their content, upload them to a Minio bucket.
2. Serve the submitted files through a GET API (accessible with curl)
3. Upload the file in a chuncks of configurable size (can be configured)
4. Configure multiple options such as:
   * upload chuncks feature enable/disable
   * chunks size
   * minio endpoint / user / password / bucketName
   * aes key to encrypt / decrypt
5. Persistence. When the service and stopped and relaunch, you can still get previous uploaded file.

# Run

Assuming minio is runing through the docker-compose:
```bash
docker-compose up
```
then
```bash
go run src/main.go
```

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
   * It has been uploaded in multiple shards with the proper size
   * If you download some files, they are encrypted

5. Retrieve the full decrypted file:
```bash 
curl localhost:3000/file/bash.txt --output new-bash.txt 
```
8. Compare them:
```bash
cmp -s bash.txt new-bash.txt && echo "Same" || echo "Different"
```

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

# Architecture / Design

I tried to structure my code in a DDD way, so I split into:
* domain
  * entities
  * services
* infrastructure
  * config
  * minio
  * redis

I made effort to find the right balance between external libraries and the standard one.

I keep the API as simple as possible, so I only have two endpoint:
* `POST /file` with the file as body.
* `GET /file/filename`

## Why Redis ?

It may be strange to choose Redis as persistence layer. I will explain this choice.

I did not need relational database, so a key/value database would fit perfectly (as I did not have complex object interacting each other).

I was curious about a few lines from their [documentation](https://redis.io/docs/management/persistence/):
>  RDB (Redis Database): RDB persistence performs point-in-time snapshots of your dataset at specified intervals.
> 
> AOF (Append Only File): AOF persistence logs every write operation received by the server. These operations can then be replayed again at server startup, reconstructing the original dataset. Commands are logged using the same format as the Redis protocol itself.
> 
> The general indication you should use both persistence methods is if you want a degree of data safety comparable to what PostgreSQL can provide you.

So I decided to give a try for this small project as I did not have strong requirement. Anyway Redis can easily be swaped for any other key/value database (e.g MongoDB)
## Possible improvement / self-criticism

* I do not have tests for some part of the code which require it (mostly `CryptoService.go`, `FileService.go` and `Config.go`)
* Cryptographic stuff is clunky
* I did not test file upload with crazy extreme value
* When upload file into multiple part, I need to generate multiple new name to upload them to Minio. I did not handle the name collision issue, so it can occur.
* I do not have strong knowledge on writing idiomatic Go code.
* The persistence layer should be updated accordingly if some object are removed from Minio (when trying to fetch if it does not work, remove the key from Redis)