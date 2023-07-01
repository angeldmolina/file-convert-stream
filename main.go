package main

import (
	"fmt"
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

	app.Post("/upload", uploadHandler)

	app.Run(iris.Addr(":8080"))
}

func uploadHandler(ctx iris.Context) {
	// Get the uploaded file
	file, info, err := ctx.FormFile("file")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.WriteString("Failed to get file from request.")
		return
	}
	defer file.Close()

	// Create a temporary file to store the uploaded file
	tempFile, err := os.CreateTemp("", "upload-*.temp")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.WriteString("Failed to create temporary file.")
		return
	}
	defer tempFile.Close()

	// Copy the uploaded file to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.WriteString("Failed to copy file to temporary location.")
		return
	}

	// Convert the uploaded file to MP4 using FFmpeg
	mp4File := strings.TrimSuffix(tempFile.Name(), ".temp") + ".mp4"
	ffmpegCmd := exec.Command("ffmpeg", "-i", tempFile.Name(), "-c:v", "libx264", "-c:a", "aac", "-strict", "-2", mp4File)
	err = ffmpegCmd.Run()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.WriteString("Failed to convert file to MP4.")
		return
	}

	// Convert the MP4 file to WebP using FFmpeg
	webpFile := strings.TrimSuffix(tempFile.Name(), ".temp") + ".webp"
	ffmpegCmd = exec.Command("ffmpeg", "-i", mp4File, webpFile)
	err = ffmpegCmd.Run()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.WriteString("Failed to convert MP4 to WebP.")
		return
	}

	// Stream the converted file to the client
	err = streamFile(ctx, webpFile)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.WriteString("Failed to stream file to client.")
		return
	}

	// Clean up the temporary files
	err = os.Remove(tempFile.Name())
	if err != nil {
		log.Println("Failed to remove temporary file:", err)
	}
	err = os.Remove(mp4File)
	if err != nil {
		log.Println("Failed to remove MP4 file:", err)
	}
	err = os.Remove(webpFile)
	if err != nil {
		log.Println("Failed to remove WebP file:", err)
	}
}

func streamFile(ctx iris.Context, filePath string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file information
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file information: %w", err)
	}

	// Set the Content-Length header
	ctx.ResponseWriter().Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))

	// Set the Content-Type header based on the file extension
	contentType := "application/octet-stream"
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".mp4":
		contentType = "video/mp4"
	case ".webp":
		contentType = "image/webp"
	}
	ctx.ResponseWriter().Header().Set("Content-Type", contentType)

	// Copy the file to the response writer
	_, err = io.Copy(ctx.ResponseWriter(), file)
	if err != nil {
		return fmt.Errorf("failed to copy file to response writer: %w", err)
	}

	return nil
}
