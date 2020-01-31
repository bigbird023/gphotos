package main

import (
	"context"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

func main() {
	creds, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	gphotoTransferFromClient := NewTransferFrom(creds)

	ctx := context.Background()

	photos, err := gphotoTransferFromClient.GetPagedLibraryContents(ctx, "")
	if err != nil {
		log.Fatalf("Unable to get photo library: %v", err)
	}

	//for photos.NextPageToken != "" {
	for _, curPhoto := range photos.MediaItems {
		log.Info("Downloading Photo " + curPhoto.Filename)
		gphotoTransferFromClient.DownloadMedia(ctx, curPhoto)
	}
	log.Info("Next Page = " + photos.NextPageToken)
	// 	photos, err = gphotoTransferFromClient.GetPagedLibraryContents(ctx, photos.NextPageToken)
	// 	if err != nil {
	// 		log.Fatalf("Unable to get next photo library: %v", err)
	// 	}
	// }
}

// func upload(filepath string, helper *GPhotos, photos *photoslibrary.Service) {
// 	filename := path.Base(filepath)
// 	log.Printf("Uploading %s", filename)
// 	file, err := os.Open(filepath)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer file.Close()
// 	uploadToken, err := helper.Upload(file, filename)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Printf("Uploaded %s as token %s", filename, uploadToken)

// 	log.Printf("Adding media %s", filename)
// 	batch, err := photos.MediaItems.BatchCreate(&photoslibrary.BatchCreateMediaItemsRequest{
// 		NewMediaItems: []*photoslibrary.NewMediaItem{
// 			&photoslibrary.NewMediaItem{
// 				Description:     filename,
// 				SimpleMediaItem: &photoslibrary.SimpleMediaItem{UploadToken: uploadToken},
// 			},
// 		},
// 	}).Do()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	for _, result := range batch.NewMediaItemResults {
// 		log.Printf("Added media %s as %s", result.MediaItem.Description, result.MediaItem.Id)
// 	}
// }
