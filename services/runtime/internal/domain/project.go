package domain

type ProjectMeta struct {
	ID         string
	Name       string
	Slug       string
	SchemaName string
	OwnerID    string
	Version    int32
}

type ProjectConfig struct {
	ProjectID  string
	Slug       string
	SchemaName string
	Version    int32
	Entities   map[string]Entity
	Modules    map[string]ModuleConfig
}

type Entity struct {
	Name string  `json:"name"`
	Pool []Field `json:"pool"`
}

type Field struct {
	Name                string   `json:"name"`
	Type                string   `json:"type"`
	Required            bool     `json:"required"`
	Unique              bool     `json:"unique"`
	System              bool     `json:"system"`
	EnumValues          []string `json:"enumValues,omitempty"`
	RelationTo          string   `json:"relationTo,omitempty"`
	RelationCardinality string   `json:"relationCardinality,omitempty"`
}

type ModuleConfig struct {
	Enabled         bool                                 `json:"enabled"`
	Functions       map[string]FunctionConfig            `json:"functions,omitempty"`
	EntityFunctions map[string]map[string]FunctionConfig `json:"entityFunctions,omitempty"`
}

type FunctionConfig struct {
	EnabledFields []string `json:"enabledFields"`
}
