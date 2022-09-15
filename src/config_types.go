// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	aw "github.com/deanishe/awgo"
	"time"
)

// Config Types

type modifierActionContent struct {
	Title        string
	Subtitle     string
	Notification string
	Action       string
	Action2      string
	Action3      string
	Arg          string
	Icon         *aw.Icon
	ActionName   string
}

type modifierActionRelation struct {
	Keys    []string
	Content modifierActionContent
}

type itemsModifierActionRelationMap = map[string]map[string]modifierActionRelation

type config struct {
	// From workflow environment variables
	AutoFetchIconCacheAge    int `default:"1440" split_words:"true"`
	AutoFetchIconMaxCacheAge time.Duration
	BwconfKeyword            string
	BwauthKeyword            string
	BwKeyword                string
	BwfKeyword               string
	BwExec                   string `split_words:"true"`
	// BwDataPath default is set in loadBitwardenJSON()
	BwDataPath         string `envconfig:"BW_DATA_PATH"`
	Debug              bool   `envconfig:"DEBUG" default:"false"`
	Email              string
	EmailMaxWait       int  `envconfig:"EMAIL_MAX_WAIT" default:"15"`
	EmptyDetailResults bool `default:"false" split_words:"true"`
	IconCacheAge       int  `default:"43200" split_words:"true"`
	IconCacheEnabled   bool `default:"true" split_words:"true"`
	IconMaxCacheAge    time.Duration
	MaxResults         int    `default:"1000" split_words:"true"`
	Mod1               string `envconfig:"MODIFIER_1" default:"alt"`
	Mod1Action         string `envconfig:"MODIFIER_1_ACTION" default:"username,code"`
	Mod2               string `envconfig:"MODIFIER_2" default:"shift"`
	Mod2Action         string `envconfig:"MODIFIER_2_ACTION" default:"url"`
	Mod3               string `envconfig:"MODIFIER_3" default:"cmd"`
	Mod3Action         string `envconfig:"MODIFIER_3_ACTION" default:"totp"`
	Mod4               string `envconfig:"MODIFIER_4" default:"cmd,alt,ctrl"`
	Mod4Action         string `envconfig:"MODIFIER_4_ACTION" default:"more"`
	NoModAction        string `envconfig:"NO_MODIFIER_ACTION" default:"password,card"`
	OpenLoginUrl       bool   `envconfig:"OPEN_LOGIN_URL" default:"true"`
	OutputFolder       string `default:"" split_words:"true"`
	Path               string
	ReorderingDisabled bool   `default:"true" split_words:"true"`
	Server             string `envconfig:"SERVER_URL" default:"https://bitwarden.com"`
	Sfa                bool   `envconfig:"2FA_ENABLED" default:"true"`
	SfaMode            int    `envconfig:"2FA_MODE" default:"0"`
	SkipTypes          string `envconfig:"SKIP_TYPES" default:""`
	TitleWithUser      bool   `envconfig:"TITLE_WITH_USER" default:"true"`
	TitleWithUrls      bool   `envconfig:"TITLE_WITH_URLS" default:"true"`
	UseApikey          bool   `envconfig:"USE_APIKEY" default:"false"`
}

type BwData struct {
	path string
	// InstalledVersion is not any longer in this location in the structure of >= 1.21
	InstalledVersion string `json:"installedVersion"`
	// UserEmail is not any longer in this location in the structure of >= 1.21
	UserEmail string `json:"userEmail"`
	// UserID is not any longer in this location in the structure of >= 1.21
	UserId       string `json:"userId"`
	ActiveUserId string `json:"activeUserId"`
	// ProtectedKey is not any longer in this location in the structure of >= 1.21
	// now "__PROTECTED__<UserId>_masterkey_auto" on root level
	ProtectedKey string `json:"__PROTECTED__key"`
	// EncKey is not any longer in this location in the structure of >= 1.21
	EncKey string `json:"encKey"`
	// Kdf is not any longer in this location in the structure of >= 1.21
	Kdf int64 `json:"kdf"`
	// KdfIterations is not any longer in this location in the structure of >= 1.21
	KdfIterations int64 `json:"kdfIterations"`
	// used in >= 1.21
	Global  BwGlobalData           `json:"global"`
	Profile BwProfileData          `json:"profile"`
	Keys    BwKeyData              `json:"keys"`
	Tokens  BwTokens               `json:"tokens"`
	Unused  map[string]interface{} `json:"-"`
}
type BwGlobalData struct {
	InstalledVersion string `json:"installedVersion"`
}
type BwProfileData struct {
	EverBeenUnlocked bool   `json:"everBeenUnlocked"`
	LastSync         string `json:"lastSync"`
	KdfIterations    int64  `json:"kdfIterations"`
	KdfType          int64  `json:"kdfType"`
	Email            string `json:"email"`
	UserId           string `json:"userId"`
}
type BwTokens struct {
	AccessToken string `json:"accessToken"`
}
type BwKeyData struct {
	ApiKeyClientSecret string               `json:"apiKeyClientSecret"`
	CryptoSymmetricKey BwCryptoSymmetricKey `json:"cryptoSymmetricKey"`
	PrivateKey         BwPrivateKey         `json:"privateKey"`
}
type BwCryptoSymmetricKey struct {
	Encrypted string `json:"encrypted"`
}
type BwPrivateKey struct {
	Encrypted string `json:"encrypted"`
}
