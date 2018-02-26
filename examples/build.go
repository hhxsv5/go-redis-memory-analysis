package main

import (
	"flag"
	"fmt"
	"github.com/hhxsv5/go-redis-memory-analysis"
	"strings"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "The host of redis")
	port := flag.Uint("port", 6379, "The port of redis")
	password := flag.String("password", "", "The password of redis")
	prefixes := flag.String("prefixes", "#,:", "The prefixes list of redis key, be split by ',', special pattern characters need to escape by '\\'")
	reportPath := flag.String("reportPath", "./reports", "The csv file path of analysis result")

	flag.Parse()

	analysis := gorma.NewAnalysis()
	err := analysis.Open(*ip, uint16(*port), *password)
	defer analysis.Close()
	if err != nil {
		fmt.Println("something wrong:", err)
		return
	}

	analysis.Start(strings.Split(*prefixes, ","))

	analysis.SaveReports(*reportPath)

	fmt.Println("Done")
}
