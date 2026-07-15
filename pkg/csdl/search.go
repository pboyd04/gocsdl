package csdl

import (
	"reflect"
	"strings"
	"sync"
)

type Searcher struct {
	parser          *Parser
	schemaCache     sync.Map // map[string]*Schema
	schemaCount     int
	schemaCacheOnce sync.Once
}

type Searchable interface {
	GetChildren() []Searchable
	GetName() string
	GetXPath() string
}

type SchmeaChild interface {
	GetSchema() *Schema
}

func NewSearcher(parser *Parser) *Searcher {
	return &Searcher{
		parser: parser,
	}
}

// initSchemaCache populates the schema cache exactly once, thread-safely.
func (s *Searcher) initSchemaCache() {
	s.schemaCacheOnce.Do(func() {
		count := 0
		for _, parsed := range s.parser.Parsed {
			for _, schema := range parsed.Edmx.DataServices.Schema {
				s.schemaCache.Store(schema.Namespace, schema)
				count++
			}
		}
		s.schemaCount = count
	})
}

func (s *Searcher) GetFiles() map[string]*Edmx {
	result := make(map[string]*Edmx)
	for key, value := range s.parser.Parsed {
		if !value.IsReference {
			result[key] = value.Edmx
		}
	}
	return result
}

func (s *Searcher) GetParser() *Parser {
	return s.parser
}

func (s *Searcher) GetAnnotations(name string) []*Annotation {
	ret := make([]*Annotation, 0, 100)
	for _, parsed := range s.parser.Parsed {
		if parsed.IsReference {
			continue
		}
		edmx := parsed.Edmx
		for _, schema := range edmx.DataServices.Schema {
			children := schema.GetChildren()
			ret = s.getAnnotationsFromChildren(children, name, ret)
		}
	}
	return ret
}

func (s *Searcher) GetAllSchemas() []*Schema {
	s.initSchemaCache()
	ret := make([]*Schema, 0, s.schemaCount)
	for _, parsed := range s.parser.Parsed {
		if parsed.IsReference {
			continue
		}
		edmx := parsed.Edmx
		ret = append(ret, edmx.DataServices.Schema...)
	}
	return ret
}

func (s *Searcher) GetSchemas(name string) []*Schema {
	if name == "" {
		return s.GetAllSchemas()
	}
	s.initSchemaCache()
	ret := make([]*Schema, 0, s.schemaCount)
	for _, parsed := range s.parser.Parsed {
		if parsed.IsReference {
			continue
		}
		edmx := parsed.Edmx
		for _, schema := range edmx.DataServices.Schema {
			if schema.Namespace == name {
				ret = append(ret, schema)
			}
		}
	}
	return ret
}

func (s *Searcher) GetEntityTypes(name string) []*EntityType {
	ret := make([]*EntityType, 0, 100)
	for _, parsed := range s.parser.Parsed {
		if parsed.IsReference && name == "" {
			continue
		}
		edmx := parsed.Edmx
		for _, schema := range edmx.DataServices.Schema {
			children := schema.GetChildren()
			for _, child := range children {
				switch t := child.(type) {
				case *EntityType:
					if name == "" || t.Name == name || name == "*" {
						ret = append(ret, t)
					}
				}
			}
		}
	}
	return ret
}

func (s *Searcher) GetComplexTypes(name string) []*ComplexType {
	ret := make([]*ComplexType, 0, 100)
	for _, parsed := range s.parser.Parsed {
		if parsed.IsReference && name == "" {
			continue
		}
		edmx := parsed.Edmx
		for _, schema := range edmx.DataServices.Schema {
			children := schema.GetChildren()
			for _, child := range children {
				switch t := child.(type) {
				case *ComplexType:
					if name == "" || t.Name == name || name == "*" {
						ret = append(ret, t)
					}
				}
			}
		}
	}
	return ret
}

