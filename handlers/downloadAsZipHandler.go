package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"hello-chi/utils"
	"net/http"
	"os"
)

type DownloadRequest struct {
	Etags []string `json:"etags"`
}

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
		var downloadRequest DownloadRequest
		err := json.NewDecoder(request.Body).Decode(&downloadRequest)

		if err != nil {
			fmt.Println("Error decoding download request:", err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		var owner string

		filePaths := make([]string, 0)
		for _, eTag := range downloadRequest.Etags {
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
			owner = fileMetaData.Owner
			filePaths = append(filePaths, uploadDir+"/"+fileMetaData.Path)
		}

		newFileEtag := uuid.New().String()
		pathToZippedFile := uploadDir + "/" + newFileEtag + ".zip"

		fileSize, err := utils.ZipFiles(filePaths, pathToZippedFile)
		if err != nil {
			fmt.Println(err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		err = utils.StoreFileMetadata(rdb, newFileEtag, utils.FileMetaData{
			Size:             fileSize,
			UploadDate:       "2021-01-01",
			Path:             pathToZippedFile,
			Owner:            owner,
			TTL:              7200,
			HASH:             "CalcHashLater",
			LastByteReceived: fileSize - 1,
			Name:             "zippedFiles.zip",
		})

		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(map[string]string{"eTag": newFileEtag})

	}

}
