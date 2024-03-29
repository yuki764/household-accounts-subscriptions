package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"google.golang.org/api/sheets/v4"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		// https://cloud.google.com/logging/docs/structured-logging
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.MessageKey {
				attr.Key = "message"
			}
			if attr.Key == slog.LevelKey {
				attr.Key = "severity"
				level := attr.Value.Any().(slog.Level)
				if level == slog.LevelWarn {
					attr.Value = slog.StringValue("WARNING")
				}
			}
			return attr
		},
	}))
	slog.SetDefault(logger)

	spreadsheetId := os.Getenv("SPREADSHEET_ID")
	sheetName := os.Getenv("SHEET_NAME")
	householdAccountsFormUrl := os.Getenv("HOUSEHOLD_ACCOUNTS_FORM_URL")

	// get current time in local time
	tz, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		slog.Default().With("error", err).Error("failed to load local time location")
		panic(err)
	}
	t := time.Now().In(tz)

	// get subscription information from spreadsheet
	ctx := context.Background()
	svc, err := sheets.NewService(ctx)
	if err != nil {
		slog.Default().With("error", err).Error("failed to create Spreadsheet service")
		panic(err)
	}
	sheetRange := "A:E"
	subsResp, err := svc.Spreadsheets.Values.Get(spreadsheetId, sheetName+"!"+sheetRange).ValueRenderOption("UNFORMATTED_VALUE").Do()
	if err != nil {
		slog.Default().With("error", err).Error("failed to get values from Spreadsheet")
		panic(err)
	}

	// check each subscription
	nothingToUpdate := true
	for i, v := range subsResp.Values {
		// skip header
		if i == 0 {
			continue
		}
		// today is payment day
		if payDay, ok := v[1].(float64); ok && int(payDay) == t.Day() {
			// monthly subscription ("*" is input) OR this month is payment month
			if payMonth, ok := v[0].(string); ok && payMonth == "*" {
				// go to "send account information"
			} else if payMonth, ok := v[0].(float64); ok && int(payMonth) == int(t.Month()) {
				// go to "send account information"
			} else {
				continue
			}

			// send account information
			accountData := url.Values{}
			accountData.Add("date", t.Format("2006-01-02"))
			if accountCategory, ok := v[2].(string); ok {
				accountData.Add("category", accountCategory)
			} else {
				slog.Default().With("error", err).Error("format error in `category` column")
				panic(err)
			}
			if accountPrice, ok := v[3].(float64); ok {
				accountData.Add("price", fmt.Sprintf("%.0f", accountPrice))
			} else {
				slog.Default().With("error", err).Error("format error in `price` column")
				panic(err)
			}
			if accountItem, ok := v[4].(string); ok {
				accountData.Add("item", accountItem)
			} else {
				slog.Default().With("error", err).Error("format error in `item` column")
				panic(err)
			}
			// run query
			accountResp, err := http.PostForm(householdAccountsFormUrl, accountData)
			if err != nil {
				slog.Default().With("error", err).Error("failed to post to account form")
				panic(err)
			}
			nothingToUpdate = false
			// print query result to stdout
			fmt.Println(accountResp.Status + " for " + v[4].(string))
		}
	}

	if nothingToUpdate {
		slog.Default().Info("nothing to update")
	}
}