func (s *Searcher) GetEnumTypes(name string) []*EnumType {
	ret := make([]*EnumType, 0, 100)
	for _, parsed := range s.parser.Parsed {
		if parsed.IsReference {
			continue
		}
		edmx := parsed.Edmx
		for _, schema := range edmx.DataServices.Schema {
			children := schema.GetChildren()
			for _, child := range children {
				switch t := child.(type) {
				case *EnumType:
					if name == "" || t.Name == name {
						ret = append(ret, t)
					}
				}
			}
		}
	}
	return ret
}

func (s *Searcher) GetProperties(name string) []*Property {
	ret := make([]*Property, 0, 100)
	for _, parsed := range s.parser.Parsed {
		if parsed.IsReference {
			continue
		}
		edmx := parsed.Edmx
		for _, schema := range edmx.DataServices.Schema {
			children := schema.GetChildren()
			for _, child := range children {
				switch t := child.(type) {
				case *Property:
					if name == "" || t.Name == name {
						ret = append(ret, t)
					}
				default:
					ret = append(ret, s.GetPropertiesUnder(child)...)
				}
			}
		}
	}
	return ret
}

func (s *Searcher) GetNavigationProperties(name string) []*NavigationProperty {
	ret := make([]*NavigationProperty, 0, 100)
	for _, parsed := range s.parser.Parsed {
		if parsed.IsReference {
			continue
		}
		edmx := parsed.Edmx
		for _, schema := range edmx.DataServices.Schema {
			children := schema.GetChildren()
			for _, child := range children {
				switch t := child.(type) {
				case *NavigationProperty:
					if name == "" || t.Name == name {
						ret = append(ret, t)
					}
				default:
					ret = append(ret, s.GetNavigationPropertiesUnder(child)...)
				}
			}
		}
	}
	return ret
}

func (s *Searcher) GetActions(name string) []*Action {
	ret := make([]*Action, 0, 100)
	for _, parsed := range s.parser.Parsed {
		if parsed.IsReference {
			continue
		}
		edmx := parsed.Edmx
		for _, schema := range edmx.DataServices.Schema {
			children := schema.GetChildren()
			for _, child := range children {
				switch t := child.(type) {
				case *Action:
					if name == "" || t.Name == name {
						ret = append(ret, t)
					}
				default:
					ret = append(ret, s.GetActionsUnder(child)...)
				}
			}
		}
	}
	return ret
}

func (s *Searcher) GetAnnotaionsUnder(haystack Searchable) []*Annotation {
	ret := make([]*Annotation, 0, 100)
	children := haystack.GetChildren()
	for _, child := range children {
		switch t := child.(type) {
		case *Annotation:
			ret = append(ret, t)
		default:
			ret = append(ret, s.GetAnnotaionsUnder(child)...)
		}
	}
	return ret
}

func (s *Searcher) GetPropertiesUnder(haystack Searchable) []*Property {
	ret := make([]*Property, 0, 100)
	children := haystack.GetChildren()
	for _, child := range children {
		switch t := child.(type) {
		case *Property:
			ret = append(ret, t)
		default:
			ret = append(ret, s.GetPropertiesUnder(child)...)
		}
	}
	return ret
}

func (s *Searcher) GetNavigationPropertiesUnder(haystack Searchable) []*NavigationProperty {
	ret := make([]*NavigationProperty, 0, 100)
	children := haystack.GetChildren()
	for _, child := range children {
		switch t := child.(type) {
		case *NavigationProperty:
			ret = append(ret, t)
		default:
			ret = append(ret, s.GetNavigationPropertiesUnder(child)...)
		}
	}
	return ret
}

func (s *Searcher) GetActionsUnder(haystack Searchable) []*Action {
	ret := make([]*Action, 0, 100)
	children := haystack.GetChildren()
	for _, child := range children {
		switch t := child.(type) {
		case *Action:
			ret = append(ret, t)
		default:
			ret = append(ret, s.GetActionsUnder(child)...)
		}
	}
	return ret
}

