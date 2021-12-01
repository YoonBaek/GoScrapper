package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id       string
	title    string
	location string
	// salary   string
	summary string
}

const NUM_PAGES = 10

var baseURL string = "https://kr.indeed.com/jobs?q=python&start="

// Scrape recruit sote
func Scrap() {
	var jobs = []extractedJob{}
	ch := make(chan []extractedJob)
	for i := 0; i < NUM_PAGES; i++ {
		go getPage(ch, i)
	}
	for i := 0; i < NUM_PAGES; i++ {
		extractedJobs := <-ch
		jobs = append(jobs, extractedJobs...)
	}
	fmt.Println("Done. extracted", len(jobs))
}

func getPage(mainCh chan<- []extractedJob, page int) {
	var jobs = []extractedJob{}
	ch := make(chan extractedJob)
	pageURL := baseURL + strconv.Itoa(page*10)
	fmt.Println("Requesting", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".tapItem")

	searchCards.Each(func(i int, s *goquery.Selection) {
		go extractJob(ch, s)
	})
	for i := 0; i < searchCards.Length(); i++ {
		jobs = append(jobs, <-ch)
	}
	mainCh <- jobs
}

func extractJob(ch chan<- extractedJob, s *goquery.Selection) {
	id, _ := s.Attr("data-jk")
	title := stringsTidy(s.Find("h2 > span").Text())
	location := stringsTidy(s.Find(".companyLocation").Text())
	summary := stringsTidy(s.Find(".job-snippet").Text())
	ch <- extractedJob{
		id:       id,
		title:    title,
		location: location,
		summary:  summary,
	}
}

// func getPages(url string) int {
// 	pages := 0
// 	res, err := http.Get(url)
// 	checkErr(err)
// 	checkCode(res)
// 	defer res.Body.Close()

// 	doc, err := goquery.NewDocumentFromReader(res.Body)
// 	checkErr(err)

// 	doc.Find(".pagination").Each(func(i int, selection *goquery.Selection) {
// 		pages = selection.Find("a").Length()
// 	})

// 	return pages
// }

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with stats:", res.StatusCode)
	}
}

func stringsTidy(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)
	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"ID", "Title", "Location", "Summary"}

	wErr := w.Write(headers)
	checkErr(wErr)

	jobCh := make(chan []string)

	for _, job := range jobs {
		go writeJob(jobCh, job)
	}
	for i := 0; i < len(jobs); i++ {
		w.Write(<-jobCh)
	}
}

// Save Jobs struct object
func writeJob(ch chan<- []string, job extractedJob) {
	jobSlice := []string{"https://kr.indeed.com/viewjob?jk=" + job.id, job.title, job.location, job.summary}
	ch <- jobSlice
}
