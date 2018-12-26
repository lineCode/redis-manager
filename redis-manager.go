package main

import (
	"log"

	"os"
	"os/signal"

	"syscall"

	"github.com/cocktail18/redis-manager/server"
	"flag"
)

var configFile string

func init()  {
	conf := flag.String("c", "", "config file")
	flag.Parse()
	if *conf == "" {
		configFile = "./conf.yaml"
	}else{
		configFile = *conf
	}
}

func main() {
	srv, err := server.NewServer(configFile)
	if err != nil {
		log.Panic(err)
	}
	c := make(chan os.Signal)
	//监听所有信号
	signal.Notify(c, syscall.SIGHUP)
	go srv.Run()

	<-c
	srv.Close()
	srv.Wait()
}
