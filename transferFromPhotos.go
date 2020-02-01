package main

import (
	gphotos "github.com/bigbird023/gphotos/gphotosclient"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
)

//NewTransferFrom - setups google photo client with approval
func NewTransferFrom(credJSON []byte) *gphotos.GPhotosClient {
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(credJSON, "https://www.googleapis.com/auth/photoslibrary.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	oauthClient := NewOauthClient(config, viper.GetString("TransferFromTokenFile"))

	return gphotos.NewGPhotos(oauthClient)
}
