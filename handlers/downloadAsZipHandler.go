package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"hello-chi/utils"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// DownloadAsZipHandlerClosure handles file downloads, supporting range requests.
//
//	@Summary		Download file
//	@Description	Downloads a file with a valid identifier. The eTag is used to retrieve the file Metadata from the Redis database. As long as the ownerId matches that in the cookie, the file is served.
//	@Tags			Files
//	@Param			identifier	path		string	true	"File identifier (etag)"
//	@Success		200			{string}	string	"File downloaded successfully"
//	@Success		206			{string}	string	"Binary data"
//	@Failure		400			{string}	string	"Bad request or invalid identifier"
//	@Failure		500			{string}	string	"Bad request or invalid identifier"
//	@Router			/download/{identifier} [get]
func DownloadAsZipHandlerClosure(rdb *redis.Client) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {
		// hardcoded file path for testing purposes

		uploadDir := os.Getenv("STOR_DIR")
		eTag := chi.URLParam(request, "identifier")

		// check if the identifier exists in redis
		fileMetaData, err := utils.GetFileMetadata(rdb, eTag)
		if err != nil {
			errorMsg := fmt.Sprintf("Invalid identifier or link expired: %s\n eTag= %s", err.Error(), eTag)
			http.Error(writer, errorMsg, http.StatusBadRequest)
			return
		}

		if fileMetaData.Path == "" || fileMetaData.Size == 0 {
			http.Error(writer, "Invalid identifier or link expired", http.StatusBadRequest)
			return
		}

		file, err := os.Open(uploadDir + "/" + fileMetaData.Path)
		if err != nil {
			http.Error(writer, "Unable to open file for reading", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		fileName := fileMetaData.Name

		fileInfo, err := file.Stat()
		if err != nil {
			http.Error(writer, "Unable to retrieve file info", http.StatusInternalServerError)
			return
		}
		fileSize := fileInfo.Size()

		rangeHeader := request.Header.Get("Range")
		if rangeHeader != "" {

			rangeHeader = strings.TrimPrefix(rangeHeader, "bytes=")
			rangeParts := strings.Split(rangeHeader, "-")

			start, err := strconv.ParseInt(rangeParts[0], 10, 64)
			if err != nil {
				http.Error(writer, "Invalid Range header", http.StatusBadRequest)
				return
			}

			var end int64
			if len(rangeParts) == 2 && rangeParts[1] != "" {
				end, err = strconv.ParseInt(rangeParts[1], 10, 64)
				if err != nil {
					http.Error(writer, "Invalid Range header", http.StatusBadRequest)
					return
				}
			} else {
				end = fileSize - 1
			}

			if start >= fileSize || end >= fileSize || start > end {
				http.Error(writer, "Requested range not satisfiable", http.StatusRequestedRangeNotSatisfiable)
				return
			}

			writer.Header().Set("Content-Type", "application/octet-stream")
			writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileMetaData.Name))
			writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
			writer.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
			writer.WriteHeader(http.StatusPartialContent)

			_, err = file.Seek(start, io.SeekStart)
			if err != nil {
				http.Error(writer, "Unable to seek to requested range", http.StatusInternalServerError)
				return
			}

			_, err = io.CopyN(writer, file, end-start+1)
			if err != nil && err != io.EOF {
				http.Error(writer, "Error streaming file", http.StatusInternalServerError)
				return
			}
		} else {
			writer.Header().Set("Content-Type", "application/octet-stream")
			writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
			writer.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
			_, err = io.Copy(writer, file)
			if err != nil {
				http.Error(writer, "Unable to copy file data to response", http.StatusInternalServerError)
				return
			}
		}
	}

}
