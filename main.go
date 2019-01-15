package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"os/exec"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

var watcher *fsnotify.Watcher

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("usage: onsave <command> [args]")
	}
	// create watcher and terminate if there is any error
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		err = watcher.Add(scanner.Text())
		if err != nil {
			log.Println("Error:", err.Error())
		} else {
			fmt.Println("+", scanner.Text())
		}
	}

	mainLoop(watcher, os.Args[1], os.Args[2:]...)

	log.Printf("exited")
}

func mainLoop(watcher *fsnotify.Watcher, callbackCommand string, callbackArgs ...string) {
	timeout := 5 * time.Second
	reset := time.Now().Add(-timeout)
	bold := color.New(color.Bold)
	for {
		select {
		case event, _ := <-watcher.Events:
			// I have a weird bug where files drop off the watcher after an event is read
			watcher.Add(event.Name)

			// only run at most once every `timeout` seconds
			log.Println("event:", event)
			if reset.Add(timeout).Before(time.Now()) {
				reset = time.Now()

				bold.Println("$", callbackCommand, strings.Join(callbackArgs, " "))

				cmd := exec.Command(callbackCommand, callbackArgs...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Start()
			}

		case err, _ := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}
