package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

// Houses is a slice of house structs
type Houses struct {
	Houses []House `json:"houses"`
}

// House represents a single home listing
type House struct {
	ID        int    `json:"id"`
	Address   string `json:"address"`
	Homeowner string `json:"homeowner"`
	Price     int    `json:"price"`
	PhotoURL  string `json:"photoURL`
}

func main() {
	baseURL := "http://app-homevision-staging.herokuapp.com/api_project/houses?page=%d"
	var houseList []Houses

	for i := 0; i < 3; i++ {
		houses, err := getHouses(fmt.Sprintf(baseURL, i+1))
		if err != nil {
			fmt.Println(err)
		} else {
			houseList = append(houseList, houses)
		}
	}

	var wg sync.WaitGroup
	fmt.Println(len(houseList))
	wg.Add(len(houseList))
	// Creating file name
	for _, h := range houseList {
		for _, hr := range h.Houses {
			go downloadFile(hr)
		}
	}
	wg.Wait()

	return

}

// err := json.Unmarshal(resp.Body, &house)
// houses = append(houses, house)

// Get response from API
// Unmarshal into House Struct
// create file name from address & url
// save file
// concurrency in saving file
// deal with flaky responses

func createFileName(h House) string {
	//id-[NNN]-[address].[ext]
	id := h.ID
	urlSplit := strings.Split(h.PhotoURL, ".")
	ext := urlSplit[len(urlSplit)-1]
	name := strings.ReplaceAll(h.Homeowner, " ", "-")
	address := strings.ReplaceAll(h.Address, ",", "")
	address = strings.ReplaceAll(address, ".", "")
	address = strings.ReplaceAll(address, " ", "-")

	return fmt.Sprintf("%d-%s-%s.%s", id, name, address, ext)

}

func saveFile(fileName string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("getting file contents: %w", err)
	}
	defer resp.Body.Close()
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("creating new file: %w", err)
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("copying file: %w", err)
	}
	return nil
}

func downloadFile(h House) error {
	fileName := createFileName(h)
	err := saveFile(fileName, h.PhotoURL)
	if err != nil {
		return err
	}
	return nil
}

func getHouses(url string) (Houses, error) {
	var houses Houses
	resp, err := http.Get(fmt.Sprintf(url))
	if err != nil {
		return Houses{}, fmt.Errorf("getting houses: %w", err)
	}
	defer resp.Body.Close()

	// Read out response body & unmarshal into Houses struct
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Houses{}, fmt.Errorf("reading response body: %w", err)
	}
	err = json.Unmarshal(body, &houses)
	if err != nil {
		return Houses{}, fmt.Errorf("marshaling to House struct: %w", err)
	}
	return houses, nil
}
