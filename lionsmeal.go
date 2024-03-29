/*
The MIT License (MIT)

Copyright (c) 2015 Lauri Koskela

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

package main

import (
	"encoding/json"
	goquery "github.com/PuerkitoBio/goquery"
	iconv "github.com/djimenez/iconv-go"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type mealDay struct {
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

func removeEmptyLines(input string) string {
	rex := regexp.MustCompile("\\n\\s+\\n")
	return rex.ReplaceAllLiteralString(input, "\n")
}

func parseMenu(url string) []mealDay {
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
		bf[i] = removeEmptyLines(strings.TrimSpace(bfSel.Eq(i + 1).Text()))
		lunch[i] = removeEmptyLines(strings.TrimSpace(lunchSel.Eq(i + 1).Text()))
		dinner[i] = removeEmptyLines(strings.TrimSpace(dinnerSel.Eq(i + 1).Text()))
		supper[i] = removeEmptyLines(strings.TrimSpace(supperSel.Eq(i + 1).Text()))
	}

	mealdays := make([]mealDay, 7)
	for i := 0; i < 7; i++ {
		mealdays[i] = mealDay{date.Unix(),
			bf[i],
			lunch[i],
			dinner[i],
			supper[i]}
		date = date.AddDate(0, 0, 1)
	}

	return mealdays
}

func main() {

	mealdays1 := parseMenu("http://www.leijonacatering.fi/ruokalista_varuskunta.php")
	mealdays2 := parseMenu("http://www.leijonacatering.fi/ruokalista_varuskunta_seur.php")

	alldays := append(mealdays1, mealdays2...)

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(alldays)
}
