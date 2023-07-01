# file-convert-stream
A simple media streaming server in Go

Build and run the microservice:
```sh
go build -o file-convert-stream
./file-convert-stream
```
The microservice will start running on localhost:8080.

You can test it by sending a file upload request to http://localhost:8080/upload with a file parameter named "file".

Build the Docker image by running the following command in the project root directory:

```sh
docker build -t file-convert-stream .
```
This command builds the Docker image based on the Dockerfile and tags it as file-convert-stream.

Run the Docker container using the following command:
```sh
docker run -p 8080:8080 file-convert-stream
```
This command starts the Docker container and maps port 8080 of the container to port 8080 of the host machine.


Note: Make sure to adjust the ffmpegCmd commands based on your specific FFmpeg installation and requirements. You may need to include additional options or filters as needed.
