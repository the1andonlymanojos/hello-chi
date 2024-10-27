package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"hello-chi/utils"
	_ "hello-chi/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// UploadChunkHandlerClosure handles chunk uploads for files, verifying content ranges.
//
//	@Summary		Upload file chunk
//	@Description	Uploads a file chunk with a valid identifier and content range
//	@Tags			Files
//	@Param			identifier		path			string			true	"File identifier (etag)"
//	@Param			Content-Range	header			string			true	"Content-Range of the chunk (e.g., bytes 0-100/500)"
//	@Success		206				{string}		string			"Chunk uploaded successfully"
//	@Failure		400				{string}		string			"Bad request or invalid chunk range"
//	@Failure		500				{string}		string			"Internal Server Error"
//	@Failure		999				{object}	utils.FileMetaData	"FileMetaData"
//	@Router			/upload/{identifier} [put]
func UploadChunkHandlerClosure(rdb *redis.Client) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var tempDir = os.Getenv("TEMP_DIR")
		//var uploadDir = os.Getenv("UPLOAD_DIR")
		identifier := chi.URLParam(request, "identifier")
		contentRange := request.Header.Get("Content-Range")

		// check if the identifier exists in redis
		fileMetaData, err := utils.GetFileMetadata(rdb, identifier)
		fmt.Println(fileMetaData)
		if err != nil {
			http.Error(writer, "Invalid identifier or link expired", http.StatusBadRequest)
			return
		}
		//check if the file is already uploaded
		fmt.Println(fileMetaData.LastByteReceived, fileMetaData.Size, "LastByteReceived and Size")
		if fileMetaData.LastByteReceived == fileMetaData.Size {
			fmt.Println("File already uploaded")
			http.Error(writer, "File already uploaded", http.StatusBadRequest)
			return
		}

		// get the range
		var start, end int64
		if contentRange != "" {
			println(contentRange)
			parts := strings.Split(contentRange, " ")
			if len(parts) == 2 {
				rangeParts := strings.Split(parts[1], "-")
				if len(rangeParts) == 2 {
					start, _ = strconv.ParseInt(rangeParts[0], 10, 64)
					if strings.Contains(rangeParts[1], "/") {
						newRangeParts := strings.Split(rangeParts[1], "/")
						end, _ = strconv.ParseInt(newRangeParts[0], 10, 64)
					} else {
						end, _ = strconv.ParseInt(rangeParts[1], 10, 64)
					}

				}
			}
		} else {
			http.Error(writer, "Missing Content-Range header", http.StatusBadRequest)
			return
		}

		// check if the range is valid
		if start != fileMetaData.LastByteReceived+1 {
			println("expected", fileMetaData.LastByteReceived+1, "got", start)
			http.Error(writer, "Invalid chunk range", http.StatusBadRequest)
			return
		}

		chunkFileName := fmt.Sprintf("%s_%d_%d", identifier, start, end)

		dir := filepath.Join(tempDir, identifier)
		fmt.Println(dir)
		err = os.MkdirAll(dir, os.ModePerm) // Create the directory and any necessary parents
		if err != nil {
			fmt.Println(err)
			http.Error(writer, "Unable to create directory "+dir, http.StatusInternalServerError)
			return
		}
		dir = filepath.Join(tempDir, identifier, chunkFileName)

		file, err := os.OpenFile(dir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(dir)
			fmt.Println(err)
			http.Error(writer, "Unable to open file for writing "+dir, http.StatusInternalServerError)
			return
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Println("Error closing file:", err)
			}
		}(file)

		// update the last byte received
		fileMetaData.UpdateLastByteReceived(end)
		err = utils.StoreFileMetadata(rdb, identifier, fileMetaData)
		if err != nil {
			http.Error(writer, "Unable to update file metadata", http.StatusInternalServerError)
			return
		}

		fmt.Println("LastByteReceived updated to ", fileMetaData.LastByteReceived)
		fmt.Println("Size of the file is ", fileMetaData.Size)
		// write the chunk data to the file
		_, err = io.Copy(file, request.Body)
		if err != nil {
			http.Error(writer, "Unable to write chunk data to file", http.StatusInternalServerError)
			return
		}

		// check if the file is fully uploaded
		if fileMetaData.LastByteReceived == fileMetaData.Size-1 {
			// file extension
			ext := filepath.Ext(fileMetaData.Name)
			finalFilePath, err := utils.ConcatenateChunks(identifier, fileMetaData.Size, ext)
			if err != nil {
				fmt.Println(err)
				http.Error(writer, "Unable to concatenate chunks ", http.StatusInternalServerError)
				return
			}
			fileMetaData.Path = finalFilePath
			err = utils.StoreFileMetadata(rdb, identifier, fileMetaData)

		} else {
			writer.WriteHeader(http.StatusPartialContent)
			write, err := writer.Write([]byte(fmt.Sprint("Chunk uploaded successfully, next expected byte is ", fileMetaData.LastByteReceived+1)))
			if err != nil {
				fmt.Println("Error writing response:", write)
				return
			}
		}
	}
}
