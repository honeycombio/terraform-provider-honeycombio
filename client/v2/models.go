package v2

import (
	"time"
)

type Environment struct {
	ID          string               `jsonapi:"primary,environments"`
	Name        string               `jsonapi:"attr,name"`
	Slug        string               `jsonapi:"attr,slug"`
	Description *string              `jsonapi:"attr,description,omitempty"`
	Color       *string              `jsonapi:"attr,color,omitempty"`
	Settings    *EnvironmentSettings `jsonapi:"attr,settings,omitempty"`
}

type EnvironmentSettings struct {
	DeleteProtected *bool `json:"delete_protected" jsonapi:"attr,delete_protected,omitempty"`
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
	SendEvents          bool `json:"send_events,omitempty" jsonapi:"attr,send_events,omitempty"`
	CreateDatasets      bool `json:"create_datasets,omitempty" jsonapi:"attr,create_datasets,omitempty"`
	ManageQueries       bool `json:"manage_columns,omitempty" jsonapi:"attr,manage_columns,omitempty"`
	RunQueries          bool `json:"run_queries,omitempty" jsonapi:"attr,run_queries,omitempty"`
	ReadServiceMaps     bool `json:"read_service_maps,omitempty" jsonapi:"attr,read_service_maps,omitempty"`
	ManagePublicBoards  bool `json:"manage_boards,omitempty" jsonapi:"attr,manage_boards,omitempty"`
	ManagePrivateBoards bool `json:"manage_privateBoards,omitempty" jsonapi:"attr,manage_privateBoards,omitempty"`
	ManageSLOs          bool `json:"manage_slos,omitempty" jsonapi:"attr,manage_slos,omitempty"`
	ManageTriggers      bool `json:"manage_triggers,omitempty" jsonapi:"attr,manage_triggers,omitempty"`
	ManageRecipients    bool `json:"manage_recipients,omitempty" jsonapi:"attr,manage_recipients,omitempty"`
	ManageMarkers       bool `json:"manage_markers,omitempty" jsonapi:"attr,manage_markers,omitempty"`
	VisibleToMembers    bool `json:"visible_team_members,omitempty" jsonapi:"attr,visible_team_members,omitempty"`
}
