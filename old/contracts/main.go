package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func main() {
	// Open Data File
	contractNames, err := os.ReadFile("contracts.txt")
	handleError(false, err)
	contracts := strings.Split(string(contractNames), ";")

	for i := 0; i < len(contracts)-1; i++ {
		selectedContract := contracts[i]
		selectedContract = removeBadCharactersIn(selectedContract)

		dataFileName := "all.ics"
		newFileName := selectedContract + ".ics"

		found := false

		// Open Data File
		dataFile, err := os.ReadFile(dataFileName)
		handleError(false, err)

		contractsWithICSFormat := strings.Split(string(dataFile), "SUMMARY:")

		newData := "BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//https://smashbangpow.dk//NONSGML v1.0//EN\n"

		for i := 0; i < len(contractsWithICSFormat); i++ {
			contractsWithNamePlace := strings.Split(contractsWithICSFormat[i], ",")
			contractsWithName := strings.Split(contractsWithNamePlace[0], "@")
			if strings.Contains(contractsWithName[0], selectedContract) {
				// Update New Contact ICS Data

				handleError(true, err)
				newData += "BEGIN:VEVENT\nSUMMARY:" + contractsWithICSFormat[i]
				found = true
				fmt.Println(contractsWithICSFormat[i])
			}
		}

		newData += "END:VCALENDAR"
		if found {
			fmt.Println("Creating: " + newFileName)
			newFile, err := os.OpenFile(newFileName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
			handleError(true, err)
			_, err = fmt.Fprintln(newFile, newData)
		}

	}
}

func removeBadCharactersIn(title string) string {
	regex, err := regexp.Compile(`[^æøåÆØÅ\w]`)
	if err != nil {
		log.Fatal(err)
	}
	title = regex.ReplaceAllString(title, "")
	return title
}

func handleError(onlyWarning bool, err error) {
	if err != nil {
		if onlyWarning {
			fmt.Print("Warning: ")
			fmt.Println(err)
		} else {
			panic(err)
		}
	}

}
