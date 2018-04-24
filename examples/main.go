package main

import (
	"fmt"
	. "github.com/hhxsv5/go-redis-memory-analysis"
)

func main() {
	analysis := NewAnalysis()
	//Open redis: 127.0.0.1:6379 without password
	err := analysis.Open("127.0.0.1", 6379, "")
	defer analysis.Close()
	if err != nil {
		fmt.Println("something wrong:", err)
		return
	}

	//Scan the keys which can be split by '#' ':'
	//Special pattern characters need to escape by '\'
	analysis.Start([]string{"#", ":"})

	//Find the csv file in default target folder: ./reports
	//CSV file name format: redis-analysis-{host:port}-{db}.csv
	//The keys order by count desc
	err = analysis.SaveReports("./reports")
	if err == nil {
		fmt.Println("done")
	} else {
		fmt.Println("error:", err)
	}
}
