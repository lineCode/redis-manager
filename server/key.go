package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"strconv"
	"strings"

	"os"

	"database/sql"

	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

var (
	ERROR_NOT_SUPPORT_TYPE = errors.New("not support type")
)

func (p KeyProto) Save(conn *redis.Client) error {
	switch p.Type {
	case "string":
		var val string
		err := json.Unmarshal(p.Data, &val)
		if err != nil {
			return err
		}
		return conn.Set(p.Key, val, 0).Err()
	case "list":
		var val []interface{}
		err := json.Unmarshal(p.Data, &val)
		if err != nil {
			return err
		}
		return conn.LPush(p.Key, val...).Err()
	case "hash":
		var val map[string]interface{}
		err := json.Unmarshal(p.Data, &val)
		if err != nil {
			return err
		}
		return conn.HMSet(p.Key, val).Err()
	case "set":
		var val []interface{}
		err := json.Unmarshal(p.Data, &val)
		if err != nil {
			return err
		}
		return conn.SAdd(p.Key, val...).Err()
	case "zset":
		var val []redis.Z
		err := json.Unmarshal(p.Data, &val)
		if err != nil {
			return err
		}
		return conn.ZAdd(p.Key, val...).Err()
	}
	return ERROR_NOT_SUPPORT_TYPE
}

func (p SaveProto) DelField(conn *redis.Client) error {
	switch p.Type {
	case "string":
		return ERROR_NOT_SUPPORT_TYPE
	case "list":
		vals, err := conn.LRange(p.Key, p.Index, p.Index).Result()
		if err != nil {
			return err
		}
		if len(vals) != 1 {
			return errors.New("the list had modified ")
		}
		if vals[0] != p.Field {
			return errors.New("the list had modified ")
		}
		removeVal := "-----TMP-----VALUE-----SHOULD-----REMOVE-----"
		err = conn.LSet(p.Key, p.Index, removeVal).Err()
		if err != nil {
			return err
		}
		return conn.LRem(p.Key, 1, removeVal).Err()
	case "hash":
		return conn.HDel(p.Key, p.Field).Err()
	case "set":
		return conn.SRem(p.Key, p.Field).Err()
	case "zset":
		return conn.ZRem(p.Key, p.Field).Err()
	}
	return ERROR_NOT_SUPPORT_TYPE
}

func (p KeyProto) Add(conn *redis.Client) error {
	// 检查是否已经存在
	t, err := conn.Type(p.Key).Result()
	if err != nil {
		return err
	}
	if t != "none" {
		return errors.New(fmt.Sprintf("key %s had already exists", p.Key))
	}
	return p.Save(conn)
}

func (srv *Server) expireRedisKey(context *gin.Context) {
	id, _ := getServerID(context)
	conn := srv.redisConn[id]
	p, err := newKeyProto(context.GetRawData())
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	if p.TTL < 0 {
		err = conn.conn.Persist(p.Key).Err()
	} else {
		err = conn.conn.Expire(p.Key, time.Second*time.Duration(p.TTL)).Err()
	}
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	context.JSON(200, newResponseProto(200, "", "success"))
}

func (srv *Server) getRedisKey(context *gin.Context) {
	id, _ := getServerID(context)
	conn := srv.redisConn[id]
	p, err := newKeyProto(context.GetRawData())
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	p.Type, err = conn.conn.Type(p.Key).Result()
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	var ttl time.Duration
	ttl, err = conn.conn.TTL(p.Key).Result()
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	p.TTL = int64(ttl.Seconds())
	switch p.Type {
	case "string":
		data, err := conn.conn.Get(p.Key).Result()
		if err != nil {
			context.JSON(200, newResponseProto(50000, err.Error(), nil))
			return
		}
		p.Data, _ = json.Marshal(data)
		p.Size = 1
	case "list":
		p.Size, err = conn.conn.LLen(p.Key).Result()
		if err != nil {
			context.JSON(200, newResponseProto(50000, err.Error(), nil))
			return
		}
	case "hash":
		p.Size, err = conn.conn.HLen(p.Key).Result()
		if err != nil {
			context.JSON(200, newResponseProto(50000, err.Error(), nil))
			return
		}
	case "set":
		p.Size, err = conn.conn.SCard(p.Key).Result()
		if err != nil {
			context.JSON(200, newResponseProto(50000, err.Error(), nil))
			return
		}
	case "zset":
		p.Size, err = conn.conn.ZCard(p.Key).Result()
		if err != nil {
			context.JSON(200, newResponseProto(50000, err.Error(), nil))
			return
		}
	case "none":
		conn.removeKey(p.Key)

	}
	context.JSON(200, newResponseProto(200, "", p))
}

