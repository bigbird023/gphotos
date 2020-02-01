package gphotos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
)

const apiVersion = "v1"
const basePath = "https://photoslibrary.googleapis.com/"

//GPhotosClient is a client for interacting with google photos api.
type GPhotosClient struct {
	client *http.Client
}

//GphotoSearch search body
type GphotoSearch struct {
	AlbumID   string        `json:"albumId,omitempty"`
	PageSize  string        `json:"pageSize,omitempty"`
	PageToken string        `json:"pageToken,omitempty"`
	Filters   GphotoFilters `json:"filters,omitempty"`
}

//GphotoFilters filters test
type GphotoFilters struct {
	DateFilter               GphotoDateFilter `json:"dateFilter,omitempty"`
	ContentFilter            interface{}      `json:"contentFilter,omitempty"`
	MediaTypeFilter          interface{}      `json:"mediaTypeFilter,omitempty"`
	FeatureFilter            interface{}      `json:"featureFilter,omitempty"`
	IncludeArchivedMedia     bool             `json:"includeArchivedMedia,omitempty"`
	ExcludeNonAppCreatedData bool             `json:"excludeNonAppCreatedData,omitempty"`
}

//GphotoDateFilter date filter
type GphotoDateFilter struct {
	Dates  []GphotoDate `json:"dates,omitempty"`
	Ranges []string     `json:"ranges,omitempty"`
}

//GphotoDate date format for searching
type GphotoDate struct {
	Day   int `json:"day"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

// NewGPhotos creates a new client.
func NewGPhotos(client *http.Client) *GPhotosClient {
	return &GPhotosClient{client}
}

//GetPagedLibraryContents - todo
func (c *GPhotosClient) GetPagedLibraryContents(ctx context.Context, search *GphotoSearch, nextPage string) (*GPhotos, error) {
	var body io.Reader = nil
	var req *http.Request = nil
	var err error = nil
	if nextPage != "" {
		log.Debug("API for nextPage " + nextPage)
		req, err = http.NewRequest("GET", fmt.Sprintf("%s/%s/mediaItems?pageToken=%s", basePath, apiVersion, nextPage), body)
	} else if search != nil {
		reqBody, err := json.Marshal(&search)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/%s/mediaItems:search", basePath, apiVersion), bytes.NewBuffer(reqBody))
		log.Debug("Search Request Body " + string(reqBody))
	} else {
		log.Debug("API for Get mediaItems " + nextPage)
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

	if res.StatusCode != 200 {
		log.Debug("Status Code Response " + res.Status)
		return nil, errors.New(res.Status)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	gphotos := GPhotos{}
	err = json.Unmarshal(b, &gphotos)
	if err != nil {
		return nil, err
	}
	d, err := json.Marshal(gphotos)
	if err != nil {
		return nil, err
	}
	log.Debug("Gphotos results " + string(d))
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

	log.Info("Trying to download " + gphoto.Filename)
	req, err := http.NewRequest("GET", url, body)
	if err != nil {
		return err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	dt := strings.Split(gphoto.MediaMetaData.CreationTime, "T")
	d := dt[0]
	ymd := strings.Split(d, "-")
	yr := ymd[0]
	mn := ymd[1]
	dy := ymd[2]
	//open a file for writing
	curPath := viper.GetString("DownloadPath") + yr + "/" + mn + "/" + dy + "/"
	os.MkdirAll(curPath, os.ModePerm)
	file, err := os.Create(curPath + gphoto.Filename)
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

//UploadMedia - todo
func (c *GPhotosClient) UploadMedia(ctx context.Context, gphoto GPhoto) (string, error) {
	filename := "/tmp/gphotos/" + gphoto.Filename
	log.Info("Filename to upload ", filename)
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	log.Info("Trying to upload " + gphoto.Filename)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/uploads", basePath, apiVersion), file)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-Goog-Upload-File-Name", filename)

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	//open a file for writing
	file, err = os.Create("/tmp/gphotos/" + gphoto.Filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	uploadToken := string(b)
	return uploadToken, nil
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
