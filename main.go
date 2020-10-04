package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"os/exec"

	"github.com/fatih/color"
	"github.com/radovskyb/watcher"
)

var w *watcher.Watcher

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("usage: onsave <command> [args]")
	}
	// create watcher and terminate if there is any error
	w = watcher.New()
	defer w.Close()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		err := w.Add(scanner.Text())
		if err != nil {
			log.Println("Error:", err.Error())
		} else {
			fmt.Println("+", scanner.Text())
		}
	}

	mainLoop(w, os.Args[1], os.Args[2:]...)
}

func mainLoop(w *watcher.Watcher, callbackCommand string, callbackArgs ...string) {
	timeout := time.Second
	reset := time.Now().Add(-timeout)
	bold := color.New(color.Bold)

	var cmd *exec.Cmd

	go func() {
		for {
			select {
			case event := <-w.Event:
				// only run at most once every `timeout` seconds
				log.Println("event:", event)
				if reset.Add(timeout).Before(time.Now()) {
					if cmd != nil {
						bold.Println("File change detected. Restarting process...")
						// signal the process to quit, and kill it after a timeout.
						go func() {
							time.Sleep(5 * time.Second)
							_ = cmd.Process.Signal(os.Kill)
							if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
								bold.Printf("[WARN] Kill %d, %d returned error: %s\n", cmd.Process.Pid, syscall.SIGKILL, err)
							}
						}()
						if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil {
							bold.Printf("[WARN] Kill %d, %d returned error: %s\n", cmd.Process.Pid, syscall.SIGTERM, err)
						}
					}

					bold.Println("$", callbackCommand, strings.Join(callbackArgs, " "))
					cmd = exec.Command(callbackCommand, callbackArgs...)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
					cmd.Start()
					reset = time.Now()
				}

			case err := <-w.Error:
				log.Println("error:", err)

			case <-w.Closed:
				return
			}
		}
	}()

	go w.TriggerEvent(watcher.Write, nil)
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}
}
