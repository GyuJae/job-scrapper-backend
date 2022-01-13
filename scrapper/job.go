package scrapper

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gyujae/jobscrapper_backend/utils"
)

var Websites = map[string]string{
	"사람인":  "https://www.saramin.co.kr",
	"잡코리아": "https://www.jobkorea.co.kr",
}

var WebsitesImages = map[string]string{
	"사람인":  "https://yt3.ggpht.com/ytc/AKedOLRnfuHawFrOlV0V9g7U2-AsqEYmqTYA7CFBlxqViQ=s900-c-k-c0x00ffffff-no-rj",
	"잡코리아": "https://image.edaily.co.kr/images/Photo/files/NP/S/2020/11/PS20111000013.gif",
}

type JobRest map[string][]Job

type Job struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Company   string `json:"company"`
	Condition string `json:"condition"`
	URL       string `json:"url"`
	Site      string `json:"site"`
}

func Filter(jobs []Job, site string) []Job {
	var result []Job
	for _, value := range jobs {
		if value.Site == site {
			result = append(result, value)
		}
	}
	return result
}

func SplitJobsBySite(keyword string) JobRest {
	jobs := JobScrapperMain(keyword)
	result := map[string][]Job{}
	for site := range Websites {
		jobResult := Filter(jobs, site)
		result[site] = jobResult
	}
	return result
}

func JobScrapperMain(keyword string) []Job {
	var jobs []Job
	c := make(chan []Job)
	for website := range Websites {
		go jobScrapper(c, keyword, website)
	}
	for i := 0; i < len(Websites); i++ {
		jobsItem := <-c
		jobs = append(jobs, jobsItem...)
	}

	return jobs
}

func jobScrapper(mainC chan<- []Job, query string, website string) {
	var jobs []Job
	totalPage := getPagination(query, website)
	c := make(chan []Job)
	if totalPage != -1 {
		for i := 1; i <= totalPage; i++ {
			go getJobData(query, website, i, c)
		}
		for i := 1; i <= totalPage; i++ {
			jobData := <-c
			jobs = append(jobs, jobData...)
		}
		mainC <- jobs
	}
}

func getJobData(query string, website string, pageNum int, mainC chan<- []Job) {
	var jobs []Job
	c := make(chan Job)
	switch website {

	case "사람인":
		keywordURL := "https://www.saramin.co.kr/zf_user/search/recruit?search_area=main&search_done=y&search_optional_item=n&searchType=search&searchword=" + query + "&recruitPage=" + strconv.Itoa(pageNum) + "&recruitSort=relation&recruitPageCount=100&inner_com_type=&company_cd=0%2C1%2C2%2C3%2C4%2C5%2C6%2C7%2C9%2C10&show_applied=&quick_apply=&except_read=&mainSearch=n"
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
		jobCard := doc.Find("#content div div div.cnt-list-wrap div div.recruit-info div.lists div div.list-default ul li")
		jobCard.Each(func(i int, s *goquery.Selection) {
			go extractedJobData(s, c, website)
		})
		for i := 0; i < jobCard.Length(); i++ {
			jobData := <-c
			jobs = append(jobs, jobData)
		}
		mainC <- jobs
	case "indeed":
		keywordURL := "https://kr.indeed.com/jobs?q=" + query + "&limit=50&start=" + strconv.Itoa((pageNum-1)*50)
		doc := getDocument(keywordURL)
		// jobCard := doc.Find("div.slider_container div div.slider_item div.job_seen_beacon")
		jobCard := doc.Find("#mosaic-provider-jobcards a.tapItem")
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
		url = Websites[website] + url
		company := s.Find(".area_corp .corp_name").Text()
		condition := cleanString(s.Find(".job_condition").Text())
		if id != "" {
			c <- Job{ID: id, Title: title, Company: company, Condition: condition, URL: url, Site: website}
		}

	case "잡코리아":
		id, _ := s.Find(".list-post").Attr("data-gno")
		company, _ := s.Find(".post a.name.dev_view").Attr("title")
		title, _ := s.Find(".post .post-list-info a.title.dev_view").Attr("title")
		url, _ := s.Find(".post .post-list-info a.title.dev_view").Attr("href")
		url = Websites[website] + url
		condition := cleanString(s.Find(".post .post-list-info p.option").Text())
		c <- Job{ID: id, Title: title, Company: company, Condition: condition, URL: url, Site: website}
	case "indeed":
		id, _ := s.Attr("data-jk")
		title := s.Find("table.jobCard_mainContent tbody tr td div.heading4.color-text-primary.singleLineTitle.tapItem-gutter h2 span").Text()
		company := s.Find("div.heading6.company_location.tapItem-gutter span.companyName").Text()
		condition := s.Find("div.heading6.company_location.tapItem-gutter div.companyLocation").Text()
		url := Websites[website] + "채용보기?jk=" + id
		if title != "" {
			c <- Job{ID: id, Title: title, Company: company, Condition: condition, URL: url, Site: website}
		}

	}
}

func getPagination(query string, website string) int {
	switch website {
	case "사람인":
		keywordURL := "https://www.saramin.co.kr/zf_user/search/recruit?search_area=main&search_done=y&search_optional_item=n&searchType=search&searchword=" + query + "&recruitPage=1&recruitSort=relation&recruitPageCount=100&inner_com_type=&company_cd=0%2C1%2C2%2C3%2C4%2C5%2C6%2C7%2C9%2C10&show_applied=&quick_apply=&except_read=&mainSearch=n"
		doc := getDocument(keywordURL)
		if doc == nil {
			return -1
		}
		count := doc.Find("#recruit_info div.header span").Text()
		count = cleanString(count)
		if count == "" {
			return -1
		}
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		submatchall := re.FindAllString(count, -1)
		if submatchall == nil {
			return -1
		}
		if len(submatchall) == 0 {
			return -1
		}
		result, err := strconv.Atoi(strings.Replace(submatchall[0], ",", "", 1))
		utils.CheckErr(err)
		return getPageCeil(result, 100)
	case "잡코리아":
		keywordURL := "https://www.jobkorea.co.kr/Search/?stext=" + query
		doc := getDocument(keywordURL)
		if doc == nil {
			return -1
		}
		count := doc.Find("#content div div div.cnt-list-wrap div div.recruit-info div.list-filter-wrap p strong").Text()
		count = cleanString(count)
		if count == "" {
			return -1
		}
		result, err := strconv.Atoi(strings.Replace(count, ",", "", 1))
		utils.CheckErr(err)
		return getPageCeil(result, 20)
	case "indeed":
		keywordURL := "https://kr.indeed.com/jobs?q=" + query + "&limit=50&start=0"
		doc := getDocument(keywordURL)
		if doc == nil {
			return -1
		}
		count := doc.Find("#searchCountPages").Text()
		count = cleanString(count)
		if count == "" {
			return -1
		}
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		submatchall := re.FindAllString(count, -1)
		if submatchall == nil {
			return -1
		}
		if len(submatchall) == 0 || len(submatchall) == 1 {
			return -1
		}
		result, err := strconv.Atoi(strings.Replace(submatchall[1], ",", "", 1))
		utils.CheckErr(err)
		return getPageCeil(result, 50)

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
