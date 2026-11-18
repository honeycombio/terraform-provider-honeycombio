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
	Description *string            `jsonapi:"attr,description,omitempty"`
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

// TelemetrySchema is a Team's custom telemetry schema.
//
// SchemaYAML is an opaque string. The API stores and returns it
// byte-for-byte, so it must never be parsed or re-serialised on the way through.
type TelemetrySchema struct {
	// ID is the slug of the Team which owns the schema.
	ID         string `jsonapi:"primary,telemetry-schema,omitempty"`
	SchemaYAML string `jsonapi:"attr,schema_yaml"`
}

// Group is a set of Team members who share the access the Group is granted.
type Group struct {
	ID          string      `jsonapi:"primary,groups,omitempty"`
	Name        string      `jsonapi:"attr,name,omitempty"`
	Description *string     `jsonapi:"attr,description,omitempty"`
	MemberCount int         `jsonapi:"attr,member_count,omitempty"`
	Timestamps  *Timestamps `jsonapi:"attr,timestamps,omitempty"`
}

// GroupMember names one user in a Group by the user's ID.
type GroupMember struct {
	ID string `jsonapi:"primary,users"`
}

// GroupDataset names one dataset a Group can read by the dataset's ID.
type GroupDataset struct {
	ID string `jsonapi:"primary,datasets"`
}

// User is a member of the Team.
type User struct {
	ID    string `jsonapi:"primary,users"`
	Email string `jsonapi:"attr,email"`
	Role  string `jsonapi:"attr,role"`
}

// Dataset is a dataset in an Environment, as the v2 API returns it.
//
// An update sends only the fields that are set, so the read-only fields
// are all omitted when empty.
type Dataset struct {
	ID                  string           `jsonapi:"primary,datasets"`
	Name                string           `jsonapi:"attr,name,omitempty"`
	Slug                string           `jsonapi:"attr,slug,omitempty"`
	Description         *string          `jsonapi:"attr,description,omitempty"`
	ExpandJSONDepth     *int             `jsonapi:"attr,expand_json_depth,omitempty"`
	Settings            *DatasetSettings `jsonapi:"attr,settings,omitempty"`
	Restricted          *bool            `jsonapi:"attr,restricted,omitempty"`
	DatasetType         string           `jsonapi:"attr,dataset_type,omitempty"`
	RegularColumnsCount int              `jsonapi:"attr,regular_columns_count,omitempty"`
	LastWrittenAt       *time.Time       `jsonapi:"attr,last_written_at,rfc3339,omitempty"`
	CreatedAt           *time.Time       `jsonapi:"attr,created_at,rfc3339,omitempty"`
}

type DatasetSettings struct {
	DeleteProtected *bool `json:"delete_protected" jsonapi:"attr,delete_protected,omitempty"`
}
