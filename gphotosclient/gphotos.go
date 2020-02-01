package gphotos

import "encoding/json"

//GPhotos list of photos and other meta data
type GPhotos struct {
	MediaItems    []GPhoto `json:"mediaItems"`
	NextPageToken string   `json:"nextPageToken"`
}

//GPhoto google photo details
type GPhoto struct {
	ID            string         `json:"id"`
	ProductURL    string         `json:"productUrl`
	BaseURL       string         `json:"BaseUrl"`
	MimeType      string         `json:"MimeType"`
	MediaMetaData gMediaMetaData `json:"mediaMetadata,omitempty"`
	Filename      string         `json:"filename"`
}

type gMediaMetaData struct {
	CreationTime string     `json:"CreationTime"`
	Width        string     `json:"Width"`
	Height       string     `json:"Height"`
	Photo        gMetaPhoto `json:"Photo,omitempty"`
	Video        gMetaVideo `json:"Video,omitempty"`
}

type gMetaPhoto struct {
	CameraModel     string      `json:"cameraModel,omitempty"`
	FocalLength     json.Number `json:"FocalLength,omitempty"`
	ApertureFNumber json.Number `json:"apertureFNumber,omitempty"`
	IsoEquivalent   json.Number `json:"isoEquivalent,omitempty"`
	ExposureTime    json.Number `json:"exposureTime,omitempty"`
}

type gMetaVideo struct {
	Fps    json.Number `json:"Fps,omitempty"`
	Status json.Number `json:"Status,omitempty"`
}
