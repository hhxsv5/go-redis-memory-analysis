Redis memory analysis
======

Analyzing memory of redis is to find the keys(prefix) which used a lot of memory, export the analysis result into csv file.


## Usage
### Run demo

```Shell
cd examples
//install dependencies
glide init
glide install
```

```Go
//cd examples
go run main.go
//find reports in current folder
```

![CSV](https://raw.githubusercontent.com/hhxsv5/go-redis-memory-analysis/master/examples/demo.png)

## Another tool implemented by PHP

[redis-memory-analysis](https://github.com/hhxsv5/redis-memory-analysis)


## License

[MIT](https://github.com/hhxsv5/go-redis-memory-analysis/blob/master/LICENSE)
