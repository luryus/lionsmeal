package main

import (
	"encoding/json"
	goquery "github.com/PuerkitoBio/goquery"
	iconv "github.com/djimenez/iconv-go"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type MealDay struct {
	Date      int64  `json:"date"`
	Breakfast string `json:"breakfast"`
	Lunch     string `json:"lunch"`
	Dinner    string `json:"dinner"`
	Supper    string `json:"supper"`
}

func parseMenuEndingDate(dateStr string) time.Time {
	// dateStr should be like "dd.mm.-dd.mm.yyyy"

	idx := strings.LastIndex(dateStr, ".")                 // find the last dot in the string
	year, err := strconv.ParseUint(dateStr[idx+1:], 10, 0) // and parse the number after it
	if err != nil {
		log.Fatal("Could not parse year: ", err)
	}

	dateStr = dateStr[:idx]                                 // remove the year from the date string
	idx = strings.LastIndex(dateStr, ".")                   // find last dot again
	month, err := strconv.ParseUint(dateStr[idx+1:], 10, 0) // and parse the number after it == month
	if err != nil {
		log.Fatal("Could not parse month: ", err)
	}

	dateStr = dateStr[:idx]                                        // remove the month from the date string
	day, err := strconv.ParseUint(dateStr[len(dateStr)-2:], 10, 0) // try to parse unsigned number using two last characters
	if err != nil {                                                // if fails, probably has dash for the first character
		day, err = strconv.ParseUint(dateStr[len(dateStr)-1:], 10, 0) // parse unsigned number using only the last character
		if err != nil {
			log.Fatal("Could not parse day: ", err)
		}
	}

	return time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
}

func parseMenu(url string) []MealDay {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	// convert to UTF-8
	utfBody, err := iconv.NewReader(res.Body, "iso-8859-1", "utf-8")
	if err != nil {
		log.Fatal(err)
	}

	// parse date
	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		log.Fatal(err)
	}

	sel := doc.Find("table tr").First().Find("td").Last()
	endDateStr := strings.TrimSpace(sel.Text())
	date := parseMenuEndingDate(endDateStr)

	weekDur, err := time.ParseDuration("-144h")
	date = date.Add(weekDur)

	// make slices for breakfast, lunch, dinner, and supper
	var bf, lunch, dinner, supper []string
	bf = make([]string, 7)
	lunch = make([]string, 7)
	dinner = make([]string, 7)
	supper = make([]string, 7)

	// parse rows for each meal
	bfSel := doc.Find("table tr").Eq(2).Find("td")
	lunchSel := doc.Find("table tr").Eq(3).Find("td")
	dinnerSel := doc.Find("table tr").Eq(4).Find("td")
	supperSel := doc.Find("table tr").Eq(5).Find("td")

	for i := 0; i < 7; i++ {
		bf[i] = strings.TrimSpace(bfSel.Eq(i + 1).Text())
		lunch[i] = strings.TrimSpace(lunchSel.Eq(i + 1).Text())
		dinner[i] = strings.TrimSpace(dinnerSel.Eq(i + 1).Text())
		supper[i] = strings.TrimSpace(supperSel.Eq(i + 1).Text())
	}

	mealdays := make([]MealDay, 7)
	for i := 0; i < 7; i++ {
		mealdays[i] = MealDay{date.Unix(),
			bf[i],
			lunch[i],
			dinner[i],
			supper[i]}
		date = date.AddDate(0, 0, 1)
	}

	return mealdays
}

func main() {

	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatal("Invalid arguments")
	}

	mealdays1 := parseMenu("http://www.leijonacatering.fi/ruokalista_varuskunta.php")
	mealdays2 := parseMenu("http://www.leijonacatering.fi/ruokalista_varuskunta_seur.php")

	alldays := append(mealdays1, mealdays2...)

	outFile, err := os.Create(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	enc := json.NewEncoder(outFile)
	enc.Encode(alldays)
}
