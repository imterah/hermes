package db

import (
	"gorm.io/gorm"
)

type Backend struct {
	gorm.Model

	UserID uint

	Name              string
	Description       *string
	Backend           string
	BackendParameters string

	Proxies []Proxy
}

type Proxy struct {
	gorm.Model

	BackendID uint
	UserID    uint

	Name            string
	Description     *string
	Protocol        string
	SourceIP        string
	SourcePort      uint16
	DestinationPort uint16
	AutoStart       bool
}

type Permission struct {
	gorm.Model

	PermissionNode string
	HasPermission  bool
	UserID         uint
}

type Token struct {
	gorm.Model

	UserID uint

	Token          string
	DisableExpiry  bool
	CreationIPAddr string
}

type User struct {
	gorm.Model

	Email    string `gorm:"unique"`
	Username string `gorm:"unique"`
	Name     string
	Password string
	IsBot    *bool

	Permissions   []Permission
	OwnedProxies  []Proxy
	OwnedBackends []Backend
	Tokens        []Token
}
