package main

import (
	"encoding/json"
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
	CurrentPrice float64
	MarketCap string
	StockPerformanceOutlookShort string
	StockPerformanceOutlookMid string
	StockPerformanceOutlookLong string
	Sector string
	SectorPerformanceOutlookShort string
	SectorPerformanceOutlookMid string
	SectorPerformanceOutlookLong string
}

var mu sync.Mutex

func main() {
	var stockData = make(map [string]*Stock)

	stockData["GWRS"] = &Stock{TickerSymbol: "GWRS"}

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
	c.OnHTML("body", func(e *colly.HTMLElement) {

		// Get the current stock in the map that matches the one on the ticker page
		stockNameTickerString := e.ChildText("#quote-header-info h1")
		justTicker :=stockNameTickerString[strings.Index(stockNameTickerString, "(")+1:strings.Index(stockNameTickerString, ")")]
		currStock := stockData[justTicker]
		log.Println(currStock.TickerSymbol)

		// Get the script tag on the page that has all the data we want
		scriptTag := e.ChildText(`script:contains("TechnicalInsights")`)
		contextObject := scriptTag[strings.Index(scriptTag, "\"context\"")+10:strings.LastIndex(scriptTag, "\"plugins\"")-1]
		var contextObjectAsStruct interface{}
		err := json.Unmarshal([]byte(contextObject), &contextObjectAsStruct)
		if err !=nil {
			log.Fatal(err)
		}

		if contextObjectAsStruct.(map[string]interface{})["dispatcher"].(map[string]interface{})["stores"].(map[string]interface{})["ResearchPageStore"].(map[string]interface{})["technicalInsights"].(map[string]interface{})["GWRS"].(map[string]interface{})["instrumentInfo"] != nil {
			// Capture stock short term outlook from script tag
			shortTermOutlook := fmt.Sprintf("%v", contextObjectAsStruct.(map[string]interface{})["dispatcher"].(map[string]interface{})["stores"].(map[string]interface{})["ResearchPageStore"].(map[string]interface{})["technicalInsights"].(map[string]interface{})["GWRS"].(map[string]interface{})["instrumentInfo"].(map[string]interface{})["technicalEvents"].(map[string]interface{})["shortTermOutlook"].(map[string]interface{})["direction"])
			currStock.StockPerformanceOutlookShort = shortTermOutlook

			// Capture stock short term outlook from script tag
			midTermOutlook := fmt.Sprintf("%v", contextObjectAsStruct.(map[string]interface{})["dispatcher"].(map[string]interface{})["stores"].(map[string]interface{})["ResearchPageStore"].(map[string]interface{})["technicalInsights"].(map[string]interface{})["GWRS"].(map[string]interface{})["instrumentInfo"].(map[string]interface{})["technicalEvents"].(map[string]interface{})["intermediateTermOutlook"].(map[string]interface{})["direction"])
			currStock.StockPerformanceOutlookMid = midTermOutlook

			// Capture stock short term outlook from script tag
			longTermOutlook := fmt.Sprintf("%v", contextObjectAsStruct.(map[string]interface{})["dispatcher"].(map[string]interface{})["stores"].(map[string]interface{})["ResearchPageStore"].(map[string]interface{})["technicalInsights"].(map[string]interface{})["GWRS"].(map[string]interface{})["instrumentInfo"].(map[string]interface{})["technicalEvents"].(map[string]interface{})["longTermOutlook"].(map[string]interface{})["direction"])
			currStock.StockPerformanceOutlookLong = longTermOutlook

			// Capture Sector from script tag
			sector := fmt.Sprintf("%v", contextObjectAsStruct.(map[string]interface{})["dispatcher"].(map[string]interface{})["stores"].(map[string]interface{})["ResearchPageStore"].(map[string]interface{})["technicalInsights"].(map[string]interface{})["GWRS"].(map[string]interface{})["instrumentInfo"].(map[string]interface{})["technicalEvents"].(map[string]interface{})["sector"])
			currStock.Sector = sector

			// Capture sector short term outlook from script tag
			secShortTermOutlook := fmt.Sprintf("%v", contextObjectAsStruct.(map[string]interface{})["dispatcher"].(map[string]interface{})["stores"].(map[string]interface{})["ResearchPageStore"].(map[string]interface{})["technicalInsights"].(map[string]interface{})["GWRS"].(map[string]interface{})["instrumentInfo"].(map[string]interface{})["technicalEvents"].(map[string]interface{})["shortTermOutlook"].(map[string]interface{})["sectorDirection"])
			currStock.SectorPerformanceOutlookShort = secShortTermOutlook

			// Capture sector short term outlook from script tag
			secMidTermOutlook := fmt.Sprintf("%v", contextObjectAsStruct.(map[string]interface{})["dispatcher"].(map[string]interface{})["stores"].(map[string]interface{})["ResearchPageStore"].(map[string]interface{})["technicalInsights"].(map[string]interface{})["GWRS"].(map[string]interface{})["instrumentInfo"].(map[string]interface{})["technicalEvents"].(map[string]interface{})["intermediateTermOutlook"].(map[string]interface{})["sectorDirection"])
			currStock.SectorPerformanceOutlookMid = secMidTermOutlook

			// Capture sector short term outlook from script tag
			secLongTermOutlook := fmt.Sprintf("%v", contextObjectAsStruct.(map[string]interface{})["dispatcher"].(map[string]interface{})["stores"].(map[string]interface{})["ResearchPageStore"].(map[string]interface{})["technicalInsights"].(map[string]interface{})["GWRS"].(map[string]interface{})["instrumentInfo"].(map[string]interface{})["technicalEvents"].(map[string]interface{})["longTermOutlook"].(map[string]interface{})["sectorDirection"])
			currStock.SectorPerformanceOutlookLong = secLongTermOutlook
		}

		// Capture current price
		priceS := e.ChildText(`#quote-header-info > div:nth-child(3) > div > div > span:first-child`)
		priceF, err := strconv.ParseFloat(priceS, 64)
		if err != nil {
			log.Fatal("Couldn't parse price for ", currStock.TickerSymbol, ": ", err)
		}
		currStock.CurrentPrice = priceF

		// Capture Market Cap
		currStock.MarketCap = e.ChildText(`td[data-test="MARKET_CAP-value"]`)
	})

	c.Visit("https://finance.yahoo.com/quote/GWRS?p=GWRS")
	c.Wait()

	for _, v := range stockData {
		fmt.Println(v.TickerSymbol, "-", v.Sector)
		fmt.Println(v.TickerSymbol, "-Sector Perf Short-", v.SectorPerformanceOutlookShort)
		fmt.Println(v.TickerSymbol, "-Sector Perf Mid-", v.SectorPerformanceOutlookMid)
		fmt.Println(v.TickerSymbol, "-Sector Perf Long-", v.SectorPerformanceOutlookLong)
		fmt.Println(v.TickerSymbol, "-Stock Perf Short-", v.StockPerformanceOutlookShort)
		fmt.Println(v.TickerSymbol, "-Stock Perf Mid-", v.StockPerformanceOutlookMid)
		fmt.Println(v.TickerSymbol, "-Stock Perf Long-", v.StockPerformanceOutlookLong)
	}
}