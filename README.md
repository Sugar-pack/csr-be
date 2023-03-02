## What is this repo?
This CSR backend API.
- `GET /api/docs` - swagger generated UI

## How to run this code
Checked from unix-compatible OS.
Go 1.19.2 is the current version.

1. Install all necessary tools
   ```bash
   make setup
   ```
2. Generate necessary go files (related to ent, swagger)
    ```shell
    make generate
    ```
3. Run database
    ```shell
    make db
    ```
4. Run the service: 
    ```shell
    make run
    ```
   The server is here - http://127.0.0.1:8080/api
   Swagger docs are here - http://127.0.0.1:8080/api/docs
5. Service cURL request example:
   ```shell
   curl -XPOST http://127.0.0.1:8080/api/v1/users/ -vvv
   *   Trying 127.0.0.1:8080...
   * Connected to 127.0.0.1 (127.0.0.1) port 8080 (#0)
   > POST /api/v1/users/ HTTP/1.1
   > Host: 127.0.0.1:8080
   > User-Agent: curl/7.77.0
   > Accept: */*
   > 
   * Mark bundle as not supporting multiuse
   < HTTP/1.1 201 Created
   < Content-Type: application/json
   < Date: Tue, 15 Feb 2022 10:16:36 GMT
   < Content-Length: 20
   < Connection: close
   < 
   {"data":{"id":"1"}}
   ```

### For developers

To draw entities relationships diagram:
```
go get -u github.com/a8m/enter
enter ./ent/schema
```

### Files workflow

Files are stored in the file system. 
The name of the folder with files is set in environment variable PHOTOS_FOLDER. 
The database stores id - names of files without an extension

<img src="images/equipments_photos.png" alt="files workflow diagrams">
