package utils

import (
	"log"
	"net/http"
)

func CheckErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func CheckResponseCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatal("❌ Request failed with Status:", res.Status)
	}
}
