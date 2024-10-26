package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"hello-chi/utils"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type UploadRequest struct {
	Size int64  `json:"size"`
	Name string `json:"name"`
	HASH string `json:"hash"`
	//Metadata Metadata `json:"metadata"`
}

//type Metadata struct {
//	OwnerID       string `json:"ownerId"`
//	LastModified  int64  `json:"lastModified"`
//	FileExtension string `json:"fileExtension"`
//}

type UploadResponse struct {
	ETag string `json:"eTag"`
}

// InitiateUploadClosure initializes a new file upload by storing metadata.
//
//	@Summary		Initiate file upload
//	@Description	Initiates the upload process by creating metadata and returning an identifier
//	@Tags			Files
//	@Accept			json
//	@Produce		json
//	@Param			uploadReq	body		UploadRequest	true	"File upload request"
//	@Success		200			{object}	UploadResponse	"eTag identifier for the file upload"
//	@Failure		400			{string}	string			"Bad request"
//	@Failure		500			{string}	string			"Internal Server Error"
//	@Router			/upload/initiate [post]
func InitiateUploadClosure(rdb *redis.Client) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		var tempDir = os.Getenv("TEMP_DIR")
		//var uploadDir = os.Getenv("UPLOAD_DIR")

		fmt.Println("HERE")
		var uploadReq UploadRequest
		err := json.NewDecoder(request.Body).Decode(&uploadReq)
		if err != nil {
			log.Println("Error decoding upload request:", err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println(uploadReq)
		identifier := uuid.New().String()
		//create directory in the temp directory for this etag
		err = os.MkdirAll(filepath.Join(tempDir, identifier), 0755)
		if err != nil {
			log.Println("Error creating directory for upload:", err)
			http.Error(writer, "Unable to create directory for upload", http.StatusInternalServerError)
			return
		}
		fileMetadata := utils.CreateFileMetadata(uploadReq.Size, "", "OWNERID FROM JWT", 3600, uploadReq.HASH, -1, uploadReq.Name)
		err = utils.StoreFileMetadata(rdb, identifier, fileMetadata)
		if err != nil {
			log.Println("Error storing file metadata:", err)
			http.Error(writer, "Unable to store file metadata", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(UploadResponse{ETag: identifier})
	}
}
