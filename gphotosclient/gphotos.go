package gphotos

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
	MediaMetaData gMediaMetaData `json:"mediaMetadata"`
	Filename      string         `json:"filename"`
}

type gMediaMetaData struct {
	CreationTime string     `json:"CreationTime"`
	Width        string     `json:"Width"`
	Height       string     `json:"Height"`
	Photo        gMetaPhoto `json:"Photo"`
	Video        gMetaVideo `json:"Video"`
}

type gMetaPhoto struct {
	CameraModel     float64 `json:"cameraModel,omitempty"`
	FocalLength     string  `json:"FocalLength"`
	ApertureFNumber string  `json:"apertureFNumber"`
	IsoEquivalent   string  `json:"isoEquivalent"`
	ExposureTime    string  `json:"exposureTime"`
}

type gMetaVideo struct {
	Fps    float64 `json:"Fps"`
	Status string  `json:"Status"`
}
