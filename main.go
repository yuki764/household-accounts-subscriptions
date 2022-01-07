package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"google.golang.org/api/sheets/v4"
)

func main() {
	spreadsheetId := os.Getenv("SPREADSHEET_ID")
	sheetName := os.Getenv("SHEET_NAME")
	householdAccountsFormUrl := os.Getenv("HOUSEHOLD_ACCOUNTS_FORM_URL")

	// get current time in local time
	tz, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		log.Fatal(err)
	}
	t := time.Now().In(tz)

	// get subscription information from spreadsheet
	ctx := context.Background()
	svc, err := sheets.NewService(ctx)
	if err != nil {
		log.Fatal(err)
	}
	sheetRange := "A:E"
	subsResp, err := svc.Spreadsheets.Values.Get(spreadsheetId, sheetName+"!"+sheetRange).Do()
	if err != nil {
		log.Fatal(err)
	}

	// check each subscription
	nothingToUpdate := true
	for i, v := range subsResp.Values {
		// skip header
		if i == 0 {
			continue
		}
		// today is payment date
		if v[1] == strconv.Itoa(t.Day()) {
			// monthly subscription ("*" is input) OR this month is payment month
			if v[0] == "*" || v[0] == strconv.Itoa(int(t.Month())) {
				// send account information
				accountData := url.Values{}
				accountData.Add("date", t.Format("2006-01-02"))
				accountData.Add("category", v[2].(string))
				accountData.Add("price", v[3].(string))
				accountData.Add("item", v[4].(string))
				accountResp, err := http.PostForm(householdAccountsFormUrl, accountData)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println(accountResp.Status + " for " + v[4].(string))
				nothingToUpdate = false
			}
		}
	}

	if nothingToUpdate {
		fmt.Println("nothing to update")
	}
}
