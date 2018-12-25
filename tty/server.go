package tty

import (
	"context"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	noesctmpl "text/template"
	"time"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/yudai/gotty/pkg/homedir"
	"github.com/yudai/gotty/webtty"
)

// Server provides a webtty HTTP endpoint.
type Server struct {
	factory Factory
	options *Options

	upgrader      *websocket.Upgrader
	indexTemplate *template.Template
	titleTemplate *noesctmpl.Template
}

// New creates a new instance of Server.
// Server will use the New() of the factory provided to handle each request.
func New(factory Factory, options *Options) (*Server, error) {
	indexData, err := Asset("resources/index.html")
	if err != nil {
		panic("index not found") // must be in bindata
	}
	if options.IndexFile != "" {
		path := homedir.Expand(options.IndexFile)
		indexData, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read custom index file at `%s`", path)
		}
	}
	indexTemplate, err := template.New("index").Parse(string(indexData))
	if err != nil {
		panic("index template parse failed") // must be valid
	}

	titleTemplate, err := noesctmpl.New("title").Parse(options.TitleFormat)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse window title format `%s`", options.TitleFormat)
	}

	var originChekcer func(r *http.Request) bool
	if options.WSOrigin != "" {
		matcher, err := regexp.Compile(options.WSOrigin)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to compile regular expression of Websocket Origin: %s", options.WSOrigin)
		}
		originChekcer = func(r *http.Request) bool {
			return matcher.MatchString(r.Header.Get("Origin"))
		}
	}

	return &Server{
		factory: factory,
		options: options,

		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			Subprotocols:    webtty.Protocols,
			CheckOrigin:     originChekcer,
		},
		indexTemplate: indexTemplate,
		titleTemplate: titleTemplate,
	}, nil
}

// Handle starts the main process of the Server.
// The cancelation of ctx will shutdown the server immediately with aborting
// existing connections. Use WithGracefullContext() to support gracefull shutdown.
func (server *Server) Handle(router *gin.RouterGroup) error {
	counter := newCounter(time.Duration(server.options.Timeout) * time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	server.setupHandlers(router, ctx, cancel, counter)
	return nil
}

func (server *Server) setupHandlers(router *gin.RouterGroup, ctx context.Context, cancel context.CancelFunc, counter *counter) {
	router.StaticFS("css/", &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "resources"})
	router.StaticFS("js/", &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "resources"})
	group := router.Group("/")
	{
		group.Any("auth_token.js", func(c *gin.Context) {
			server.handleAuthToken(c.Writer, c.Request)
		})
		group.Any("config.js", func(c *gin.Context) {
			server.handleConfig(c.Writer, c.Request)
		})
		handleWS := server.generateHandleWS(ctx, cancel, counter)
		group.Any("ws", func(c *gin.Context) {
			handleWS.ServeHTTP(c.Writer, c.Request)
		})
		group.Any("/", func(c *gin.Context) {
			server.handleIndex(c.Writer, c.Request)
		})
	}
}
