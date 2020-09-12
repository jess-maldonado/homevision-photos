package main

import (
	"testing"
)

func TestGetHouses(t *testing.T) {
	testCases := []struct {
		desc string
	}{
		{
			desc: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {

		})
	}
}

func TestSaveFile(t *testing.T) {
	testCases := []struct {
		desc string
	}{
		{
			desc: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {

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
			fileName := createFileName(tc.house)
			if fileName != tc.expected {
				t.Errorf("names did not match, got: %s, wanted: %s", fileName, tc.expected)
			}

		})
	}
}
