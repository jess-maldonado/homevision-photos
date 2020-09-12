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

/*
Thoughts on improving for an actual production service
- Add a custom HTTP client with logging, auth, etc
- Add a max # of retries for a single page
- Ability to save into a custom folder instead of working directory
- Use Regex to parse out all punctuation from names and addresses
*/

func main() {
	baseURL := "http://app-homevision-staging.herokuapp.com/api_project/houses?page=%d"
	var houseList []Houses

	// Getting all of the houses
	for i := 0; i < 10; i++ {
		houses, err := getHouses(fmt.Sprintf(baseURL, i+1))
		// while error returns, try again
		// getHouses will error if a non-200 is returned
		for err != nil {
			fmt.Printf("trying again for page %d \n", i+1)
			houses, err = getHouses(fmt.Sprintf(baseURL, i+1))
		}
		houseList = append(houseList, houses)
	}

	// Using a waitgroup & goroutines to take advantage of concurrency in saving files
	// WG is houseList * 10 because each page unmarshals to a Houses struct with [10]House
	// Just using len(houseList) will make the waitgroup finish before everything is saved
	var wg sync.WaitGroup
	wg.Add(len(houseList) * 10)
	// Creating file name
	for _, h := range houseList {
		for _, hr := range h.Houses {
			go func(h House) {
				err := downloadPhoto(h)
				if err != nil {
					fmt.Println(err)
				}
				defer wg.Done()
			}(hr)
		}
	}
	wg.Wait()
	fmt.Println("All houses saved successfully.")

	return

}

func createFileName(h House) string {
	//Format: id-[NNN]-[address].[ext]
	id := h.ID
	urlSplit := strings.Split(h.PhotoURL, ".")
	ext := urlSplit[len(urlSplit)-1]
	name := strings.ReplaceAll(h.Homeowner, "'", "")
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ToUpper(name)
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

func downloadPhoto(h House) error {
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
	if resp.StatusCode != 200 {
		return Houses{}, fmt.Errorf("status code: %d", resp.StatusCode)
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
