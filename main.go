package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
)

func main() {
	app := iris.New()
	app.Use(logger.New())
	app.Use(recover.New())

	// Serve static files from the "public" directory
	app.HandleDir("/", "./public")

	app.Post("/upload", uploadHandler)

	app.Get("/stream/{fileName}", streamHandler)

	app.Run(iris.Addr(":8080"))
}

func uploadHandler(ctx iris.Context) {
	// Get the uploaded file
	file, info, err := ctx.FormFile("file")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(map[string]interface{}{
			"success": false,
			"message": "Failed to get file from request.",
		})
		return
	}
	defer file.Close()

	// Create a temporary file to store the uploaded file
	tempFile, err := os.CreateTemp("", "upload-*.temp")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(map[string]interface{}{
			"success": false,
			"message": "Failed to create temporary file.",
		})
		return
	}
	defer tempFile.Close()

	// Copy the uploaded file to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(map[string]interface{}{
			"success": false,
			"message": "Failed to copy file to temporary location.",
		})
		return
	}

	// Convert the uploaded file to MP4 using FFmpeg
	mp4File := strings.TrimSuffix(tempFile.Name(), ".temp") + ".mp4"
	ffmpegCmd := exec.Command("ffmpeg", "-i", tempFile.Name(), "-c:v", "libx264", "-c:a", "aac", "-strict", "-2", mp4File)
	err = ffmpegCmd.Run()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(map[string]interface{}{
			"success": false,
			"message": "Failed to convert file to MP4.",
		})
		return
	}

	// Convert the MP4 file to WebM using FFmpeg
	webmFile := strings.TrimSuffix(tempFile.Name(), ".temp") + ".webm"
	ffmpegCmd = exec.Command("ffmpeg", "-i", mp4File, "-c:v", "libvpx", "-c:a", "libvorbis", "-cpu-used", "5", "-deadline", "realtime", webmFile)
	err = ffmpegCmd.Run()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(map[string]interface{}{
			"success": false,
			"message": "Failed to convert MP4 to WebM.",
		})
		return
	}

	// Provide the preview URL to the client
	previewURL := "/stream/" + filepath.Base(webmFile)

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(map[string]interface{}{
		"success":    true,
		"message":    "File uploaded and converted successfully.",
		"previewURL": previewURL,
	})
}

func streamHandler(ctx iris.Context) {
	fileName := ctx.Params().Get("fileName")

	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		ctx.StatusCode(http.StatusNotFound)
		return
	}
	defer file.Close()

	// Get file information
	info, err := file.Stat()
	if err != nil {
		log.Println("Failed to get file information:", err)
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}

	// Set the Content-Length header
	ctx.ResponseWriter().Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))

	// Set the Content-Type header based on the file extension
	contentType := "application/octet-stream"
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".mp4":
		contentType = "video/mp4"
	case ".webm":
		contentType = "video/webm"
	}
	ctx.ResponseWriter().Header().Set("Content-Type", contentType)

	// Copy the file to the response writer
	_, err = io.Copy(ctx.ResponseWriter(), file)
	if err != nil {
		log.Println("Failed to copy file to response writer:", err)
		ctx.StatusCode(http.StatusInternalServerError)
	}
}
