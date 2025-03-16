package megaapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestChunkDownloader() *ChunkDownloader {
	return &ChunkDownloader{
		URL:    "http://test.com/file",
		Client: &http.Client{Timeout: 1 * time.Second},
	}
}

func TestCreateRequest(t *testing.T) {
	cd := newTestChunkDownloader()
	start, end := int64(0), int64(1024)

	req, err := cd.createRequest(start, end)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if req.URL.String() != cd.URL {
		t.Errorf("Expected URL %s, got %s", cd.URL, req.URL.String())
	}

	expectedRange := fmt.Sprintf("bytes=%d-%d", start, end)
	if req.Header.Get("Range") != expectedRange {
		t.Errorf("Expected Range header %s, got %s", expectedRange, req.Header.Get("Range"))
	}
}

func TestSetRangeHeader(t *testing.T) {
	cd := newTestChunkDownloader()
	req := httptest.NewRequest(http.MethodGet, cd.URL, nil)
	start, end := int64(0), int64(1024)

	cd.setRangeHeader(req, start, end)

	expectedRange := fmt.Sprintf("bytes=%d-%d", start, end)
	if req.Header.Get("Range") != expectedRange {
		t.Errorf("Expected Range header %s, got %s", expectedRange, req.Header.Get("Range"))
	}
}

func TestSendRequest(t *testing.T) {
	cd := newTestChunkDownloader()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	cd.URL = mockServer.URL
	req, err := cd.createRequest(0, 1024)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	resp, err := cd.sendRequest(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestCheckResponseStatus(t *testing.T) {
	cd := newTestChunkDownloader()

	resp := &http.Response{StatusCode: http.StatusOK}
	if err := cd.checkResponseStatus(resp); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	resp.StatusCode = http.StatusPartialContent
	if err := cd.checkResponseStatus(resp); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	resp.StatusCode = http.StatusNotFound
	if err := cd.checkResponseStatus(resp); err == nil {
		t.Errorf("Expected error, got none")
	}
}

func TestGetContentLength(t *testing.T) {
	cd := newTestChunkDownloader()

	resp := &http.Response{Header: make(http.Header)}
	_, err := cd.getContentLength(resp)
	if err == nil {
		t.Errorf("Expected error, got none")
	}

	resp.Header.Set("Content-Length", "invalid")
	_, err = cd.getContentLength(resp)
	if err == nil {
		t.Errorf("Expected error, got none")
	}

	resp.Header.Set("Content-Length", "1024")
	length, err := cd.getContentLength(resp)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if length != 1024 {
		t.Errorf("Expected length 1024, got %d", length)
	}
}

func TestReadResponseBody(t *testing.T) {
	cd := newTestChunkDownloader()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := []byte("This is a test body.") // Simulate a short body
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer mockServer.Close()

	cd.URL = mockServer.URL

	_, err := cd.DownloadRange(0, int64(len("This is a test body. This is a test body.")))
	if err == nil {
		t.Errorf("Expected error, got none")
	}

	data, err := cd.DownloadRange(0, int64(len("This is a test body.")))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if string(data) != "This is a test body." {
		t.Errorf("Expected body 'This is a test body.', got %s", string(data))
	}
}

func TestDownloadRange(t *testing.T) {
	cd := newTestChunkDownloader()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := []byte("This is a test body. This is a test body. This is a test body.")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer mockServer.Close()

	cd.URL = mockServer.URL

	start, end := int64(0), int64(len("This is a test body. This is a test body. This is a test body."))
	data, err := cd.DownloadRange(start, end)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if string(data) != string([]byte("This is a test body. This is a test body. This is a test body.")) {
		t.Errorf("Expected body %s, got %s", string([]byte("This is a test body. This is a test body. This is a test body.")), string(data))
	}

	_, err = cd.DownloadRange(start, end+1024)
	if err == nil {
		t.Errorf("Expected error for range larger than available content, got none")
	}

	_, err = cd.DownloadRange(start, int64(len("This is a test body.")))
	if err != nil {
		t.Errorf("Expected no error for partial read, got %v", err)
	}
}
