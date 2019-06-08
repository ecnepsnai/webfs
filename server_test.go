package webfs_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ecnepsnai/security"
	"github.com/ecnepsnai/webfs"
)

var testTempDir *string

func TestMain(m *testing.M) {
	tmp, err := ioutil.TempDir("", "certbox")
	if err != nil {
		panic(err)
	}
	testTempDir = &tmp
	go webfs.Start(tmp, "127.0.0.1:8080")

	// Sleep just a bit to let the server start
	time.Sleep(10 * time.Millisecond)

	retCode := m.Run()
	os.RemoveAll(tmp)
	os.Exit(retCode)
}

func TestUploadAndGetFile(t *testing.T) {
	name := randomString(12)
	body := []byte(randomString(12))

	resp, err := http.Post("http://127.0.0.1:8080/"+name, "text/plain", bytes.NewReader(body))
	if err != nil {
		t.Fatalf(err.Error())
	}
	if resp.StatusCode != 200 {
		t.Errorf("Error uploading file: HTTP %d", resp.StatusCode)
	}

	resp, err = http.Get("http://127.0.0.01:8080/" + name)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if resp.StatusCode != 200 {
		t.Errorf("Error getting file: HTTP %d", resp.StatusCode)
	}
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading HTTP body: %s", err.Error())
	}

	if string(result) != string(body) {
		t.Errorf("Unexpected HTTP body. Expected '%s' got '%s'", body, result)
	}

	contentType := resp.Header.Get("Content-Type")
	disposition := resp.Header.Get("Content-Disposition")

	if contentType != "application/octet-stream" {
		t.Errorf("Unexpected response value for header %s '%s'. Expected: '%s'", "Content-Type", contentType, "application/octet-stream")
	}

	if disposition != "attachment" {
		t.Errorf("Unexpected response value for header %s '%s'. Expected: '%s'", "Content-Disposition", disposition, "attachment")
	}
}

func TestDeleteFile(t *testing.T) {
	name := randomString(12)
	body := []byte(randomString(12))

	resp, err := http.Post("http://127.0.0.1:8080/"+name, "text/plain", bytes.NewReader(body))
	if err != nil {
		t.Fatalf(err.Error())
	}
	if resp.StatusCode != 200 {
		t.Errorf("Error uploading file: HTTP %d", resp.StatusCode)
	}

	req, _ := http.NewRequest("DELETE", "http://127.0.0.01:8080/"+name, nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if resp.StatusCode != 200 {
		t.Errorf("Error deleting file: HTTP %d", resp.StatusCode)
	}
}

func TestGetNonexistingFile(t *testing.T) {
	name := "this file does not exist"

	resp, err := http.Get("http://127.0.0.01:8080/" + name)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if resp.StatusCode == 200 {
		t.Errorf("No error seen when one expected")
	}
}

func randomString(length uint16) string {
	return security.RandomString(length)
}
