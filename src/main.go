// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	ps "github.com/mitchellh/go-ps"
	"github.com/soellman/pidfile"
)

const (
	issueTrackerURL   = "https://github.com/blacs30/bitwarden-alfred-workflow/issues"
	forumThreadURL    = "https://www.alfredforum.com/topic/11705-bitwarden-cli-get-passwords-username-and-totp-from-bitwarden/"
	repo              = "blacs30/bitwarden-alfred-workflow"
	CACHE_NAME        = "bw-items"
	ICON_CACHE_NAME   = "icon-items"
	FOLDER_CACHE_NAME = "bw-items-folders"
	WORKFLOW_NAME     = "bitwarden-alfred-workflow"
	AUTO_FETCH_CACHE  = "auto-fetch"
	LAST_USAGE_CACHE  = "last-usage"
	SYNC_CACHE_NAME   = "sync-cache"
)

var (
	wf *aw.Workflow
)

func init() {
	if len(os.Getenv("alfred_workflow_bundleid")) == 0 {
		if err := os.Setenv("alfred_workflow_bundleid", "com.lisowski-development.alfred.bitwarden"); err != nil {
			fmt.Println(err)
		}
	}
	if len(os.Getenv("alfred_workflow_data")) == 0 {
		if err := os.Setenv("alfred_workflow_data", "./tmp/data"); err != nil {
			fmt.Println(err)
		}
	}
	if len(os.Getenv("alfred_workflow_cache")) == 0 {
		if err := os.Setenv("alfred_workflow_cache", "./tmp/cache"); err != nil {
			fmt.Println(err)
		}
	}
	if len(os.Getenv("alfred_workflow_version")) == 0 {
		wf = aw.New()
	} else {
		wf = aw.New(update.GitHub(repo), aw.HelpURL(issueTrackerURL))
	}
	loadConfig()
}

func checkRunningProcesses(processName string) error {
	myPid := os.Getpid()
	processList, err := ps.Processes()
	if err != nil {
		return err
	}

	for x := range processList {
		process := processList[x]
		process.Executable()
		// See if there is another bitwarden process hanging which is not our own
		// ...and kill that stale process
		if process.Executable() == processName && process.Pid() != myPid {
			// log.Printf("Found stale process %d\t%s\n", process.Pid(), process.Executable())
			err := killProcess(process.Pid())
			if err != nil {
				log.Println(err)
			}
		}
	}
	return nil
}

func killProcess(pid int) error {
	// log.Printf("PID: %d will be killed.\n", pid)
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	// Kill the process
	err = proc.Kill()
	if err != nil {
		log.Println(err)
	}
	return nil
}

func pidHandler(pidfilePath string) {
	err := pidfile.Write(pidfilePath)
	if err != nil && (err == pidfile.ErrFileInvalid || err == pidfile.ErrFileStale) {
		err = os.Remove(pidfilePath)
		if err != nil {
			log.Println(err)
			return
		}
	} else if err != nil && err == pidfile.ErrProcessRunning {
		pid, err := pidfileContents(pidfilePath)
		if err != nil {
			log.Println(err)
			return
		}
		err = killProcess(pid)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func pidfileContents(filename string) (int, error) {
	ErrFileInvalid := errors.New("pidfile has invalid contents")
	contents, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(contents)))
	if err != nil {
		return 0, ErrFileInvalid
	}

	return pid, nil
}

func checkIfJobRuns() {
	if wf.IsRunning("sync") {
		wf.Rerun(0.3)
		wf.NewItem("Syncing Bitwarden secrets…").
			Icon(ReloadIcon())
		wf.SendFeedback()
		return
	}
	if wf.IsRunning("icons") {
		wf.Rerun(0.3)
		wf.NewItem("Refreshing Icon cache…").
			Icon(ReloadIcon())
	}
}

func run() {
	var err error
	if err = cli.Parse(wf.Args()); err != nil {
		if err == flag.ErrHelp {
			return
		}
		wf.FatalError(err)
	}
	opts.Query = cli.Arg(0)

	// log.Printf("%#v", opts)
	if wf.Debug() {
		// log.Printf("args=%#v => %#v", wf.Args(), cli.Args())
		// log.Print(spew.Sdump(conf))
	}

	exists := commandExists(conf.BwExec)
	if !exists && !opts.Open {
		wf.NewItem(fmt.Sprintf("Error the Bitwarden command %q wasn't found.", conf.BwExec)).
			Subtitle("Set \"BW_EXEC\" or \"PATH\" in the Workflow. Press ↩ or ⇥ for more info.").
			Valid(true).
			Arg("README.html").
			Valid(true).
			Icon(iconWarning).
			Var("action", "-open")
		wf.SendFeedback()
		return
	}

	if conf.Email == "" && !opts.SetConfigs {
		wf.NewItem("Enter your Bitwarden Email").
			Subtitle("Email not yet set. Configure your Bitwarden login email").
			UID("email").
			Valid(true).
			Icon(iconWarning).
			Var("action", "-setconfigs").
			Var("action2", "email").
			Var("notification", fmt.Sprintf("Set Email to: \n%s", opts.Query)).
			Var("title", "Set Email").
			Var("subtitle", fmt.Sprintf("Currently set to: %q (remove \"email\" from the beginning if exist)", conf.Email)).
			Arg(opts.Query)
		wf.SendFeedback()
		return
	}

	checkIfJobRuns()

	if !wf.IsRunning("sync") && !wf.IsRunning("icons") {
		pidfilePath := fmt.Sprintf("/tmp/%s", WORKFLOW_NAME)
		processName := WORKFLOW_NAME
		pidHandler(pidfilePath)
		err = checkRunningProcesses(processName)
		if err != nil {
			log.Print(err)
		}
		defer func() {
			err := pidfile.Remove(pidfilePath)
			if err != nil {
				log.Println(err)
			}
		}()
	}

	if opts.Config {
		runConfig()
		return
	}

	if opts.Auth {
		runAuth()
		return
	}

	if opts.AuthConfig {
		runAuthConfig()
		return
	}

	if opts.SetConfigs {
		runSetConfigs()
		return
	}

	if opts.Open {
		runOpen()
		return
	}

	if opts.Sync {
		runSync(opts.Force, opts.Last)
		return
	}

	if opts.Lock {
		runLock()
		return
	}

	if opts.Unlock {
		runUnlock()
		return
	}

	if opts.Login {
		runLogin()
		return
	}

	if opts.Logout {
		runLogout()
		return
	}

	if opts.Icons {
		// log.Println("Start getting icons")
		runGetIcons("", "")
		return
	}

	if opts.Search {
		// log.Print("Number of flags", cli.NArg())
		var argString []string
		for i := 0; i < cli.NArg(); i++ {
			argString = append(argString, cli.Arg(i))
		}
		// log.Printf("argstring is %q", strings.Join(argString, " "))
		searchAlfred(strings.Join(argString, " "))
		return
	}

	if opts.GetItem {
		runGetItem()
		return
	}
	runSearch(opts.Folder, opts.Id)
}

func main() {
	wf.Run(run)
}
