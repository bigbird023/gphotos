package main

import (
	"context"
	"io/ioutil"
	"time"

	gphotos "github.com/bigbird023/gphotos/gphotosclient"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	SetupViper()

	creds, err := ioutil.ReadFile(viper.GetString("CredFile"))
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	gphotoTransferFromClient := NewTransferFrom(creds)
	//gphotoTransferToClient := NewTransferTo(creds)

	ctx := context.Background()

	currentConfigDate := viper.GetString("CurrentDate")
	currentDate := StringToDate(currentConfigDate)
	configYear := currentDate.Year()

	search := gphotos.GphotoSearch{}
	search.Filters.DateFilter.Dates = append(search.Filters.DateFilter.Dates, StringToGphotoDate(currentConfigDate))

	for configYear == currentDate.Year() {
		log.Info("Downloading Photos for day " + DateToString(currentDate))
		photos, err := gphotoTransferFromClient.GetPagedLibraryContents(ctx, &search, "")
		if err != nil {
			log.Fatalf("Unable to get photo library: %v", err)
		}
		processPhotos(ctx, photos, gphotoTransferFromClient)

		for photos.NextPageToken != "" {
			log.Info("Next Page = " + photos.NextPageToken)
			photos, err = gphotoTransferFromClient.GetPagedLibraryContents(ctx, nil, photos.NextPageToken)
			if err != nil {
				log.Fatalf("Unable to get next photo library: %v", err)
			}
			processPhotos(ctx, photos, gphotoTransferFromClient)
		}
		currentDate = nextDay(currentDate)
	}
}

func processPhotos(ctx context.Context, photos *gphotos.GPhotos, gphotoTransferFromClient *gphotos.GPhotosClient) {
	log.Info("Processing " + string(len(photos.MediaItems)) + " MediaItems")
	for _, curPhoto := range photos.MediaItems {
		err := gphotoTransferFromClient.DownloadMedia(ctx, curPhoto)
		if err != nil {
			log.Error("Error Downloading", err)
		} else {
			//_, err = gphotoTransferToClient.UploadMedia(ctx, curPhoto)
			//if err != nil {
			//	log.Error("Error Uploading ", err)
			//}
		}

	}
}

func nextDay(curDate time.Time) time.Time {
	curDate = curDate.AddDate(0, 0, -1)
	viper.Set("CurrentDate", DateToString(curDate))
	viper.WriteConfig()
	return curDate
}
