package main


import (
    "net/http"
    "log"
    iconv "github.com/djimenez/iconv-go"
    goquery "github.com/PuerkitoBio/goquery"
    "strings"
    "time"
    "encoding/json"
    "os"
)

type MealDay struct {
    Date        string  `json:"date"`
    Breakfast   string  `json:"breakfast"`
    Lunch       string  `json:"lunch"`
    Dinner      string  `json:"dinner"`
    Supper      string  `json:"supper"`
}

func main() {

    args := os.Args[1:]
    if (len(args) != 1) {
        log.Fatal("Invalid arguments")
    }

    res, err := http.Get("http://www.leijonacatering.fi/ruokalista_varuskunta.php")
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

    sel := doc.Find("table tr").First().Find("td").Last();
    dateSl := strings.Split(sel.Text(), "-")
    endDateStr := strings.TrimSpace(strings.Replace(dateSl[len(dateSl) - 1], ".", "/", -1))
    date, err := time.Parse("2/1/2006", endDateStr)
    if err != nil {
        log.Fatal(err)
    }
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
        bf[i] = bfSel.Eq(i+1).Text()
        lunch[i] = lunchSel.Eq(i+1).Text()
        dinner[i] = dinnerSel.Eq(i+1).Text()
        supper[i] = supperSel.Eq(i+1).Text()
    }
    
    mealdays := make([]MealDay, 7)
    dateFormat := "2-1-2006"
    for i := 0; i < 7; i++ {
        mealdays[i] = MealDay{  date.Format(dateFormat),
                                bf[i],
                                lunch[i],
                                dinner[i],
                                supper[i] }
        date = date.AddDate(0,0,1)
    }

    /* jsonMeals, err := json.Marshal(mealdays)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s \n", jsonMeals) */

    outFile, err := os.Create(args[0])
    if err != nil {
        log.Fatal(err)
    }
    defer outFile.Close()

    enc := json.NewEncoder(outFile)
    enc.Encode(mealdays)
}