func (s *Searcher) GetPath(searchable Searchable) (string, string) {
	for fileName, parsed := range s.parser.Parsed {
		for _, schema := range parsed.Edmx.DataServices.Schema {
			switch t := searchable.(type) {
			case *Schema:
				if reflect.DeepEqual(t, schema) {
					return fileName + ".xml", "/edmx:Edmx/edmx:DataServices/Schema[@Namespace='" + schema.Namespace + "']"
				}
			default:
				schemaName := schema.Namespace
				children := schema.GetChildren()
				for _, child := range children {
					tmp := s.getPathFromChild(child, searchable)
					if tmp != "" {
						return fileName + ".xml", "/edmx:Edmx/edmx:DataServices/Schema[@Namespace='" + schemaName + "']/" + tmp
					}
				}
			}
		}
	}
	return "", ""
}

func (s *Searcher) GetSchemaByNamespace(namespace string) *Schema {
	s.initSchemaCache()
	value, ok := s.schemaCache.Load(namespace)
	if !ok {
		return nil
	}
	return value.(*Schema)
}

func (s *Searcher) GetByTypeName(typeName string) Searchable {
	if strings.HasPrefix(typeName, "Redfish.") {
		typeName = strings.TrimPrefix(typeName, "Redfish.")
		typeName = "RedfishExtensions.v1_0_0." + typeName
	}
	if strings.HasPrefix(typeName, "Validation.") {
		typeName = strings.TrimPrefix(typeName, "Validation.")
		typeName = "Validation.v1_0_0." + typeName
	}
	if strings.HasPrefix(typeName, "Collection(") {
		typeName = strings.TrimPrefix(typeName, "Collection(")
		typeName = strings.TrimSuffix(typeName, ")")
	}
	index := strings.LastIndex(typeName, ".")
	typeNamespace := ""
	typeLocalName := typeName
	if index != -1 {
		typeNamespace = typeName[:index]
		typeLocalName = typeName[index+1:]
	}
	if typeNamespace == "Edm" {
		return nil
	}
	schema := s.GetSchemaByNamespace(typeNamespace)
	if schema == nil {
		return nil
	}
	ret := s.getByTypeNameFromSchema(schema, typeLocalName)
	if ret != nil {
		return ret
	}
	// Look in the parent schema
	if strings.Contains(typeName, "v1_0_0.") {
		return s.GetByTypeName(strings.ReplaceAll(typeName, "v1_0_0.", ""))
	}
	return nil
}

func (s *Searcher) getByTypeNameFromSchema(schema *Schema, typeName string) Searchable {
	children := schema.GetChildren()
	for _, child := range children {
		if child.GetName() == typeName {
			return child
		}
	}
	return nil
}

func (s *Searcher) getAnnotationsFromChildren(children []Searchable, name string, ret []*Annotation) []*Annotation {
	if len(children) == 0 {
		return ret
	}
	for _, child := range children {
		switch t := child.(type) {
		case *Annotation:
			if t.Term == name || name == "" {
				ret = append(ret, t)
			}
		default:
			ret = s.getAnnotationsFromChildren(child.GetChildren(), name, ret)
		}
	}
	return ret
}

func (s *Searcher) getPathFromChild(haystack, needle Searchable) string {
	if haystack == needle {
		return haystack.GetXPath()
	}
	children := haystack.GetChildren()
	for _, child := range children {
		if child == needle {
			return haystack.GetXPath() + "/" + child.GetXPath()
		}
		ret := s.getPathFromChild(child, needle)
		if ret != "" {
			return haystack.GetXPath() + "/" + ret
		}
	}
	return ""
}

func utilAppendTypeAsSearchable[T Searchable](ret []Searchable, types ...T) []Searchable {
	for _, t := range types {
		ret = append(ret, t)
	}
	return ret
}
