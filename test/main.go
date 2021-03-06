package main

import (
"fmt"
"github.com/gocolly/colly/v2"
"log"
"strings"
"sync"
"time"
)

type Stock struct {
	Url            string
	TickerSymbol   string
	CompanyName    string
	PerformanceOutlookShort string
	PerformanceOutlookMid string
	PerformanceOutlookLong string
}

var mu sync.Mutex

func main() {
	var stockData = make(map [string]*Stock)

	stockData["WB"] = &Stock{TickerSymbol: "WB"}

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

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())

	})
	
	// On each stock ticker page collect the relevant information and update each map item
	c.OnHTML(".finance.US", func(e *colly.HTMLElement) {
		

		// Get the current stock in the map that matches the one on the ticker page
		stockNameTickerString := e.ChildText("#quote-header-info h1")
		justTicker :=stockNameTickerString[strings.Index(stockNameTickerString, "(")+1:strings.Index(stockNameTickerString, ")")]
		currStock := stockData[justTicker]
		log.Println(currStock.TickerSymbol)


		// The LAST 3 Stats are not always captured for each stock... why?
		// Capture Short Term Outlook
		if currStock.PerformanceOutlookShort == "" {
			currStock.PerformanceOutlookShort = e.ChildAttr(`#chrt-evts-mod > div:nth-child(3) > ul > li:first-child > a svg`, "style")
		}

		// Capture Mid Term Outlook
		if currStock.PerformanceOutlookMid == "" {
			currStock.PerformanceOutlookMid = e.ChildAttr(`#chrt-evts-mod > div:nth-child(3) > ul > li:nth-child(2)> a svg`, "style")
		}

		// Capture Short Term Outlook
		if currStock.PerformanceOutlookLong == "" {
			currStock.PerformanceOutlookLong = e.ChildAttr(`#chrt-evts-mod > div:nth-child(3) > ul > li:nth-child(3)> a svg`, "style")
		}

		if currStock.PerformanceOutlookShort == "" || currStock.PerformanceOutlookMid == "" || currStock.PerformanceOutlookLong == "" {
			e.Request.Visit("https://finance.yahoo.com/quote/WB?p=WB")
		}

	})

	c.Visit("https://finance.yahoo.com/quote/WB?p=WB")
	c.Wait()

	for _, v := range stockData {
		fmt.Println(v.TickerSymbol, " - ", v.PerformanceOutlookShort)
		fmt.Println(v.TickerSymbol, " - ", v.PerformanceOutlookMid)
		fmt.Println(v.TickerSymbol, " - ", v.PerformanceOutlookLong)
	}
}