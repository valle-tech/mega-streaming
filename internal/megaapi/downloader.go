package megaapi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type ChunkDownloader struct {
	URL     string
	Client  *http.Client
	Timeout time.Duration
}

func NewChunkDownloader(url string, timeout time.Duration) *ChunkDownloader {
	return &ChunkDownloader{
		URL:     url,
		Client:  &http.Client{Timeout: timeout},
		Timeout: timeout,
	}
}

func (cd *ChunkDownloader) DownloadRange(start, end int64) ([]byte, error) {
	req, err := cd.createRequest(start, end)
	if err != nil {
		return nil, err
	}

	resp, err := cd.sendRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := cd.checkResponseStatus(resp); err != nil {
		return nil, err
	}

	length, err := cd.getContentLength(resp)
	if err != nil {
		return nil, err
	}

	if length < int(end-start) {
		return nil, fmt.Errorf("requested range exceeds content length: requested %d, available %d", end-start, length)
	}

	data, err := cd.readResponseBody(resp, length)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (cd *ChunkDownloader) createRequest(start, end int64) (*http.Request, error) {
	req, err := http.NewRequest("GET", cd.URL, nil)
	if err != nil {
		return nil, err
	}
	cd.setRangeHeader(req, start, end)
	return req, nil
}

func (cd *ChunkDownloader) setRangeHeader(req *http.Request, start, end int64) {
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
}

func (cd *ChunkDownloader) sendRequest(req *http.Request) (*http.Response, error) {
	resp, err := cd.Client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (cd *ChunkDownloader) checkResponseStatus(resp *http.Response) error {
	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (cd *ChunkDownloader) getContentLength(resp *http.Response) (int, error) {
	totalLength := resp.Header.Get("Content-Length")
	if totalLength == "" {
		return 0, errors.New("missing Content-Length in response")
	}

	length, err := strconv.Atoi(totalLength)
	if err != nil || length <= 0 {
		return 0, errors.New("invalid Content-Length")
	}

	return length, nil
}

func (cd *ChunkDownloader) readResponseBody(resp *http.Response, length int) ([]byte, error) {
	data := make([]byte, length)
	totalRead := 0

	for totalRead < length {
		n, err := resp.Body.Read(data[totalRead:])
		totalRead += n
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("read error: %w", err)
		}
	}

	if totalRead != length {
		return nil, fmt.Errorf("incomplete read: expected %d, got %d", length, totalRead)
	}

	return data, nil
}
