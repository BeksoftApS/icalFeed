package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

//var eventsJsonFilePath = "/var/www/icalContractFeed/contracts.json" // TODO: Server
//var leadsJsonFilePath = "/var/www/icalContractFeed/leads.json" // TODO: Server
//var feedPath = "/var/www/html/ical/contract/" // TODO: Server
//var configPath = "./config.txt" // TODO: Server

var contractsJsonFilePath = "./contracts.json" // Local
var leadsJsonFilePath = "./leads.json"         // Local
var feedPath = "./feeds/"                      // Local
var configPath = "./config.txt"

var podioWebHookLink string
var domainName string

// this is a comment

type Artist struct {
	Title          []string
	Artist         []string
	Date           []string
	PodioAppItemID []int
	CreatedDate    []string
}

type Contracts []struct {
	Title          string `json:"title"`
	Artist         string `json:"artist"`
	Date           string `json:"show-date-calendar"`
	PodioAppItemID int    `json:"podio-app-item-id"`
	CreatedDate    string `json:"created-date"`
}

type Leads []struct {
	Title          string `json:"band-event"`
	Artist         string `json:"kontrahent-2"`
	Date           string `json:"datovaelger"`
	PodioAppItemID int    `json:"podio-app-item-id"`
	CreatedDate    string `json:"created-date"`
}

func main() {
	loadConfig()
	contracts := getContracts()
	leads := getLeads()
	makeFeedsForContractsAndLeads(contracts, leads)
}

func loadConfig() {
	configData, err := os.ReadFile(configPath) // open file
	if err != nil {
		panic(err)
	}

	config := strings.Split(string(configData), ";")
	if len(config) >= 2 {
		domainName = config[0]
		podioWebHookLink = config[1]
	} else {
		panic("Configuration needs to hold: [domainName],[podioWebHookLink]")
	}
}

// Generic function (Contracts, Leads)
func loadJSONFileDataToStruct[T any](fileName string, dataStruct T) T {
	fileData, err := os.ReadFile(fileName) // open file
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(fileData, &dataStruct) // Unmarshal json - put data from json to constructed variable
	if err != nil {
		panic(err)
	}

	return dataStruct
}

func getContracts() Contracts {
	// Load data from json file
	contracts := loadJSONFileDataToStruct[Contracts](contractsJsonFilePath, Contracts{})
	return contracts
}

func getLeads() Leads {
	leads := loadJSONFileDataToStruct[Leads](leadsJsonFilePath, Leads{}) // Load data from json file
	return leads
}

