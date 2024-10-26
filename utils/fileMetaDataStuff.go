package utils

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
)

var ctx = context.Background()

// FileMetaData represents metadata for a file, as stored in Redis as a stringified JSON object.
// @Description	FileMetaData represents metadata for a file upload.
// @Name			FileMetaData
type FileMetaData struct {
	Size             int64  `json:"size"`
	UploadDate       string `json:"upload_date"`
	Path             string `json:"path"`
	Owner            string `json:"owner"`
	TTL              int64  `json:"ttl"`
	HASH             string `json:"hash"`
	LastByteReceived int64  `json:"last_byte_received"`
	Name             string `json:"name"`
}

func CreateFileMetadata(size int64, path, owner string, ttl int64, hash string, lastByteReceived int64, name string) FileMetaData {
	return FileMetaData{
		Size:             size,
		UploadDate:       time.Now().Format(time.RFC3339), // Set upload date to current time
		Path:             path,
		Owner:            owner,
		TTL:              ttl,
		HASH:             hash,
		LastByteReceived: lastByteReceived,
		Name:             name,
	}
}

func (f *FileMetaData) UpdateLastByteReceived(lastByteReceived int64) {
	f.LastByteReceived = lastByteReceived
}
func StoreFileMetadata(rdb *redis.Client, etag string, metadata FileMetaData) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	err = rdb.Set(ctx, etag, metadataJSON, time.Duration(metadata.TTL)*time.Second).Err()
	if err != nil {
		return err
	}

	//log.Println("File metadata updated in Redis:", etag)
	return nil
}

func GetFileMetadata(rdb *redis.Client, etag string) (FileMetaData, error) {
	val, err := rdb.Get(ctx, etag).Result()
	if err != nil {
		return FileMetaData{}, err
	}

	var metadata FileMetaData
	err = json.Unmarshal([]byte(val), &metadata)
	if err != nil {
		return FileMetaData{}, err
	}

	return metadata, nil
}
