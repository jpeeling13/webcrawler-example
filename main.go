package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Post struct {
	Url            string
	Title          string
	User           string
	CommentsURL    string
	CrawledAt      time.Time
	DescriptionSrc string
}

var Posts []Post

func main() {
	// Get the src at reddit.com/r/programming

	httpRes, err := http.Get("https://www.reddit.com")

	if err != nil {
		log.Fatal("Could not reach reddit: ", err)
	}

	if httpRes.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", httpRes.StatusCode, httpRes.Status)
	}

	fmt.Println(httpRes.StatusCode)

	defer httpRes.Body.Close()
	initialSrc, err := goquery.NewDocumentFromReader(httpRes.Body)
	if err != nil {
		log.Fatal("Could not resolve response: ", err)
	}

	firstPost := Post{}
	firstPostSrc := initialSrc.Find(".scrollerItem").First()
	fmt.Println(firstPostSrc.Html())

	firstPost.CrawledAt = time.Now()
	firstPost.Title = firstPostSrc.Find("._eYtD2XCVieq6emjKBH3m").Text()
	firstPost.Url = firstPostSrc.Find(".SQnoC3ObvgnGjWt90zD9Z").AttrOr("href", "")
	userHtml, err := firstPostSrc.Find(".oQctV4n0yUb0uiHDdGnmE").Html()
	if err != nil {
		fmt.Println("User selector is wrong")
		log.Fatal(err)
	}
	fmt.Println("The User Html: ", userHtml)
	firstPost.User = userHtml

	fmt.Printf("First Post - Title: %v, Url: %v, User: %v", firstPost.Title, firstPost.Url, firstPost.User)

	// use goquery to get the first Post
	// Visit the first post to get the Description Src

	// ---- REPEAT ----
	// start hitting the API to load posts (starting with first ID in the initial HTML)
	// convert the returned JSON to posts (skip sponsored content)
	// Visit each permalink URL to get the Description Src
	// ---- END REPEAT ----

}
