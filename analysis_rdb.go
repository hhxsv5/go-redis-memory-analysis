package gorma

import (
	"os"
	"fmt"
	"github.com/vrischmann/rdbtools"
	"strings"
	"time"
	"strconv"
	"sort"
	"github.com/hhxsv5/go-redis-memory-analysis/storages"
)

type AnalysisRDB struct {
	rdb     *os.File
	Reports DBReports
}

func NewAnalysisRDB(rdb string) (*AnalysisRDB, error) {
	fp, err := os.Open(rdb)
	if err != nil {
		return nil, err
	}
	return &AnalysisRDB{fp, DBReports{}}, nil
}

func (analysis *AnalysisRDB) Close() {
	if analysis.rdb != nil {
		analysis.rdb.Close()
	}
}

func (analysis AnalysisRDB) Start(delimiters []string) {
	fmt.Println("Starting analysis")
	var (
		r   Report
		sr  SortByCountReports
		mr  KeyReports
		cdb = -1
	)

	ctx := rdbtools.ParserContext{
		DbCh:                make(chan int),
		StringObjectCh:      make(chan rdbtools.StringObject),
		ListMetadataCh:      make(chan rdbtools.ListMetadata),
		ListDataCh:          make(chan interface{}),
		SetMetadataCh:       make(chan rdbtools.SetMetadata),
		SetDataCh:           make(chan interface{}),
		HashMetadataCh:      make(chan rdbtools.HashMetadata),
		HashDataCh:          make(chan rdbtools.HashEntry),
		SortedSetMetadataCh: make(chan rdbtools.SortedSetMetadata),
		SortedSetEntriesCh:  make(chan rdbtools.SortedSetEntry),
	}
	p := rdbtools.NewParser(ctx)

	go func() {
		analyze := func(ko rdbtools.KeyObject) {
			key := rdbtools.DataToString(ko.Key)
			fd, fp, tmp, nk := "", 0, 0, ""
			for _, delimiter := range delimiters {
				tmp = strings.Index(key, delimiter)
				if tmp != -1 && (tmp < fp || fp == 0) {
					fd, fp = delimiter, tmp
				}
			}

			if fp == 0 {
				return
			}

			nk = key[0:fp] + fd + "*"

			if _, ok := mr[nk]; ok {
				r = mr[nk]
			} else {
				r = Report{nk, 0, 0, 0, 0}
			}
			if ko.ExpiryTime.IsZero() {
				r.NeverExpire++
			} else {
				if ko.Expired() {
					//Ignore
				} else {
					ff := float64(r.AvgTtl*(r.Count-r.NeverExpire)+uint64(ko.ExpiryTime.Unix()-time.Now().Unix())) / float64(r.Count+1-r.NeverExpire)
					ttl, _ := strconv.ParseUint(fmt.Sprintf("%0.0f", ff), 10, 64)
					r.AvgTtl = ttl
				}
			}
			r.Count++
			mr[nk] = r
		}

		for {
			select {
			case t, ok := <-ctx.DbCh:
				if cdb >= 0 {
					//Sort by size
					sr = SortByCountReports{}
					for _, report := range mr {
						sr = append(sr, report)
					}
					sort.Sort(sr)
					analysis.Reports[uint64(cdb)] = sr
				}

				if !ok {
					ctx.DbCh = nil
					break
				}
				fmt.Println("Analyzing db", t)
				cdb = t
				mr = KeyReports{}
			case t, ok := <-ctx.StringObjectCh:
				if !ok {
					ctx.StringObjectCh = nil
					break
				}
				analyze(t.Key)
				//fmt.Println("string object", t)
			case t, ok := <-ctx.ListMetadataCh:
				if !ok {
					ctx.ListMetadataCh = nil
					break
				}
				analyze(t.Key)
				//fmt.Println("list meta data", t)
			case _, ok := <-ctx.ListDataCh:
				if !ok {
					ctx.ListDataCh = nil
					break
				}
				//fmt.Println("list data", rdbtools.DataToString(t))
			case t, ok := <-ctx.SetMetadataCh:
				if !ok {
					ctx.SetMetadataCh = nil
					break
				}
				analyze(t.Key)
				//fmt.Println("set meta data", t)
			case _, ok := <-ctx.SetDataCh:
				if !ok {
					ctx.SetDataCh = nil
					break
				}
				//fmt.Println("set data", rdbtools.DataToString(t))
			case t, ok := <-ctx.HashMetadataCh:
				if !ok {
					ctx.HashMetadataCh = nil
					break
				}
				analyze(t.Key)
				//fmt.Println("hash meta data", t)
			case _, ok := <-ctx.HashDataCh:
				if !ok {
					ctx.HashDataCh = nil
					break
				}
				//fmt.Println("hash data", t)
			case t, ok := <-ctx.SortedSetMetadataCh:
				if !ok {
					ctx.SortedSetMetadataCh = nil
					break
				}
				analyze(t.Key)
				//fmt.Println("sorted set meta data", t)
			case _, ok := <-ctx.SortedSetEntriesCh:
				if !ok {
					ctx.SortedSetEntriesCh = nil
					break
				}
				//fmt.Println("sorted set entries", t)
			}

			if ctx.Invalid() {
				break
			}
		}

		//fmt.Println(analysis.Reports)
	}()

	if err := p.Parse(analysis.rdb); err != nil {
		panic(err)
	}
}

func (analysis AnalysisRDB) SaveReports(folder string) error {
	fmt.Println("Saving the results of the analysis into", folder)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, os.ModePerm)
	}

	var (
		str      string
		filename string
		size     float64
		unit     string
	)
	template := fmt.Sprintf("%s%sredis-analysis-%s%s", folder, string(os.PathSeparator), strings.Replace(analysis.rdb.Name(), "/", "-", -1), "-%d.csv")
	for db, reports := range analysis.Reports {
		filename = fmt.Sprintf(template, db)
		fp, err := storages.NewFile(filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		fp.Append([]byte("Key,Count,Size,NeverExpire,AvgTtl(excluded never expire)\n"))
		for _, value := range reports {
			size, unit = HumanSize(value.Size)
			str = fmt.Sprintf("%s,%d,%s,%d,%d\n",
				value.Key,
				value.Count,
				fmt.Sprintf("%0.3f %s", size, unit),
				value.NeverExpire,
				value.AvgTtl)
			fp.Append([]byte(str))
		}
		fp.Close()
	}
	return nil
}
