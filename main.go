package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Stock struct {
	Url            string
	TickerSymbol   string
	CompanyName    string
	MarketCap string
	CurrentPrice string
	NextEarningsDate   string
	PerformanceOutlookShort string
	PerformanceOutlookMid string
	PerformanceOutlookLong string
	AnalystPriceTarget int64
	NumberOfAnalysts int64
	RecommendationRating int64
	CompanyAddress string
	CompanyDescription string
	CrawledDtm      time.Time
}

func main() {
	// Get the src at yahoo finance where the earnings calendar is, for a particular date
	httpRes, err := http.Get("https://finance.yahoo.com/calendar/earnings?from=2021-02-28&to=2021-03-06&day=2021-03-02")

	if err != nil {
		log.Fatal("Could not reach yahoo: ", err)
	}

	if httpRes.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", httpRes.StatusCode, httpRes.Status)
	}

	defer httpRes.Body.Close()
	initialSrc, err := goquery.NewDocumentFromReader(httpRes.Body)
	if err != nil {
		log.Fatal("Could not resolve response: ", err)
	}

	fmt.Println(initialSrc.Html())

	// Refactor the above code to use polly to visit yahoo finance where the earnings calendar is, for a particular date (see above url)
	// Collect the list of stocks, initializing each one in a map
	// Visit each stock ticker page and collect the relevant information and update each map item
	// Sort the list by some criteria and spit out

}
