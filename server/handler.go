//go:generate  go-bindata -pkg redis_manager -o asset.go -prefix frontend/dist/ frontend/dist/...
package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
	"github.com/elazarl/go-bindata-assetfs"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		var headerKeys []string
		for k, _ := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma, X-Token")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar")
			c.Set("content-type", "application/json")
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		// 处理请求
		c.Next() //  处理请求
	}
}

func AuthHandleFunc(srv *Server) gin.HandlerFunc {
	return func(context *gin.Context) {
		sessionID := srv.sessionMgr.CheckTokenValid(context.Writer, context.Request)
		userInfo, b := srv.sessionMgr.GetSessionVal(sessionID, "userinfo")
		if !b {
			context.AbortWithStatusJSON(200, newResponseProto(50014, "Token expire, please login in agin!", nil))
		} else {
			context.Set("userinfo", userInfo)
		}
	}
}

func getServerID(context *gin.Context) (int, error) {
	id := context.Param("server_id")
	return strconv.Atoi(id)
}

func CheckRedisAuthHandleFunc(srv *Server) gin.HandlerFunc {
	return func(context *gin.Context) {
		_, b := context.Get("userinfo")
		if !b {
			context.AbortWithStatusJSON(200, newResponseProto(50014, "Token expire, please login in agin!!", nil))
			return
		}
		id, err := getServerID(context)
		if err != nil {
			context.AbortWithStatusJSON(200, newResponseProto(50014, "Token expire, please login in agin!!!", nil))
			return
		}
		if _, b := srv.redisConn[id]; b {
			return
		}
		context.AbortWithStatusJSON(200, newResponseProto(50014, "Token expire, please login in agin!!!!", nil))
	}
}

func (srv *Server) initHandler() {
	router := gin.Default()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	router.Use(gin.Recovery())
	if gin.Mode() != "release" {
		router.Use(corsMiddleware())
	} else {
		log.SetLevel(log.WarnLevel)
	}
	authMiddleware := AuthHandleFunc(srv)
	checkRedisMiddleware := CheckRedisAuthHandleFunc(srv)

	// login and logout
	router.POST("/user/login", func(context *gin.Context) {
		// 检查账号密码
		p, err := newLoginProto(context.GetRawData())
		if err != nil {
			context.JSON(200, newResponseProto(50000, err.Error(), nil))
			return
		}
		for _, userInfo := range srv.cfg.User {
			if p.Username == userInfo.Username && p.Password == userInfo.Password {
				sessionID := srv.sessionMgr.StartSession(context.Writer, context.Request)
				srv.sessionMgr.SetSessionVal(sessionID, "userinfo", userInfo)
				context.AbortWithStatusJSON(200, newResponseProto(200, "", gin.H{"token": sessionID}))
				return
			}
		}
		context.AbortWithStatusJSON(200, newResponseProto(50020, "username or password is invalid!", nil))
	})

	router.POST("/user/logout", func(context *gin.Context) {
		srv.sessionMgr.EndSession(context.Writer, context.Request)
		context.AbortWithStatusJSON(200, newResponseProto(200, "", "success"))
	})

	fs := &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: ""}
	router.GET("/", func(context *gin.Context) {
		bytes, _ := fs.Asset("index.html")
		context.Writer.Write(bytes)
	})
	router.GET("/favicon.ico", func(context *gin.Context) {
		bytes, _ := fs.Asset("favicon.ico")
		context.Writer.Write(bytes)
	})
	fs.Prefix = "/static"
	router.StaticFS("/static", fs)
	authorized := router.Group("/")
	authorized.Use(authMiddleware)
	{
		srv.runTTYServer(authorized.Group("/tty"))
		authorized.GET("/user/info", func(context *gin.Context) {
			userInfo, _ := context.Get("userinfo")
			context.AbortWithStatusJSON(200, newResponseProto(200, "", userInfo))
		})
		authorized.GET("/server/list", srv.listRedisServer)

		checkedRedisServer := authorized.Group("/server").Use(checkRedisMiddleware)
		checkedRedisServer.POST("/scan/:server_id", srv.scanRedisKey)
		checkedRedisServer.POST("/add/:server_id", srv.addRedisKey)
		checkedRedisServer.POST("/save/:server_id", srv.saveRedisKey)
		checkedRedisServer.POST("/saveRow/:server_id", srv.saveRowRedisKey)
		checkedRedisServer.POST("/del/:server_id", srv.delRedisKey)
		checkedRedisServer.POST("/expire/:server_id", srv.expireRedisKey)
		checkedRedisServer.POST("/get/:server_id", srv.getRedisKey)

		checkedRedisServer.POST("/rename/:server_id", srv.renameRedisKey)
		checkedRedisServer.POST("/del-row/:server_id", srv.delRowRedisKey)
		checkedRedisServer.POST("/scan-value/:server_id", srv.valueScanRedisKey)
		checkedRedisServer.POST("/stop-del-task/:server_id", srv.stopDelTask)
		checkedRedisServer.POST("/remove-del-task/:server_id", srv.removeDelTask)
		checkedRedisServer.POST("/del-background/:server_id", srv.delBackgroundRedisKey)
		checkedRedisServer.GET("/info/:server_id", srv.redisServerInfo)
		checkedRedisServer.GET("/dbSize/:server_id", srv.redisDBSize)
		checkedRedisServer.POST("/refresh-keys/:server_id", srv.refreshKeys)

	}
	srv.router = router
}

func (srv *Server) listRedisServer(context *gin.Context) {
	var list []*ServerListProto
	for id, _ := range srv.cfg.RedisServer {
		conn := srv.redisConn[id]
		proto := &ServerListProto{}
		proto.RedisBaseCfg = RedisBaseCfg{conn.cfg.ID, conn.cfg.Name}
		proto.ID = id
		proto.Initialize = conn.Initialize
		// 获取内存使用量
		memoryInfo, err := conn.conn.Info("memory").Result()
		if err != nil {
			context.JSON(200, newResponseProto(500, err.Error(), nil))
			return
		}
		var usedMemory float64
		var maxMemory float64
		var totalSystemMemory float64
		for _, row := range strings.Split(memoryInfo, "\r\n") {
			fields := strings.Split(row, ":")
			if fields[0] == "used_memory" {
				usedMemory, _ = strconv.ParseFloat(fields[1], 10)
			}
			if fields[0] == "maxmemory" {
				maxMemory, _ = strconv.ParseFloat(fields[1], 10)
			}
			if fields[0] == "total_system_memory" {
				totalSystemMemory, _ = strconv.ParseFloat(fields[1], 10)
			}
		}
		if maxMemory == 0 {
			if totalSystemMemory == 0 {
				m, _ := mem.VirtualMemory()
				totalSystemMemory = float64(m.Total)
			}
			if totalSystemMemory != 0 {
				proto.MemoryUsedPercent = int64(usedMemory / totalSystemMemory * 100)
			}
		} else {
			proto.MemoryUsedPercent = int64(usedMemory / maxMemory * 100)
		}
		proto.LastRefreshTime = conn.lastRefreshTime.Unix()
		proto.RefreshStatus = conn.refreshStatus
		proto.Preload = conn.cfg.PreLoad
		// 获取删除key 的list
		tasks := make([]*DelTask, 0, 10)
		for e := conn.delQueue.Front(); e != nil; e = e.Next() {
			if e != nil {
				tasks = append(tasks, e.Value.(*DelTask))
			}
		}
		proto.DelTask = tasks
		list = append(list, proto)
	}
	context.JSON(200, newResponseProto(200, "", list))
}
