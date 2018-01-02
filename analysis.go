package gorma

import (
	"fmt"
	. "github.com/hhxsv5/go-redis-memory-analysis/storages"
	"os"
	"sort"
	"strconv"
	"strings"
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

type SortReports []Report

type Analysis struct {
	redis   *RedisClient
	Reports DBReports
}

func (sr SortReports) Len() int {
	return len(sr)
}

func (sr SortReports) Less(i, j int) bool {
	return sr[i].Size > sr[j].Size
}

func (sr SortReports) Swap(i, j int) {
	sr[i], sr[j] = sr[j], sr[i]
}

func NewAnalysis(redis *RedisClient) *Analysis {
	return &Analysis{redis, DBReports{}}
}

func (analysis Analysis) Start(delimiters []string) {

	match := "*[" + strings.Join(delimiters, "") + "]*"
	databases, _ := analysis.redis.GetDatabases()

	var (
		cursor uint64
		r      Report
		f      float64
		ttl    int64
		length uint64
		sr     SortReports
		mr     KeyReports
	)

	for db, _ := range databases {
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
		sr = SortReports{}
		for _, report := range mr {
			sr = append(sr, report)
		}
		sort.Sort(sr)

		analysis.Reports[db] = sr
	}
}

func (analysis Analysis) SaveReports(folder string) error {
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, os.ModePerm)
	}

	var template = fmt.Sprintf("%s%sredis-analysis-%s%s", folder, string(os.PathSeparator), analysis.redis.Id, "-%d.csv")
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
