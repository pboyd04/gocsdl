package csdl

import "sync/atomic"

type Edmx struct {
	Reference    []*Reference `xml:"Reference"`
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
	Schema []*Schema `xml:"Schema"`
}

type Schema struct {
	Namespace       string             `xml:"Namespace,attr"`
	Alias           string             `xml:"Alias,attr"`
	Action          []*Action          `xml:"Action"`
	Annotations     []*Annotations     `xml:"Annotations"`
	Annotation      []*Annotation      `xml:"Annotation"`
	ComplexType     []*ComplexType     `xml:"ComplexType"`
	EntityContainer []*EntityContainer `xml:"EntityContainer"`
	EntityType      []*EntityType      `xml:"EntityType"`
	EnumType        []*EnumType        `xml:"EnumType"`
	Function        []*Function        `xml:"Function"`
	Term            []*Term            `xml:"Term"`
	TypeDefinition  []*TypeDefinition  `xml:"TypeDefinition"`
	Other           []any              `xml:",any"`

	children atomic.Pointer[[]Searchable]
}

func (s *Schema) GetChildren() []Searchable {
	retVal := s.children.Load()
	if retVal != nil {
		return *retVal
	}
	ret := make([]Searchable, 0, 100)
	ret = utilAppendTypeAsSearchable(ret, s.Action...)
	ret = utilAppendTypeAsSearchable(ret, s.Annotations...)
	ret = utilAppendTypeAsSearchable(ret, s.Annotation...)
	ret = utilAppendTypeAsSearchable(ret, s.ComplexType...)
	ret = utilAppendTypeAsSearchable(ret, s.EntityContainer...)
	ret = utilAppendTypeAsSearchable(ret, s.EntityType...)
	ret = utilAppendTypeAsSearchable(ret, s.EnumType...)
	// ret = utilAppendTypeAsSearchable(ret, s.Function...)
	ret = utilAppendTypeAsSearchable(ret, s.Term...)
	ret = utilAppendTypeAsSearchable(ret, s.TypeDefinition...)
	s.children.Store(&ret)
	return ret
}

func (s *Schema) GetName() string {
	return s.Namespace
}

func (s *Schema) GetXPath() string {
	return s.Namespace
}

type Action struct {
	Name       string        `xml:"Name,attr"`
	IsBound    bool          `xml:"IsBound,attr"`
	Annotation []*Annotation `xml:"Annotation"`
	Parameter  []*Parameter  `xml:"Parameter"`
	Other      []any         `xml:",any"`
}

func (a Action) GetChildren() []Searchable {
	ret := make([]Searchable, 0, 100)
	ret = utilAppendTypeAsSearchable(ret, a.Annotation...)
	ret = utilAppendTypeAsSearchable(ret, a.Parameter...)
	return ret
}

func (a Action) GetName() string {
	return a.Name
}

func (a Action) GetXPath() string {
	return a.Name
}

type Annotation struct {
	Term       string  `xml:"Term,attr"`
	Qualifier  string  `xml:"Qualifier,attr"`
	String     string  `xml:"String,attr"`
	EnumMember string  `xml:"EnumMember,attr"`
	Bool       bool    `xml:"Bool,attr"`
	Int        int64   `xml:"Int,attr"`
	Decimal    float64 `xml:"Decimal,attr"`
	Record     Record  `xml:"Record"`
}

func (a Annotation) GetChildren() []Searchable {
	return nil
}

func (a Annotation) GetName() string {
	return a.Term
}

func (a Annotation) GetXPath() string {
	return "Annotation[@Term='" + a.Term + "']"
}

type Annotations struct {
	Target     string        `xml:"Target,attr"`
	Qualifier  string        `xml:"Qualifier,attr"`
	Annotation []*Annotation `xml:"Annotation"`
}

func (a Annotations) GetChildren() []Searchable {
	ret := make([]Searchable, 0, len(a.Annotation))
	ret = utilAppendTypeAsSearchable(ret, a.Annotation...)
	return ret
}

func (a Annotations) GetName() string {
	return a.Target
}

