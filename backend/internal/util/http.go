package util

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPClient wraps http.Client with cookie support
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetWithCookies performs a GET request and returns body + set-cookie headers
func (h *HTTPClient) GetWithCookies(url string, headers map[string]string) (string, []*http.Cookie, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	ApplyCookiesToRequest(req)

	resp, err := h.client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 412 {
		return "", nil, fmt.Errorf("HTTP 412 - 请求被拒绝")
	}

	body, err := readBody(resp)
	if err != nil {
		return "", nil, err
	}
	return body, resp.Cookies(), nil
}

// Get performs a GET request and returns the response body as string
func (h *HTTPClient) Get(url string, headers map[string]string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	ApplyCookiesToRequest(req)

	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 412 {
		return "", fmt.Errorf("HTTP 412 - 请求被拒绝")
	}

	return readBody(resp)
}

// Post performs a POST request with form data
func (h *HTTPClient) Post(url string, headers map[string]string, body string) (string, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	ApplyCookiesToRequest(req)

	resp, err := h.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 412 {
		return "", fmt.Errorf("HTTP 412 - 请求被拒绝")
	}

	return readBody(resp)
}

// Download downloads a file with resume support
func (h *HTTPClient) Download(url, savePath string, headers map[string]string, onProgress func(downloaded, total int64)) (bool, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	ApplyCookiesToRequest(req)

	resp, err := h.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	total := resp.ContentLength
	buf := make([]byte, 1024*1024) // 1MB buffer
	var downloaded int64

	out, err := createFile(savePath)
	if err != nil {
		return false, err
	}
	defer out.Close()

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return false, werr
			}
			downloaded += int64(n)
			if onProgress != nil {
				onProgress(downloaded, total)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return false, readErr
		}
	}
	return true, nil
}

func readBody(resp *http.Response) (string, error) {
	var reader io.Reader = resp.Body
	encoding := resp.Header.Get("Content-Encoding")
	if strings.Contains(encoding, "gzip") {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", err
		}
		defer gz.Close()
		reader = gz
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return "", err
	}
	return buf.String(), nil
}
