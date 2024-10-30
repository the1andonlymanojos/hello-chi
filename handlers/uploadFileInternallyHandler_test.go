package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"hello-chi/dbShit"
	"hello-chi/utils"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestUploadFileInternal(t *testing.T) {
	utils.Init()
	//var tempDir = os.Getenv("TEMP_DIR")
	var uploadDir = os.Getenv("STOR_DIR")

	randUUID := uuid.New().String()

	println(uploadDir)
	println(fmt.Sprintf("%s/%s.png", uploadDir, randUUID))

	//make a rando file in the volume
	_, err := os.Create(fmt.Sprintf("%s/%s.png", uploadDir, randUUID))
	if err != nil {
		t.Fatal(err)
	}
	//write 1024 bytes to the file
	file, err := os.OpenFile(fmt.Sprintf("%s/%s.png", uploadDir, randUUID), os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(rand.Intn(256))
	}
	_, err = file.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	uploadReq := UploadFileInternalRequest{
		Size:  1024,
		Name:  "rando.png",
		Owner: "sameOwner",
		UUID:  randUUID,
		Path:  fmt.Sprintf("%s.png", randUUID),
	}

	// Create a new HTTP request for the /upload/initiate endpoint
	r := chi.NewRouter()
	rdb := dbShit.InitializeRedisClient()
	r.Post("/upload/internal", UploadFileInternalClosure(rdb)) // Pass your redis client
	reqBody, _ := json.Marshal(uploadReq)
	req, err := http.NewRequest("POST", "/upload/internal", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, status)
	}

	//check by downloading the file
	req, err = http.NewRequest("GET", fmt.Sprintf("/download/%s", randUUID), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	r.Get("/download/{identifier}", DownloadHandlerClosure(rdb))
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status code %v, got %v", http.StatusOK, status)
		t.Errorf("Response Body: %v", rr.Body.String())
	}
	//print the headers
	fmt.Println(rr.Header())

	//cleanup
	os.Remove(fmt.Sprintf("%s/%s.png", uploadDir, randUUID))
	rdb.Del(ctx, randUUID)

}
