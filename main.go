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
	//gphotoTransferToClient := NewTransferTo(creds)

	ctx := context.Background()

	photos, err := gphotoTransferFromClient.GetPagedLibraryContents(ctx, "")
	if err != nil {
		log.Fatalf("Unable to get photo library: %v", err)
	}

	for photos.NextPageToken != "" {
		for _, curPhoto := range photos.MediaItems {
			err = gphotoTransferFromClient.DownloadMedia(ctx, curPhoto)
			if err != nil {
				log.Error("Error Downloading", err)
			} else {
				//_, err = gphotoTransferToClient.UploadMedia(ctx, curPhoto)
				//if err != nil {
				//	log.Error("Error Uploading ", err)
				//}
			}

		}
		log.Info("Next Page = " + photos.NextPageToken)
		photos, err = gphotoTransferFromClient.GetPagedLibraryContents(ctx, photos.NextPageToken)
		if err != nil {
			log.Fatalf("Unable to get next photo library: %v", err)
		}
	}
}
