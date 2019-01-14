package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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

var cmd *exec.Cmd

func startFluentBit() {
	glog.Info("Start fluent-bit...\n")
	if cmd != nil {
		//glog.Infof("Cmd PID %v\n", cmd.Process.Pid)
		syscall.Kill(-cmd.Process.Pid, syscall.SIGHUP)
	}

	if Exists("/fluent-bit/app-config/fluent-bit.conf") {
		cmd = exec.Command("/fluent-bit/bin/fluent-bit", "-c", "/fluent-bit/etc/fluent-bit-custom.conf", ">&1", "2>&2")
	} else {
		cmd = exec.Command("/fluent-bit/bin/fluent-bit", "-c", "/fluent-bit/etc/fluent-bit.conf", ">&1", "2>&2")
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Start()
}

var ch = make(chan int, 10)

func fluentBitDaemon() {
	for op := range ch {
		switch op {
		case 1:
			startFluentBit()
		}
	}
}

func configReloadHandler(w http.ResponseWriter, r *http.Request) {
	ch <- 1
	fmt.Fprintf(w, `{"ok": true}`)
}

func main() {
	flag.Parse()
	defer glog.Flush()

	glog.Info("Start Fluent-Bit daemon...\n")
	go fluentBitDaemon()
	ch <- 1
	http.HandleFunc("/api/config.reload", configReloadHandler)
	glog.Fatal(http.ListenAndServe(":24444", nil))
}