func (srv *Server) addRedisKey(context *gin.Context) {
	id, _ := getServerID(context)
	conn := srv.redisConn[id]
	p, err := newKeyProto(context.GetRawData())
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	err = p.Add(conn.conn)
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}

	conn.addKey(p.Key)
	context.JSON(200, newResponseProto(200, "", "success"))
}

func (srv *Server) saveRedisKey(context *gin.Context) {
	id, _ := getServerID(context)
	conn := srv.redisConn[id]
	p, err := newSaveProto(context.GetRawData())
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	err = p.Save(conn.conn)
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}

	conn.addKey(p.Key)
	context.JSON(200, newResponseProto(200, "", "success"))
}

func (p *SaveProto) saveRow(conn *redis.Client) error {
	switch p.Type {
	case "string":
		var val string
		err := json.Unmarshal(p.Data, &val)
		if err != nil {
			return err
		}
		return conn.Set(p.Key, val, 0).Err()
	case "list":
		var val listValProto
		err := json.Unmarshal(p.Data, &val)
		if err != nil {
			return err
		}
		vals, err := conn.LRange(p.Key, val.Index, val.Index).Result()
		if err != nil {
			return err
		}
		if len(vals) != 1 {
			return errors.New("the list had modified ")
		}
		if vals[0] != p.Field {
			return errors.New("the list had modified ")
		}
		return conn.LSet(p.Key, val.Index, val.Val).Err()
	case "hash":
		var val map[string]interface{}
		err := json.Unmarshal(p.Data, &val)
		if err != nil {
			return err
		}
		if len(val) != 1 {
			return errors.New("params error")
		}
		err = conn.HMSet(p.Key, val).Err()
		if err != nil {
			return err
		}
		for k, _ := range val {
			if k != p.Field {
				return conn.HDel(p.Key, p.Field).Err()
			}
		}
		return nil
	case "set":
		var val []interface{}
		err := json.Unmarshal(p.Data, &val)
		if err != nil {
			return err
		}
		if len(val) != 1 {
			return errors.New("params error")
		}
		err = conn.SAdd(p.Key, val...).Err()
		if err != nil {
			return err
		}
		if val[0] != p.Field {
			return conn.SRem(p.Key, p.Field).Err()
		}
		return nil
	case "zset":
		var val []redis.Z
		err := json.Unmarshal(p.Data, &val)
		if err != nil {
			return err
		}
		if len(val) != 1 {
			return errors.New("params error")
		}
		err = conn.ZAdd(p.Key, val...).Err()
		if err != nil {
			return err
		}
		if val[0].Member.(string) != p.Field {
			return conn.ZRem(p.Key, p.Field).Err()
		}
		return nil
	}
	return ERROR_NOT_SUPPORT_TYPE
}

func (srv *Server) saveRowRedisKey(context *gin.Context) {
	id, _ := getServerID(context)
	conn := srv.redisConn[id]
	p, err := newSaveProto(context.GetRawData())
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	err = p.saveRow(conn.conn)
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}

	conn.addKey(p.Key)
	context.JSON(200, newResponseProto(200, "", "success"))
}

func (srv *Server) delRowRedisKey(context *gin.Context) {
	id, _ := getServerID(context)
	conn := srv.redisConn[id]
	p, err := newSaveProto(context.GetRawData())
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	err = p.DelField(conn.conn)
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	context.JSON(200, newResponseProto(200, "", "success"))
}

func (srv *Server) renameRedisKey(context *gin.Context) {
	id, _ := getServerID(context)
	p, err := newRenameProto(context.GetRawData())
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	conn := srv.redisConn[id]
	err = conn.conn.Rename(p.OldKey, p.NewKey).Err()
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}

	conn.removeKey(p.OldKey)
	conn.addKey(p.NewKey)
	context.JSON(200, newResponseProto(200, "", "success"))
}

