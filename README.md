Redis memory analysis
======

🔎  Analyzing memory of redis is to find the keys(prefix) which used a lot of memory, export the analysis result into csv file.

## Usage

1. Install

```Shell
//cd your-root-folder-of-project
//Create the file glide.yaml if not exist
//touch glide.yaml
glide get github.com/hhxsv5/go-redis-memory-analysis#~2.0.0
//glide: THANK GO15VENDOREXPERIMENT
```

2. Run

```Go
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
analysis.SaveReports("./reports")
```

![CSV](https://raw.githubusercontent.com/hhxsv5/go-redis-memory-analysis/master/examples/demo.png)

## Another tool implemented by PHP

[redis-memory-analysis](https://github.com/hhxsv5/redis-memory-analysis)


## License

[MIT](https://github.com/hhxsv5/go-redis-memory-analysis/blob/master/LICENSE)
