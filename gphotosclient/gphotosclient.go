package gphotos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/labstack/gommon/log"
)

const apiVersion = "v1"
const basePath = "https://photoslibrary.googleapis.com/"

// GPhotosClient is a client for interacting with google photos api.
type GPhotosClient struct {
	client *http.Client
}

// NewGPhotos creates a new client.
func NewGPhotos(client *http.Client) *GPhotosClient {
	return &GPhotosClient{client}
}

//GetPagedLibraryContents - todo
func (c *GPhotosClient) GetPagedLibraryContents(ctx context.Context, nextPage string) (*GPhotos, error) {
	var body io.Reader = nil
	var req *http.Request = nil
	var err error = nil
	if nextPage != "" {
		req, err = http.NewRequest("GET", fmt.Sprintf("%s/%s/mediaItems?pageToken=%s", basePath, apiVersion, nextPage), body)
	} else {
		req, err = http.NewRequest("GET", fmt.Sprintf("%s/%s/mediaItems", basePath, apiVersion), body)
	}
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	gphotos := GPhotos{}
	err = json.Unmarshal(b, &gphotos)
	if err != nil {
		return nil, err
	}
	return &gphotos, nil
}

//DownloadMedia - todo
func (c *GPhotosClient) DownloadMedia(ctx context.Context, gphoto GPhoto) error {
	var body io.Reader = nil

	url := gphoto.BaseURL
	if gphoto.MediaMetaData.Photo != (gMetaPhoto{}) {
		url = url + "=d" //w" + gphoto.MediaMetaData.Width + "-h" + gphoto.MediaMetaData.Height + "=
	} else if gphoto.MediaMetaData.Video != (gMetaVideo{}) {
		url = url + "=dv"
	}

	log.Info("Trying to download " + url)
	req, err := http.NewRequest("GET", url, body)
	if err != nil {
		return err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	//open a file for writing
	file, err := os.Create("/tmp/gphotos/" + gphoto.Filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, res.Body)
	if err != nil {
		return err
	}

	return nil
}

func ToString(a interface{}) string {
	out, err := json.Marshal(a)
	if err != nil {
		return "ERROR CONVERTING"
	}

	return string(out)
}

// // Upload sends the media and returns the UploadToken.
// func (c *GPhotos) Upload(r io.Reader, filename string) (string, error) {
// 	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/uploads", basePath, apiVersion), r)
// 	if err != nil {
// 		return "", err
// 	}
// 	req.Header.Add("X-Goog-Upload-File-Name", filename)

// 	res, err := c.client.Do(req)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer res.Body.Close()

// 	b, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return "", err
// 	}
// 	uploadToken := string(b)
// 	return uploadToken, nil
// }
