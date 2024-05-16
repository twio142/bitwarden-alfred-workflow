// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
	"sort"

	"github.com/blacs30/bitwarden-alfred-workflow/alfred"
	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/util"
)

var (
	opts = &options{}
	cli  = flag.NewFlagSet("bitwarden-alfred-workflow", flag.ContinueOnError)
)

// CLI flags
type options struct {
	// Commands
	Search       bool
	Config       bool
	SetConfigs   bool
	Auth         bool
	OnOffConfigs bool
	AuthConfig   bool
	Lock         bool
	Icons        bool
	Folder       bool
	Favorites    bool
	Unlock       bool
	Login        bool
	Logout       bool
	Sync         bool
	Open         bool
	GetItem      bool
	GetTotp      bool

	// Options
	Force      bool
	Totp       bool
	Last       bool
	Background bool

	// Arguments
	Id         string
	Query      string
	Attachment string
	Output     string
}

func init() {
	cli.BoolVar(&opts.Search, "search", false, "run a new search with options")
	cli.BoolVar(&opts.Config, "conf", false, "show/filter configuration")
	cli.BoolVar(&opts.SetConfigs, "setconfigs", false, "set configs")
	cli.BoolVar(&opts.Auth, "auth", false, "show/filter auth configuration")
	cli.BoolVar(&opts.AuthConfig, "authconfig", false, "Display Auth config options")
	cli.BoolVar(&opts.Open, "open", false, "open specified file in default app")
	cli.BoolVar(&opts.Lock, "lock", false, "lock Bitwarden")
	cli.BoolVar(&opts.Unlock, "unlock", false, "unlock Bitwarden")
	cli.BoolVar(&opts.Icons, "icons", false, "Get favicons")
	cli.BoolVar(&opts.Folder, "folder", false, "Filter Bitwarden Folders")
	cli.BoolVar(&opts.Favorites, "favorites", false, "Search Bitwarden Favorite Items")
	cli.StringVar(&opts.Id, "id", "", "Get item by id")
	cli.StringVar(&opts.Attachment, "attachment", "", "set attachment id")
	cli.BoolVar(&opts.Login, "login", false, "login to Bitwarden")
	cli.BoolVar(&opts.Logout, "logout", false, "logout Bitwarden")
	cli.BoolVar(&opts.Sync, "sync", false, "sync secrets")
	cli.BoolVar(&opts.Background, "background", false, "Run job in background")
	cli.BoolVar(&opts.Last, "last", false, "last sync")
	cli.BoolVar(&opts.Force, "force", false, "force full sync")
	cli.BoolVar(&opts.Totp, "totp", false, "get totp for item id")
	cli.BoolVar(&opts.GetTotp, "gettotp", false, "get totp the other way")
	cli.BoolVar(&opts.GetItem, "getitem", false, "get item and an object of it")

	cli.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: bitwarden-alfred-workflow [options] [arguments]

Alfred workflow to get secrets from Bitwarden.

Usage:
    bitwarden-alfred-workflow [<query>]
    bitwarden-alfred-workflow -auth [<query>]
    bitwarden-alfred-workflow -conf [<query>]
    bitwarden-alfred-workflow -folder [<query>]
	bitwarden-alfred-workflow -favorites
    bitwarden-alfred-workflow -getitem -id <id> [-totp] [-attachment <id>] [<query>] (query is used as jsonpath)
    bitwarden-alfred-workflow -icons [-background]
    bitwarden-alfred-workflow -lock
    bitwarden-alfred-workflow -login
    bitwarden-alfred-workflow -logout
    bitwarden-alfred-workflow -open [<query>]
    bitwarden-alfred-workflow -output <query>
    bitwarden-alfred-workflow -search <query>
    bitwarden-alfred-workflow -setsfaconfig [<setting>]
    bitwarden-alfred-workflow -authconfig [<query>]
    bitwarden-alfred-workflow -sync [-force|-last] [-background]
    bitwarden-alfred-workflow -unlock
    bitwarden-alfred-workflow -h|-help

Options:
`)
		cli.PrintDefaults()
	}
}

func BitwardenAuthChecks() (loginErr error, unlockErr error) {
	args := fmt.Sprintf("%s login --quiet --check", conf.BwExec)
	if wf.Debug() {
		args = fmt.Sprintf("%s login --check", conf.BwExec)
	}
	_, loginErr = runCmd(args, NOT_LOGGED_IN_MSG)
	if wf.Debug() {
		if loginErr != nil {
			log.Println("[ERROR] ==> ", loginErr)
		}
	}

	noQuiet := "--quiet"
	if wf.Debug() {
		noQuiet = ""
	}
	token, err := alfred.GetToken(wf)
	if err != nil {
		args = fmt.Sprintf("%s unlock %s --check", conf.BwExec, noQuiet)
	} else {
		// workaround for https://github.com/bitwarden/clients/issues/2729
		// args = fmt.Sprintf("%s unlock %s --check --session %s", conf.BwExec, noQuiet, token)
		args = fmt.Sprintf("%s list folders --nointeraction --session %s", conf.BwExec, token)
		// end workaround
	}
	_, unlockErr = runCmd(args, NOT_UNLOCKED_MSG)
	if wf.Debug() {
		if unlockErr != nil {
			log.Println("[ERROR] ==> ", unlockErr)
		}
	}
	return
}

// Filter configuration in Alfred
func runConfig() {

	// prevent Alfred from re-ordering results
	if opts.Query == "" || conf.ReorderingDisabled {
		wf.Configure(aw.SuppressUIDs(true))
	}

	// get current email
	email := conf.Email
	server := conf.Server

	wf.NewItem("Enter Login Email").
		UID("email").
		Valid(true).
		Icon(iconEmailAt).
		Var("action", "-setconfigs").
		Var("action2", "email").
		Var("notification", fmt.Sprintf("Set Email to: \n%s", opts.Query)).
		Var("title", "Set Email").
		Var("subtitle", fmt.Sprintf("Currently set to: %q (remove \"email\" from the beginning if exist)", email)).
		Arg(opts.Query)

	wf.NewItem("Set Server URL").
		Subtitle("Only for self hosting users").
		UID("server").
		Valid(true).
		Icon(iconServer).
		Var("action", "-setconfigs").
		Var("action2", "server").
		Var("notification", fmt.Sprintf("Set Server to: \n%s", opts.Query)).
		Var("title", "Set Server").
		Var("subtitle", fmt.Sprintf("Currently set to: %q", server)).
		Arg(opts.Query)

	wf.NewItem("Set WebUI URL").
		Subtitle("Only for self hosting users").
		UID("webui").
		Valid(true).
		Icon(iconBw).
		Var("action", "-setconfigs").
		Var("action2", "webui").
		Var("notification", fmt.Sprintf("Set WebUI URL to: \n%s", opts.Query)).
		Var("title", "Set WebUI URL").
		Var("subtitle", fmt.Sprintf("Currently set to: %q", conf.WebUiURL)).
		Arg(opts.Query)

	wf.NewItem("Enable or disable 2FA").
		Subtitle("Configure Bitwarden to use or not use 2 Factor Authentication").
		UID("sfa").
		Valid(true).
		Icon(iconUserClock).
		Var("action", "-authconfig").
		Var("action2", "-id on-off-sfa")

	wf.NewItem("Enable or disable API Key login").
		Subtitle("Configure Bitwarden to use API keys to login").
		UID("apikeyauth").
		Valid(true).
		Icon(iconUserClock).
		Var("action", "-authconfig").
		Var("action2", "-id on-off-apikey")

	wf.NewItem("Set the 2FA method").
		Subtitle("Configure which 2 Factor Authentication Method you use").
		UID("sfamode").
		Valid(true).
		Icon(iconUserClock).
		Var("action", "-authconfig").
		Var("action2", "-id Use")

	wf.NewItem("Delete Workflow cache").
		Subtitle("↩ or ⇥ to clean cached items and icons").
		Valid(false).
		UID("delcache").
		Autocomplete("workflow:delcache").
		Icon(aw.IconTrash)

	wf.NewItem("View Help File").
		Subtitle("Open workflow help in your browser").
		Arg("README.html").
		UID("help").
		Valid(true).
		Icon(iconHelp).
		Var("action", "-open")

	wf.NewItem("Report Issue").
		Subtitle("Open workflow issue tracker in your browser").
		Arg(issueTrackerURL).
		UID("issue").
		Valid(true).
		Icon(iconIssue).
		Var("action", "-open")

	wf.NewItem("Visit Forum Thread").
		Subtitle("Open workflow thread on alfredforum.com in your browser").
		Arg(forumThreadURL).
		UID("forum").
		Valid(true).
		Icon(iconLink).
		Var("action", "-open")

	wf.NewItem("Sync Bitwarden Secrets").
		Valid(true).
		UID("sync").
		Icon(iconReload).
		Var("action", "-sync").
		Var("action2", "-force").
		Var("notification", "Syncing Bitwarden secrets").
		Arg("-background")

	wf.NewItem("Remove Workflow cache").
		Valid(false).
		UID("delcache").
		Autocomplete("workflow:delcache").
		Icon(aw.IconTrash)

	wf.NewItem("Enable or Disable API Key login").
		UID("apikeyauth").
		Valid(true).
		Icon(iconUserClock).
		Var("action", "-authconfig").
		Var("action2", "-id on-off-apikey")

	wf.NewItem("Enable or Disable 2FA").
		UID("sfa").
		Valid(true).
		Icon(iconUserClock).
		Var("action", "-authconfig").
		Var("action2", "-id on-off-sfa")

	wf.NewItem("Set 2FA Method").
		UID("sfamode").
		Valid(true).
		Icon(iconUserClock).
		Var("action", "-authconfig").
		Var("action2", "-id Use")

	wf.NewItem("Report an Issue").
		Arg(issueTrackerURL).
		UID("issue").
		Valid(true).
		Icon(iconIssue).
		Var("action", "-open")

	wf.NewItem("Download/ Update Favicon for URLs").
		Valid(true).
		UID("icons").
		Icon(iconReload).
		Var("action", "-icons").
		Var("notification", "Downloading Favicons for URLs").
		Arg("-background")

	if opts.Query != "" {
		wf.Filter(opts.Query)
	}

	wf.WarnEmpty("No Config Found", "Try a different query?")
	wf.SendFeedback()
}

// Open path/URL
func runOpen() {
	wf.Configure(aw.TextErrors(true))

	var args []string
	args = append(args, opts.Query)

	cmd := exec.Command("/usr/bin/open", args...)
	if _, err := util.RunCmd(cmd); err != nil {
		wf.Fatalf("/usr/bin/open %q: %v", opts.Query, err)
	}
}

// Filter auth config in Alfred
func runAuth() {

	// prevent Alfred from re-ordering results
	if opts.Query == "" {
		wf.Configure(aw.SuppressUIDs(true))
	}
	email := conf.Email
	sfaMode := -1
	if conf.Sfa {
		sfaMode = conf.SfaMode
	}

	if bwData.UserId == "" {
		addLoginItem(email, sfaMode)
	} else {
		wf.NewItem("Logout").
			UID("logout").
			Valid(true).
			Icon(iconOff).
			Var("action", "-logout")
	}

	if bwData.UserId != "" && bwData.ProtectedKey != "" {
		addUnlockItem(email)
	} else {
		wf.NewItem("Lock").
			UID("lock").
			Valid(true).
			Icon(iconOff).
			Var("action", "-lock")
	}

	if opts.Query != "" {
		wf.Filter(opts.Query)
	}

	wf.WarnEmpty("No Auth Config Found", "Try a different query?")
	wf.SendFeedback()
}

func addLoginItem(email string, sfaMode int) {
	wf.NewItem("Login to Bitwarden").
		Valid(true).
		UID("login").
		Icon(iconOn).
		Var("action", "-login").
		Var("type", "login").
		Var("email", email).
		Var("sfamode", fmt.Sprintf("%d", sfaMode)).
		Var("mapsfamode", map2faMode(sfaMode))
}

func addUnlockItem(email string) {
	wf.NewItem("Unlock").
		UID("unlock").
		Valid(true).
		Icon(iconOn).
		Var("action", "-unlock").
		Var("type", "unlock").
		Var("email", email)
}

// Logout from Bitwarden
func runSetConfigs() {
	wf.Configure(aw.TextErrors(true))

	if cli.NFlag() > 0 {
		var err error
		mode := cli.Arg(0)
		value := cli.Arg(1)
		switch mode {
		case "email":
			err = alfred.SetEmail(wf, value)
		case "server":
			if value == "" {
				value = "https://bitwarden.com"
			}
			if cli.NArg() > 2 {
				value = cli.Arg(1)
				for i := 2; i < cli.NArg(); i++ {
					value = fmt.Sprintf("%s %s", value, cli.Arg(i))
				}
			}
			command := fmt.Sprintf("%s config server %s", conf.BwExec, value)
			message := fmt.Sprintf("Unable to set Bitwarden server %s", value)
			_, err := runCmd(command, message)

			if err != nil {
				wf.FatalError(err)
			}
			err = alfred.SetServer(wf, value)
			if err != nil {
				wf.FatalError(err)
			}
		case "webui":
			if value == "" {
				value = "https://vault.bitwarden.com"
			}
			if cli.NArg() > 2 {
				value = cli.Arg(1)
				for i := 2; i < cli.NArg(); i++ {
					value = fmt.Sprintf("%s %s", value, cli.Arg(i))
				}
			}
			err = alfred.SetWebUiUrl(wf, value)
			if err != nil {
				wf.FatalError(err)
			}
		case "2fa":
			err = alfred.SetSfa(wf, value)
		case "2famode":
			err = alfred.SetSfaMode(wf, value)
			if err != nil {
				wf.FatalError(err)
			}
			sfaModeValue, err := strconv.Atoi(value)
			if err != nil {
				log.Println(err)
			}
			sfamode := map2faMode(sfaModeValue)
			fmt.Printf("DONE: Set %s to \n%s", mode, sfamode)
			searchAlfred(conf.BwconfKeyword)
			return
		case "apikey":
			err = alfred.SetApikey(wf, value)
		}
		if err != nil {
			wf.FatalError(err)
		}
		fmt.Printf("DONE: Set %s to: \n%s", mode, value)
		searchAlfred(conf.BwconfKeyword)
	}
}

// Logout from Bitwarden
func runAuthConfig() {
	wf.Configure(aw.TextErrors(true))

	if opts.Id == "Use" {
		// https://github.com/bitwarden/jslib/blob/master/common/src/enums/twoFactorProviderType.ts
		factorMap := []struct {
			title string
			uid   string
			icon  *aw.Icon
			name  string
			enum  string
		}{
			{
				title: "Use Authenticator app",
				uid:   "totp",
				icon:  iconApp,
				name:  "Authenticator app",
				enum:  "0",
			},
			{
				title: "Use Email",
				uid:   "email",
				icon:  iconEmail,
				name:  "Email",
				enum:  "1",
			},
			{
				title: "Use Yubikey OTP",
				uid:   "yubikey",
				icon:  iconYubi,
				name:  "Yubikey",
				enum:  "3",
			},
		}
		for _, item := range factorMap {
			wf.NewItem(item.title).
				Subtitle(fmt.Sprintf("Currently set to: %s", map2faMode(conf.SfaMode))).
				UID(item.uid).
				Valid(true).
				Icon(item.icon).
				Var("notification", fmt.Sprintf("2FA set to %s", item.name)).
				Var("action", "-setconfigs").
				Var("action2", "2famode").
				Arg(item.enum)
		}
	} else if opts.Id == "on-off-sfa" {
		wf.NewItem("ON/OFF: Enable 2FA for Bitwarden").
			Subtitle(fmt.Sprintf("Currently set to: %t", conf.Sfa)).
			UID("sfaon").
			Valid(true).
			Icon(iconOn).
			Var("notification", "Enabled 2FA").
			Var("action", "-setconfigs").
			Var("action2", "2fa").
			Arg("true")

		wf.NewItem("ON/OFF: Disable 2FA for Bitwarden").
			Subtitle(fmt.Sprintf("Currently set to: %t", conf.Sfa)).
			UID("sfaoff").
			Valid(true).
			Icon(iconOff).
			Var("notification", "Disabled 2FA").
			Var("action", "-setconfigs").
			Var("action2", "2fa").
			Arg("false")
	} else if opts.Id == "on-off-apikey" {
		wf.NewItem("ON/OFF: Enable APIKEY login for Bitwarden").
			Subtitle(fmt.Sprintf("Currently set to: %t", conf.UseApikey)).
			UID("apikeyon").
			Valid(true).
			Icon(iconOn).
			Var("notification", "Enabled APIKEY login").
			Var("action", "-setconfigs").
			Var("action2", "apikey").
			Arg("true")

		wf.NewItem("ON/OFF: Disable APIKEY login for Bitwarden").
			Subtitle(fmt.Sprintf("Currently set to: %t", conf.UseApikey)).
			UID("sfaoff").
			Valid(true).
			Icon(iconOff).
			Var("notification", "Disabled APIKEY login").
			Var("action", "-setconfigs").
			Var("action2", "apikey").
			Arg("false")
	}

	if opts.Query != "" {
		wf.Filter(opts.Query)
	}

	wf.SendFeedback()
}

// Filter Bitwarden secrets in Alfred
func runSearch(folderSearch bool, itemId string, favoritesSearch bool) {
	email := conf.Email
	sfaMode := -1
	if conf.Sfa {
		sfaMode = conf.SfaMode
	}

	wf.Configure(aw.SuppressUIDs(true))
	if bwData.UserId == "" {
		// message := "Need to login first."
		// if wf.Cache.Exists(CACHE_NAME) && wf.Cache.Exists(FOLDER_CACHE_NAME) {
		// 	message = "Need to login first to get secrets, reading cached items without the secret."
		// }
		// wf.NewWarningItem("Not logged in to Bitwarden.", message)
		addLoginItem(email, sfaMode)
	}

	if bwData.UserId != "" && bwData.ProtectedKey == "" {
		// message := "Need to unlock first to get secrets, reading cached items without the secrets."
		// wf.NewWarningItem("Bitwarden is locked.", message)
		addUnlockItem(email)
	}

	if conf.ReorderingDisabled {
		wf.Configure(aw.SuppressUIDs(true))
	} else {
		wf.Configure(aw.SuppressUIDs(false))
	}

	wf.Configure(aw.MaxResults(conf.MaxResults))

	// Load data
	var items []Item
	var folders []Folder

	// check if the data cache exists
	if wf.Cache.Exists(CACHE_NAME) && wf.Cache.Exists(FOLDER_CACHE_NAME) {
		data, err := Decrypt()
		if err != nil {
			log.Printf("Error decrypting data: %s", err)
		}
		if err := json.Unmarshal(data, &items); err != nil {
			log.Printf("Couldn't load the items cache, error: %s", err)
		}
		if err := wf.Cache.LoadJSON(FOLDER_CACHE_NAME, &folders); err != nil {
			log.Printf("Couldn't load the folders cache, error: %s", err)
		}
	}

	// Check if the sync cache exists
	if !wf.Cache.Exists(SYNC_CACHE_NAME) && !wf.Cache.Exists(CACHE_NAME) {
		if !wf.IsRunning("sync") {
			wf.NewItem("Cache Expired or not Existed").
				Subtitle("Sync now").
				Valid(true).
				UID("sync").
				Icon(iconReload).
				Var("action", "-sync").
				Var("action2", "-force").
				Var("notification", "Syncing Bitwarden secrets").
				Arg("-background")
			wf.SendFeedback()
			return
		}
		wf.NewItem("Refreshing Cache…").
			Icon(ReloadIcon())
		wf.SendFeedback()
		return
	}

	// If iconcache enabled and the cache is expired (or doesn't exist)
	if conf.IconCacheEnabled && (wf.Data.Expired(ICON_CACHE_NAME, conf.IconMaxCacheAge) || !wf.Data.Exists(ICON_CACHE_NAME)) {
		// getIcon(wf)
		wf.NewItem("Favicon Cache Expired or not Existed").
			Subtitle("Download Favicons now").
			Valid(true).
			UID("icons").
			Icon(iconReload).
			Var("action", "-icons").
			Var("notification", "Downloading Favicons for URLs").
			Arg("-background")
		wf.SendFeedback()
		return
	}

	// set lastUsageCache after all the config and auth options and cache checks ran
	// it's only set when a search  is successfully ready to be executed
	timestamp := time.Now().Unix()
	err := wf.Cache.Store(LAST_USAGE_CACHE, []byte(strconv.FormatInt(timestamp, 10)))
	if err != nil {
		log.Println(err)
	}

	if folderSearch && itemId == "" {
		runSearchFolder(items, folders)
	}

	autoFetchCache := false
	if wf.Cache.Expired(AUTO_FETCH_CACHE, conf.AutoFetchIconMaxCacheAge) || !wf.Cache.Exists(AUTO_FETCH_CACHE) {
		autoFetchCache = true
		err := wf.Cache.Store(AUTO_FETCH_CACHE, []byte(string("auto-fetch-cache")))
		if err != nil {
			log.Println(err)
		}
	}

	if itemId != "" && !folderSearch {
		// Add item to workflow for itemId
		for _, item := range items {
			if item.Id == itemId {
				addItemDetails(item, autoFetchCache)
				wf.SendFeedback()
				return
			}
		}
	}

	if itemId != "" && folderSearch {
		// Add item to search folders
		for _, item := range items {
			if item.FolderId == itemId {
				addItemsToWorkflow(item, autoFetchCache)
			}
			if itemId == "null" {
				if item.FolderId == "" {
					addItemsToWorkflow(item, autoFetchCache)
				}
			}
		}
		wf.NewItem("Go Back to Folder Search").
			Valid(true).
			UID("").
			Icon(iconFolder).
			Var("action", "-search").
			Arg(conf.BwfKeyword).
			Match(".")
	}

	if len(items) == 0 && len(folders) == 0 {
		wf.NewItem("No Secrets Found").Subtitle("Try a different query or sync manually").Icon(iconWarning).Valid(false)
	}

	if !folderSearch && itemId == "" && !favoritesSearch {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Favorite && !items[j].Favorite
		})
		for _, item := range items {
			addItemsToWorkflow(item, autoFetchCache)
		}
	}

	if favoritesSearch {
		for _, item := range items {
			if item.Favorite {
				addItemsToWorkflow(item, autoFetchCache)
			}
		}
		wf.NewItem("Go Back to Folder Search").
			Valid(true).
			UID("").
			Icon(iconFolder).
			Var("action", "-search").
			Arg(conf.BwfKeyword).
			Match(".")
	}

	wf.SendFeedback()
}

// Filter Bitwarden secrets in Alfred
func runSearchFolder(items []Item, folders []Folder) {
	wf.NewItem("Favorites").
		Subtitle(fmt.Sprintf("%d items", getFavoriteItemsCount(items))).
		Valid(true).
		UID("favorites").
		Icon(iconStar).
		Var("action", "-favorites")

	for _, folder := range folders {
		itemCount := getItemsInFolderCount(folder.Id, items)
		id := "null"
		if folder.Id != "" {
			id = folder.Id
		}
		if opts.Query != "" {
			wf.NewItem(folder.Name).
				Subtitle(fmt.Sprintf("%d items", itemCount)).Valid(true).
				UID(id).
				Icon(iconFolderOpen).
				Var("action", "-folder").
				Var("action2", fmt.Sprintf("-id %s ", id))
		} else {
			wf.NewItem(folder.Name).
				Subtitle(fmt.Sprintf("%d items", itemCount)).Valid(true).
				UID(id).
				Icon(iconFolderOpen).
				Var("action", "-folder").
				Var("action2", fmt.Sprintf("-id %s ", id))
		}
	}

	if len(items) == 0 && len(folders) == 0 {
		wf.WarnEmpty("No Secrets Found", "Try a different query or sync manually")
	}

	addBackToNormalSearchItem()
	wf.WarnEmpty("No Folders Found", "Try a different query")
	wf.SendFeedback()
}
