package main

import (
	"github.com/gocolly/colly/v2"
	"log"
	"strings"
	"time"
)

type Stock struct {
	Url            string
	TickerSymbol   string
	CompanyName    string
	MarketCap string
	CurrentPrice float64
	NextEarningsDate   string
	NextEarningsCallTime string
	NextEarningsEstimate string
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
	var stockData = make(map [string]*Stock)

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36"),
		colly.AllowedDomains("finance.yahoo.com"),
		colly.Async(true),
	)

	// Set max Parallelism and introduce a Random Delay
	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		Delay: 5 * time.Second,
	})

	log.Println("User Agent: ", c.UserAgent)

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())

	})

	// Unique Identifier for the Earnings Page
	c.OnHTML(`.simpTblRow`, func(e *colly.HTMLElement){
		temp := Stock{}
		temp.CompanyName = e.ChildText(`td[aria-label="Company"]`)
		temp.TickerSymbol = e.ChildText(`td[aria-label="Symbol"]`)
		temp.NextEarningsCallTime = e.ChildText(`td[aria-label="Earnings Call Time"]`)
		temp.NextEarningsEstimate = e.ChildText(`td[aria-label="EPS Estimate"]`)
		temp.Url =e.ChildAttr("td>a", "href")
		stockData[temp.CompanyName] = &temp
		e.Request.Visit("https://finance.yahoo.com" + temp.Url)
	})

	// Unique Identifier for the Individual Stock Quote Page
	c.OnHTML(".finance.US", func(e *colly.HTMLElement) {
		if !strings.Contains(e.Request.URL.Path, "/quote/") {
			return
		}
		log.Println("Got Stock Quote Page: ", e.ChildText("#quote-header-info"))
	})


	c.Visit("https://finance.yahoo.com/calendar/earnings?from=2021-02-28&to=2021-03-06&day=2021-03-03")
	c.Wait()

	for _, v := range stockData {
		log.Println(v.CompanyName, v.TickerSymbol, v.NextEarningsCallTime, v.NextEarningsEstimate, v.Url)
	}

	log.Println("Total Stocks: ", len(stockData))

	// Collect the list of stocks, initializing each one in a map
	// Visit each stock ticker page and collect the relevant information and update each map item
	// Sort the list by some criteria and spit out

}
