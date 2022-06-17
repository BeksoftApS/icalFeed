package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

var eventsJsonFilePath = "./contracts.json"

type ConcertArray []struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Date   string `json:"show-date-calendar"`
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
	makeFeedWithAllConcerts()
}

func makeFeedWithAllConcerts() {
	concertArray, _ := loadJSONFileData(eventsJsonFilePath)
	file, useFileError := useFile("./all.ics")
	defer file.Close()
	if useFileError != nil {
		panic(useFileError)
	}

	writeLineInFile(
		"BEGIN:VCALENDAR\r\n"+
			"VERSION:2.0\r\n"+
			"DESCRIPTION:Contract Calender\r\n"+
			"PRODID:-//https://smashbangpow.dk//NONSGML v1.0//EN\r\n",
		file)
	for _, concert := range concertArray {
		randomLetters := random6Letters()
		concertDate, _ := time.Parse("2 Jan 2006", concert.Date)
		// 20210610T172345Z
		// yyyymmddThhmmssZ
		concertDateString := concertDate.Format("20060102T150405Z")
		const max = 59
		const min = 0
		randomNumber := rand.Intn(max-min) + min
		concertDateStamp := concertDate.Add(time.Duration(int(time.Second) * randomNumber))
		concertDateStampString := concertDateStamp.Format("20060102T150405Z")
		writeLineInFile(
			"BEGIN:VEVENT\r\n"+
				"UID:"+concertDateStampString+"-"+randomLetters+"@smashbangpow.dk\r\n"+
				"DTSTAMP:"+concertDateStampString+"\r\n"+
				"DTSTART:"+concertDateString+"\r\n"+
				"DTEND:"+concertDateString+"\r\n"+
				"SUMMARY:"+concert.Title+"\r\n"+
				"END:VEVENT\r\n",
			file)
	}
	writeLineInFile("END:VCALENDAR", file)
}

func makeAllFeedsForConcerts() {
	// Load data from json file
	concertArray, _ := loadJSONFileData(eventsJsonFilePath)

	makeFeedForConcert(concertArray)
}

type Concert struct {
	Title []string
	Date  []string
}

func makeFeedForConcert(concertArray ConcertArray) {

	concerts := map[string]Concert{}
	var concertTitles []string

	for _, concert := range concertArray {
		sanitizedTitle := sanitizeFileTitle(concert.Title)
		_, exist := concerts[sanitizedTitle]
		// check if duplicate exist
		if !exist {
			// if duplicate does not exist, then just insert concert data
			concertTitles = append(concertTitles, sanitizedTitle)
			concerts[sanitizedTitle] = Concert{[]string{concert.Title}, []string{concert.Date}}
		} else {
			// else append concert data to existing array data
			var titleArray = concerts[sanitizedTitle].Title
			var dateArray = concerts[sanitizedTitle].Date

			titleArray = append(titleArray, concert.Title)
			dateArray = append(dateArray, concert.Date)
			// insert array
			concerts[sanitizedTitle] = Concert{titleArray, dateArray}
		}
	}

	// loop assign concertTitle
	for i := 0; i < len(concertTitles); i++ {
		sanitizedTitle := concertTitles[i]
		// Create file and/or open file, if it does not exist
		file, createFileError := useFile("./feeds/" + sanitizedTitle + ".ics")
		if createFileError != nil {
			log.Fatal(createFileError)
		}
		defer file.Close()

		fillWithICSData(file, sanitizedTitle, concerts[sanitizedTitle].Title, concerts[sanitizedTitle].Date)
	}

}
func random6Letters() string {
	// Random6Letters
	n := 3
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	s := fmt.Sprintf("%X", b)
	return s
}
func fillWithICSData(file *os.File, sanitizedTitle string, title []string, date []string) {
	writeLineInFile(
		"BEGIN:VCALENDAR\r\n"+
			"VERSION:2.0\r\n"+
			"DESCRIPTION:Contract Calender\r\n"+
			"PRODID:-//https://smashbangpow.dk//NONSGML v1.0//EN\r\n",
		file)

	for i := 0; i < len(date); i++ {
		randomLetters := random6Letters()
		//concertDate, _ := time.Parse("2006-02-01 15:04:05", date[i])
		concertDate, _ := time.Parse("2 Jan 2006", date[i])
		fmt.Println(date[i] + "->" + concertDate.String())
		// 20210610T172345Z
		// yyyymmddThhmmssZ
		concertDateString := concertDate.Format("20060102T150405Z")
		fmt.Println(concertDateString)

		const max = 59
		const min = 0
		randomNumber := rand.Intn(max-min) + min
		concertDateStamp := concertDate.Add(time.Duration(int(time.Second) * randomNumber))
		concertDateStampString := concertDateStamp.Format("20060102T150405Z")
		writeLineInFile(
			"BEGIN:VEVENT\r\n"+
				"UID:"+concertDateStampString+"-"+randomLetters+"@smashbangpow.dk\r\n"+
				"DTSTAMP:"+concertDateStampString+"\r\n"+
				"DTSTART:"+concertDateString+"\r\n"+
				"DTEND:"+concertDateString+"\r\n"+
				"SUMMARY:"+title[i]+"\r\n"+
				"END:VEVENT\r\n",
			file)
	}

	writeLineInFile("END:VCALENDAR", file)
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

func useFile(fileName string) (*os.File, error) {
	// If the file doesn't exist, create it, or truncate the file if it exists
	file, fileError := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if fileError != nil {
		return file, fileError
	}

	return file, nil
}

func writeLineInFile(data string, file *os.File) error {
	// Write a line with data to the file
	_, filePrintLineError := fmt.Fprint(file, data)
	if filePrintLineError != nil {
		return filePrintLineError
	}
	return nil
}
