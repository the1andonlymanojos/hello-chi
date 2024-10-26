package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"hello-chi/dbShit"
	"hello-chi/utils"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInitiateUpload(t *testing.T) {
	utils.Init()
	// Create a sample upload request payload
	uploadReq := UploadRequest{
		Size: 1024,
		Name: "test_file.txt",
		HASH: "testhash",
	}
	rdb := dbShit.InitializeRedisClient()
	reqBody, _ := json.Marshal(uploadReq)

	// Create a new HTTP request for the /upload/initiate endpoint
	req, err := http.NewRequest("POST", "/upload/initiate", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Initialize the router and handler for the test
	r := chi.NewRouter()
	r.Post("/upload/initiate", InitiateUploadClosure(rdb)) // Pass your redis client

	// Serve the request
	r.ServeHTTP(rr, req)

	// Check the status code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, status)
	}

	// Check that the response body contains the identifier
	var resp map[string]string
	err = json.NewDecoder(rr.Body).Decode(&resp)
	if err != nil {
		t.Errorf("Unable to decode response body: %v", err)
	}
	if _, ok := resp["eTag"]; !ok {
		t.Errorf("expected identifier in response")
	}

	fmt.Println("Response Body:", resp)

	//cleanup
	//dir := filepath.Join(utils.TempDir, resp["eTag"])
	//os.Remove(dir)
	//rdb.Del(ctx, resp["eTag"])

}
