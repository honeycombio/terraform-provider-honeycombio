package v2

import (
	"time"
)

type Environment struct {
	ID   string `jsonapi:"primary,environments"`
	Name string `jsonapi:"attr,name"`
	Slug string `jsonapi:"attr,slug"`
}

type Team struct {
	ID   string `jsonapi:"primary,teams"`
	Name string `jsonapi:"attr,name"`
	Slug string `jsonapi:"attr,slug"`
}

type Timestamps struct {
	CreatedAt time.Time `jsonapi:"attr,created,rfc3339,omitempty"`
	UpdatedAt time.Time `jsonapi:"attr,updated,rfc3339,omitempty"`
}

type AuthMetadata struct {
	ID         string      `jsonapi:"primary,api-keys"`
	Name       string      `jsonapi:"attr,name"`
	KeyType    string      `jsonapi:"attr,key_type"`
	Disabled   bool        `jsonapi:"attr,disabled"`
	Scopes     []string    `jsonapi:"attr,scopes"`
	Timestamps *Timestamps `jsonapi:"attr,timestamps"`
	Team       *Team       `jsonapi:"relation,team"`
}

type APIKey struct {
	ID          string             `jsonapi:"primary,api-keys,omitempty"`
	Name        *string            `jsonapi:"attr,name,omitempty"`
	KeyType     string             `jsonapi:"attr,key_type,omitempty"`
	Disabled    *bool              `jsonapi:"attr,disabled,omitempty"`
	Secret      string             `jsonapi:"attr,secret,omitempty"`
	Permissions *APIKeyPermissions `jsonapi:"attr,permissions,omitempty"`
	Timestamps  *Timestamps        `jsonapi:"attr,timestamps,omitempty"`
	Environment *Environment       `jsonapi:"relation,environment"`
}

type APIKeyPermissions struct {
	CreateDatasets bool `json:"create_datasets" jsonapi:"attr,create_datasets"`
}
