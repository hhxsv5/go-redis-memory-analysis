package storages

import (
	"bytes"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"strings"
)

type RedisClient struct {
	Id   string
	conn redis.Conn
}

func NewRedisClient(host string, port uint16, password string) (*RedisClient, error) {
	var addr bytes.Buffer
	addr.WriteString(host)
	addr.WriteString(":")
	addr.WriteString(strconv.Itoa(int(port)))

	conn, err := redis.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}

	if password != "" {
		_, err := conn.Do("AUTH", password)
		if err != nil {
			return nil, err
		}
	}

	return &RedisClient{addr.String(), conn}, err
}

func (client RedisClient) Select(db uint64) error {
	_, err := client.conn.Do("SELECT", db)
	return err
}

func (client RedisClient) GetDatabases() (map[uint64]string, error) {
	var databases = make(map[uint64]string)

	reply, err := client.conn.Do("INFO", "Keyspace")
	keyspace, err := redis.String(reply, err)
	keyspace = strings.Trim(keyspace[12:], "\n")
	keyspaces := strings.Split(keyspace, "\r")

	for _, db := range keyspaces {
		strs := strings.Split(db, ":")
		strs[0] = strings.Trim(strs[0], "\n")
		if strs[0] == "" {
			continue
		}

		dbi, _ := strconv.ParseUint(strs[0][2:], 10, 64)
		databases[dbi] = strs[1]
	}
	return databases, err
}

func (client RedisClient) Scan(cursor *uint64, match string, limit uint64) ([]string, error) {
	reply, err := client.conn.Do("SCAN", *cursor, "MATCH", match, "COUNT", limit)
	result, err := redis.Values(reply, err)

	var keys []string

	for _, v := range result {
		switch v.(type) {
		case []uint8:
			*cursor, _ = redis.Uint64(v, nil)
		case []interface{}:
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

	if err != nil {
		return 0, err
	}

	debugs := strings.Split(debug, " ")
	items := strings.Split(debugs[4], ":")

	return strconv.ParseUint(items[1], 10, 64)
}

func (client RedisClient) Close() error {
	return client.conn.Close()
}