func (srv *Server) scanRedisKey(context *gin.Context) {
	id, _ := getServerID(context)
	p, err := newScanProto(context.GetRawData())
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	conn := srv.redisConn[id]
	if p.Regex { //@todo
		context.JSON(200, newResponseProto(50000, ERROR_NOT_SUPPORT_TYPE.Error(), nil))
		return
	}
	if p.Num <= 0 {
		p.Num = 1000
	}
	var cursor int64
	keys := make([]string, 0, p.Num)
	if conn.cfg.PreLoad { // 预加载模式，从 trie 里面查找
		cursor, keys, err = conn.scanFromDB(p)
	} else { // 直接查找redis
		cursor, keys, err = p.scanKeys(conn.conn, conn.cfg.ScanRowNumEachLoop)
	}
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	context.JSON(200, newResponseProto(200, "", gin.H{"keys": keys, "cursor": cursor}))
}

func (redisConn *RedisConn) scanFromDB(p ScanProto) (int64, []string, error) {
	var db *sql.DB = redisConn.db
	keys := make([]string, 0, p.Num)
	rows, err := db.Query("SELECT key FROM data WHERE key LIKE ? ORDER BY key ASC LIMIT ? OFFSET ? ", p.Keyword+"%", p.Num, p.Iter)
	if err != nil {
		return 0, keys, err
	}
	defer rows.Close()
	for rows.Next() {
		var key string
		err := rows.Scan(&key)
		if err != nil {
			return 0, keys, err
		}
		keys = append(keys, key)
	}
	if len(keys) > 0 {
		return int64(len(keys)) + p.Iter, keys, nil
	} else {
		return 0, keys, nil
	}
}

func (redisConn *RedisConn) removeKey(key string) error {
	if !redisConn.cfg.PreLoad {
		return nil
	}
	var err error
	_, err = redisConn.db.Exec("DELETE FROM data WHERE key = ?", key)
	if err != nil {
		log.Warningf("del key %s error: %s", key, err.Error())
	}

	return err
}

func (redisConn *RedisConn) addKey(key string) error {
	if !redisConn.cfg.PreLoad {
		return nil
	}
	var err error
	_, err = redisConn.db.Exec("INSERT OR IGNORE INTO data(key) VALUES(?)", key)
	if err != nil {
		log.Warningf("add key %s error: %s", key, err.Error())
	}

	return err
}

func (redisConn *RedisConn) scanKeys2db(stop <-chan struct{}, maxRows int) (int64, error) {
	db, err := redisConn.reCreateDB(true)
	if err != nil {
		return 0, err
	}

	var cursor uint64
	var total int64

	cacheKeys := make([]string, 0, maxRows+int(redisConn.cfg.ScanRowNumEachLoop*3))
	save2db := func(keys []string) error {
		if len(keys) == 0 {
			return nil
		}
		keysLen := len(keys)
		for l := 0; l < keysLen; l += 900 {
			var allocLen int
			if l+900 > keysLen {
				allocLen = keysLen - l
			} else {
				allocLen = 900
			}
			valStrArr := make([]string, allocLen, allocLen)
			valInterface := make([]interface{}, allocLen, allocLen)
			i := 0
			for iter := l; iter < l+allocLen; iter++ {
				valStrArr[i] = "(?)"
				valInterface[i] = keys[iter]
				i++
			}
			_, err = db.Exec("INSERT OR IGNORE INTO data(key) values"+strings.Join(valStrArr, ","), valInterface...)
			if err != nil {
				return err
			}
		}
		log.Debugf("save to db, length: %d", keysLen)
		return nil
	}
	for {

		select {
		case <-stop:
			log.Debug("stop")
			return 0, nil
		default:
		}
		var keys []string
		var err error
		keys, cursor, err = redisConn.conn.Scan(cursor, "*", redisConn.cfg.ScanRowNumEachLoop).Result()
		if err != nil {
			return 0, err
		}

		if len(keys) > 0 {
			total += int64(len(keys))
			cacheKeys = append(cacheKeys, keys...)
			if len(cacheKeys) > int(maxRows) {
				err = save2db(cacheKeys)
				if err != nil {
					return 0, err
				}
				cacheKeys = make([]string, 0, maxRows+int(redisConn.cfg.ScanRowNumEachLoop*3))
			}
		}

		if cursor == 0 {
			err = save2db(cacheKeys)
			if err != nil {
				return 0, err
			}

			var keyTotal sql.NullInt64
			err = db.QueryRow("SELECT count(*) FROM data LIMIT 1;").Scan(&keyTotal)
			if err != nil {
				return 0, err
			}
			redisConn.changeDB(db)
			if keyTotal.Valid {
				return keyTotal.Int64, nil
			}
			return 0, nil
		}
	}
}

