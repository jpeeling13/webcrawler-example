package main

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Stock struct {
	Url            string
	TickerSymbol   string
	CompanyName    string
	MarketCap string
	PrevDayVolume int64
	AvgVolume int64
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

var mu sync.Mutex

func main() {
	var stockData = make(map [string]*Stock)

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36"),
		colly.AllowedDomains("finance.yahoo.com"),
		colly.MaxBodySize(0),
		colly.AllowURLRevisit(),
		colly.Async(true),
	)

	// Set max Parallelism and introduce a Random Delay
	c.Limit(&colly.LimitRule{
		DomainGlob: "*",
		Parallelism: 2,
		Delay: 100 * time.Millisecond,
	})

	log.Println("User Agent: ", c.UserAgent)

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())

	})

	// Unique Identifier for the Earnings Page
	// Collect the list of stocks, initializing each one in a map
	c.OnHTML(`.simpTblRow`, func(e *colly.HTMLElement){
		temp := Stock{}
		temp.CompanyName = e.ChildText(`td[aria-label="Company"]`)
		temp.TickerSymbol = e.ChildText(`td[aria-label="Symbol"]`)
		temp.NextEarningsCallTime = e.ChildText(`td[aria-label="Earnings Call Time"]`)
		temp.NextEarningsEstimate = e.ChildText(`td[aria-label="EPS Estimate"]`)
		temp.Url =e.ChildAttr("td>a", "href")
		stockData[temp.TickerSymbol] = &temp
		e.Request.Visit("https://finance.yahoo.com" + temp.Url)
	})

	// On each stock ticker page collect the relevant information and update each map item
	c.OnHTML(".finance.US", func(e *colly.HTMLElement) {

		// Skip this callback if we are on the earnings calendar page
		if !strings.Contains(e.Request.URL.Path, "/quote/"){
			return
		}

		// Get the current stock in the map that matches the one on the ticker page
		stockNameTickerString := e.ChildText("#quote-header-info h1")
		justTicker :=stockNameTickerString[strings.Index(stockNameTickerString, "(")+1:strings.Index(stockNameTickerString, ")")]
		currStock := stockData[justTicker]
		log.Println(currStock.TickerSymbol)


		// Capture current price
		priceS := e.ChildText(`#quote-header-info > div:nth-child(3) > div > div > span:first-child`)
		priceF, err := strconv.ParseFloat(priceS, 64)
		if err != nil {
			log.Fatal("Couldn't parse price for ", currStock.TickerSymbol, ": ", err)
		}
		currStock.CurrentPrice = priceF

		// Capture Market Cap
		currStock.MarketCap = e.ChildText(`td[data-test="MARKET_CAP-value"]`)

		// Capture Previous Day Volume
		prevDayVolS := strings.Replace(e.ChildText(`td[data-test="TD_VOLUME-value"]`), ",", "", -1)
		prevDayVolI, err := strconv.ParseInt(prevDayVolS, 10, 64)
		if err != nil {
			log.Fatal("Couldn't parse prev day vol for ", currStock.TickerSymbol, ": ", err)
		}
		currStock.PrevDayVolume = prevDayVolI


		// Capture Average Daily Volume
		avgVolS := strings.Replace(e.ChildText(`td[data-test="AVERAGE_VOLUME_3MONTH-value"]`), ",", "", -1)
		if avgVolS == "N/A" {
			currStock.AvgVolume = -1
		} else {
			avgVolI, err := strconv.ParseInt(avgVolS, 10, 64)
			if err != nil {
				log.Fatal("Couldn't parse avg vol for ", currStock.TickerSymbol, ": ", err)
			}
			currStock.AvgVolume = avgVolI
		}

		// The LAST 3 Stats are not always captured for each stock... why?
		// Capture Short Term Outlook
		currStock.PerformanceOutlookShort = e.ChildAttr(`#chrt-evts-mod > div:nth-child(3) > ul > li:first-child > a svg`, "style")

		// Capture Mid Term Outlook
		currStock.PerformanceOutlookMid = e.ChildAttr(`#chrt-evts-mod > div:nth-child(3) > ul > li:nth-child(2)> a svg`, "style")

		// Capture Short Term Outlook
		currStock.PerformanceOutlookLong = e.ChildAttr(`#chrt-evts-mod > div:nth-child(3) > ul > li:nth-child(3)> a svg`, "style")

	})

	c.Visit("https://finance.yahoo.com/calendar/earnings?from=2021-02-28&to=2021-03-06&day=2021-03-04")
	c.Wait()

	for _, v := range stockData {
		fmt.Println(v.TickerSymbol, " - ", v.PerformanceOutlookShort)
		fmt.Println(v.TickerSymbol, " - ", v.PerformanceOutlookMid)
		fmt.Println(v.TickerSymbol, " - ", v.PerformanceOutlookLong)
	}
}