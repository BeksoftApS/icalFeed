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
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
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
	makeFeedWithAllConcerts()
	//Write data to file
	//	writeLineInFileError := writeLineInFile("bob", file)
	//	if writeLineInFileError != nil {
	//		log.Fatal(writeLineInFileError)
	//	}
	fmt.Println("test")
}

func makeFeedWithAllConcerts() {
	concertArray, _ := loadJSONFileData(eventsJsonFilePath)
	file, useFileError := useFile("./all.ics")
	defer file.Close()
	if useFileError != nil {
		panic(useFileError)
	}

	writeLineInFile(
		"BEGIN:VCALENDAR\n"+
			"VERSION:2.0\n"+
			"PRODID:-//https://smashbangpow.dk//NONSGML v1.0//EN\n",
		file)
	for _, concert := range concertArray {
		randomLetters := random6Letters()
		concertDate, _ := time.Parse("2006-02-01 15:04:05", concert.ConcertDate)
		// 20210610T172345Z
		// yyyymmddThhmmssZ
		concertDateString := concertDate.Format("20060201T150405Z")
		const max = 59
		const min = 0
		randomNumber := rand.Intn(max-min) + min
		concertDateStamp := concertDate.Add(time.Duration(int(time.Second) * randomNumber))
		concertDateStampString := concertDateStamp.Format("20060201T150405Z")
		writeLineInFile(
			"BEGIN:VEVENT\n"+
				"UID:"+concertDateStampString+"-"+randomLetters+"@smashbangpow.dk\n"+
				"DTSTAMP:"+concertDateStampString+"\n"+
				"DTSTART:"+concertDateString+"\n"+
				"DTEND:"+concertDateString+"\n"+
				"SUMMARY:"+concert.Koncert+"\n"+
				"END:VEVENT\n",
			file)
	}
	writeLineInFile(

		"END:VCALENDAR",
		file)
}

func makeAllFeedsForConcerts() {
	// Load data from json file
	concertArray, _ := loadJSONFileData(eventsJsonFilePath)

	// 1) struct
	// 2) koncert - prevent duplicates
	// array(   KONCERT - title, date
	//			, KONCERT 2
	//			, KONCERT 3

	makeFeedForConcert(concertArray)
}

type Concert struct {
	title []string
	date  []string
}

func makeFeedForConcert(concertArray ConcertArray) {

	concerts := map[string]Concert{}
	var concertTitles []string

	for _, concert := range concertArray {
		sanitizedTitle := sanitizeFileTitle(concert.Koncert)
		_, exist := concerts[sanitizedTitle]
		// check if duplicate exist
		if !exist {
			// if duplicate does not exist, then just insert concert data
			concertTitles = append(concertTitles, sanitizedTitle)
			concerts[sanitizedTitle] = Concert{[]string{concert.Koncert}, []string{concert.ConcertDate}}
		} else {
			// else append concert data to existing array data
			var titleArray = concerts[sanitizedTitle].title
			var dateArray = concerts[sanitizedTitle].date
			titleArray = append(titleArray, concert.Koncert)
			dateArray = append(dateArray, concert.ConcertDate)
			// insert array
			concerts[sanitizedTitle] = Concert{titleArray, dateArray}
		}
	}

	//fmt.Println(concerts["artistNavn3"].title)
	//fmt.Println(len(concerts["artistNavn3"].title))

	// loop assign concertTitle
	for i := 0; i < len(concertTitles); i++ {
		sanitizedTitle := concertTitles[i]
		// Create file and/or open file, if it does not exist
		file, createFileError := useFile("./feeds/" + sanitizedTitle + ".ics")
		if createFileError != nil {
			log.Fatal(createFileError)
		}
		defer file.Close()

		fillWithICSData(file, sanitizedTitle, concerts[sanitizedTitle].title, concerts[sanitizedTitle].date)
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
	// TODO:
	// FILL WITH ICS DATA
	// TEMPLATE!!!
	writeLineInFile(
		"BEGIN:VCALENDAR\n"+
			"VERSION:2.0\n"+
			"PRODID:-//https://smashbangpow.dk//NONSGML v1.0//EN\n",
		file)

	//for _, data := range concertArray {
	//	makeFeedForConcert(data.Koncert)
	//}
	for i := 0; i < len(date); i++ {
		randomLetters := random6Letters()
		concertDate, _ := time.Parse("2006-02-01 15:04:05", date[i])
		// 20210610T172345Z
		// yyyymmddThhmmssZ
		concertDateString := concertDate.Format("20060201T150405Z")
		const max = 59
		const min = 0
		randomNumber := rand.Intn(max-min) + min
		concertDateStamp := concertDate.Add(time.Duration(int(time.Second) * randomNumber))
		concertDateStampString := concertDateStamp.Format("20060201T150405Z")
		writeLineInFile(
			"BEGIN:VEVENT\n"+
				"UID:"+concertDateStampString+"-"+randomLetters+"@smashbangpow.dk\n"+
				"DTSTAMP:"+concertDateStampString+"\n"+
				"DTSTART:"+concertDateString+"\n"+
				"DTEND:"+concertDateString+"\n"+
				"SUMMARY:"+title[i]+"\n"+
				"END:VEVENT\n",
			file)
	}

	writeLineInFile(

		"END:VCALENDAR",
		file)

	//writeLineInFile("BEGIN:VCALENDAR", file)
	//writeLineInFile("VERSION:2.0", file)
	//writeLineInFile("END:VCALENDAR", file)
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
	_, filePrintLineError := fmt.Fprintln(file, data)
	if filePrintLineError != nil {
		return filePrintLineError
	}
	return nil
}
