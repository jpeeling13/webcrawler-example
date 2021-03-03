package main

import (
	"github.com/gocolly/colly/v2"
	"log"
	"strconv"
	"strings"
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

func main() {
	var stockData = make(map [string]*Stock)

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36"),
		colly.AllowedDomains("finance.yahoo.com"),
		colly.MaxBodySize(0),
		colly.Async(false),
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

	// Unique Identifier for the Individual Stock Quote Page
	// On each stock ticker page collect the relevant information and update each map item
	c.OnHTML(".finance.US", func(e *colly.HTMLElement) {
		if !strings.Contains(e.Request.URL.Path, "/quote/") {
			return
		}

		// Get the current stock in the map that matches the one on the ticker page
		stockNameTickerString := e.ChildText("#quote-header-info h1")
		justTicker :=stockNameTickerString[strings.Index(stockNameTickerString, "(")+1:strings.Index(stockNameTickerString, ")")]
		currStock := stockData[justTicker]

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
		avgVolI, err := strconv.ParseInt(avgVolS, 10, 64)
		if err != nil {
			log.Fatal("Couldn't parse avg vol for ", currStock.TickerSymbol, ": ", err)
		}
		currStock.AvgVolume = avgVolI

		// Visit the Technical Page
		e.Request.Visit("https://finance.yahoo.com/chart/" + currStock.TickerSymbol + "?technical=true")
	})


	// On each stock Technical page collect the short, mid, and long-term outlook
	c.OnHTML("#chart-header", func(e *colly.HTMLElement) {
		if !strings.Contains(e.Request.URL.Path, "chart") {
			return
		}
		log.Println("Finally GOT here")
		// Get the current stock in the map that matches the one on the ticker page
		stockNameTickerString := e.ChildText("#chart-header h1 > translate")
		justTicker :=stockNameTickerString[strings.Index(stockNameTickerString, "(")+1:strings.Index(stockNameTickerString, ")")]
		currStock := stockData[justTicker]
		log.Println("Technical Page - Ticker: ", currStock.TickerSymbol)
	})


	c.Visit("https://finance.yahoo.com/calendar/earnings?from=2021-02-28&to=2021-03-06&day=2021-03-03")
	c.Wait()

	//for _, v := range stockData {
	//	log.Println(v.CompanyName, v.TickerSymbol, v.NextEarningsCallTime, v.NextEarningsEstimate, v.Url)
	//}

	log.Println("Total Stocks: ", len(stockData))

	// Sort the list by some criteria and spit out

}
