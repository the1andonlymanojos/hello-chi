package handlers

import (
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"hello-chi/utils"
	"net/http"
	"time"
)

type UploadFileInternalRequest struct {
	Size  int64  `json:"size"`
	Name  string `json:"name"`
	HASH  string `json:"hash"`
	Owner string `json:"owner"`
	UUID  string `json:"uuid"`
	Path  string `json:"path"`
}

// UploadFileInternalClosure handles file uploads internally.
//
//	@Summary		Upload file internally
//	@Description	Used by file conversion services to upload files internally
//	@Tags			Internal API
//	@Accept			json
//	@Produce		json
//	@Param			uploadReq	body		UploadFileInternalRequest	true	"File upload request"
//	@Success		200			{string}	string						"File uploaded successfully"
//	@Failure		400			{string}	string						"Bad request"
//	@Failure		500			{string}	string						"Internal Server Error"
//	@Router			/upload/internal [post]
func UploadFileInternalClosure(rdb *redis.Client) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		var uploadReq UploadFileInternalRequest
		err := json.NewDecoder(request.Body).Decode(&uploadReq)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		newUUID := uploadReq.UUID
		newFileMetadata := utils.FileMetaData{
			Size:             uploadReq.Size,
			UploadDate:       time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Path:             uploadReq.Path,
			Owner:            uploadReq.Owner,
			TTL:              7200,
			HASH:             uploadReq.HASH,
			LastByteReceived: uploadReq.Size,
			Name:             uploadReq.Name,
		}
		err = utils.StoreFileMetadata(rdb, newUUID, newFileMetadata)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(map[string]string{"uuid": newUUID})

	}
}