func associateArtists(contracts Contracts, leads Leads, artistContracts map[string]Artist) ([]string, map[string]Artist) {
	var concertTitles []string
	for _, contract := range contracts {
		sanitizedTitle := sanitizeFileTitle(contract.Artist)
		_, exist := artistContracts[sanitizedTitle]
		// check if duplicate exist
		if !exist {
			// if duplicate does not exist, then just insert concert data
			concertTitles = append(concertTitles, sanitizedTitle)
			artistContracts[sanitizedTitle] = Artist{
				[]string{contract.Title + " CONFIRMED"},
				[]string{contract.Artist},
				[]string{contract.Date},
				[]int{contract.PodioAppItemID},
				[]string{contract.CreatedDate},
			}
		} else {
			// else append concert data to existing array data
			var titleArray = artistContracts[sanitizedTitle].Title
			var artistArray = artistContracts[sanitizedTitle].Artist
			var dateArray = artistContracts[sanitizedTitle].Date
			var itemIDArray = artistContracts[sanitizedTitle].PodioAppItemID
			var createdDateArray = artistContracts[sanitizedTitle].CreatedDate

			titleArray = append(titleArray, contract.Title+" CONFIRMED")
			artistArray = append(artistArray, contract.Artist)
			dateArray = append(dateArray, contract.Date)
			itemIDArray = append(itemIDArray, contract.PodioAppItemID)
			createdDateArray = append(createdDateArray, contract.CreatedDate)
			// insert array
			artistContracts[sanitizedTitle] = Artist{titleArray, artistArray, dateArray, itemIDArray, createdDateArray}
		}

	}
	for _, lead := range leads {
		sanitizedTitle := sanitizeFileTitle(lead.Artist)
		_, exist := artistContracts[sanitizedTitle]
		// check if duplicate exist
		if !exist {
			// if duplicate does not exist, then just insert concert data
			concertTitles = append(concertTitles, sanitizedTitle)
			artistContracts[sanitizedTitle] = Artist{
				[]string{"HOLD: " + lead.Title},
				[]string{lead.Artist},
				[]string{lead.Date},
				[]int{lead.PodioAppItemID},
				[]string{lead.CreatedDate},
			}
		} else {
			// else append concert data to existing array data
			var titleArray = artistContracts[sanitizedTitle].Title
			var artistArray = artistContracts[sanitizedTitle].Artist
			var dateArray = artistContracts[sanitizedTitle].Date
			var itemIDArray = artistContracts[sanitizedTitle].PodioAppItemID
			var createdDateArray = artistContracts[sanitizedTitle].CreatedDate

			titleArray = append(titleArray, "HOLD: "+lead.Title)
			artistArray = append(artistArray, lead.Artist)
			dateArray = append(dateArray, lead.Date)
			itemIDArray = append(itemIDArray, lead.PodioAppItemID)
			createdDateArray = append(createdDateArray, lead.CreatedDate)
			// insert array
			artistContracts[sanitizedTitle] = Artist{titleArray, artistArray, dateArray, itemIDArray, createdDateArray}
		}

	}
	return concertTitles, artistContracts
}

func makeFeedsForContractsAndLeads(contracts Contracts, leads Leads) {
	concertTitles, artistContracts := associateArtists(contracts, leads, map[string]Artist{})

	// loop assign concertTitle
	for i := 0; i < len(concertTitles); i++ {

		sanitizedTitle := concertTitles[i]

		// secretKey
		// 2022-05-25 07:18:11
		// 2022-05-19 13:13:13
		// 2006-02-01
		createdDate, _ := time.Parse("2006-01-02 15:04:05", artistContracts[sanitizedTitle].CreatedDate[0])
		secretKey := createdDate.Format("0502040205011505010402011520060401")
		folderName := secretKey

		err := os.MkdirAll(feedPath+folderName, os.ModePerm)
		if err != nil {
			panic(err)
		}

		// Create file and/or open file, if it does not exist
		file, createFileError := useFile(feedPath + folderName + "/" + sanitizedTitle + ".ics")
		if createFileError != nil {
			log.Fatal(createFileError)
		}
		defer file.Close()

		fillWithICSData(file, sanitizedTitle, artistContracts[sanitizedTitle].Title, artistContracts[sanitizedTitle].Date)
		// WEBHOOK TO PODIO
		request, err := http.NewRequest("GET", podioWebHookLink, nil)
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}

		query := request.URL.Query()
		query.Add("artistName", artistContracts[sanitizedTitle].Artist[0])
		query.Add("url", domainName+"/ical/artist/"+folderName+"/"+sanitizedTitle+".ics")
		request.URL.RawQuery = query.Encode()

		http.Get(request.URL.String())
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

func removeBadCharactersIn(title string) string {
	regex, err := regexp.Compile(`[^æøåÆØÅ\w]`)
	if err != nil {
		log.Fatal(err)
	}
	title = regex.ReplaceAllString(title, "")
	return title
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

		// 20210610T172345Z
		// yyyymmddThhmmssZ
		concertDateString := concertDate.Format("20060102")

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

func oldFillWithICSData(file *os.File, sanitizedTitle string, title []string, date []string) {
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
				"SUMMARY:"+title[i]+"\r\n"+
				"END:VEVENT\r\n",
			file)
	}

	writeLineInFile("END:VCALENDAR", file)
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

/*
func makeFeedWithAllConcerts() {
	concertArray, _ := loadJSONFileData(eventsJsonFilePath)
	file, useFileError := useFile("/var/www/icalContractFeed/all.ics")
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
*/
