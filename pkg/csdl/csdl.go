package csdl

type Edmx struct {
	Reference    []Reference  `xml:"Reference"`
	DataServices DataServices `xml:"DataServices"`
}

type Reference struct {
	Uri     string    `xml:"Uri,attr"`
	Include []Include `xml:"Include"`
	// Also could have IncludeAnnotations, but I don't have a use case for that right now
}

type Include struct {
	Namespace string `xml:"Namespace,attr"`
	Alias     string `xml:"Alias,attr"`
	Other     []any  `xml:",any"`
}

type DataServices struct {
	Schema []Schema `xml:"Schema"`
}

type Schema struct {
	Namespace       string            `xml:"Namespace,attr"`
	Alias           string            `xml:"Alias,attr"`
	Action          []Action          `xml:"Action"`
	Annotations     []Annotations     `xml:"Annotations"`
	Annotation      []Annotation      `xml:"Annotation"`
	ComplexType     []ComplexType     `xml:"ComplexType"`
	EntityContainer []EntityContainer `xml:"EntityContainer"`
	EntityType      []EntityType      `xml:"EntityType"`
	EnumType        []EnumType        `xml:"EnumType"`
	Function        []Function        `xml:"Function"`
	Term            []Term            `xml:"Term"`
	TypeDefinition  []TypeDefinition  `xml:"TypeDefinition"`
	Other           []any             `xml:",any"`
}

type Action struct {
	Name       string       `xml:"Name,attr"`
	IsBound    bool         `xml:"IsBound,attr"`
	Annotation []Annotation `xml:"Annotation"`
	Parameter  []Parameter  `xml:"Parameter"`
	Other      []any        `xml:",any"`
}

type Annotation struct {
	Term       string  `xml:"Term,attr"`
	Qualifier  string  `xml:"Qualifier,attr"`
	String     string  `xml:"String,attr"`
	EnumMember string  `xml:"EnumMember,attr"`
	Bool       bool    `xml:"Bool,attr"`
	Int        int64   `xml:"Int,attr"`
	Decimal    float64 `xml:"Decimal,attr"`
}

type Annotations struct {
	Target     string       `xml:"Target,attr"`
	Qualifier  string       `xml:"Qualifier,attr"`
	Annotation []Annotation `xml:"Annotation"`
}

type ComplexType struct {
	Name               string               `xml:"Name,attr"`
	BaseType           string               `xml:"BaseType,attr"`
	Abstract           bool                 `xml:"Abstract,attr"`
	OpenType           bool                 `xml:"OpenType,attr"`
	Property           []Property           `xml:"Property"`
	NavigationProperty []NavigationProperty `xml:"NavigationProperty"`
	Annotation         []Annotation         `xml:"Annotation"`
}

type EntityContainer struct {
	Other []any `xml:",any"`
}

type EntityType struct {
	Name               string               `xml:"Name,attr"`
	BaseType           string               `xml:"BaseType,attr"`
	Abstract           bool                 `xml:"Abstract,attr"`
	OpenType           bool                 `xml:"OpenType,attr"`
	HasStream          bool                 `xml:"HasStream,attr"`
	Key                Key                  `xml:"Key"`
	Property           []Property           `xml:"Property"`
	NavigationProperty []NavigationProperty `xml:"NavigationProperty"`
}

type EnumType struct {
	Name           string `xml:"Name,attr"`
	UnderlyingType string `xml:"UnderlyingType,attr"`
	IsFlags        bool   `xml:"IsFlags,attr"`
	Annotation     []Annotation
	Member         []Member
}

type Function struct {
	Other []any `xml:",any"`
}

type Key struct {
}

type Member struct {
	Name       string       `xml:"Name,attr"`
	Value      *string      `xml:"Value,attr"`
	Annotation []Annotation `xml:"Annotation"`
}

type NavigationProperty struct {
	Name           string       `xml:"Name,attr"`
	Type           string       `xml:"Type,attr"`
	Nullable       *bool        `xml:"Nullable,attr"`
	Partner        string       `xml:"Partner,attr"`
	ContainsTarget bool         `xml:"ContainsTarget,attr"`
	Annotation     []Annotation `xml:"Annotation"`
}

type Parameter struct {
	Other []any `xml:",any"`
}

type Property struct {
	Name         string       `xml:"Name,attr"`
	Type         string       `xml:"Type,attr"`
	Nullable     *bool        `xml:"Nullable,attr"`
	MaxLength    int          `xml:"MaxLength,attr"`
	Unicode      bool         `xml:"Unicode,attr"`
	Precision    int          `xml:"Precision,attr"`
	Scale        int          `xml:"Scale,attr"`
	SRID         string       `xml:"SRID,attr"`
	DefaultValue string       `xml:"DefaultValue,attr"`
	Annotation   []Annotation `xml:"Annotation"`
}

type Term struct {
	Other []any `xml:",any"`
}

type TypeDefinition struct {
	Name           string       `xml:"Name,attr"`
	UnderlyingType string       `xml:"UnderlyingType,attr"`
	Annotation     []Annotation `xml:"Annotation"`
}
