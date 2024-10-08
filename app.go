package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	Megascans "megascansDownloader/megascans"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	DownloadsFolder     string                   `json:"downloadsFolder"`
	SuccessfulDownloads Megascans.AcquiredAssets `json:"SuccessfulDownloads"`
}

func getConfigPath() string {
	executable, err := os.Executable()
	if err != nil {
		fmt.Println("Could not determine executable path for this application")
		log.Fatal(err)
	}
	const configLocation = "downloadedContent.json"
	configFilePath := filepath.Join(filepath.Dir(executable), configLocation)
	return configFilePath
}

func textPrompt(promptText string) (*string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(promptText)
	response, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	response = strings.TrimSpace(response)
	return &response, nil
}

func getBaseConfig() (*Config, error) {
	downloadsFolder, err := textPrompt("Enter a downloads folder: ")
	if err != nil {
		fmt.Println("Downloads folder invalid")
		return nil, err
	}
	config := Config{}
	config.DownloadsFolder = *downloadsFolder
	config.SuccessfulDownloads = Megascans.AcquiredAssets{}
	return &config, nil
}

func createDownloadConfigFile() error {
	baseConfig, err := getBaseConfig()
	if err != nil {
		fmt.Println("Failed to get base download config")
		return err
	}
	baseConfigJSON, err := json.MarshalIndent(*baseConfig, "", "    ")
	if err != nil {
		fmt.Println("Failed to marshal base download config json")
		return err
	}
	downloadConfig, err := os.Create(getConfigPath())
	defer downloadConfig.Close()
	if err != nil {
		fmt.Println("Error creating download config file")
		return err
	}
	_, err = downloadConfig.Write(baseConfigJSON)
	if err != nil {
		fmt.Println("Error writing initial config to json file")
		return err
	}
	return nil
}
func getDownloadConfigFileData() (*Config, error) {
	downloadConfig, err := os.ReadFile(getConfigPath())
	if os.IsNotExist(err) {
		err := createDownloadConfigFile()
		if err != nil {
			return nil, err
		}
		downloadConfig, err = os.ReadFile(getConfigPath())
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		fmt.Println("Error opening download config file")
		return nil, err
	}
	var downloadConfigStruct Config
	err = json.Unmarshal(downloadConfig, &downloadConfigStruct)
	if err != nil {
		fmt.Printf("Error unmarshalling download config file %v \n", err)
		return nil, err
	}

	return &downloadConfigStruct, nil
}

func addAssetToSuccessfulDownloads(asset Megascans.AcquiredAsset) error {
	configData, err := getDownloadConfigFileData()
	if err != nil {
		return err
	}
	if len(configData.SuccessfulDownloads) > 0 {
		configData.SuccessfulDownloads = append(configData.SuccessfulDownloads, asset)
	} else {
		configData.SuccessfulDownloads = []Megascans.AcquiredAsset{asset}
	}
	configDataJSON, err := json.MarshalIndent(configData, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling json for config file")
		return err
	}
	err = ioutil.WriteFile(getConfigPath(), configDataJSON, 0644)
	if err != nil {
		fmt.Println("Error writing json to file")
	}
	return nil
}

func getDownloadedAssets() ([]Megascans.AcquiredAsset, error) {
	downloadConfig, err := getDownloadConfigFileData()
	if err != nil {
		return nil, err
	}
	return downloadConfig.SuccessfulDownloads, nil
}

func isAlreadyDownloaded(asset Megascans.AcquiredAsset) (bool, error) {
	downloadedAssets, err := getDownloadedAssets()
	if len(downloadedAssets) < 1 {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, downloadedAsset := range downloadedAssets {
		if asset.AssetID == downloadedAsset.AssetID {
			return true, nil
		}
	}
	return false, nil
}

func main() {
	//Just using this to verify config data
	configData, err := getDownloadConfigFileData()
	if err != nil {
		fmt.Printf("Error getting download config data %v", err)
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	user, err := Megascans.NewUser()
	if err != nil {
		log.Fatal(err)
	}

	user.TestEmail()

	err = user.Authenticate()
	if err != nil {
		log.Fatal(err)
	}

	acquiredAssets, err := user.GetAcquiredAssets()
	if err != nil {
		log.Fatal()
	}

	var needsToBeDownloaded Megascans.AcquiredAssets
	for _, asset := range *acquiredAssets {
		assetIsDownloaded, err := isAlreadyDownloaded(asset)
		if err != nil {
			log.Fatal(err)
		}
		if assetIsDownloaded {
			continue
		}
		needsToBeDownloaded = append(needsToBeDownloaded, asset)
	}
	fmt.Printf("\nDOWNLOADING %v Assets\n", len(needsToBeDownloaded))
	for i, asset := range needsToBeDownloaded {
		fmt.Printf("--------%v--------\n", asset.AssetID)
		fmt.Printf("Downloading asset %v of %v\n", i+1, len(needsToBeDownloaded))
		downloadedFile, err := user.DownloadAsset(asset, configData.DownloadsFolder)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("SUCCESSFULLY DOWNLOADED: ", *downloadedFile)
		fmt.Println("Writing to successful downloads")
		err = addAssetToSuccessfulDownloads(asset)
		if err != nil {
			fmt.Println("Error adding asset to successful downloads")
			log.Fatal(err)
		}
	}
}