func (a Annotations) GetXPath() string {
	return "Annotations[@Target='" + a.Target + "']"
}

type ComplexType struct {
	Name               string                `xml:"Name,attr"`
	BaseType           string                `xml:"BaseType,attr"`
	Abstract           bool                  `xml:"Abstract,attr"`
	OpenType           bool                  `xml:"OpenType,attr"`
	Property           []*Property           `xml:"Property"`
	NavigationProperty []*NavigationProperty `xml:"NavigationProperty"`
	Annotation         []*Annotation         `xml:"Annotation"`
	schema             *Schema
}

func (c ComplexType) GetChildren() []Searchable {
	ret := make([]Searchable, 0, 100)
	ret = utilAppendTypeAsSearchable(ret, c.Annotation...)
	ret = utilAppendTypeAsSearchable(ret, c.Property...)
	ret = utilAppendTypeAsSearchable(ret, c.NavigationProperty...)
	return ret
}

func (c ComplexType) GetName() string {
	return c.Name
}

func (c ComplexType) GetSchema() *Schema {
	return c.schema
}

func (c ComplexType) GetXPath() string {
	return "ComplexType[@Name='" + c.Name + "']"
}

func (c *ComplexType) SetSchema(schema *Schema) {
	c.schema = schema
}

type EntityContainer struct {
	Name    string `xml:"Name,attr"`
	Extends string `xml:"Extends,attr"`
	Other   []any  `xml:",any"`
}

func (e *EntityContainer) GetName() string {
	return e.Name
}

func (e *EntityContainer) GetChildren() []Searchable {
	return nil
}

func (e *EntityContainer) GetXPath() string {
	return "EntityContainer[@Name='" + e.Name + "']"
}

type EntityType struct {
	Name               string                `xml:"Name,attr"`
	BaseType           string                `xml:"BaseType,attr"`
	Abstract           bool                  `xml:"Abstract,attr"`
	OpenType           bool                  `xml:"OpenType,attr"`
	HasStream          bool                  `xml:"HasStream,attr"`
	Key                Key                   `xml:"Key"`
	Property           []*Property           `xml:"Property"`
	NavigationProperty []*NavigationProperty `xml:"NavigationProperty"`
	Annotation         []*Annotation         `xml:"Annotation"`
	schema             *Schema
	children           atomic.Pointer[[]Searchable]
}

func (e *EntityType) GetChildren() []Searchable {
	if children := e.children.Load(); children != nil {
		return *children
	}
	ret := make([]Searchable, 0, 100)
	ret = utilAppendTypeAsSearchable(ret, e.Property...)
	ret = utilAppendTypeAsSearchable(ret, e.NavigationProperty...)
	ret = utilAppendTypeAsSearchable(ret, e.Annotation...)
	e.children.Store(&ret)
	return ret
}

func (e *EntityType) GetName() string {
	return e.Name
}

func (e *EntityType) GetSchema() *Schema {
	return e.schema
}

func (e *EntityType) GetXPath() string {
	return "EntityType[@Name='" + e.Name + "']"
}

func (e *EntityType) SetSchema(schema *Schema) {
	e.schema = schema
}

type EnumType struct {
	Name           string `xml:"Name,attr"`
	UnderlyingType string `xml:"UnderlyingType,attr"`
	IsFlags        bool   `xml:"IsFlags,attr"`
	Annotation     []*Annotation
	Member         []*Member
}

func (e EnumType) GetChildren() []Searchable {
	ret := make([]Searchable, 0, 100)
	ret = utilAppendTypeAsSearchable(ret, e.Annotation...)
	ret = utilAppendTypeAsSearchable(ret, e.Member...)
	return ret
}

func (e EnumType) GetName() string {
	return e.Name
}

func (e EnumType) GetXPath() string {
	return "EnumType[@Name='" + e.Name + "']"
}

type Function struct {
	Other []any `xml:",any"`
}

type Key struct {
}

type Member struct {
	Name       string        `xml:"Name,attr"`
	Value      *string       `xml:"Value,attr"`
	Annotation []*Annotation `xml:"Annotation"`
}

