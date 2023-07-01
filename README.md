# file-convert-stream
A simple media streaming server in Go

Build and run the microservice:
```sh
go build -o file-convert-stream
./file-convert-stream
```
The microservice will start running on localhost:8080.

You can test it by sending a file upload request to http://localhost:8080/upload with a file parameter named "file".
