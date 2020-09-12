package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

// Stub for the API response that provides a server with a given response & status code
func apiResponseStub(response string, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status != http.StatusOK {
			http.Error(w, "service available", http.StatusNotFound)
			return
		}
		w.Write([]byte(response))

	}))

}

var okResponse = `{"houses":[{"id":0,"address":"4 Pumpkin Hill Street Antioch, TN 37013","homeowner":"Nicole Bone","price":105124,"photoURL":"https://image.shutterstock.com/image-photo/big-custom-made-luxury-house-260nw-374099713.jpg"},{"id":1,"address":"495 Marsh Road Portage, IN 46368","homeowner":"Rheanna Walsh","price":161856,"photoURL":"https://media-cdn.tripadvisor.com/media/photo-s/09/7c/a2/1f/patagonia-hostel.jpg"}]}`
var okHouse = Houses{
	Houses: []House{
		{
			ID:        0,
			Address:   "4 Pumpkin Hill Street Antioch, TN 37013",
			Homeowner: "Nicole Bone",
			Price:     105124,
			PhotoURL:  "https://image.shutterstock.com/image-photo/big-custom-made-luxury-house-260nw-374099713.jpg",
		},
		{
			ID:        1,
			Address:   "495 Marsh Road Portage, IN 46368",
			Homeowner: "Rheanna Walsh",
			Price:     161856,
			PhotoURL:  "https://media-cdn.tripadvisor.com/media/photo-s/09/7c/a2/1f/patagonia-hostel.jpg",
		},
	},
}

func TestGetHouses(t *testing.T) {
	testCases := []struct {
		desc      string
		code      int
		response  string
		house     Houses
		expectErr bool
	}{
		{
			desc:     "VALID_RESPONSE",
			code:     http.StatusOK,
			response: okResponse,
			house:    okHouse,
		},
		{
			desc:     "EMPTY_HOUSES",
			code:     http.StatusOK,
			response: `{"houses": []}`,
			house: Houses{
				Houses: []House{},
			},
		},
		{
			desc:      "INVALID_JSON",
			code:      http.StatusOK,
			response:  `""`,
			expectErr: true,
		},
		{
			desc:      "NON_200_RESPONSE",
			code:      http.StatusNotFound,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			server := apiResponseStub(tc.response, tc.code)
			defer server.Close()
			testClient := API{
				Client:  server.Client(),
				BaseURL: fmt.Sprintf("%s/api_project/houses?page=%s", server.URL, "%s"),
			}
			houses, err := testClient.getHouses(25)
			if err != nil && !tc.expectErr {
				t.Errorf("Got unexpected error: %w", err)
			}
			if err == nil && tc.expectErr {
				t.Errorf("Expected an error and got none")
			}
			if !tc.expectErr && !reflect.DeepEqual(tc.house, houses) {
				t.Errorf("Houses did not return as expected. Expected: %v, got: %v", tc.house, houses)

			}

		})
	}
}

func TestDownloadFile(t *testing.T) {
	testCases := []struct {
		desc      string
		fileName  string
		url       string
		directory string
		expectErr bool
	}{
		{
			desc:      "VALID_PHOTO_FILE",
			fileName:  "file1.jpg",
			directory: "photos",
			url:       "https://image.shutterstock.com/image-photo/big-custom-made-luxury-house-260nw-374099713.jpg",
		},
		{
			desc:      "INVALID_URL",
			fileName:  "file1.jpg",
			directory: "photos",
			url:       "empty string",
			expectErr: true,
		},
		{
			desc:      "INVALID_FILE_NAME",
			fileName:  "file1%!;..jpg",
			directory: "photos",
			url:       "https://image.shutterstock.com/image-photo/big-custom-made-luxury-house-260nw-374099713.jpg",
			expectErr: true,
		},
		{
			desc:      "INVALID_DIRECTORY",
			fileName:  "file1.jpg",
			directory: "abc/.xk;%",
			url:       "https://image.shutterstock.com/image-photo/big-custom-made-luxury-house-260nw-374099713.jpg",
			expectErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tempDir, err := ioutil.TempDir("", tc.directory)
			filePath := fmt.Sprintf("%s/%s", tempDir, tc.fileName)

			if err != nil && tc.desc != "INVALID_DIRECTORY" {
				t.Errorf("Unable to get temp directory: %w", err)
			}
			err = downloadFile(tempDir, tc.fileName, tc.url)

			if err != nil && !tc.expectErr {
				t.Errorf("Unexpected error: %w", err)
			}
			if !tc.expectErr {
				_, err := os.Stat(filePath)
				if err != nil {
					t.Errorf("Expected to find saved file: %w", err)
				}
			}

			defer os.Remove(filePath)
			defer os.Remove(tempDir)

		})
	}
}

func TestCreateFileName(t *testing.T) {
	testCases := []struct {
		desc     string
		house    House
		expected string
	}{
		{
			desc: "BASIC_HOUSE",
			house: House{
				ID:        1,
				Homeowner: "John Smith",
				Address:   "123 Main St. Cary, NC 27513",
				PhotoURL:  "google.com/photo.jpg",
			},
			expected: "1-JOHN-SMITH-123-Main-St-Cary-NC-27513.jpg",
		},
		{
			desc: "LOTS_OF_SUBDOMAINS_IN_URL",
			house: House{
				ID:        1,
				Homeowner: "John Smith",
				Address:   "123 Main St. Cary, NC 27513",
				PhotoURL:  "photos.drive.homevision.google.com/photo.jpg",
			},
			expected: "1-JOHN-SMITH-123-Main-St-Cary-NC-27513.jpg",
		},
		{
			desc: "NO_HOMEOWNER",
			house: House{
				ID:       1,
				Address:  "123 Main St. Cary, NC 27513",
				PhotoURL: "photos.drive.homevision.google.com/photo.jpg",
			},
			expected: "1--123-Main-St-Cary-NC-27513.jpg",
		},
		{
			desc: "MULTIPLE_COMMAS_IN_ADDRESS",
			house: House{
				ID:        1,
				Homeowner: "John Smith",
				Address:   "123 Main St., Cary, NC 27513",
				PhotoURL:  "photos.drive.homevision.google.com/photo.jpg",
			},
			expected: "1-JOHN-SMITH-123-Main-St-Cary-NC-27513.jpg",
		},
		{
			desc: "MULTIPLE_PERIODS_IN_ADDRESS",
			house: House{
				ID:        1,
				Homeowner: "John Smith",
				Address:   "123 N.W. Main St., Cary, NC 27513",
				PhotoURL:  "photos.drive.homevision.google.com/photo.jpg",
			},
			expected: "1-JOHN-SMITH-123-NW-Main-St-Cary-NC-27513.jpg",
		},
		{
			desc: "EXTRA_CHARACTERS_IN_NAME",
			house: House{
				ID:        1,
				Homeowner: "Kevin O'Leary",
				Address:   "123 Main St., Cary, NC 27513",
				PhotoURL:  "photos.drive.homevision.google.com/photo.jpg",
			},
			expected: "1-KEVIN-OLEARY-123-Main-St-Cary-NC-27513.jpg",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			fileName, err := createFileName(tc.house)
			if err != nil {
				t.Error(err)
			}
			if fileName != tc.expected {
				t.Errorf("names did not match, got: %s, wanted: %s", fileName, tc.expected)
			}

		})
	}
}