func (conn *RedisConn) reCreateDB(isRemove bool) (*sql.DB, error) {
	var filename string
	if conn.currentDB == 0 {
		filename = conn.cacheDir + "/cache_a.db"
	} else {
		filename = conn.cacheDir + "/cache_b.db"
	}
	var err error
	if isRemove {
		err = os.Remove(filename)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	sqlStmt := `
CREATE TABLE IF NOT EXISTS data ( key TEXT PRIMARY KEY );
`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}
	return db, err
}

func (conn *RedisConn) changeDB(db *sql.DB) {
	if conn.db != nil {
		conn.db.Close()
	}
	conn.db = db
	if conn.currentDB == 0 {
		atomic.SwapInt32(&conn.currentDB, 1)
	} else {
		atomic.SwapInt32(&conn.currentDB, 0)
	}
}

func (srv *Server) refreshRedisKeyDaemon() {
	for id, conn := range srv.redisConn {
		if !conn.cfg.PreLoad {
			continue
		}
		cacheDir := strings.TrimRight(srv.cfg.CacheDir, "/") + "/" + strconv.Itoa(id)
		b, err := PathExists(cacheDir)
		if err != nil {
			log.Fatal(err)
		}
		if !b {
			err = os.MkdirAll(cacheDir, 0755)
			if err != nil {
				log.Fatal(err)
			}
		}
		conn.cacheDir = cacheDir
		b, err = PathExists(cacheDir + "/cache_a.db") // 使用之前的缓存
		if err != nil {
			log.Fatal(err)
		}
		if !b {
			conn.currentDB = 1
		}
		db, err := conn.reCreateDB(false)
		if err != nil {
			log.Fatalf("error create db %s", err.Error())
		}
		conn.changeDB(db)
		srv.waitGroup.Add(1)
		go func(conn *RedisConn) {
			defer srv.waitGroup.Done()
			refresh := func() {
				atomic.SwapInt32(&conn.refreshStatus, 1)
				startTime := time.Now()
				total, err := conn.scanKeys2db(srv.exit, srv.cfg.CacheSize)
				atomic.SwapInt32(&conn.refreshStatus, 0)
				if err != nil {
					log.Warningf("refresh redis keys error %s", err.Error())
				} else {
					conn.Initialize = true
					conn.lastRefreshTime = time.Now()
					usedTime := conn.lastRefreshTime.Sub(startTime)
					log.Debugf("refresh %s success, use %.2f seconds, total %d \n", conn.cfg.Name, usedTime.Seconds(), total)
				}
			}
			refresh()
			for {
				select {
				case <-time.After(time.Duration(conn.cfg.RefreshDuration) * time.Second):
					if conn.cfg.RefreshDuration > 0 {
						refresh()
					} else {
						<-time.After(time.Second)
					}

				case <-conn.refreshChan:
					refresh()
				case <-srv.exit:
					err = conn.db.Close()
					if err != nil {
						log.Warningf("close db error %s", err.Error())
					}
					log.Debug("refresh stop ")
					return
				}
			}
		}(conn)
	}
}

func (srv *Server) refreshKeys(context *gin.Context) {
	id, _ := getServerID(context)
	conn := srv.redisConn[id]
	select {
	case conn.refreshChan <- struct{}{}:
		context.JSON(200, newResponseProto(200, "", "refresh running at backend!"))
	case <-time.After(time.Second):
		context.JSON(200, newResponseProto(10005, "refreshing, please wait!", ""))
	}
}
