package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func ConcatenateChunks(identifier string, totalSize int64, fileExtension string) (string, error) {
	var tempDir = os.Getenv("TEMP_DIR")
	var uploadDir = os.Getenv("STOR_DIR")

	chunkDir := fmt.Sprintf("%s/%s", tempDir, identifier)
	files, err := os.ReadDir(chunkDir)
	if err != nil {
		return "", fmt.Errorf("unable to read chunk directory: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		iStart := getStartByteFromFilename(files[i].Name())
		jStart := getStartByteFromFilename(files[j].Name())
		return iStart < jStart
	})

	// Create the final file
	finalFilePath := fmt.Sprintf("%s/%s%s", uploadDir, identifier, fileExtension)
	finalFilePathWrtToUploadDir := fmt.Sprintf("%s%s", identifier, fileExtension)
	errr := os.MkdirAll(uploadDir, os.ModePerm)
	if errr != nil {
		return "", fmt.Errorf("unable to create upload directory: %w", errr)
	}
	finalFile, err := os.Create(finalFilePath)
	if err != nil {
		return "", fmt.Errorf("unable to create final file: %w", err)
	}
	defer func(finalFile *os.File) {
		err := finalFile.Close()
		if err != nil {
			_ = fmt.Errorf("unable to close final file: %w", err)
		}
	}(finalFile)
	// Create a hash object to compute hash while writing
	hash := sha256.New()

	// Concatenate each chunk
	for _, chunkFile := range files {
		chunkPath := filepath.Join(chunkDir, chunkFile.Name())
		chunkFileHandle, err := os.Open(chunkPath) // Open the chunk file for reading
		if err != nil {
			return "", fmt.Errorf("unable to open chunk file %s: %w", chunkFile.Name(), err)
		}
		defer chunkFileHandle.Close()

		// Copy chunk data to final file and simultaneously to the hash
		if _, err := io.Copy(finalFile, io.TeeReader(chunkFileHandle, hash)); err != nil {
			return "", fmt.Errorf("unable to write to final file: %w", err)
		}
	}

	//check if the final file size is equal to the total size
	fileInfo, err := finalFile.Stat()
	if err != nil {
		return "", fmt.Errorf("unable to get file info: %w", err)
	}
	if fileInfo.Size() != totalSize {
		fmt.Println("size of the final file: ", fileInfo.Size(), "total size: ", totalSize)
		return "", fmt.Errorf("final file size does not match total size")
	} else {
		//check hash of the final file
		fmt.Println("Hash of the final file: ", fmt.Sprintf("%x", hash.Sum(nil)))
		//delete the chunk directory
		err = os.RemoveAll(chunkDir)
		if err != nil {
			fmt.Println("Error deleting chunk directory: ", err)
			return "", fmt.Errorf("unable to delete chunk directory: %w", err)
		}
	}

	return finalFilePathWrtToUploadDir, nil
}

func getStartByteFromFilename(filename string) int64 {
	// Assuming filename format is `identifier_start_end`, extract `start`
	parts := strings.Split(filename, "_")
	if len(parts) >= 2 {
		start, _ := strconv.ParseInt(parts[1], 10, 64)
		return start
	}
	return 0
}

func TestGetStartByteFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		expected int64
	}{
		{"identifier_0_100", 0},
		{"identifier_100_200", 100},
		{"identifier_200_300", 200},
		{"identifier_invalid", 0},
		{"identifier_abc_300", 0},
	}

	for _, test := range tests {
		result := getStartByteFromFilename(test.filename)
		if result != test.expected {
			t.Errorf("getStartByteFromFilename(%s) = %d; want %d", test.filename, result, test.expected)
		}
	}
}
