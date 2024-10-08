package Megascans

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const acquiredAssetsEndpoint = "https://quixel.com/v1/assets/acquired"

type User struct {
	AuthenticationToken     string
	email                   string
	previouslyAuthenticated bool
}

// AcquiredAsset Any asset you have purchased
type AcquiredAsset struct {
	// This is the uuid of the asset
	AssetID string `json:"assetID"`
	// This is NOT the resolution available for an asset, this is essentially an 8k license.
	Resolution int    `json:"resolution"`
	ExrAccess  string `json:"exrAccess"`
}

// AcquiredAssets List of purchased assets
type AcquiredAssets []AcquiredAsset

// DownloadManifest is used to get a download link
type DownloadManifest struct {
	Asset  string `json:"asset"`
	Config struct {
		Highpoly        bool   `json:"highpoly"`
		Ztool           bool   `json:"ztool"`
		LowerlodNormals bool   `json:"lowerlod_normals"`
		AlbedoLods      bool   `json:"albedo_lods"`
		MeshMimeType    string `json:"meshMimeType"`
		Brushes         bool   `json:"brushes"`
	} `json:"config"`
}

// There is a ton of info in the payload,
// But the only useful thing is the payload id
// To request a new download

func NewDownloadManifest(asset AcquiredAsset) *DownloadManifest {
	payload := DownloadManifest{}
	payload.Asset = asset.AssetID
	payload.Config.Highpoly = true
	payload.Config.Ztool = true
	payload.Config.AlbedoLods = true
	payload.Config.MeshMimeType = "application/x-fbx"
	payload.Config.Brushes = true
	return &payload
}

func NewUser() (*User, error) {
	email, err := EmailPrompt()
	if err != nil {
		return nil, err
	}
	token, err := AuthTokenPrompt()
	if err != nil {
		return nil, err
	}

	return &User{
		AuthenticationToken:     *token,
		email:                   *email,
		previouslyAuthenticated: false,
	}, nil
}

func (u *User) SayHello() {
	fmt.Println("HELLOOOOO")
}
func (u *User) TestEmail() {
	x := fmt.Sprintf("https://accounts.quixel.com/api/v1/users/%s", u.email)
	fmt.Println(x)
}

func (u *User) Authenticate() error {
	time.Sleep(1 * time.Second)
	fmt.Println("Authenticating...")
	client := &http.Client{}

	base := "https://accounts.quixel.com/api/v1/users/%s"
	email := u.email

	baseUrl, err := url.Parse(fmt.Sprintf(base, u.email))
	if err != nil {
		fmt.Println("Error parsing user url")
		return err
	}

	query := baseUrl.Query()
	query.Set("email", email)
	baseUrl.RawQuery = query.Encode()

	finalUrl := baseUrl.String()

	req, err := http.NewRequest("GET", finalUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+u.AuthenticationToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request for authentication")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		fmt.Println(resp.Status)
		if u.previouslyAuthenticated {
			email, err := EmailPrompt()
			if err != nil {
				return err
			}
			u.email = *email
		} else {
			token, err := AuthTokenPrompt()
			if err != nil {
				return err
			}
			u.AuthenticationToken = *token
			email, err := EmailPrompt()
			if err != nil {
				return err
			}
			u.email = *email
		}
		// Verify newly provided credentials
		err = u.Authenticate()
		if err != nil {
			return err
		}
	}
	fmt.Println("Authentication Successful")
	u.previouslyAuthenticated = true
	return nil
}

func EmailPrompt() (*string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please enter your email address associated with your Megascans account:")
	email, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	email = strings.TrimSpace(email)
	return &email, nil
}

func AuthTokenPrompt() (*string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter an authentication token Bearer:")
	token, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	token = strings.TrimSpace(token)
	return &token, nil
}

func (u *User) GetAcquiredAssets() (*AcquiredAssets, error) {
	time.Sleep(1 * time.Second)
	fmt.Println("Retrieving Acquired Assets")
	client := &http.Client{}
	err := u.Authenticate()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", acquiredAssetsEndpoint, nil)
	if err != nil {
		fmt.Println("Error making request for acquired assets")
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+u.AuthenticationToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	var acquiredAssets AcquiredAssets
	err = json.Unmarshal(body, &acquiredAssets)
	if err != nil {
		return nil, err
	}
	if len(acquiredAssets) == 0 {
		return nil, errors.New("No acquired assets were found")
	}
	return &acquiredAssets, nil
}

func (u *User) DownloadAsset(asset AcquiredAsset, downloadsFolder string) (*string, error) {
	time.Sleep(1 * time.Second)
	fmt.Println("Initiating download")
	client := &http.Client{}
	payload := NewDownloadManifest(asset)
	jsonPayload, err := json.MarshalIndent(*payload, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling payload")
		return nil, err
	}
	req, err := http.NewRequest(
		"POST",
		"https://quixel.com/v1/downloads",
		bytes.NewReader(jsonPayload),
	)
	if err != nil {
		fmt.Println("Error initiating request for download payload")
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+u.AuthenticationToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		log.Fatal("UNAUTHORIZED: Most likely your token just expired, "+
			"so update that and then rerun the application ", resp.StatusCode, resp.Status)
	}
	if resp == nil {
		return nil, errors.New("No download payload found, most likely quixel deleted this asset.")
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error making request for download payload %v\n", resp.Status)
		if resp.StatusCode == http.StatusBadRequest {
			fmt.Printf(string(jsonPayload))
		}
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response from download payload")
		return nil, err
	}
	// If megascans returns a nil body, it must not be dereferenced.
	if resp.Body != nil {
		resp.Body.Close()
	}
	type Payload struct {
		Id string `json:"id"`
	}
	var p Payload
	err = json.Unmarshal(body, &p)
	if err != nil {
		fmt.Println("Error unmarshalling response from download payload")
		return nil, err
	}
	zipData, err := requestDownload(u, p.Id)
	if err != nil {
		fmt.Println("Error downloading asset")
		return nil, err
	}
	fp := filepath.Join(downloadsFolder, asset.AssetID+".zip")
	err = saveDownloadToDisk(*zipData, fp)
	if err != nil {
		return nil, err
	}
	return &fp, nil
}

func requestDownload(u *User, downloadId string) (*[]byte, error) {
	time.Sleep(1 * time.Second)
	fmt.Println("Requesting Download")
	client := &http.Client{}
	downloadURL := fmt.Sprintf(
		"https://assetdownloads.quixel.com/download/"+
			"%s"+
			"?preserveStructure=true&url=https%%3A%%2F%%2Fquixel.com%%2Fv1%%2Fdownloads",
		downloadId,
	)
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, err
	}
	// This accept encoding type might help fix eof errors
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Authorization", u.AuthenticationToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Error making request to download asset" + resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Println("Error reading download body")
		return nil, err
	}
	if len(body) == 0 {
		return nil, errors.New("Download body is empty")
	}
	return &body, nil
}

func saveDownloadToDisk(data []byte, filepath string) error {
	fmt.Println("Saving download to disk...")
	outFile, err := os.Create(filepath)
	defer outFile.Close()
	if err != nil {
		fmt.Println("Error creating download file")
		return err
	}
	bytesWritten, err := outFile.Write(data)
	if err != nil {
		fmt.Println("Error writing download file")
		return err
	}
	if bytesWritten != len(data) {
		fmt.Println("Download did not complete entirely.")
		return errors.New("Download did not complete")
	}
	fmt.Println("Download saved successfully!")
	return nil
}
