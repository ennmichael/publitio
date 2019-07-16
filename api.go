package publitio

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// API is used to make all API calls.
type API struct {
	Key    string
	Secret string
}

// Response is the parsed JSON server response.
type Response interface{}

// UploadFile uploads a media file to the server using the filename.
// To upload a file from memory, use api.UploadFile(fileReader, url.Values{"title": {"My file"}}).
// To upload a file from a remote URL, use api.UploadFile(nil, url.Values{"file_url": {"https://example.com/file.png"}, "title": {"My file"}}).
func (api *API) UploadFile(file io.Reader, values url.Values) (result Response, err error) {
	url, err := api.publitioURL("/files/create", values)
	if err != nil {
		return nil, fmt.Errorf("error while creating Publitio url: %v", err)
	}

	var res *http.Response

	if file == nil {
		res, err = http.Post(url, "multipart/form-data", &bytes.Buffer{})
	} else {
		content, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("error while reading file %v", err)
		}

		requestBody := &bytes.Buffer{}
		multipartWriter := multipart.NewWriter(requestBody)
		w, err := multipartWriter.CreateFormFile("file", "new ")
		if err != nil {
			return nil, fmt.Errorf("error while creating multipart writer: %v", err)
		}

		_, err = w.Write(content)
		if err != nil {
			return nil, fmt.Errorf("error while writing multipart data: %v", err)
		}

		err = multipartWriter.Close()
		if err != nil {
			return nil, fmt.Errorf("error while closing the multipart writer: %v", err)
		}

		client := http.Client{}
		res, err = client.Post(url, "multipart/form-data; boundary="+multipartWriter.Boundary(), requestBody)
		if err != nil {
			return nil, fmt.Errorf("error while performing HTTP request: %v", err)
		}
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			result = nil
		}
	}()

	result, err = parseResponse(res)
	if err != nil {
		return nil, fmt.Errorf("error while parsing the response: %v", err)
	}

	return result, nil
}

// Get performs a GET request to the server, for example when listing all files.
func (api *API) Get(path string, values url.Values) (Response, error) {
	res, err := api.Call("GET", path, values)
	if err != nil {
		return nil, fmt.Errorf("error while performing Publitio API GET: %v", err)
	}

	return res, nil
}

// Put performs a PUT request to the server, for example when updating files.
func (api *API) Put(path string, values url.Values) (Response, error) {
	res, err := api.Call("PUT", path, values)
	if err != nil {
		return nil, fmt.Errorf("error while performing Publitio API PUT: %v", err)
	}

	return res, nil
}

// Delete performs a DELETE request to the server, for example when deleting files.
func (api *API) Delete(path string, values url.Values) (Response, error) {
	res, err := api.Call("DELETE", path, values)
	if err != nil {
		return nil, fmt.Errorf("error while performing Publitio API DELETE: %v", err)
	}

	return res, nil
}

// Call performs any request to the server; use Get, Put and Delete for convenience.
// If you need a post request, you should probably use Upload or UploadFile.
func (api *API) Call(method, path string, values url.Values) (result Response, err error) {
	url, err := api.publitioURL(path, values)
	if err != nil {
		return nil, fmt.Errorf("error while creating Publitio URL: %v", err)
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating HTTP request: %v", err)
	}
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while performing HTTP request: %v", err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			result = nil
		}
	}()

	result, err = parseResponse(res)
	if err != nil {
		return nil, fmt.Errorf("error while parsing the Publitio response: %v", err)
	}

	return result, nil
}

func parseResponse(res *http.Response) (Response, error) {
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %v", err)
	}

	var r interface{}
	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("error while parsing JSON %s: %v", data, err)
	}

	return r, nil
}

func (api *API) publitioURL(path string, values url.Values) (string, error) {
	const baseURL = "https://api.publit.io/v1"
	var u *url.URL
	var err error

	if path[0] == '/' {
		u, err = url.Parse(baseURL + path)
	} else {
		u, err = url.Parse(baseURL + "/" + path)
	}
	if err != nil {
		return "", err
	}

	nonce, err := generateNonce()
	if err != nil {
		return "", fmt.Errorf("error while generating nonce: %v", err)
	}

	// Apparently this has to be a 32-bit number, but Unix() returns a 64-bit number
	timestamp := strconv.FormatInt(time.Now().Unix()%0xFFFFFFFF, 10)

	queryValues := make(url.Values)
	queryValues["api_nonce"] = []string{nonce}
	queryValues["api_timestamp"] = []string{timestamp}
	queryValues["api_key"] = []string{api.Key}
	queryValues["api_signature"] = []string{signature(api.Secret, timestamp, nonce)}
	for k, v := range values {
		queryValues[k] = v
	}

	u.RawQuery = queryValues.Encode()
	return u.String(), nil
}

func signature(secret, timestamp, nonce string) string {
	sum := sha1.Sum([]byte(timestamp + nonce + secret))
	hexSum := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(hexSum, sum[:])
	return string(hexSum[:])
}

func generateNonce() (string, error) {
	r, err := rand.Int(rand.Reader, big.NewInt(89999999))
	if err != nil {
		return "", fmt.Errorf("error while generating nonce: %v", err)
	}
	r.Add(r, big.NewInt(10000000))
	return r.String(), nil
}
