package redis_manager

import (
	"errors"
	"fmt"

	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

var incrTaskID int64

type DelTask struct {
	stopChan    chan struct{}
	TaskName    string   `json:"task_name"`
	TaskID      int64    `json:"task_id"`
	Keys        []string `json:"keys"`
	Process     float32  `json:"process"`
	ErrMsg      string   `json:"err"`
	Status      int64    `json:"status"` // 0 进行中，1 完成，2 停止中，3失败, 4 被动停止
	HadTryTimes int      `json:"had_try_times"`
}

func (srv *Server) delRedisKey(context *gin.Context) {
	id, _ := getServerID(context)
	p, err := newKeyProto(context.GetRawData())
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	conn := srv.redisConn[id]
	err = conn.conn.Del(p.Key).Err()
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	conn.removeKey(p.Key)
	context.JSON(200, newResponseProto(200, "", "success"))
}

func (srv *Server) delBackgroundRedisKey(context *gin.Context) {
	id, _ := getServerID(context)
	p, err := newDelKeysProto(context.GetRawData())
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	if len(p.Keys) == 0 {
		context.JSON(200, newResponseProto(50003, "keys should not be empty", nil))
		return
	}
	if p.Name == "" {
		p.Name = fmt.Sprintf("%s - %d", p.Keys[0], len(p.Keys))
	}
	conn := srv.redisConn[id]
	srv.delRedisKeyDaemon(conn, p.Keys, p.Name)
	context.JSON(200, newResponseProto(200, "", "success"))
}

func (srv *Server) removeDelTask(context *gin.Context) {
	id, _ := getServerID(context)
	p, err := newRemoveDelTaskProto(context.GetRawData())
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	conn := srv.redisConn[id]
	for e := conn.delQueue.Front(); e != nil; e = e.Next() {
		task := e.Value.(*DelTask)
		if task.TaskID == p.TaskID {
			if task.Status == 0 {
				context.JSON(200, newResponseProto(50004, "task is running", nil))
				return
			}
			conn.delQueue.Remove(e)
			break
		}
	}
	context.JSON(200, newResponseProto(200, "", "success"))
}

func (srv *Server) stopDelTask(context *gin.Context) {
	id, _ := getServerID(context)
	p, err := newRemoveDelTaskProto(context.GetRawData())
	if err != nil {
		context.JSON(200, newResponseProto(50000, err.Error(), nil))
		return
	}
	conn := srv.redisConn[id]
	for e := conn.delQueue.Front(); e != nil; e = e.Next() {
		task := e.Value.(*DelTask)
		if task.TaskID == p.TaskID {
			if task.Status == 2 {
				context.JSON(200, newResponseProto(50005, "task is stopping!", nil))
				return
			}
			if task.Status == 1 || task.Status == 3 {
				context.JSON(200, newResponseProto(50005, "task is completed!", nil))
				return
			}
			atomic.SwapInt64(&task.Status, 2)
			_, isClose := <-task.stopChan
			if !isClose {
				close(task.stopChan)
			}
			break
		}
	}
	context.JSON(200, newResponseProto(200, "", "success"))
}

func (srv *Server) listDelTask(context *gin.Context) {
	id, _ := getServerID(context)
	conn := srv.redisConn[id]
	var tasks []DelTask
	for e := conn.delQueue.Front(); e == nil; e = e.Next() {
		tasks = append(tasks, e.Value.(DelTask))
	}
	context.JSON(200, newResponseProto(200, "", tasks))
}

func (srv *Server) delRedisKeyDaemon(conn *RedisConn, keys []string, name string) {
	task := &DelTask{
		TaskID:   atomic.AddInt64(&incrTaskID, 1),
		Keys:     keys,
		stopChan: make(chan struct{}, 0),
		TaskName: name,
	}
	conn.delQueue.PushFront(task)
	go func(task *DelTask) {
		keyLen := len(task.Keys)
		for i := 0; i < keyLen; i++ {
			key := task.Keys[i]
			err := delKey(conn.conn, key, conn.cfg.ScanRowNumEachLoop, task.stopChan)
			if err != nil {
				log.Errorf("conn %s del key %s error %s", conn.cfg.Name, key, err.Error())
				task.ErrMsg = err.Error()
				atomic.SwapInt64(&task.Status, 3)
				return
			} else {
				select {
				case <-srv.exit:
					return
				case <-task.stopChan:
					atomic.SwapInt64(&task.Status, 4)
					return
				default:
					conn.removeKey(key)
					task.Process = task.Process + float32(1)/float32(keyLen)*100
				}
			}
		}
		task.Process = 100
		atomic.SwapInt64(&task.Status, 1)
		return
	}(task)

}

func delKey(conn *redis.Client, key string, maxRowsEachLoop int64, stopChan chan struct{}) error {
	var err error
	var keyType string
	keyType, err = conn.Type(key).Result()
	if err != nil {
		goto END
	}
	switch keyType {
	case "string": // 在下面删除
	case "list":
		var sizes int64
		sizes, err = conn.LLen(key).Result()
		if err != nil {
			goto END
		}
		var i int64 = 0
		for ; i < sizes; i += maxRowsEachLoop {
			select {
			case <-stopChan:
				goto END
			default:
			}
			err := conn.LTrim(key, i, i+maxRowsEachLoop).Err()
			if err != nil {
				goto END
			}
		}
	case "hash":
		var cursor uint64
		var fields []string
		for {
			select {
			case <-stopChan:
				goto END
			default:
			}
			fields, cursor, err = conn.HScan(key, cursor, "*", maxRowsEachLoop).Result()
			if err != nil {
				goto END
			}
			if len(fields) > 0 {
				tmp := make([]string, len(fields)/2)
				for j := 0; j < len(fields); j += 2 {
					tmp[j/2] = fields[j]
				}
				_, err = conn.HDel(key, tmp...).Result()
				if err != nil {
					goto END
				}
			}

			if cursor == 0 {
				break
			}
		}
	case "set":
		var cursor uint64
		var fields []string
		for {
			select {
			case <-stopChan:
				goto END
			default:
			}
			fields, cursor, err = conn.SScan(key, cursor, "*", maxRowsEachLoop).Result()
			if err != nil {
				goto END
			}
			if len(fields) > 0 {
				var members []interface{}
				members = make([]interface{}, len(fields))
				for i := 0; i < len(fields); i++ {
					members[i] = fields[i]
				}
				_, err = conn.SRem(key, members...).Result()
				if err != nil {
					goto END
				}
			}

			if cursor == 0 {
				break
			}
		}
	case "zset":
		var cursor uint64
		var fields []string
		for {
			select {
			case <-stopChan:
				goto END
			default:
			}
			fields, cursor, err = conn.ZScan(key, cursor, "*", maxRowsEachLoop).Result()
			if err != nil {
				goto END
			}
			if len(fields) > 0 {
				var members []interface{}
				members = make([]interface{}, len(fields)/2)
				for i := 0; i < len(fields); i += 2 {
					members[i/2] = fields[i]
				}
				_, err = conn.ZRem(key, members...).Result()
				if err != nil {
					goto END
				}
			}

			if cursor == 0 {
				break
			}
		}

	case "none":
		// 已经删除了的
		goto END
	default:
		return errors.New("未知的类型：" + keyType)
	}
	err = conn.Del(key).Err()
END:
	if err != nil {
		if err == redis.Nil { // 空key
			return nil
		}
		return err
	}
	return nil
}
