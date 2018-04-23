package main

import (
	"flag"
	"fmt"
	"strings"
	"github.com/hhxsv5/go-redis-memory-analysis"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "The host of redis")
	port := flag.Uint("port", 6379, "The port of redis")
	password := flag.String("password", "", "The password of redis")
	rdb := flag.String("rdb", "", "The rdb file of redis")
	prefixes := flag.String("prefixes", "#//:", "The prefixes list of redis key, be split by '//', special pattern characters need to escape by '\\'")
	reportPath := flag.String("reportPath", "./reports", "The csv file path of analysis result")

	flag.Parse()

	analysis := gorma.NewAnalysis()
	if len(*rdb) > 0 {
		err := analysis.OpenRDB(*rdb)
		defer analysis.CloseRDB()
		if err != nil {
			fmt.Println("something wrong:", err)
			return
		}
		analysis.StartRDB(strings.Split(*prefixes, "//"))
	} else {
		err := analysis.Open(*ip, uint16(*port), *password)
		defer analysis.Close()
		if err != nil {
			fmt.Println("something wrong:", err)
			return
		}
		analysis.Start(strings.Split(*prefixes, "//"))
	}
	analysis.SaveReports(*reportPath)

	fmt.Println("done")
}
