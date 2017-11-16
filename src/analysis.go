package gorma

import (
	"strings"
	"fmt"
	"strconv"
	. "./models"
)

type Report struct {
	Key         string
	Count       uint64
	Size        uint64
	NeverExpire uint64
	AvgTtl      uint64
}

type Analysis struct {
	redis   *RedisClient
	Reports map[int]map[string]Report
}

func NewAnalysis(redis *RedisClient) (*Analysis) {
	return &Analysis{redis, map[int]map[string]Report{}}
}

func (analysis *Analysis) Start(delimiters []string, limit uint64) {
	match := "*[" + strings.Join(delimiters, "") + "]*"
	var cursor uint64 = 0
	keys, _ := analysis.redis.Scan(&cursor, match, 3000)

	fd, fp, tmp, nk := "", 0, 0, ""
	var r Report
	var f float64
	var ttl int64
	var length uint64
	for _, key := range keys {
		fd, fp, tmp, nk, ttl, ne = "", 0, 0, "", 0, 0
		for _, delimiter := range delimiters {
			tmp = strings.Index(key, delimiter)
			if tmp < fp {
				fd, fp = delimiter, tmp
			}
		}

		if _, ok := analysis.Reports[db]; !ok {
			analysis.Reports[db] = map[string]Report{}
		}

		nk = key[fp:] + fd + "*"

		if _, ok := analysis.Reports[db][nk]; ok {
			r = analysis.Reports[db][nk]
		} else {
			r = Report{nk, 1, 0, 0, 0}
		}

		ttl, _ = analysis.redis.Ttl(nk)

		switch ttl {
		case -2:
			continue
		case -1:
			r.NeverExpire++
		default:
			f = float64(r.AvgTtl*(r.Count-r.NeverExpire)+uint64(ttl)) / float64(r.Count+1)
			ttl, _ := strconv.ParseUint(fmt.Sprintf("%0.0f", f), 10, 64)
			r.AvgTtl = ttl
			r.Count++
		}

		length, _ = analysis.redis.SerializedLength(nk)
		r.Size += length

		analysis.Reports[db][nk] = r
	}
}
