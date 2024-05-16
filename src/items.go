// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT License - http://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
)

func checkIconExistence(item Item, autoFetchCache bool) *aw.Icon {
	icon := iconLink
	if len(item.Login.Uris) > 0 && conf.IconCacheEnabled {
		iconPath := fmt.Sprintf("%s/%s/%s.png", wf.DataDir(), "urlicon", item.Id)
		if _, err := os.Stat(iconPath); err != nil {
			// log.Println("Couldn't load the cached icon, error: ", err)
			if autoFetchCache {
				// log.Println("Getting icons.")
				runGetIcons(item.Login.Uris[0].Uri, item.Id)
			}
		}
		icon = &aw.Icon{Value: iconPath}
	}
	return icon
}

func addBackToNormalSearchItem() {
	wf.NewItem("Go Back to Item Search").
		Valid(true).
		Icon(iconLevelUp).
		Var("action", "-search").
		Arg(conf.BwKeyword).
		Match(".")
}

func addItemDetails(item Item, autoFetchCache bool) {
	wf.Configure(aw.SuppressUIDs(true))
	if (conf.EmptyDetailResults && item.Type != 2) || (item.Type != 2 && item.Notes != "") {
		wf.NewItem(fmt.Sprintf("Note: %s", item.Notes)).
			Arg(item.Notes).
			Icon(iconNote).
			Var("sound", "true").
			Var("action", "output").Valid(true)
	} else if (item.Type == 2 && conf.EmptyDetailResults) || (item.Type == 2 && item.Notes != "") {
		wf.NewItem(fmt.Sprintf("Secure Note: %s", item.Notes)).
			Icon(iconNote).
			Var("sound", "true").
			Var("action", "-getitem").
			Var("action2", fmt.Sprintf("-id %s", item.Id)).
			Arg("notes").Valid(true) // used as jsonpath
	}
	if conf.EmptyDetailResults || item.Favorite {
		wf.NewItem("Favorite").
			Arg(strconv.FormatBool(item.Favorite)).
			Icon(iconStar).Valid(false)
	}

	// item.Fields
	if len(item.Fields) > 0 {
		for k, field := range item.Fields {
			// counter := k + 1
			// it's a secret type so we need to fetch the secret from Bitwarden
			if field.Type == 1 {
				if field.Name == "TOTP" {
					wf.NewItem("TOTP: ✳✳✳✳✳︎︎︎︎︎").
						Icon(iconUserClock).
						Var("sound", "true").
						Var("action", "-gettotp").
						Var("action2", fmt.Sprintf("-id %s", item.Id)).
						Arg(fmt.Sprintf("fields[%d].value", k)).
						Valid(true)
				} else {
					wf.NewItem(fmt.Sprintf("%s: %s", field.Name, field.Value)).
						Icon(iconBars).
						Var("sound", "true").
						Var("action", "-getitem").
						Var("action2", fmt.Sprintf("-id %s", item.Id)).
						Arg(fmt.Sprintf("fields[%d].value", k)). // used as jsonpath
						Valid(true)
				}
			} else {
				wf.NewItem(fmt.Sprintf("%s: %s", field.Name, field.Value)).
					Arg(field.Value).
					Icon(iconBars).
					Var("sound", "true").
					Var("action", "output").Valid(true)
			}
		}
	}
	// item.Attachments
	if len(item.Attachments) > 0 {
		for _, att := range item.Attachments {
			// it's a secret type so we need to fetch the secret from Bitwarden
			wf.NewItem(fmt.Sprintf("Attachment: %s (%s)", att.FileName, att.SizeName)).
				Subtitle(fmt.Sprintf("Save attachment to %s", conf.OutputFolder)).
				Icon(iconPaperClip).
				Valid(true).
				Var("notification", fmt.Sprintf("Attachment saved to:\n%s%s", conf.OutputFolder, att.FileName)).
				Var("action", "-getitem").
				Var("action2", fmt.Sprintf("-attachment %s", att.Id)).
				Var("action3", fmt.Sprintf("-id %s", item.Id))
		}
	}
	// item.CollectionIds
	if conf.EmptyDetailResults || len(item.CollectionIds) > 0 {
		wf.NewItem(fmt.Sprintf("Collection Ids: %s", strings.Join(item.CollectionIds, ", "))).
			Arg(fmt.Sprintf("%q", strings.Join(item.CollectionIds, ","))).
			Icon(iconBoxes).
			Var("sound", "true").
			Var("action", "output").Valid(true)

	}
	// item.RevisionDate
	if conf.EmptyDetailResults || fmt.Sprint(item.RevisionDate) != "" {
		dateSlice := strings.Split(fmt.Sprint(item.RevisionDate), " ")
		wf.NewItem(fmt.Sprintf("Revised on %s", dateSlice[0])).
			Arg(fmt.Sprint(dateSlice[0])).
			Icon(iconCalDay).
			Valid(false)
	}
	// specific items for login type
	// item.Type 1
	if item.Type == 1 {
		// get icons from cache
		icon := checkIconExistence(item, autoFetchCache)

		// item.Login.Username
		if conf.EmptyDetailResults || item.Login.Username != "" {
			wf.NewItem(fmt.Sprintf("Username: %s", item.Login.Username)).
				Valid(true).
				Arg(item.Login.Username).
				Icon(iconUser).
				Var("action", "output").Valid(true).
				Var("sound", "true")
		}
		// item.Login.Uris[*].Uri
		if len(item.Login.Uris) > 0 {
			for _, uri := range item.Login.Uris {
				// counter := k + 1
				wf.NewItem(fmt.Sprintf("URL: %s", uri.Uri)).
					Valid(true).
					Arg(uri.Uri).
					Icon(icon).
					Var("action", "-open").Valid(true)
			}
		}
		// TOTP
 		/* if item.Login.Totp != "" {
			wf.NewItem(fmt.Sprintf("TOTP: %s", item.Login.Totp)).
				Valid(true).
				Icon(iconUserClock).
				Var("sound", "true").
				Var("action", "-getitem").
				Var("action2", "-totp").
				Var("action3", fmt.Sprintf("-id %s", item.Id))
		} */
		// Password Revision Date
		// check if the set value matches the initial value of time, then we know the passwordRevisionDate hasn't been set by Bitwarden
		d1 := time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC)
		datesEqual := d1.Equal(item.Login.PasswordRevisionDate)
		if !datesEqual {
			dateSlice := strings.Split(fmt.Sprint(item.Login.PasswordRevisionDate), " ")
			wf.NewItem(fmt.Sprintf("Password Revised on %s", dateSlice[0])).
				Valid(true).
				Icon(iconDate).
				Arg(fmt.Sprint(dateSlice[0])).
				Valid(false)
		}
	} else if item.Type == 3 {
		if conf.EmptyDetailResults || item.Card.Number != "" {
			wf.NewItem(fmt.Sprintf("Card Number: %s", item.Card.Number)).
				Valid(true).
				Icon(iconCreditCard).
				Var("sound", "true").
				Var("action", "-getitem").
				Var("action2", fmt.Sprintf("-id %s", item.Id)).
				Arg("card.number")
		}
		if conf.EmptyDetailResults || item.Card.Code != "" {
			wf.NewItem(fmt.Sprintf("Card Security Code: %s", item.Card.Code)).
				Valid(true).
				Icon(iconPassword).
				Var("sound", "true").
				Var("action", "-getitem").
				Var("action2", fmt.Sprintf("-id %s", item.Id)).
				Arg("card.code")
		}
		if item.Card.ExpMonth != "" && item.Card.ExpYear != "" {
			wf.NewItem(fmt.Sprintf("Expiration Date: %s%s", item.Card.ExpMonth, item.Card.ExpYear[2:])).
				Valid(true).
				Icon(iconDate).
				Arg(fmt.Sprintf("%s%s", item.Card.ExpMonth, item.Card.ExpYear[2:])).
				Var("sound", "true").
				Var("action", "output")
		} else {
			if conf.EmptyDetailResults || item.Card.ExpMonth != "" {
				wf.NewItem(fmt.Sprintf("Expiration Month: %s", item.Card.ExpMonth)).
					Valid(true).
					Icon(iconDate).
					Arg(item.Card.ExpMonth).
					Var("sound", "true").
					Var("action", "output")
			}
			if conf.EmptyDetailResults || item.Card.ExpYear != "" {
				wf.NewItem(fmt.Sprintf("Expiration Year: %s", item.Card.ExpYear)).
					Valid(true).
					Icon(iconDate).
					Arg(item.Card.ExpYear).
					Var("sound", "true").
					Var("action", "output")
			}
		}
		if conf.EmptyDetailResults || item.Card.CardHolderName != "" {
			wf.NewItem(fmt.Sprintf("Card Holder: %s", item.Card.CardHolderName)).
				Valid(true).
				Icon(iconUser).
				Arg(item.Card.CardHolderName).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Card.Brand != "" {
			wf.NewItem(fmt.Sprintf("Card Brand: %s", item.Card.Brand)).
				Valid(true).
				Icon(iconCreditCardRegular).
				Arg(item.Card.Brand).
				Var("sound", "true").
				Var("action", "output")
		}
	} else if item.Type == 4 {
		if conf.EmptyDetailResults || item.Identity.Title != "" {
			wf.NewItem(fmt.Sprintf("Title: %s", item.Identity.Title)).
				Valid(true).
				Icon(iconIdBatch).
				Arg(item.Identity.Title).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.FirstName != "" {
			wf.NewItem(fmt.Sprintf("First Name: %s", item.Identity.FirstName)).
				Valid(true).
				Icon(iconIdBatch).
				Arg(item.Identity.FirstName).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.MiddleName != "" {
			wf.NewItem(fmt.Sprintf("Middle Name: %s", item.Identity.MiddleName)).
				Valid(true).
				Icon(iconIdBatch).
				Arg(item.Identity.MiddleName).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.LastName != "" {
			wf.NewItem(fmt.Sprintf("Last Name: %s", item.Identity.LastName)).
				Valid(true).
				Icon(iconIdBatch).
				Arg(item.Identity.LastName).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Address1 != "" {
			wf.NewItem(fmt.Sprintf("Address 1: %s", item.Identity.Address1)).
				Valid(true).
				Icon(iconHome).
				Arg(item.Identity.Address1).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Address2 != "" {
			wf.NewItem(fmt.Sprintf("Address 2: %s", item.Identity.Address2)).
				Valid(true).
				Icon(iconHome).
				Arg(item.Identity.Address2).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Address3 != "" {
			wf.NewItem(fmt.Sprintf("Address 3: %s", item.Identity.Address3)).
				Valid(true).
				Icon(iconHome).
				Arg(item.Identity.Address3).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.City != "" {
			wf.NewItem(fmt.Sprintf("City: %s", item.Identity.City)).
				Valid(true).
				Icon(iconCity).
				Arg(item.Identity.City).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.State != "" {
			wf.NewItem(fmt.Sprintf("State: %s", item.Identity.State)).
				Valid(true).
				Icon(iconMap).
				Arg(item.Identity.State).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.PostalCode != "" {
			wf.NewItem(fmt.Sprintf("Postal Code: %s", item.Identity.PostalCode)).
				Valid(true).
				Icon(iconMap).
				Arg(item.Identity.PostalCode).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Country != "" {
			wf.NewItem(fmt.Sprintf("Country: %s", item.Identity.Country)).
				Valid(true).
				Icon(iconMap).
				Arg(item.Identity.Country).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Company != "" {
			wf.NewItem(fmt.Sprintf("Company: %s", item.Identity.Company)).
				Valid(true).
				Icon(iconOrg).
				Arg(item.Identity.Company).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Email != "" {
			wf.NewItem(fmt.Sprintf("Email: %s", item.Identity.Email)).
				Valid(true).
				Icon(iconEmailAt).
				Arg(item.Identity.Email).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Phone != "" {
			wf.NewItem(fmt.Sprintf("Phone: %s", item.Identity.Phone)).
				Valid(true).
				Icon(iconPhone).
				Arg(item.Identity.Phone).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Ssn != "" {
			wf.NewItem(fmt.Sprintf("Social Security Number: %s", item.Identity.Ssn)).
				Valid(true).
				Icon(iconIdCard).
				Arg(item.Identity.Ssn).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Username != "" {
			wf.NewItem(fmt.Sprintf("Username: %s", item.Identity.Username)).
				Valid(true).
				Icon(iconUser).
				Arg(item.Identity.Username).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.PassportNumber != "" {
			wf.NewItem(fmt.Sprintf("Passport Number: %s", item.Identity.PassportNumber)).
				Valid(true).
				Icon(iconIdBatch).
				Arg(item.Identity.PassportNumber).
				Var("sound", "true").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.LicenseNumber != "" {
			wf.NewItem(fmt.Sprintf("License Number: %s", item.Identity.LicenseNumber)).
				Valid(true).
				Icon(iconIdBatch).
				Arg(item.Identity.LicenseNumber).
				Var("sound", "true").
				Var("action", "output")
		}
	}
	addBackToNormalSearchItem()
}

func addItemsToWorkflow(item Item, autoFetchCache bool) {
	var template = map[string]modifierActionRelation{
		"nomod": {}, "mod1": {}, "mod2": {}, "mod3": {}, "mod4": {},
	}
	var itemModSet = map[string]map[string]modifierActionRelation{
		"item1": template, "item2": template, "item3": template, "item4": template,
	}

	if item.Type == 1 {
		// get icons from cache
		icon := checkIconExistence(item, autoFetchCache)

		totpEmoji, err := getTypeEmoji("totp")
		if err != nil {
			log.Fatal(err.Error())
		}
		totp := fmt.Sprintf("%s *TOTP, ", totpEmoji)
		if len(item.Login.Totp) == 0 {
			totp = ""
		}

		getModifierActionRelations(itemModSet, item, "item1", icon, totp)
		addNewItem(itemModSet["item1"], item.Name)
	} else if item.Type == 2 {
		getModifierActionRelations(itemModSet, item, "item2", nil, "")
		addNewItem(itemModSet["item2"], item.Name)
	} else if item.Type == 3 {
		getModifierActionRelations(itemModSet, item, "item3", nil, "")
		addNewItem(itemModSet["item3"], item.Name)
	} else if item.Type == 4 {
		getModifierActionRelations(itemModSet, item, "item4", nil, "")
		addNewItem(itemModSet["item4"], item.Name)
	}
}

func addNewItem(item map[string]modifierActionRelation, name string) *aw.Item {
	sound := ""
	if item["nomod"].Content.Sound {
		sound = "true"
	}
	it := wf.NewItem(item["nomod"].Content.Title).
		Subtitle(item["nomod"].Content.Subtitle).Valid(true).
		Arg(item["nomod"].Content.Arg).
		UID(name).
		Var("action", item["nomod"].Content.Action).
		Var("action2", item["nomod"].Content.Action2).
		Var("action3", item["nomod"].Content.Action3).
		Var("sound", sound).
		Arg(item["nomod"].Content.Arg).
		Icon(item["nomod"].Content.Icon)
	if item["mod1"].Keys != nil {
		addNewModifierItem(it, item["mod1"])
	}
	if item["mod2"].Keys != nil {
		addNewModifierItem(it, item["mod2"])
	}
	if item["mod3"].Keys != nil {
		addNewModifierItem(it, item["mod3"])
	}
	if item["mod4"].Keys != nil {
		addNewModifierItem(it, item["mod4"])
	}
	if item["mod5"].Keys != nil {
		addNewModifierItem(it, item["mod5"])
	}
	return it
}

func addNewModifierItem(item *aw.Item, modifier modifierActionRelation) {
	sound := ""
	if modifier.Content.Sound {
		sound = "true"
	}
	item.NewModifier(modifier.Keys[0:]...).
		Subtitle(modifier.Content.Subtitle).
		Arg(modifier.Content.Arg).
		Var("action", modifier.Content.Action).
		Var("action2", modifier.Content.Action2).
		Var("action3", modifier.Content.Action3).
		Var("sound", sound).
		Arg(modifier.Content.Arg).
		Icon(modifier.Content.Icon)
}
