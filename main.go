/**

Objective:
1) write to file
2) create information by template (ical example file)
3) insert data from json example file

**/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

var eventsJsonFilePath = "./events.json"

type ConcertArray []struct {
	Koncert     string `json:"Koncert"`
	ConcertDate string `json:"Concert date"`
}

func removeBadCharactersIn(title string) string {
	regex, err := regexp.Compile(`[^æøåÆØÅ\w]`)
	if err != nil {
		log.Fatal(err)
	}
	title = regex.ReplaceAllString(title, "")
	return title
}

func main() {
	makeAllFeedsForConcerts()
	//Write data to file
	//	writeLineInFileError := writeLineInFile("bob", file)
	//	if writeLineInFileError != nil {
	//		log.Fatal(writeLineInFileError)
	//	}
}

func createFile(fileName string) (*os.File, error) {
	// If the file doesn't exist, create it, or truncate the file if it exists
	file, fileError := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if fileError != nil {
		return file, fileError
	}

	return file, nil
}

func writeLineInFile(data string, file *os.File) error {
	// Write a line with data to the file
	_, filePrintLineError := fmt.Fprintln(file, data)
	if filePrintLineError != nil {
		return filePrintLineError
	}
	return nil
}
