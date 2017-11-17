package main

import (
	"fmt"
	. "github.com/hhxsv5/go-redis-memory-analysis/storages"
	. "github.com/hhxsv5/go-redis-memory-analysis"
)

func main() {
	redis, err := NewRedisClient("127.0.0.1", 6379, "")
	if err != nil {
		fmt.Println("Connect redis fail", err)
		return
	}
	defer redis.Close()

	analysis := NewAnalysis(redis)
	analysis.Start([]string{"#", ":"}, 3000)

	fmt.Println(analysis.Reports)
}
