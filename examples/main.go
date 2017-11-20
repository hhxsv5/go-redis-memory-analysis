package main

import (
	"fmt"
	. "github.com/hhxsv5/go-redis-memory-analysis"
	. "github.com/hhxsv5/go-redis-memory-analysis/storages"
)

func main() {
	redis, err := NewRedisClient("127.0.0.1", 6379, "")
	if err != nil {
		fmt.Println("Connect redis fail", err)
		return
	}
	defer redis.Close()

	analysis := NewAnalysis(redis)

	//Scan the keys which can be split by '#' ':'
	//Special pattern characters need to escape by '\'
	analysis.Start([]string{"#", ":"}, 3000)

	//Find the csv file in default target folder: ./reports
	//CSV file name format: redis-analysis-{host:port}-{db}.csv
	//The keys order by count desc
	analysis.SaveReports("./reports")
}
