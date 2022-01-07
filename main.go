package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gyujae/jobscrapper_backend/utils"
)

var koreaWebsites = map[string]string{
	"사람인":  "https://www.saramin.co.kr/",
	"잡코리아": "https://www.jobkorea.co.kr",
}

type Job struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Company   string `json:"company"`
	Condition string `json:"condition"`
	URL       string `json:"url"`
	Site      string `json:"site"`
}

func main() {
	var jobs []Job
	totalPage := getPagination("python", "잡코리아")
	c := make(chan []Job)
	for i := 1; i <= totalPage; i++ {
		go getJobData("python", "잡코리아", i, c)
	}
	for i := 1; i <= totalPage; i++ {
		jobData := <-c
		jobs = append(jobs, jobData...)
	}
	fmt.Println(jobs)
}

func getJobData(query string, website string, pageNum int, mainC chan<- []Job) {
	var jobs []Job
	c := make(chan Job)
	switch website {

	case "사람인":
		keywordURL := "https://www.saramin.co.kr/zf_user/search/recruit?search_area=main&search_done=y&search_optional_item=n&searchType=search&searchword=" + query + "&recruitPage=" + strconv.Itoa(pageNum) + "&recruitSort=relation&recruitPageCount=40&inner_com_type=&company_cd=0%2C1%2C2%2C3%2C4%2C5%2C6%2C7%2C9%2C10&show_applied=&quick_apply=&except_read=&mainSearch=n"
		doc := getDocument(keywordURL)
		jobCard := doc.Find(".item_recruit")
		jobCard.Each(func(i int, s *goquery.Selection) {
			go extractedJobData(s, c, website)
		})
		for i := 0; i < jobCard.Length(); i++ {
			jobData := <-c
			jobs = append(jobs, jobData)
		}
		mainC <- jobs

	case "잡코리아":
		keywordURL := "https://www.jobkorea.co.kr/Search/?stext=" + query + "&Page_No=" + strconv.Itoa(pageNum)
		doc := getDocument(keywordURL)
		jobCard := doc.Find(".list-default .clear")
		jobCard.Each(func(i int, s *goquery.Selection) {
			go extractedJobData(s, c, website)
		})
		for i := 0; i < jobCard.Length(); i++ {
			jobData := <-c
			jobs = append(jobs, jobData)
		}
		mainC <- jobs

	}
}

func extractedJobData(s *goquery.Selection, c chan<- Job, website string) {
	switch website {

	case "사람인":
		id, _ := s.Attr("value")
		title, _ := s.Find("a.data_layer").Attr("title")
		url, _ := s.Find("a.data_layer").Attr("href")
		url = koreaWebsites[website] + url
		company := s.Find(".area_corp .corp_name").Text()
		condition := cleanString(s.Find(".job_condition").Text())
		c <- Job{ID: id, Title: title, Company: company, Condition: condition, URL: url, Site: website}

	case "잡코리아":
		id, _ := s.Find(".list-post").Attr("data-gno")
		company, _ := s.Find(".post a.name.dev_view").Attr("title")
		title, _ := s.Find(".post .post-list-info a.title.dev_view").Attr("title")
		url, _ := s.Find(".post .post-list-info a.title.dev_view").Attr("href")
		url = koreaWebsites[website] + url
		condition := cleanString(s.Find(".post .post-list-info p.option").Text())
		c <- Job{ID: id, Title: title, Company: company, Condition: condition, URL: url, Site: website}

	}
}

func getPagination(query string, website string) int {
	switch website {

	case "사람인":
		keywordURL := "https://www.saramin.co.kr/zf_user/search/recruit?search_area=main&search_done=y&search_optional_item=n&searchType=search&searchword=" + query + "&recruitPage=1&recruitSort=relation&recruitPageCount=100&inner_com_type=&company_cd=0%2C1%2C2%2C3%2C4%2C5%2C6%2C7%2C9%2C10&show_applied=&quick_apply=&except_read=&mainSearch=n"
		doc := getDocument(keywordURL)
		count := doc.Find("#recruit_info div.header span").Text()
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		submatchall := re.FindAllString(count, -1)
		result, err := strconv.Atoi(strings.Replace(submatchall[0], ",", "", 1))
		utils.CheckErr(err)
		return getPageCeil(result, 100)

	case "잡코리아":
		keywordURL := "https://www.jobkorea.co.kr/Search/?stext=" + query
		doc := getDocument(keywordURL)
		count := doc.Find("#content div div div.cnt-list-wrap div div.recruit-info div.list-filter-wrap p strong").Text()
		result, err := strconv.Atoi(strings.Replace(count, ",", "", 1))
		utils.CheckErr(err)
		return getPageCeil(result, 20)
	}

	return -1
}

func getPageCeil(num int, ceil int) int {
	if (num % ceil) != 0 {
		return (num / ceil) + 1
	}
	return num / ceil
}

func getDocument(keywordURL string) *goquery.Document {
	res, err := http.Get(keywordURL)
	utils.CheckErr(err)
	defer res.Body.Close()
	utils.CheckResponseCode(res)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	utils.CheckErr(err)

	return doc
}

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
