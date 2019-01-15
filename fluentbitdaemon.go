package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

var fluentPID int
var running bool

func fluentBitRunner() {
	for running {
		var cmd *exec.Cmd
		if Exists("/fluent-bit/app-config/fluent-bit.conf") {
			cmd = exec.Command("/fluent-bit/bin/fluent-bit", "-c", "/fluent-bit/etc/fluent-bit-custom.conf")
		} else {
			cmd = exec.Command("/fluent-bit/bin/fluent-bit", "-c", "/fluent-bit/etc/fluent-bit.conf")
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Start()
		fluentPID = cmd.Process.Pid
		cmd.Wait()
		fluentPID = -1
	}
}

func killFluentBit() {
	if fluentPID > 0 {
		syscall.Kill(-fluentPID, syscall.SIGHUP)
		fluentPID = -1
	}
}

var ch = make(chan int, 10)

func fluentBitDaemon() {
	for op := range ch {
		switch op {
		case 1:
			killFluentBit()
		}
	}
}

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
	killFluentBit()
	os.Exit(0)
}

func configReloadHandler(w http.ResponseWriter, r *http.Request) {
	ch <- 1
	fmt.Fprintf(w, `{"ok": true}`)
}

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()
	defer glog.Flush()

	glog.Info("Start Fluent-Bit daemon...\n")

	exitHandler()

	fluentPID = -1
	running = true
	go fluentBitRunner()
	go fluentBitDaemon()

	http.HandleFunc("/api/config.reload", configReloadHandler)
	glog.Fatal(http.ListenAndServe(":24444", nil))
}
