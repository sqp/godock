package upload

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

// PostForm creates a new http POST request with optional extra params.
//
func PostForm(url string, values url.Values) (string, error) {
	r, e := http.PostForm(url, values)
	if e != nil {
		// Log.Info("error posting %s", e)
		return "", e
	}
	defer r.Body.Close()
	body, er := ioutil.ReadAll(r.Body)
	if er != nil {
		// Log.Info("error reading %s", er)
		return "", er
	}

	if len(body) == 0 {
		return "", errors.New("POST output empty")
	}

	return string(body), nil
}

// File creates a new http PUT request with the given file path.
//
func File(method, url, file string) (string, error) {
	f, e := os.OpenFile(file, os.O_RDONLY, 0644)
	if e != nil {
		return "", e
	}
	defer f.Close()
	stat, _ := os.Stat(file)
	return Reader(method, url, f, stat.Size())
}

// Reader creates a new http request with the given reader and method (PUT / POST).
//
func Reader(method, url string, r io.Reader, size int64) (string, error) {
	req, e := http.NewRequest(method, url, r)
	if e != nil {
		return "", e
	}
	req.ContentLength = size
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return "", e
	}
	defer resp.Body.Close()
	ret, e := ioutil.ReadAll(resp.Body)
	return string(ret), e
}

// ReaderMultipart creates a new http multipart POST upload request with optional extra params.
//
// Thanks to https://gist.github.com/mattetti/5914158/f4d1393d83ebedc682a3c8e7bdc6b49670083b84
//
func ReaderMultipart(url, fileref, filename string, limitRate int, rdr io.Reader, size int64, params map[string]string) (string, error) {
	bodyOut := new(bytes.Buffer)
	writer := multipart.NewWriter(bodyOut)
	part, e := writer.CreateFormFile(fileref, filename)
	if e != nil {
		return "", e
	}

	io.Copy(part, rdr)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	e = writer.Close()
	if e != nil {
		return "", e
	}

	request, e := http.NewRequest("POST", url, bodyOut)
	if e != nil {
		return "", e
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, e := client.Do(request)
	if e != nil {
		return "", e
	}
	defer resp.Body.Close()

	bodyIn, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return "", e
	}
	return string(bodyIn), nil
}
