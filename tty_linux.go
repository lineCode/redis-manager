package redis_manager

import (
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cocktail18/redis-manager/tty"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (srv *Server) runTTYServer(router *gin.RouterGroup) {

	backendOptions := &Options{CloseTimeout: 5}
	factory, err := NewFactory(srv, "redis-cli", nil, backendOptions)
	if err != nil {
		log.Fatalf("new factory error: %s", err.Error())
	}
	listenArr := strings.Split(srv.cfg.Listen, ":")
	if len(listenArr) != 2 {
		log.Fatalf("config listen should be ip:port")
	}

	srv.ttySrv, err = tty.New(factory, &tty.Options{
		PermitWrite:     true,
		PermitArguments: true,
		EnableReconnect: true,
		Preferences:     &tty.HtermPrefernces{CloseOnExit: false},
	})
	if err != nil {
		log.Fatalf("new tty server error: %s", err.Error())
	}
	srv.ttySrv.Handle(router)
}

type Options struct {
	CloseSignal  int `hcl:"close_signal" flagName:"close-signal" flagSName:"" flagDescribe:"Signal sent to the command process when gotty close it (default: SIGHUP)" default:"1"`
	CloseTimeout int `hcl:"close_timeout" flagName:"close-timeout" flagSName:"" flagDescribe:"Time in seconds to force kill process after client is disconnected (default: -1)" default:"-1"`
}

type Factory struct {
	command string
	argv    []string
	options *Options
	opts    []Option
	srv     *Server
}

func NewFactory(srv *Server, command string, argv []string, options *Options) (*Factory, error) {
	opts := []Option{WithCloseSignal(syscall.Signal(options.CloseSignal))}
	if options.CloseTimeout >= 0 {
		opts = append(opts, WithCloseTimeout(time.Duration(options.CloseTimeout)*time.Second))
	}

	return &Factory{
		command: command,
		argv:    argv,
		options: options,
		opts:    opts,
		srv:     srv,
	}, nil
}

func (factory *Factory) Name() string {
	return "local command"
}

func (factory *Factory) New(params map[string][]string) (tty.Slave, error) {
	if params["server_id"] != nil && len(params["server_id"]) > 0 {
		serverID, err := strconv.Atoi(params["server_id"][0])
		if err != nil {
			return nil, errors.New("server id error")
		}
		cfg := factory.srv.GetCfg(serverID)
		argv := make([]string, 0, 10)
		argv = append(argv, "-h")
		argv = append(argv, cfg.Host)
		argv = append(argv, "-p")
		argv = append(argv, strconv.Itoa(cfg.Port))
		if cfg.Password != "" {
			argv = append(argv, "-a")
			argv = append(argv, cfg.Password)
		}
		argv = append(argv, "-n")
		argv = append(argv, strconv.Itoa(cfg.DB))
		return New(factory.command, argv, factory.opts...)
	}
	return nil, errors.New("server id error")
}
