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

func makeAllFeedsForConcerts() {
	// Load data from json file
	concertArray, _ := loadJSONFileData(eventsJsonFilePath)
	var title string
	for _, data := range concertArray {
		title = makeFeedForConcert(data.Koncert)
		fmt.Println(title)
	}

}

func makeFeedForConcert(data string) string {
	title := sanitizeFileTitle(data)
	// Create file and/or open file, if it does not exist
	file, createFileError := createFile("./feeds/" + title + ".txt")
	if createFileError != nil {
		log.Fatal(createFileError)
	}
	defer file.Close()
	fillWithICSDATA(file)

	return title
}

func fillWithICSDATA(file *os.File) {
	// TODO:
	// FILL WITH ICS DATA
	// TEMPLATE!!!
	writeLineInFile("test1", file)
	writeLineInFile("test2", file)
}

func loadJSONFileData(fileName string) (ConcertArray, error) {
	// open file
	fileData, readFileError := os.ReadFile(fileName)
	if readFileError != nil {
		return nil, readFileError
	}

	// Unmarshal json - put data from json to constructed variable
	var concerts ConcertArray
	err := json.Unmarshal(fileData, &concerts)
	if err != nil {
		return nil, err
	}
	return concerts, nil
}

func sanitizeFileTitle(title string) string {
	// limit the values to A-Z and 0-9
	title = strings.Split(title, "@")[0] // take only title in the left in the example: concertTitle @ concertPlace
	title = removeBadCharactersIn(title) // remove everything that is not A-Z or 0-9
	return title
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
