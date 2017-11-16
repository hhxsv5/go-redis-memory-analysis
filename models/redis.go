package models

import (
	"github.com/garyburd/redigo/redis"
	"fmt"
	"bytes"
	"strings"
	"strconv"
)

type RedisClient struct {
	conn redis.Conn
}

func NewRedisClient(host string, port uint16, password string) (*RedisClient, error) {
	var addr bytes.Buffer
	addr.WriteString(host)
	addr.WriteString(":")
	addr.WriteString(string(port))

	conn, err := redis.Dial("tcp", addr.String())
	if err != nil {
		fmt.Println("connect redis fail: ", err)
		return nil, err
	}

	if password != "" {
		_, err := conn.Do("AUTH", password)
		if err != nil {
			fmt.Println("auth fail:", err)
			return nil, err
		}
	}

	return &RedisClient{conn}, err
}

func (client RedisClient) GetDatabases() ([]string, error) {
	reply, err := client.conn.Do("INFO", "Keyspace")
	dbs, err := redis.Strings(reply, err)

	var databases []string

	for _, db := range dbs {
		strs := strings.Split(db, ":")
		dbi, _ := strconv.Atoi(strs[0][2:])
		databases[dbi] = strs[1]
	}
	return databases, err
}

func (client RedisClient) Scan(cursor *uint64, match string, limit uint64) ([]string, error) {
	reply, err := client.conn.Do("SCAN", cursor, match, limit)
	result, err := redis.Values(reply, err)

	var keys []string

	for _, v := range result {
		switch v.(type) {
		case uint8:
			*cursor, _ = redis.Uint64(v, nil)
		case []string:
			keys, _ = redis.Strings(v, nil)
		}
	}
	return keys, err
}

func (client RedisClient) Ttl(key string) (int64, error) {
	reply, err := client.conn.Do("TTL", key)
	ttl, err := redis.Int64(reply, err)
	return ttl, err
}

func (client RedisClient) SerializedLength(key string) (uint64, error) {
	reply, err := client.conn.Do("DEBUG", "OBJECT", key)
	debug, err := redis.String(reply, err)
	debugs := strings.Split(debug, " ")
	items := strings.Split(debugs[4], ":")
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(items[1], 10, 64)
}
