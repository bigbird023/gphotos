package main

import (
	"fmt"
	"path"
	"time"

	gphotos "github.com/bigbird023/gphotos/gphotosclient"
	"github.com/spf13/viper"
)

const appName = "gphotos"
const configFileName = "gphotos"

//SetupViper function to configure the configuration
func SetupViper() {
	viper.SetConfigName(configFileName)            // name of config file (without extension)
	viper.SetConfigType("yaml")                    // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/" + appName + "/")   // path to look for the config file in
	viper.AddConfigPath("$HOME/." + appName + "/") // call multiple times to add many search paths
	viper.AddConfigPath(".")                       // optionally look for config in the working directory
	err := viper.ReadInConfig()                    // Find and read the config file
	if err != nil {                                // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	viper.SetDefault("Threads", 1)
	viper.SetDefault("CurrentDate", DateToString(time.Now()))
	viper.SetDefault("CredFile", path.Dir(viper.ConfigFileUsed())+"/credentials.json")
	viper.SetDefault("TransferFromTokenFile", path.Dir(viper.ConfigFileUsed())+"/transferFromToken.json")
	viper.SetDefault("TransferToTokenFile", path.Dir(viper.ConfigFileUsed())+"/transferToToken.json")
	viper.SetDefault("DownloadPath", "/tmp/gphotos/")
}

//DateToString will convert the dateTime to string format
func DateToString(datetime time.Time) string {
	return datetime.Format("2006-01-02T00:00:00.000Z")
}

//StringToDate will convert at rest string into date (from gphotos
func StringToDate(datetime string) time.Time {
	layout := "2006-01-02T15:04:05.000Z"
	t, err := time.Parse(layout, datetime)

	if err != nil {
		fmt.Println(err)
	}

	return t
}

//StringToGphotoDate converter for string to gphotodate
func StringToGphotoDate(datetime string) gphotos.GphotoDate {
	t := StringToDate(datetime)
	d := gphotos.GphotoDate{}
	d.Day = t.Day()
	m := t.Month()
	d.Month = int(m)
	d.Year = t.Year()
	return d
}
