package redis_manager

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

func (srv *Server) valueScanRedisKey(context *gin.Context) {
	id, _ := getServerID(context)
	conn := srv.redisConn[id]
	p, err := newValueScanProto(context.GetRawData())
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	// 获取redis key
	cursor, val, err := p.scanValue(conn.conn, conn.cfg.ScanRowNumEachLoop)
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	context.JSON(200, newResponseProto(200, "", gin.H{"val": val, "cursor": cursor}))
}

func (p *ValueScanProto) scanValue(conn *redis.Client, scanRowNumEachLoop int64) (int64, interface{}, error) {
	var err error
	p.Type, err = conn.Type(p.Key).Result()
	if err != nil {
		return p.Start, nil, err
	}
	switch p.Type {
	case "string":
		ret, err := conn.Get(p.Key).Result()
		return 0, ret, err
	case "list":
		ret := make([]listValProto, 0, p.Num)
		vals, err := conn.LRange(p.Key, p.Start, p.Num+p.Start).Result()
		if len(vals) == 0 {
			return 0, ret, err
		}
		for i, val := range vals {
			ret = append(ret, listValProto{
				p.Start + int64(i),
				val,
			})
		}
		return p.Start + int64(len(ret)), ret, err
	case "hash":
		var cursor uint64 = uint64(p.Iter)
		var results = make(map[string]interface{})
		if p.Keyword != "" { //获取一个
			val, err := conn.HGet(p.Key, p.Keyword).Result()
			if err != nil {
				return 0, nil, err
			}
			results[p.Keyword] = val
			return 0, results, nil
		}
		for {
			var fields []string
			fields, cursor, err = conn.HScan(p.Key, cursor, "*", scanRowNumEachLoop).Result()
			if err != nil {
				return p.Start, nil, err
			}
			if len(fields) > 0 {
				for j := 0; j < len(fields); j += 2 {
					results[fields[j]] = fields[j+1]
				}
			}
			if len(results) >= int(p.Num) {
				break
			}
			if cursor == 0 {
				break
			}
		}
		return int64(cursor), results, nil
	case "set":
		var cursor uint64 = uint64(p.Iter)
		var results = make([]string, 0, p.Num)
		if p.Keyword != "" { //获取一个
			b, err := conn.SIsMember(p.Key, p.Keyword).Result()
			if err != nil {
				return 0, nil, err
			}
			if b {
				results = append(results, p.Keyword)
			}
			return 0, results, nil
		}
		for {
			var fields []string
			fields, cursor, err = conn.SScan(p.Key, cursor, "*", scanRowNumEachLoop).Result()
			if err != nil {
				return p.Start, nil, err
			}
			if len(fields) > 0 {
				for j := 0; j < len(fields); j++ {
					results = append(results, fields[j])
				}
			}
			if len(results) >= int(p.Num) {
				break
			}
			if cursor == 0 {
				break
			}
		}
		return int64(cursor), results, nil
	case "zset":
		var cursor int64
		var results = make([]redis.Z, 0, p.Num)
		if p.Keyword != "" { //获取一个
			score, err := conn.ZScore(p.Key, p.Keyword).Result()
			if err != nil {
				return 0, nil, err
			}
			results = append(results, redis.Z{Member: p.Keyword, Score: score})
			return 0, results, nil
		}
		cursor = p.Start + p.Num - 1
		if p.Asc {
			results, err = conn.ZRangeWithScores(p.Key, p.Start, cursor).Result()
		} else {
			results, err = conn.ZRevRangeWithScores(p.Key, p.Start, cursor).Result()
		}
		return cursor, results, nil
	}
	return p.Start, nil, ERROR_NOT_SUPPORT_TYPE
}

func (p *ScanProto) scanKeys(conn *redis.Client, scanRowNumEachLoop int64) (int64, []string, error) {
	var err error
	var cursor uint64 = uint64(p.Iter)
	results := make([]string, 0, p.Num)
	for {
		var fields []string
		fields, cursor, err = conn.Scan(cursor, p.Keyword+"*", scanRowNumEachLoop).Result()
		if err != nil {
			return p.Start, nil, err
		}
		if len(fields) > 0 {
			results = append(results, fields...)
		}
		if len(results) >= int(p.Num) {
			break
		}
		if cursor == 0 {
			break
		}
	}
	return int64(cursor), results, nil

}
