## GoScrapper

GoRoutine을 활용한 **FAST SCRAPER**

## 명세

고루틴을 활용해 아래 3가지 기능을 `Concurrent`하게 처리합니다.

![goscrapper](./README.assets/goscrapper.jpg)

### 1. Page Hit

```go
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
```

### 2. Scrap (정보 추출)

```go
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
```

### 3. Write (csv 파일에 jobs 쓰기)

```go
func writeJob(ch chan<- error, w *csv.Writer, job extractedJob) {
	jobSlice := []string{"https://kr.indeed.com/viewjob?jk=" + job.id, job.title, job.location, job.summary}
	jwErr := w.Write(jobSlice)
	ch <- jwErr
```

## 느낀 점

### 1. 밴당했다

`concurrency` 이전의 코드로는 밴 당하지 않았는데, 이후 워낙 빠르게 접근하다보니
자주 많이 요청을 보내게 되어 서버에서 밴 당했다. 확실히 그 정도로 빠르다.

### 2. 엄청나게 빠르다.

`javascript` `ajax`를 통해서도 동시성의 위력을 느꼈지만 scrapper에서 직접 활용 해보니 `for loop`를 동시에 처리한 다는 것의 위력을 직접 체감했다. 컴퓨팅 자원이 허락한다면 모든 루프를 도는 시간이 원래는 sum of every case여야 하는데  worst case 하나만큼 밖에 걸리지 않는다...!

### 3. 수신 채널을 통한 블로킹이 중요하다.

의도적으로 수신 채널을 만들어서 블로킹을 유도하면 동시적 처리 과정이 끝난 뒤 다음 과정으로 넘어갈 때 유용할 것 같다. 
