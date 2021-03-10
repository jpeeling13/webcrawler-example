package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"math/rand"
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
	Sector string
	StockPerformanceOutlookShort string
	StockPerformanceOutlookMid string
	StockPerformanceOutlookLong string
	SectorPerformanceOutlookShort string
	SectorPerformanceOutlookMid string
	SectorPerformanceOutlookLong string
	AnalystPriceTarget int64
	NumberOfAnalysts int64
	RecommendationRating int64
	CompanyAddress string
	CompanyDescription string
	CrawledDtm      time.Time
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
func RandomString() string {
	b := make([]byte, rand.Intn(10)+10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func main() {
	var stockData = make(map [string]*Stock)

	c := colly.NewCollector(
		colly.AllowedDomains("finance.yahoo.com"),
		colly.MaxBodySize(0),
		colly.AllowURLRevisit(),
		colly.Async(true),
	)

	// Set max Parallelism and Delay of 10 seconds
	// Annoying, but seems that Yahoo has a strict rate limit for wherever it's grabbing the data to display the
	// Short, Mid, and Long term technicals. Most likely is tied by IP address
	c.Limit(&colly.LimitRule{
		DomainGlob: "*",
		Parallelism: 1,
		Delay: 10 * time.Second,
	})

	log.Println("User Agent: ", c.UserAgent)

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
		r.Headers.Set("User-Agent", RandomString())
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
		c.Visit("https://finance.yahoo.com" + temp.Url)
	})

	// On each stock ticker page collect the relevant information and update each map item
	c.OnHTML("body", func(e *colly.HTMLElement) {

		// Skip this callback if we are on the earnings calendar page
		if !strings.Contains(e.Request.URL.Path, "/quote/"){
			return
		}

		// Get the current stock in the map that matches the one on the ticker page
		stockNameTickerString := e.ChildText("#quote-header-info h1")
		justTicker :=stockNameTickerString[strings.Index(stockNameTickerString, "(")+1:strings.Index(stockNameTickerString, ")")]
		currStock := stockData[justTicker]
		log.Println(currStock.TickerSymbol)

		// Capture Crawled DTM
		currStock.CrawledDtm = time.Now()

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
	})

	// On each stock ticker page within the script tag collect the technical indicator information
	c.OnHTML(`script:contains("ResearchPageStore")`, func(e *colly.HTMLElement) {

		// Get the ticker symbol from the ticker
		scriptTag := e.DOM.Text()
		ticker := scriptTag[strings.Index(scriptTag,"\"originUrl\":\"\\u002Fquote\\u002F")+30:strings.Index(scriptTag, "?p=")]
		log.Println("Got ticker from script tag: ", ticker)
		currStock := stockData[ticker]


		contextObject := scriptTag[strings.Index(scriptTag, "\"context\"")+10:strings.LastIndex(scriptTag, "\"plugins\"")-1]
		var contextObjectAsStruct interface{}
		err := json.Unmarshal([]byte(contextObject), &contextObjectAsStruct)
		if err !=nil {
			log.Fatal("Could not Unmarshal Script Context Object", err)
		}

		if researchPageStore, ok := contextObjectAsStruct.(map[string]interface{})["dispatcher"].(map[string]interface{})["stores"].(map[string]interface{})["ResearchPageStore"].(map[string]interface{}); ok {

			if technicalInsights, ok := researchPageStore["technicalInsights"].(map[string]interface{}); ok{

				if ticker, ok := technicalInsights[currStock.TickerSymbol].(map[string]interface{}); ok {

					if instrumentInfo, ok := ticker["instrumentInfo"].(map[string]interface{}); ok {
						// Capture stock short term outlook from script tag
						shortTermOutlook := fmt.Sprintf("%v", instrumentInfo["technicalEvents"].(map[string]interface{})["shortTermOutlook"].(map[string]interface{})["direction"])
						currStock.StockPerformanceOutlookShort = shortTermOutlook

						// Capture stock short term outlook from script tag
						midTermOutlook := fmt.Sprintf("%v", instrumentInfo["technicalEvents"].(map[string]interface{})["intermediateTermOutlook"].(map[string]interface{})["direction"])
						currStock.StockPerformanceOutlookMid = midTermOutlook

						// Capture stock short term outlook from script tag
						longTermOutlook := fmt.Sprintf("%v", instrumentInfo["technicalEvents"].(map[string]interface{})["longTermOutlook"].(map[string]interface{})["direction"])
						currStock.StockPerformanceOutlookLong = longTermOutlook

						// Capture Sector from script tag
						sector := fmt.Sprintf("%v", instrumentInfo["technicalEvents"].(map[string]interface{})["sector"])
						currStock.Sector = sector

						// Capture sector short term outlook from script tag
						secShortTermOutlook := fmt.Sprintf("%v", instrumentInfo["technicalEvents"].(map[string]interface{})["shortTermOutlook"].(map[string]interface{})["sectorDirection"])
						currStock.SectorPerformanceOutlookShort = secShortTermOutlook

						// Capture sector short term outlook from script tag
						secMidTermOutlook := fmt.Sprintf("%v", instrumentInfo["technicalEvents"].(map[string]interface{})["intermediateTermOutlook"].(map[string]interface{})["sectorDirection"])
						currStock.SectorPerformanceOutlookMid = secMidTermOutlook

						// Capture sector short term outlook from script tag
						secLongTermOutlook := fmt.Sprintf("%v", instrumentInfo["technicalEvents"].(map[string]interface{})["longTermOutlook"].(map[string]interface{})["sectorDirection"])
						currStock.SectorPerformanceOutlookLong = secLongTermOutlook
					} else {
							log.Println(currStock.TickerSymbol, ": No instrumentInfo in Source")
					}
				} else {
					if strings.Contains(scriptTag, "LOAD_TECH_INSIGHTS_FAIL") {
						log.Println(currStock.TickerSymbol, ": COULD NOT LOAD RESEARCH PAGE STORE... SKIP")
					} else {
						log.Println(currStock.TickerSymbol, ": No Ticker in Source")
					}
				}
			} else {
				log.Fatal(currStock.TickerSymbol, ": No Technical Insights in Source")
			}
		} else {
			log.Fatal(currStock.TickerSymbol, ": No Research Page Store in Source")
		}
	})

	c.Visit("https://finance.yahoo.com/calendar/earnings?from=2021-02-28&to=2021-03-06&day=2021-03-04")
	c.Wait()

	fmt.Println("Total Stocks: ", len(stockData))
	for _, v := range stockData {
		fmt.Println(v.TickerSymbol, " - ", v.Url)
		fmt.Println(v.TickerSymbol, " - Current Price: ", v.CurrentPrice)
		fmt.Println(v.TickerSymbol, " - Previous Day Volume: ", v.PrevDayVolume)
		fmt.Println(v.TickerSymbol, " - Avg. Volume: ", v.AvgVolume)
		fmt.Println(v.TickerSymbol, " - Sector: ", v.Sector)
		fmt.Println(v.TickerSymbol, " - Sector Performance (short): ", v.SectorPerformanceOutlookShort)
		fmt.Println(v.TickerSymbol, " - Stock Performance (short): ", v.StockPerformanceOutlookShort)
		fmt.Println(v.TickerSymbol, " - Sector Performance (mid): ", v.SectorPerformanceOutlookMid)
		fmt.Println(v.TickerSymbol, " - Stock Performance (mid): ", v.StockPerformanceOutlookMid)
		fmt.Println(v.TickerSymbol, " - Sector Performance (long): ", v.SectorPerformanceOutlookLong)
		fmt.Println(v.TickerSymbol, " - Stock Performance (long): ", v.StockPerformanceOutlookLong)
	}
}