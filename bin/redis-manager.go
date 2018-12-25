package main

import (
	"log"

	"os"
	"os/signal"

	"syscall"

	"github.com/cocktail18/redis-manager"
)

func main() {
	server, err := redis_manager.NewServer("conf.yaml")
	if err != nil {
		log.Panic(err)
	}
	c := make(chan os.Signal)
	//监听所有信号
	signal.Notify(c, syscall.SIGHUP)
	go server.Run()

	<-c
	server.Close()
	server.Wait()
}
