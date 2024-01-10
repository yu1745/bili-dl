package api

import (
	"io"
	"net/http"
	"testing"
)

func TestWbi(t *testing.T) {
	urlStr := "https://api.bilibili.com/x/space/wbi/acc/info?mid=1850091"
	newUrlStr, err := sign(urlStr)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	req, err := http.NewRequest("GET", newUrlStr, nil)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("Request failed: %s", err)
		return
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Errorf("Failed to read response: %s", err)
		return
	}
	t.Log(string(body))
}
