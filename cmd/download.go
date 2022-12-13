package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	gphotosclient "github.com/bigbird023/gphotos/gphotosclient"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downloadCmd)
}

const appName = "gphotos"
const configFileName = "gphotos"

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "download photos based on configuration",
	Long:  `download photos based on configuration`,
	Run: func(cmd *cobra.Command, args []string) {

		ctx, cancel := context.WithCancel(context.Background())
		go watchForSignal(ctx, cancel)

		setupViper()
		creds, err := os.ReadFile(viper.GetString("CredFile"))
		if err != nil {
			log.Fatalf("Unable to read client secret file: %v", err)
		}

		gphotoTransferFromClient := newTransferFrom(creds)
		//gphotoTransferToClient := NewTransferTo(creds)

		downloadChannel := make(chan gphotosclient.GPhoto, 1)

		threads := viper.GetInt("Threads")
		wg := sync.WaitGroup{}
		if threads < 2 {
			//no threads
		} else {
			for i := 0; i < threads; i++ {
				wg.Add(1)
				go processDownloads(ctx, &wg, downloadChannel, gphotoTransferFromClient)
			}
		}

		configStopYear := viper.GetInt("configStopYear")
		currentConfigDate := viper.GetString("CurrentDate")
		currentDate := stringToDate(currentConfigDate)
		configYear := currentDate.Year()

		search := gphotosclient.Search{}

		for configYear == currentDate.Year() {

			select {
			case <-time.After(10 * time.Millisecond):

				currentConfigDate = dateToString(currentDate)
				search.PageToken = ""
				search.Filters.DateFilter.Dates = search.Filters.DateFilter.Dates[:0]
				search.Filters.DateFilter.Dates = append(search.Filters.DateFilter.Dates, stringToGphotoDate(currentConfigDate))

				log.Info("Downloading Photos for day " + dateToString(currentDate))
				photos, err := gphotoTransferFromClient.GetPagedLibraryContents(ctx, &search, "")
				if err != nil {
					log.Fatalf("Unable to get photo library: %v", err)
				}
				processPhotos(ctx, downloadChannel, photos, gphotoTransferFromClient)

				for photos.NextPageToken != "" {
					search.PageToken = photos.NextPageToken
					photos, err = gphotoTransferFromClient.GetPagedLibraryContents(ctx, &search, "")
					if err != nil {
						log.Fatalf("Unable to get next photo library: %v", err)
					}
					processPhotos(ctx, downloadChannel, photos, gphotoTransferFromClient)
				}
				currentDate = nextDay(currentDate)
				if configYear != currentDate.Year() {
					if configStopYear < currentDate.Year() {
						configYear = currentDate.Year()
					}

				}
			case <-ctx.Done():
				configYear = 0
			}
		}
		cancel()
		wg.Wait()
		close(downloadChannel)
	},
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

func processPhotos(ctx context.Context, downloadChannel chan gphotosclient.GPhoto, photos *gphotosclient.GPhotos, gphotoTransferFromClient *gphotosclient.Client) {
	log.Info("Processing MediaItems ", len(photos.MediaItems))
	for _, curPhoto := range photos.MediaItems {

		if viper.GetInt("Threads") < 2 {
			err := gphotoTransferFromClient.DownloadMedia(ctx, curPhoto)
			if err != nil {
				log.Error("Error Downloading", err)
			}
		} else {
			select {
			case downloadChannel <- curPhoto:
			case <-ctx.Done():
			}
		}
	}
}

func processDownloads(ctx context.Context, wg *sync.WaitGroup, downloadChannel chan gphotosclient.GPhoto, gphotoTransferFromClient *gphotosclient.Client) {
foreverloop:
	for {
		select {
		case curPhoto := <-downloadChannel:

			err := gphotoTransferFromClient.DownloadMedia(ctx, curPhoto)
			if err != nil {
				log.Error("Error Downloading", err)
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
	viper.Set("CurrentDate", dateToString(curDate))
	viper.WriteConfig()
	return curDate
}

// SetupViper function to configure the configuration
func setupViper() {
	viper.SetConfigName(configFileName)            // name of config file (without extension)
	viper.SetConfigType("yaml")                    // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/" + appName + "/")   // path to look for the config file in
	viper.AddConfigPath("$HOME/." + appName + "/") // call multiple times to add many search paths
	viper.AddConfigPath(".")                       // optionally look for config in the working directory
	err := viper.ReadInConfig()                    // Find and read the config file
	if err != nil {                                // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	viper.SetDefault("Threads", 1)
	viper.SetDefault("CurrentDate", dateToString(time.Now()))
	viper.SetDefault("configStopYear", dateToString(time.Now().AddDate(-1, 0, 0)))
	viper.SetDefault("CredFile", path.Dir(viper.ConfigFileUsed())+"/credentials.json")
	viper.SetDefault("TransferFromTokenFile", path.Dir(viper.ConfigFileUsed())+"/transferFromToken.json")
	viper.SetDefault("TransferToTokenFile", path.Dir(viper.ConfigFileUsed())+"/transferToToken.json")
	viper.SetDefault("DownloadPath", "/tmp/gphotos/")
}

// DateToString will convert the dateTime to string format
func dateToString(datetime time.Time) string {
	return datetime.Format("2006-01-02T00:00:00.000Z")
}

// StringToDate will convert at rest string into date (from gphotos
func stringToDate(datetime string) time.Time {
	layout := "2006-01-02T15:04:05.000Z"
	t, err := time.Parse(layout, datetime)

	if err != nil {
		fmt.Println(err)
	}

	return t
}

// StringToGphotoDate converter for string to gphotodate
func stringToGphotoDate(datetime string) gphotosclient.GphotoDate {
	t := stringToDate(datetime)
	d := gphotosclient.GphotoDate{}
	d.Day = t.Day()
	m := t.Month()
	d.Month = int(m)
	d.Year = t.Year()
	return d
}

func newTransferFrom(credJSON []byte) *gphotosclient.Client {
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(credJSON, "https://www.googleapis.com/auth/photoslibrary.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	oauthClient := newOauthClient(config, viper.GetString("TransferFromTokenFile"))

	return gphotosclient.NewGPhotos(oauthClient)
}

// func newTransferTo(credJSON []byte) *gphotosclient.Client {
// 	// If modifying these scopes, delete your previously saved token.json.
// 	config, err := google.ConfigFromJSON(credJSON, "https://www.googleapis.com/auth/photoslibrary")
// 	if err != nil {
// 		log.Fatalf("Unable to parse client secret file to config: %v", err)
// 	}
// 	oauthClient := newOauthClient(config, viper.GetString("TransferToTokenFile"))

// 	return gphotosclient.NewGPhotos(oauthClient)
// }

func newOauthClient(config *oauth2.Config, localTokenFile string) *http.Client {
	tok, err := tokenFromFile(localTokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(localTokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Requests a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache OAuth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
