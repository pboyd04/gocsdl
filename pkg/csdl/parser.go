package csdl

import (
	"cmp"
	"encoding/xml"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
)

type Parser struct {
	IgnoreCollections bool
	Files             map[string]io.ReadCloser
	Replacements      map[string]string
}

func NewParser() *Parser {
	return &Parser{
		Files:        make(map[string]io.ReadCloser),
		Replacements: make(map[string]string),
	}
}

func (p *Parser) Close() error {
	for _, file := range p.Files {
		_ = file.Close()
	}
	return nil
}

func (p *Parser) AddFile(name string, file io.ReadCloser) {
	name = strings.TrimSuffix(name, ".xml")
	p.Files[name] = file
}

func (p *Parser) Parse() (map[string]*Type, error) {
	types := map[string]*Type{}
	for _, reader := range p.Files {
		dec := xml.NewDecoder(reader)
		edmx := Edmx{}
		err := dec.Decode(&edmx)
		if err != nil {
			return nil, err
		}
		for _, schema := range edmx.DataServices.Schema {
			for _, entityType := range schema.EntityType {
				types[schema.Namespace+"."+entityType.Name] = NewTypeFromEntityType(entityType, schema.Namespace)
			}
			for _, enumType := range schema.EnumType {
				types[schema.Namespace+"."+enumType.Name] = NewTypeFromEnumType(enumType, schema.Namespace)
			}
			for _, complexType := range schema.ComplexType {
				types[schema.Namespace+"."+complexType.Name] = NewTypeFromComplexType(complexType, schema.Namespace)
			}
			for _, typeDefinition := range schema.TypeDefinition {
				p.Replacements[schema.Namespace+"."+typeDefinition.Name] = typeDefinition.UnderlyingType
			}
		}
	}
	return types, nil
}

// Fold will consolidate types, adding properties from base types to derived types
func (p *Parser) Fold(types map[string]*Type) {
	keys := maps.Keys(types)
	sortedKeys := slices.SortedFunc(keys, sortNamespace)
	slices.Reverse(sortedKeys)
	for _, name := range sortedKeys {
		t, ok := types[name]
		if !ok {
			// we removed this type in the fold
			continue
		}
		types[name] = t.Fold(types, p.Replacements)
		if types[name] == nil {
			delete(types, name)
		}
	}
}

// need to do this because simple string sort doesn't work...
func sortNamespace(a, b string) int {
	a1, a2, a3 := splitNamespace(a)
	b1, b2, b3 := splitNamespace(b)
	if a1 != b1 {
		return strings.Compare(a1, b1)
	}
	if a3 == "" {
		return strings.Compare(a2, b2)
	}
	// The second component is where we have issues...
	aMajor, aMinor, aRev := splitVersion(a2)
	bMajor, bMinor, bRev := splitVersion(b2)
	if aMajor != bMajor {
		return cmp.Compare(aMajor, bMajor)
	}
	if aMinor != bMinor {
		return cmp.Compare(aMinor, bMinor)
	}
	if aRev != bRev {
		return cmp.Compare(aRev, bRev)
	}
	return strings.Compare(a3, b3)
}

func splitNamespace(namespace string) (string, string, string) {
	index := strings.Index(namespace, ".")
	if index == -1 {
		return namespace, "", ""
	}
	comp1 := namespace[:index]
	comp2 := namespace[index+1:]
	index = strings.Index(comp2, ".")
	if index == -1 {
		return comp1, comp2, ""
	}
	return comp1, comp2[:index], comp2[index+1:]
}

func splitVersion(version string) (int, int, int) {
	version = strings.TrimPrefix(version, "v")
	index := strings.Index(version, "_")
	if index == -1 {
		intVer, err := strconv.Atoi(version)
		if err != nil {
			return 0, 0, 0
		}
		return intVer, 0, 0
	}
	comp1 := version[:index]
	major, err := strconv.Atoi(comp1)
	if err != nil {
		return 0, 0, 0
	}
	comp2 := version[index+1:]
	index = strings.Index(comp2, "_")
	if index == -1 {
		minor, err := strconv.Atoi(comp2)
		if err != nil {
			return major, 0, 0
		}
		return major, minor, 0
	}
	minor, err := strconv.Atoi(comp2[:index])
	if err != nil {
		return major, 0, 0
	}
	rev, err := strconv.Atoi(comp2[index+1:])
	if err != nil {
		return major, minor, 0
	}
	return major, minor, rev
}
