package server

import (
	"io/ioutil"
	"os"

	"container/list"

	"fmt"

	"net/http"
	_ "net/http/pprof"

	"sync"

	"database/sql"

	"time"

	"github.com/cocktail18/redis-manager/tty"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

type RedisConn struct {
	Initialize      bool
	delQueue        *list.List
	conn            *redis.Client
	cfg             RedisConfig
	db              *sql.DB
	currentDB       int32
	cacheDir        string
	refreshStatus   int32 // 1 刷新中
	refreshChan     chan struct{}
	lastRefreshTime time.Time
}

type Server struct {
	cfg        Config
	cfgPath    string
	redisConn  map[int]*RedisConn
	sessionMgr *SessionMgr
	router     *gin.Engine
	exit       chan struct{}
	waitGroup  sync.WaitGroup
	ttySrv     *tty.Server
}

func NewServer(cfgPath string) (srv *Server, err error) {
	data, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return
	}
	srv = &Server{}
	err = yaml.Unmarshal(data, &(srv.cfg))
	if err != nil {
		return
	}
	if srv.cfg.Debug {
		go func() {
			log.Println(http.ListenAndServe("127.0.0.1:6060", nil))
		}()
	}
	if srv.cfg.CacheDir == "" {
		srv.cfg.CacheDir = os.TempDir() + "/redis-manager-cache"
	}
	if srv.cfg.CacheSize == 0 {
		srv.cfg.CacheSize = 1000000
	}
	srv.exit = make(chan struct{}, 1)
	srv.sessionMgr = NewSessionMgr(3600)
	return
}

func (srv *Server) GetCfg(serverID int) RedisConfig {
	return srv.cfg.RedisServer[serverID]
}

func (srv *Server) Run() (err error) {
	srv.redisConn = make(map[int]*RedisConn)
	defer func() {
		if err != nil {
			srv.Close()
		}
	}()
	// 连接redis
	for id, redisCfg := range srv.cfg.RedisServer {
		var conn *redis.Client
		conn, err = newRedisConn(redisCfg)
		if err != nil {
			log.Fatalf("connect to redis error: %s", err.Error())
		}
		if redisCfg.Name == "" {
			redisCfg.Name = fmt.Sprintf("%s:%d:%d", redisCfg.Host, redisCfg.Port, redisCfg.DB)
		}
		if redisCfg.ScanRowNumEachLoop == 0 {
			redisCfg.ScanRowNumEachLoop = DEFAULT_SCAN_ROW_NUM_EACH_LOOP
		}
		srv.redisConn[id] = &RedisConn{conn: conn, cfg: redisCfg, delQueue: list.New(), Initialize: !redisCfg.PreLoad, refreshChan: make(chan struct{})}
	}
	srv.initHandler()
	srv.refreshRedisKeyDaemon()
	return srv.router.Run(srv.cfg.Listen)
}

func (srv *Server) Close() {
	close(srv.exit)
	for _, conn := range srv.redisConn {
		conn.conn.Close()
	}

}

func (srv *Server) Wait() {
	srv.waitGroup.Wait()
}

func (srv *Server) redisServerInfo(context *gin.Context) {
	id, _ := getServerID(context)
	conn := srv.redisConn[id]
	info, err := conn.conn.Info().Result()
	if err != nil {
		context.JSON(200, newResponseProto(500, err.Error(), nil))
		return
	}
	context.JSON(200, newResponseProto(200, "success", info))
}

func (srv *Server) redisDBSize(context *gin.Context) {
	id, _ := getServerID(context)
	conn := srv.redisConn[id]
	size, err := conn.conn.DbSize().Result()
	if err != nil {
		context.JSON(200, newResponseProto(500, err.Error(), nil))
		return
	}
	context.JSON(200, newResponseProto(200, "success", size))
}
