// Testing project main.go
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var recipe []struct {
	Name        string `json:"name"`
	Ingredients []struct {
		Item   string `json:"item"`
		Amount string `json:"amount"`
		Unit   string `json:"unit"`
	} `json:"ingredients"`
}

type RecipeCount struct {
	Name            string
	IngredientCount int
	Count           int
}

type fridgeRecord struct {
	Item   string
	Amount int
	Unit   string
	Useby  string
}

func TextStr(csv []fridgeRecord) string {
	var rCount, fCount RecipeCount
	var ingCount int = 0
	var year, month, day int
	var tmonth time.Month = 2
	t := time.Now()
	for _, recip := range recipe {
		for _, fridge := range csv {
			rCount.IngredientCount = 0
			rCount.Count = 0
			for _, ingredient := range recip.Ingredients {
				ingCount, _ = strconv.Atoi(ingredient.Amount)
				var dateSlice = strings.Split(fridge.Useby, "/")
				if len(dateSlice) > 1 {
					year, _ = strconv.Atoi(dateSlice[2])
					month, _ = strconv.Atoi(dateSlice[1])
					day, _ = strconv.Atoi(dateSlice[0])

					tmonth = time.Month(month)
					//time.Month
					fmt.Printf(" month is %d ", month)
					var Time = time.Date(year, tmonth, day, 23, 59, 59, 0, time.UTC)

					if fridge.Item == ingredient.Item && fridge.Amount >= ingCount && t.Before(Time) {
						rCount.IngredientCount += ingCount
						rCount.Count += fridge.Amount
						rCount.Name = recip.Name
						fmt.Printf("MATCH : %s is recip name %s : %s and unit amount %s in fridge %s \n", Time, fridge.Useby, ingredient.Item, ingredient.Amount, fridge.Amount)
					}

				} else {
					//fmt.Printf("NON MATCH: %s is recip name %s : %s  and unit amount %s in fridge %s \n", recip.Name,fridge.Item, ingredient.Item,ingredient.Amount,fridge.Amount)
				}
			}
			// end of ingredients now compare and keep with highest
			if rCount.IngredientCount > fCount.IngredientCount {
				fCount = rCount
			}
		}
		fmt.Printf(" counts %s and rCount %d \n", fCount.Name, fCount.IngredientCount)
		fmt.Printf(" rcounts %s and rCount %d \n", rCount.Name, rCount.IngredientCount)
	}
	if len(fCount.Name) > 3{
	return fCount.Name
	}else{
		return "Order Takeout";
	}
	
}

func parseCsvFile(csvfile multipart.File) ([]fridgeRecord, error) {

	reader := csv.NewReader(csvfile)

	reader.FieldsPerRecord = -1

	rawCSVdata, err := reader.ReadAll()

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// sanity check, display to standard output
	/*       for _, each := range rawCSVdata {
	                 //fmt.Printf("val 1 : %s and val 2 : %s\n", each[0], each[1])
	         }
	*/
	// now, safe to move raw CSV data to struct

	var oneRecord fridgeRecord

	var allRecords []fridgeRecord

	for _, each := range rawCSVdata {
		oneRecord.Item = each[0]
		oneRecord.Amount, _ = strconv.Atoi(each[1])
		oneRecord.Unit = each[2]
		oneRecord.Useby = each[3]
		allRecords = append(allRecords, oneRecord)
	}

	return allRecords, nil

}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t, _ := template.ParseFiles("layout/main.gtpl")
		t.Execute(w, nil)
	} else {
		file, err := UplHandler(w, r, "jsonfile")

		if file != nil {
			jsonParser := json.NewDecoder(file)
			if err = jsonParser.Decode(&recipe); err != nil {
				fmt.Println(err.Error())
			}

			fmt.Println(reflect.TypeOf(recipe))
			if len(recipe) > 0 {
				//fmt.Printf("return Name is %s and ingred is %s", recipe[0].Name, recipe[0].Ingredients[0].Item)
			}
			file2, err := UplHandler(w, r, "csvfile")
			if file2 != nil {
				csv, err := parseCsvFile(file2)
				//fmt.Printf("CSV is %s ", file)

				if len(csv) > 0 && err == nil {
					var Text string = TextStr(csv)
					//fmt.Printf(" the text is %s ",Text)
					//w.Write([]byte("<h1>  </h1>"))
					w.Write([]byte(Text))

					//fmt.Printf("return Name is %s and ingred is %s", csv[0].Amount, csv[0].Item)
				}
			}
			if err != nil {
				//w.Write([]byte(err.Error()))
				//fmt.Printf("return Name is %s ", err.Error())
			}

		}

		//r.ParseForm()
		// logic part of log in
		//fmt.Println("username:", r.Form["username"])
		//fmt.Println("password:", r.Form["password"])
	}
	// w.Write([]byte("Hello, World!"))
}

func UplHandler(w http.ResponseWriter, r *http.Request, n string) (multipart.File, error) {

	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile(n)
	return file, err
}

func main() {
	fmt.Println("Test! started")
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	http.ListenAndServe(":8800", r)
}
