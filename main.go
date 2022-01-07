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

	tz, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		log.Fatal(err)
	}
	t := time.Now().In(tz)
	ctx := context.Background()
	svc, err := sheets.NewService(ctx)
	if err != nil {
		log.Fatal(err)
	}
	sheetRange := "A:D"
	resp, err := svc.Spreadsheets.Values.Get(spreadsheetId, sheetName+"!"+sheetRange).Do()
	if err != nil {
		log.Fatal(err)
	}

	nothingToUpdate := true
	for i, v := range resp.Values {
		// skip header
		if i == 0 {
			continue
		}
		// today is payment date
		if v[0] == strconv.Itoa(t.Day()) {
			// send account information
			accountData := url.Values{}
			accountData.Add("date", t.Format("2006-01-02"))
			accountData.Add("category", v[1].(string))
			accountData.Add("price", v[2].(string))
			accountData.Add("item", v[3].(string))
			accountResp, err := http.PostForm(householdAccountsFormUrl, accountData)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(accountResp.Status + " for " + v[3].(string))
			nothingToUpdate = false
		}
	}

	if nothingToUpdate {
		fmt.Println("nothing to update")
	}
}
