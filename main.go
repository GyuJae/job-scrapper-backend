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

// var koreaWebsites = map[string]string{
// 	"사람인":  "https://www.saramin.co.kr/zf_user/",
// 	"잡코리아": "https://www.jobkorea.co.kr/",
// 	"커리어":  "https://www.career.co.kr/Default.asp",
// 	"인크루트": "https://www.incruit.com/",
// 	"피플앤잡": "https://www.peoplenjob.com/jobs",
// }

func main() {
	pages := getPagination("python", "원티드")
	fmt.Println(pages)
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
		fmt.Println(keywordURL)
		doc := getDocument(keywordURL)
		count := doc.Find("#content div div div.cnt-list-wrap div div.recruit-info div.list-filter-wrap p strong").Text()
		result, err := strconv.Atoi(strings.Replace(count, ",", "", 1))
		utils.CheckErr(err)
		return getPageCeil(result, 20)
	case "커리어":
		keywordURL := "https://search.career.co.kr/jobs?kw=" + query
		fmt.Println(keywordURL)
		doc := getDocument(keywordURL)
		count := doc.Find("#container div div div div div.totSehWrap div.totSehLt div.txContBoxWrap.clearfix div div.txTit.MT45 small").Text()
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		result, err := strconv.Atoi(re.FindAllString(count, -1)[0])
		utils.CheckErr(err)
		return getPageCeil(result, 10)
	case "인크루트":
		keywordURL := "https://search.incruit.com/list/search.asp?col=job&il=y&kw=" + query
		fmt.Println(keywordURL)
		doc := getDocument(keywordURL)
		count := doc.Find("#content div.section h2 span.numall").Text()
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		result, err := strconv.Atoi(re.FindAllString(count, -1)[0])
		utils.CheckErr(err)
		return getPageCeil(result, 20)
	case "피플앤잡":
		keywordURL := "https://www.peoplenjob.com/jobs?field=all&q=" + query
		fmt.Println(keywordURL)
		doc := getDocument(keywordURL)
		count := doc.Find("#content-main div div.page-header div.page-title div span").Text()
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		result, err := strconv.Atoi(re.FindAllString(count, -1)[0])
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
