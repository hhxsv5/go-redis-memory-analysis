package gorma

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"github.com/vrischmann/rdbtools"
	. "github.com/hhxsv5/go-redis-memory-analysis/storages"
	"time"
)

type Report struct {
	Key         string
	Count       uint64
	Size        uint64
	NeverExpire uint64
	AvgTtl      uint64
}

type DBReports map[uint64][]Report

type KeyReports map[string]Report

type SortBySizeReports []Report
type SortByCountReports []Report

type Analysis struct {
	redis   *RedisClient
	rdb     *os.File
	Reports DBReports
}

func (sr SortBySizeReports) Len() int {
	return len(sr)
}

func (sr SortBySizeReports) Less(i, j int) bool {
	return sr[i].Size > sr[j].Size
}

func (sr SortBySizeReports) Swap(i, j int) {
	sr[i], sr[j] = sr[j], sr[i]
}

func (sr SortByCountReports) Len() int {
	return len(sr)
}

func (sr SortByCountReports) Less(i, j int) bool {
	return sr[i].Count > sr[j].Count
}

func (sr SortByCountReports) Swap(i, j int) {
	sr[i], sr[j] = sr[j], sr[i]
}

func NewAnalysis() *Analysis {
	return &Analysis{nil, nil, DBReports{}}
}

func (analysis *Analysis) Open(host string, port uint16, password string) error {
	redis, err := NewRedisClient(host, port, password)
	if err != nil {
		return err
	}
	analysis.redis = redis
	return nil
}

func (analysis *Analysis) OpenRDB(rdb string) error {
	fp, err := os.Open(rdb)
	if err != nil {
		return err
	}
	analysis.rdb = fp
	return nil
}

func (analysis *Analysis) Close() {
	if analysis.redis != nil {
		analysis.redis.Close()
	}
}

func (analysis *Analysis) CloseRDB() {
	if analysis.rdb != nil {
		analysis.rdb.Close()
	}
}

func (analysis Analysis) Start(delimiters []string) {
	fmt.Println("Starting analysis")
	match := "*[" + strings.Join(delimiters, "") + "]*"
	databases, _ := analysis.redis.GetDatabases()

	var (
		cursor uint64
		r      Report
		f      float64
		ttl    int64
		length uint64
		sr     SortBySizeReports
		mr     KeyReports
	)

	for db, _ := range databases {
		fmt.Println("Analyzing db", db)
		cursor = 0
		mr = KeyReports{}

		analysis.redis.Select(db)

		for {
			keys, _ := analysis.redis.Scan(&cursor, match, 3000)
			fd, fp, tmp, nk := "", 0, 0, ""
			for _, key := range keys {
				fd, fp, tmp, nk, ttl = "", 0, 0, "", 0
				for _, delimiter := range delimiters {
					tmp = strings.Index(key, delimiter)
					if tmp != -1 && (tmp < fp || fp == 0) {
						fd, fp = delimiter, tmp
						break
					}
				}

				if fp == 0 {
					continue
				}

				nk = key[0:fp] + fd + "*"

				if _, ok := mr[nk]; ok {
					r = mr[nk]
				} else {
					r = Report{nk, 0, 0, 0, 0}
				}

				ttl, _ = analysis.redis.Ttl(key)

				switch ttl {
				case -2:
					continue
				case -1:
					r.NeverExpire++
					r.Count++
				default:
					f = float64(r.AvgTtl*(r.Count-r.NeverExpire)+uint64(ttl)) / float64(r.Count+1-r.NeverExpire)
					ttl, _ := strconv.ParseUint(fmt.Sprintf("%0.0f", f), 10, 64)
					r.AvgTtl = ttl
					r.Count++
				}

				length, _ = analysis.redis.SerializedLength(key)
				r.Size += length

				mr[nk] = r
			}

			if cursor == 0 {
				break
			}
		}

		//Sort by size
		sr = SortBySizeReports{}
		for _, report := range mr {
			sr = append(sr, report)
		}
		sort.Sort(sr)

		analysis.Reports[db] = sr
	}
}

func (analysis Analysis) StartRDB(delimiters []string) {
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
					break
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

func (analysis Analysis) SaveReports(folder string) error {
	fmt.Println("Saving the results of the analysis into", folder)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, os.ModePerm)
	}

	var template string
	if analysis.redis != nil {
		template = fmt.Sprintf("%s%sredis-analysis-%s%s", folder, string(os.PathSeparator), strings.Replace(analysis.redis.Id, ":", "-", 1), "-%d.csv")
	} else if analysis.rdb != nil {
		template = fmt.Sprintf("%s%sredis-analysis-%s%s", folder, string(os.PathSeparator), strings.Replace(analysis.rdb.Name(), "/", "-", 1), "-%d.csv")
	} else {
		template = fmt.Sprintf("%s%sredis-analysis-%s", folder, string(os.PathSeparator), "-%d.csv")
	}
	var (
		str      string
		filename string
		size     float64
		unit     string
	)
	for db, reports := range analysis.Reports {
		filename = fmt.Sprintf(template, db)
		fp, err := NewFile(filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
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
