package main

// Dichiarazione RW e RT senza sforzo con foglio Excel automatico
// https://www.youtube.com/watch?v=IttTqwAOk2s

// Foglio excel di esempio
// https://docs.google.com/spreadsheets/d/1bZHdkXXtCJhCA--OKJiDKlrRNT_Y8nJP/edit?gid=1998825627#gid=1998825627

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Record struct {
	Day      int
	Buy      bool
	Quantity float64
	Price    float64
}

type Value struct {
	Name       string
	DayStart   int
	DayEnd     int
	Days       int
	Quantity   float64
	PriceBegin float64
	PriceEnd   float64
}

type Database struct {
	Records    map[string][]Record
	FinalPrice map[string]float64
}

func NewDatabase() *Database {
	return &Database{
		Records:    make(map[string][]Record),
		FinalPrice: make(map[string]float64),
	}
}

func (r *Database) Push(name string, record Record) {
	_, ok := r.Records[name]
	if !ok {
		value := make([]Record, 0)
		value = append(value, record)
		r.Records[name] = value
	} else {
		r.Records[name] = append(r.Records[name], record)
	}
}

func (r *Database) Compute() []Value {
	results := make([]Value, 0)
	for key, value := range r.Records {
		values := Compute(key, value, r.FinalPrice[key])
		for _, value := range values {
			value.Name = key
			results = append(results, value)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Days > results[j].Days
	})
	return results
}

func Compute(name string, records []Record, finalPrice float64) []Value {
	result := make([]Value, 0)
	current := make([]Record, 0)
	for index := 0; index < len(records); index++ {

		// log.Printf("Processing (%s): %+v", name, records[index])

		if records[index].Buy {

			current = append(current, Record{
				Day:      records[index].Day,
				Buy:      records[index].Buy,
				Quantity: records[index].Quantity,
				Price:    records[index].Price,
			})

		} else {

			// log.Printf("Buy (%s): %+v", name, records[index])

			backIndex := len(current)
			for records[index].Quantity > 0 {
				backIndex = backIndex - 1

				// log.Printf("Analyzing: %+v", records[index].Quantity)

				if current[backIndex].Quantity <= records[index].Quantity {

					record := Value{
						DayStart:   current[backIndex].Day,
						DayEnd:     records[index].Day,
						Days:       records[index].Day - current[backIndex].Day,
						Quantity:   current[backIndex].Quantity,
						PriceBegin: current[backIndex].Price,
						PriceEnd:   records[index].Price,
					}

					// log.Printf("Established: %+v", record)
					result = append(result, record)

					records[index].Quantity = records[index].Quantity - record.Quantity
				} else {

					quantity := current[backIndex].Quantity - records[index].Quantity

					// Split in two
					record := Value{
						DayStart:   current[backIndex].Day,
						DayEnd:     records[index].Day,
						Days:       records[index].Day - current[backIndex].Day,
						Quantity:   records[index].Quantity,
						PriceBegin: current[backIndex].Price,
						PriceEnd:   records[index].Price,
					}
					result = append(result, record)
					// log.Printf("Established: %+v", record)

					current[backIndex].Quantity = quantity
					records[index].Quantity = 0

				}
			}

			current = current[0 : backIndex+1]

		}
	}

	lastDay := 365

	for _, item := range current {
		record := Value{
			DayStart:   item.Day,
			DayEnd:     lastDay,
			Days:       lastDay - item.Day,
			Quantity:   item.Quantity,
			PriceBegin: item.Price,
			PriceEnd:   finalPrice,
		}

		result = append(result, record)
	}

	return result

}

func DaysFromStartOfYear(dateString string) (int, error) {
	components := strings.Split(dateString, "/")
	day, err := strconv.Atoi(components[0])
	if err != nil {
		return 0, err
	}
	month, err := strconv.Atoi(components[1])
	if err != nil {
		return 0, err
	}
	year, err := strconv.Atoi(components[2])
	if err != nil {
		return 0, err
	}
	selectedDay := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	startOfYear := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	days := int(selectedDay.Sub(startOfYear).Hours() / 24)
	return days, nil // Adding 1 to include the current day
}

func main() {

	file, err := os.Open("dati.csv")
	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	// Closes the file
	defer file.Close()

	// The csv.NewReader() function is called in
	// which the object os.File passed as its parameter
	// and this creates a new csv.Reader that reads
	// from the file
	reader := csv.NewReader(file)

	// ReadAll reads all the records from the CSV file
	// and Returns them as slice of slices of string
	// and an error if any
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading records")
	}

	database := NewDatabase()

	// Loop to iterate through
	// and print each of the string slice
	for _, item := range records {

		name := item[1]
		if name == "" {
			continue
		}

		days, err := DaysFromStartOfYear(item[0])
		if err != nil {
			// log.Println("Error reading records (date): ", err)
			continue
		}

		// End Line
		if item[2] == "" {

			price, err := strconv.ParseFloat(item[3], 64)
			if err != nil {
				// log.Fatal("Error reading records (price): ", err)
				continue
			}

			database.FinalPrice[name] = price

		} else {

			quantity, err := strconv.ParseFloat(item[2], 64)
			if err != nil {
				// log.Println("Error reading records (quantity): ", err)
				continue
			}

			price, err := strconv.ParseFloat(item[3], 64)
			if err != nil {
				// log.Fatal("Error reading records (price): ", err)
				continue
			}

			sell := quantity < 0
			if sell {
				quantity = -quantity
			}

			record := Record{
				Day:      days,
				Buy:      !sell,
				Quantity: quantity,
				Price:    price,
			}

			// log.Printf("%s: %+v", name, record)

			database.Push(name, record)

		}

	}

	results := database.Compute()

	for _, item := range results {
		fmt.Print(color.RedString("%30s", item.Name))
		fmt.Print(color.CyanString("%10s", fmt.Sprintf("%+v", item.Quantity)))
		fmt.Print(color.CyanString("%10s", fmt.Sprintf("%+v", item.PriceBegin)))
		fmt.Print(color.CyanString("%10s", fmt.Sprintf("%+v", item.PriceEnd)))
		fmt.Print(color.CyanString("%10s", fmt.Sprintf("%+v", item.Days)))
		fmt.Println()
		// color.Cyan(fmt.Sprintf("%s (%+v), price: %+v => %+v, days: %d", item.Name, item.Quantity, item.PriceBegin, item.PriceEnd, item.Days))

		// color.Red()
		// log.Printf("%+v", item)
	}

}
