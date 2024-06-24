// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"context"
	"fmt"
	aw "github.com/deanishe/awgo"
	"github.com/go-cmd/cmd"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func checkReturn(status cmd.Status, message string) ([]string, error) {
	exitCode := status.Exit
	if exitCode == 127 {
		if wf.Debug() {
			log.Printf("[ERROR] ==> Exit code 127. %q not found in path %q\n", conf.BwExec, os.Getenv("PATH"))
		}
		return []string{}, fmt.Errorf("%q not found in path %q", conf.BwExec, os.Getenv("PATH"))
	} else if exitCode == 126 {
		if wf.Debug() {
			log.Printf("[ERROR] ==> Exit code 126. %q has wrong permissions. Must be executable.\n", conf.BwExec)
		}
		return []string{}, fmt.Errorf("%q has wrong permissions. Must be executable", conf.BwExec)
	} else if exitCode == 1 {
		if wf.Debug() {
			log.Println("[ERROR] ==> ", status.Stderr)
		}
		for _, stderr := range status.Stderr {
			if strings.Contains(stderr, "User cancelled.") {
				if wf.Debug() {
					log.Println("[ERROR] ==> ", stderr)
				}
				// return []string{}, fmt.Errorf("User cancelled.")
			}
		}
		errorString := strings.Join(status.Stderr[:], "")
		if wf.Debug() {
			log.Printf("[ERROR] ==> Exit code 1. %s Err: %s\n", message, errorString)
		}
		return []string{}, fmt.Errorf(fmt.Sprintf("%s Error:\n%s", message, errorString))
	} else if exitCode == 0 {
		return status.Stdout, nil
	} else {
		if wf.Debug() {
			log.Println("[DEBUG] Unexpected exit code: => ", exitCode)
			// Print each line of STDOUT and STDERR from Cmd
			for _, line := range status.Stdout {
				log.Println("[DEBUG] Stdout: => ", line)
			}
			for _, line := range status.Stderr {
				log.Println("[DEBUG] Stderr: => ", line)
			}
		}
		errMessage := ""
		for _, line := range status.Stderr {
			errMessage += fmt.Sprintf(" %s", line)
			if exitCode == -1 && strings.Contains(errMessage, "Two-step login code") {
				return []string{"Two-step login code"}, nil
			}
		}
		return []string{}, fmt.Errorf("unexpected error. Exit code %d. Has the session key changed?\n[ERROR] %s", exitCode, errMessage)
	}
}

func runCmd(args string, message string) ([]string, error) {
	// Start a long-running process, capture stdout and stderr
	argSet := strings.Fields(args)
	runCmd := cmd.NewCmd(argSet[0], argSet[1:]...)
	runCmd.Env = append(os.Environ(), "NODE_NO_WARNINGS=1")
	status := <-runCmd.Start()

	return checkReturn(status, message)
}

func runCmdWithContext(emailMaxWait int, args string, message string) ([]string, error) {
	// Start a long-running process, capture stdout and stderr
	c, cancel := context.WithTimeout(context.Background(), time.Duration(emailMaxWait)*time.Second)
	defer cancel()

	argSet := strings.Fields(args)
	runCmd := cmd.NewCmd(argSet[0], argSet[1:]...)
	runCmd.Env = append(os.Environ(), "NODE_NO_WARNINGS=1")

	select {
	case status := <-runCmd.Start():
		return checkReturn(status, message)
	case <-c.Done():
		log.Print(c.Err())
		// return runCmd.Status().Stderr, nil
		return checkReturn(runCmd.Status(), message)
	}
}

func searchAlfred(search string) {
	// Open Alfred
	a := aw.NewAlfred()
	err := a.Search(search)
	if err != nil {
		log.Println(err)
	}
}

func getItemsInFolderCount(folderId string, items []Item) int {
	counter := 0
	for _, item := range items {
		if item.FolderId == folderId {
			counter += 1
		}
	}
	return counter
}

func getFavoriteItemsCount(items []Item) int {
	counter := 0
	for _, item := range items {
		if item.Favorite {
			counter += 1
		}
	}
	return counter
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func map2faMode(mode int) string {
	switch mode {
	case 0:
		return "Authenticator-app"
	case 1:
		return "Email"
	case 2:
		return "Duo"
	case 3:
		return "YubiKey"
	case 4:
		return "U2F"
	}
	return " "
}

func clearCache() error {
	err := wf.Cache.StoreJSON(CACHE_NAME, nil)
	if err != nil {
		return err
	}
	err = wf.Cache.StoreJSON(FOLDER_CACHE_NAME, nil)
	if err != nil {
		return err
	}
	err = wf.Cache.StoreJSON(AUTO_FETCH_CACHE, nil)
	if err != nil {
		return err
	}
	return nil
}

func debugLog(message string) {
	if conf.Debug {
		log.Print("[DEBUG] ", message)
	}
}
