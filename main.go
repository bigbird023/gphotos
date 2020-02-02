package main

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	gphotos "github.com/bigbird023/gphotos/gphotosclient"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go watchForSignal(ctx, cancel)

	SetupViper()
	creds, err := ioutil.ReadFile(viper.GetString("CredFile"))
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	gphotoTransferFromClient := NewTransferFrom(creds)
	//gphotoTransferToClient := NewTransferTo(creds)

	downloadChannel := make(chan gphotos.GPhoto, 1)

	threads := viper.GetInt("Threads")
	if threads < 1 {
		log.Fatal("Threads must be greater than 0")
	}
	wg := sync.WaitGroup{}
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go processDownloads(ctx, &wg, downloadChannel, gphotoTransferFromClient)
	}

	currentConfigDate := viper.GetString("CurrentDate")
	currentDate := StringToDate(currentConfigDate)
	configYear := currentDate.Year()

	search := gphotos.GphotoSearch{}

	for configYear == currentDate.Year() {

		select {
		case <-time.After(10 * time.Millisecond):

			currentConfigDate = DateToString(currentDate)
			search.Filters.DateFilter.Dates = search.Filters.DateFilter.Dates[:0]
			search.Filters.DateFilter.Dates = append(search.Filters.DateFilter.Dates, StringToGphotoDate(currentConfigDate))

			log.Info("Downloading Photos for day " + DateToString(currentDate))
			photos, err := gphotoTransferFromClient.GetPagedLibraryContents(ctx, &search, "")
			if err != nil {
				log.Fatalf("Unable to get photo library: %v", err)
			}
			processPhotos(ctx, downloadChannel, photos, gphotoTransferFromClient)

			for photos.NextPageToken != "" {
				photos, err = gphotoTransferFromClient.GetPagedLibraryContents(ctx, nil, photos.NextPageToken)
				if err != nil {
					log.Fatalf("Unable to get next photo library: %v", err)
				}
				processPhotos(ctx, downloadChannel, photos, gphotoTransferFromClient)
			}
			currentDate = nextDay(currentDate)
		case <-ctx.Done():
			configYear = 0
		}
	}
	cancel()
	wg.Wait()
	close(downloadChannel)
}

func watchForSignal(ctx context.Context, cancel func()) {
	sigint := make(chan os.Signal, 1)

	// interrupt signal sent from terminal
	signal.Notify(sigint, os.Interrupt)
	// sigterm signal sent from kubernetes
	signal.Notify(sigint, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case sig := <-sigint:
		log.WithContext(ctx).WithField("OS Signal", sig).Info("OS Signal Received")
		cancel()
	}

}

func processPhotos(ctx context.Context, downloadChannel chan gphotos.GPhoto, photos *gphotos.GPhotos, gphotoTransferFromClient *gphotos.GPhotosClient) {
	log.Info("Processing MediaItems ", len(photos.MediaItems))
	for _, curPhoto := range photos.MediaItems {
		select {
		case downloadChannel <- curPhoto:
		case <-ctx.Done():
		}
	}
}

func processDownloads(ctx context.Context, wg *sync.WaitGroup, downloadChannel chan gphotos.GPhoto, gphotoTransferFromClient *gphotos.GPhotosClient) {
foreverloop:
	for {
		select {
		case curPhoto := <-downloadChannel:

			err := gphotoTransferFromClient.DownloadMedia(ctx, curPhoto)
			if err != nil {
				log.Error("Error Downloading", err)
			} else {
				//_, err = gphotoTransferToClient.UploadMedia(ctx, curPhoto)
				//if err != nil {
				//	log.Error("Error Uploading ", err)
				//}
			}
		case <-time.After(time.Duration(60) * time.Second):
			log.Info("No Downloads Received after a 60 Seconds Timeout")

		case <-ctx.Done():
			log.Info("Context Sent Done Signal")
			break foreverloop
		}

	}
	wg.Done()
}

func nextDay(curDate time.Time) time.Time {
	curDate = curDate.AddDate(0, 0, -1)
	viper.Set("CurrentDate", DateToString(curDate))
	viper.WriteConfig()
	return curDate
}
