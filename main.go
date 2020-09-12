package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

// API holds the client & base URL for the API
type API struct {
	Client  *http.Client
	BaseURL string
}

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
- Add all House structs to one Houses - could possibly run faster than the existing loop,
would need to benchmark & understand use case
- Use Regex to parse out all punctuation from names and addresses
- Use go modules for dependencies...I ran into issues with the go installation on my personal computer not recognizing external packages,
so for the sake of time I'm using just the default packages. I usually use testify/require for testing.
*/

func main() {
	client := API{
		Client:  &http.Client{},
		BaseURL: "http://app-homevision-staging.herokuapp.com/api_project/houses?page=%d",
	}
	var houseList []Houses

	// Getting all of the houses
	for i := 0; i < 10; i++ {
		houses, err := client.getHouses(i + 1)
		// while error returns, try again
		// getHouses will return an error if a non-200 is returned from the API
		for err != nil {
			fmt.Printf("trying again for page %d \n", i+1)
			houses, err = client.getHouses(i + 1)
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
				fileName, err := createFileName(h)
				if err != nil {
					fmt.Println(err)
				}
				err = downloadFile("photos", fileName, h.PhotoURL)
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

func createFileName(h House) (string, error) {
	//Format: id-[NNN]-[address].[ext]
	id := h.ID
	urlSplit := strings.Split(h.PhotoURL, ".")
	ext := urlSplit[len(urlSplit)-1]

	// Regex to match only strings, integers, and spaces
	reg, err := regexp.Compile("[^a-zA-Z0-9 ]+")
	if err != nil {
		return "", fmt.Errorf("Unable to compile regex: %w", err)
	}

	// Use regex to remove all punctuation from address & name fields
	address := reg.ReplaceAllString(h.Address, "")
	address = strings.ReplaceAll(address, " ", "-")

	name := reg.ReplaceAllString(h.Homeowner, "")
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ToUpper(name)

	return fmt.Sprintf("%d-%s-%s.%s", id, name, address, ext), nil

}

func downloadFile(directory string, fileName string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("getting file contents: %w", err)
	}
	defer resp.Body.Close()

	// Create directory if not exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, 0755)
	}
	if err != nil {
		return fmt.Errorf("creating photo directory: %w", err)
	}
	// Creating file at given path
	file, err := os.Create(fmt.Sprintf("%s/%s", directory, fileName))
	if err != nil {
		return fmt.Errorf("creating new file: %w", err)
	}
	defer file.Close()

	// Copying contents to file path
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("copying file: %w", err)
	}
	return nil
}

func (api API) getHouses(page int) (Houses, error) {
	var houses Houses

	// Creating request
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(api.BaseURL, page), nil)
	if err != nil {
		return Houses{}, fmt.Errorf("creating request: %w", err)
	}
	// Executing request
	resp, err := api.Client.Do(req)
	if err != nil {
		return Houses{}, fmt.Errorf("getting houses: %w", err)
	}
	// If status code not 200, return error
	if resp.StatusCode != http.StatusOK {
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