func (m Member) GetChildren() []Searchable {
	ret := make([]Searchable, 0, 100)
	ret = utilAppendTypeAsSearchable(ret, m.Annotation...)
	return ret
}

func (m Member) GetName() string {
	return m.Name
}

func (m Member) GetXPath() string {
	return "Member[@Name='" + m.Name + "']"
}

type NavigationProperty struct {
	Name           string        `xml:"Name,attr"`
	Type           string        `xml:"Type,attr"`
	Nullable       *bool         `xml:"Nullable,attr"`
	Partner        string        `xml:"Partner,attr"`
	ContainsTarget bool          `xml:"ContainsTarget,attr"`
	Annotation     []*Annotation `xml:"Annotation"`
}

func (n NavigationProperty) GetChildren() []Searchable {
	ret := make([]Searchable, 0, 100)
	ret = utilAppendTypeAsSearchable(ret, n.Annotation...)
	return ret
}

func (n NavigationProperty) GetName() string {
	return n.Name
}

func (n NavigationProperty) GetXPath() string {
	return "NavigationProperty[@Name='" + n.Name + "']"
}

type Parameter struct {
	Name       string        `xml:"Name,attr"`
	Type       string        `xml:"Type,attr"`
	Annotation []*Annotation `xml:"Annotation"`
	Other      []any         `xml:",any"`
}

func (p Parameter) GetChildren() []Searchable {
	ret := make([]Searchable, 0, 100)
	ret = utilAppendTypeAsSearchable(ret, p.Annotation...)
	return ret
}

func (p Parameter) GetName() string {
	return p.Name
}

func (p Parameter) GetXPath() string {
	return "Parameter[@Name='" + p.Name + "']"
}

type Property struct {
	Name         string        `xml:"Name,attr"`
	Type         string        `xml:"Type,attr"`
	Nullable     *bool         `xml:"Nullable,attr"`
	MaxLength    int           `xml:"MaxLength,attr"`
	Unicode      bool          `xml:"Unicode,attr"`
	Precision    int           `xml:"Precision,attr"`
	Scale        int           `xml:"Scale,attr"`
	SRID         string        `xml:"SRID,attr"`
	DefaultValue string        `xml:"DefaultValue,attr"`
	Annotation   []*Annotation `xml:"Annotation"`
}

func (p Property) GetChildren() []Searchable {
	ret := make([]Searchable, 0, len(p.Annotation))
	ret = utilAppendTypeAsSearchable(ret, p.Annotation...)
	return ret
}

func (p Property) GetName() string {
	return p.Name
}

func (p Property) GetXPath() string {
	return "Property[@Name='" + p.Name + "']"
}

type Term struct {
	Name       string        `xml:"Name,attr"`
	Type       string        `xml:"Type,attr"`
	Annotation []*Annotation `xml:"Annotation"`
	Other      []any         `xml:",any"`
}

func (t Term) GetChildren() []Searchable {
	ret := make([]Searchable, 0, 100)
	ret = utilAppendTypeAsSearchable(ret, t.Annotation...)
	return ret
}

func (t Term) GetName() string {
	return t.Name
}

func (t Term) GetXPath() string {
	return "Term[@Name='" + t.Name + "']"
}

type TypeDefinition struct {
	Name           string        `xml:"Name,attr"`
	UnderlyingType string        `xml:"UnderlyingType,attr"`
	Annotation     []*Annotation `xml:"Annotation"`
}

func (t TypeDefinition) GetChildren() []Searchable {
	ret := make([]Searchable, 0, 100)
	ret = utilAppendTypeAsSearchable(ret, t.Annotation...)
	return ret
}

func (t TypeDefinition) GetName() string {
	return t.Name
}

func (t TypeDefinition) GetXPath() string {
	return t.Name
}

type Record struct {
	PropertyValue []*PropertyValue `xml:"PropertyValue"`
	Other         []any            `xml:",any"`
}

type PropertyValue struct {
	Property string `xml:"Property,attr"`
	Bool     bool   `xml:"Bool,attr"`
	Other    []any  `xml:",any"`
}
