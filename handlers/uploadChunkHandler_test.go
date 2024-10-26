package handlers

import (
	"bytes"
	"context"
	"github.com/go-chi/chi/v5"
	"hello-chi/dbShit"
	"hello-chi/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var ctx = context.Background()

func TestSingleFileUpload(t *testing.T) {
	utils.Init()
	// Create a sample upload request payload
	chunkData := []byte("sample chunk data")
	reqBody := bytes.NewBuffer(chunkData)
	rdb := dbShit.InitializeRedisClient()

	utils.StoreFileMetadata(rdb, "testIdentifer", utils.FileMetaData{Size: int64(len(chunkData)), Name: "test.txt", LastByteReceived: -1})

	// Create a new HTTP request for the /upload/chunk endpoint
	req, err := http.NewRequest("PUT", "/upload/chunk/testIdentifer", reqBody)
	req.Header.Set("Content-Range", "bytes 0-16") // Adjust based on your expected range
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Initialize the router and handler for the test
	r := chi.NewRouter()
	r.Put("/upload/chunk/{identifier}", UploadChunkHandlerClosure(rdb)) // Pass your redis client

	// Serve the request
	r.ServeHTTP(rr, req)

	// Check the status code is 206 Partial Content
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, status)
		t.Errorf("Response Body: %v", rr.Body.String())
	}

	//cleanup
	//rdb.Del(ctx, "testIdentifer")
	//os.Remove("/Users/manojos/Projects/Go/file-service-polyfile/volume/STORAGE/testIdentifer")
}

func TestMultipleChunkUpload(t *testing.T) {
	utils.Init()
	// Create a sample upload request payload
	filePath := "../mocks/glimpse-of-us.mp3"
	file, err := os.Open(filePath)
	FileInfo, _ := file.Stat()
	size := FileInfo.Size()
	if err != nil {
		t.Fatal(err)
	}
	rdb := dbShit.InitializeRedisClient()

	utils.StoreFileMetadata(rdb, "testIdentifer", utils.FileMetaData{Size: size, Name: "glimpse-of-us.mp3", LastByteReceived: -1})
	// Initialize the router and handler for the test
	r := chi.NewRouter()
	r.Put("/upload/chunk/{identifier}", UploadChunkHandlerClosure(rdb)) // Pass your redis client
	const chunkSize = 1024 * 1024                                       // Example chunk size (in bytes)
	var totalUploaded int64

	for start := int64(0); start < size; start += chunkSize {
		// Read the chunk from the file
		chunkData := make([]byte, chunkSize)
		n, err := file.Read(chunkData)
		if err != nil && err.Error() != "EOF" {
			t.Fatal(err)
		}
		chunkData = chunkData[:n] // Adjust the chunk size based on the read

		// Create a new HTTP request for the /upload/chunk endpoint
		reqBody := bytes.NewBuffer(chunkData)
		req, err := http.NewRequest("PUT", "/upload/chunk/testIdentifer", reqBody)
		if err != nil {
			t.Fatal(err)
		}

		// Set the Content-Range header for the current chunk
		req.Header.Set("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(start+int64(n)-1, 10))

		// Create a ResponseRecorder to capture the response
		rr := httptest.NewRecorder()

		// Serve the request
		r.ServeHTTP(rr, req)

		if start >= size-chunkSize {
			// Check the status code is 200 OK
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("expected status code %v, got %v", http.StatusOK, status)
				t.Errorf("Response Body: %v", rr.Body.String())
			}
		} else if status := rr.Code; status != http.StatusPartialContent {
			t.Errorf("expected status code %v, got %v", http.StatusPartialContent, status)
			t.Errorf("Response Body: %v", rr.Body.String())
		}

		totalUploaded += int64(n) // Update total uploaded size
	}

	//cleanup
	//rdb.Del(ctx, "testIdentifer")
	//os.Remove("/Users/manojos/Projects/Go/file-service-polyfile/volume/STORAGE/testIdentifer")
}

func TestDownloadHandlerClosure(t *testing.T) {
	utils.Init()
	stor_dir := os.Getenv("STOR_DIR")
	rdb := dbShit.InitializeRedisClient()
	// Create a new HTTP request for the /download endpoint
	req, err := http.NewRequest("GET", "/download/testIdentifer", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Range", "bytes=0-20")
	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Initialize the router and handler for the test
	r := chi.NewRouter()
	r.Get("/download/{identifier}", DownloadHandlerClosure(rdb)) // Pass your redis client

	// Serve the request
	r.ServeHTTP(rr, req)

	// Check the status code is 206 Partial Content
	range2 := rr.Header().Get("Content-Range")
	if range2 != "bytes 0-20/4685409" {
		t.Errorf("expected Content-Range %v, got %v", "bytes 0-20/4685409", range2)
	}

	//check if the response is a partial content and the is correct

	if status := rr.Code; status != http.StatusPartialContent {
		t.Errorf("expected status code %v, gasdfsot %v", http.StatusPartialContent, status)
		t.Errorf("Response Body: %v", rr.Body.String())
	}

	//cleanup
	rdb.Del(ctx, "testIdentifer")

	os.Remove(stor_dir + "/" + "testIdentifer.mp3")

}
