package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
)

var running bool

func exitHandler() {
	c := make(chan os.Signal)

	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				ExitFunc()
			case syscall.SIGUSR1:
			case syscall.SIGUSR2:
			default:
			}
		}
	}()
}

func ExitFunc() {
	running = false
	os.Exit(0)
}

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()
	defer glog.Flush()

	glog.Info("Fluent-Bit disabled...\n")

	exitHandler()

	running = true

	for running {
		time.Sleep(24 * 3600 * time.Second)
	}
}
