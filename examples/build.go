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

	var (
		analysis gorma.AnalysisInterface
		err      error
	)
	if len(*rdb) > 0 {
		analysis, err = gorma.NewAnalysisRDB(*rdb)
	} else {
		analysis, err = gorma.NewAnalysisConnection(*ip, uint16(*port), *password)
	}
	if err != nil {
		fmt.Println("something wrong:", err)
		return
	}
	defer analysis.Close()
	analysis.Start(strings.Split(*prefixes, "//"))
	err = analysis.SaveReports(*reportPath)
	if err == nil {
		fmt.Println("done")
	} else {
		fmt.Println("error:", err)
	}
}
