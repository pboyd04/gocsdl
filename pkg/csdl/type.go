package csdl

import (
	"fmt"
	"go/ast"
	"go/token"
	"maps"
	"slices"
	"strconv"
	"strings"
)

type Type struct {
	Name         string
	Namespace    string
	BaseType     string
	Properties   map[string]PropType
	Members      map[string]MemberType
	Wildcard     bool              // true if this is a wildcard type, like Attributes
	Replacements map[string]string // used to replace types with their underlying type
	ComplexType  bool
}

func NewTypeFromEntityType(entityType EntityType, nameSpace string) *Type {
	myType := &Type{
		Name:        entityType.Name,
		Namespace:   nameSpace,
		Properties:  make(map[string]PropType),
		BaseType:    entityType.BaseType,
		ComplexType: false,
	}
	for _, property := range entityType.Property {
		canBeNull := true
		if property.Nullable != nil {
			canBeNull = *property.Nullable
		}
		propType := PropType{
			Navigation: false,
			Type:       property.Type,
			CanBeNull:  canBeNull,
		}
		myType.Properties[property.Name] = propType
	}
	for _, navProp := range entityType.NavigationProperty {
		canBeNull := true
		if navProp.Nullable != nil {
			canBeNull = *navProp.Nullable
		}
		propType := PropType{
			Navigation: true,
			Type:       navProp.Type,
			CanBeNull:  canBeNull,
		}
		myType.Properties[navProp.Name] = propType
	}
	return myType
}

func NewTypeFromComplexType(complexType ComplexType, nameSpace string) *Type {
	myType := &Type{
		Name:        complexType.Name,
		Namespace:   nameSpace,
		Properties:  make(map[string]PropType),
		BaseType:    complexType.BaseType,
		ComplexType: true,
	}
	for _, property := range complexType.Property {
		canBeNull := true
		if property.Nullable != nil {
			canBeNull = *property.Nullable
		}
		propType := PropType{
			Navigation: false,
			Type:       property.Type,
			CanBeNull:  canBeNull,
		}
		myType.Properties[property.Name] = propType
	}
	for _, navProp := range complexType.NavigationProperty {
		canBeNull := true
		if navProp.Nullable != nil {
			canBeNull = *navProp.Nullable
		}
		propType := PropType{
			Navigation: true,
			Type:       navProp.Type,
			CanBeNull:  canBeNull,
		}
		myType.Properties[navProp.Name] = propType
	}
	if len(myType.Properties) == 0 {
		// Search for annotations that inidicate this is a wildcard type
		for _, annotation := range complexType.Annotation {
			if annotation.Term == "Redfish.DynamicPropertyPatterns" {
				myType.Wildcard = true
				break
			}
			if annotation.Term == "OData.AdditionalProperties" && annotation.Bool {
				myType.Wildcard = true
				break
			}
		}
	}
	return myType
}

func NewTypeFromEnumType(enumType EnumType, nameSpace string) *Type {
	myType := &Type{
		Name:      enumType.Name,
		Namespace: nameSpace,
		Members:   make(map[string]MemberType),
	}
	for _, member := range enumType.Member {
		memType := MemberType{
			Name: member.Name,
		}
		if member.Value != nil {
			memType.Value = *member.Value
		}
		myType.Members[member.Name] = memType
	}
	return myType
}

func handleUnknownType(t *Type, types map[string]*Type) *Type {
	switch t.BaseType {
	case "Resource.v1_0_0.Resource":
		// This can happen if we are doing a single file...
		// Just handle it
		t.Properties["ID"] = PropType{Type: "Edm.String", CanBeNull: false, Navigation: false, JsonName: "@odata.id"}
		t.Properties["Type"] = PropType{Type: "Edm.String", CanBeNull: false, Navigation: false, JsonName: "@odata.type"}
		t.Properties["Name"] = PropType{Type: "Edm.String", CanBeNull: false, Navigation: false}
		t.Properties["Description"] = PropType{Type: "Edm.String", CanBeNull: true, Navigation: false}
		return t
	case "Resource.Links":
		t.Properties["Oem"] = PropType{Type: "Resource.Oem", CanBeNull: true, Navigation: false}
		return t
	case "":
		// No base type, we're done
		return t
	default:
		// dead end... this can happen with trivial namespaces that we don't need
		delete(types, t.Namespace+"."+t.Name)
		return nil
	}
}

func (t *Type) Fold(types map[string]*Type, replacements map[string]string) *Type {
	baseType, ok := types[t.BaseType]
	if !ok {
		replacement, ok := replacements[t.BaseType]
		if ok {
			// This is a type that has been replaced, so we need to find the new type
			baseType, ok = types[replacement]
			if !ok {
				return handleUnknownType(t, types)
			}
		} else {
			return handleUnknownType(t, types)
		}
	}
	for name, prop := range baseType.Properties {
		if _, ok := t.Properties[name]; ok {
			continue
		}
		t.Properties[name] = prop
	}
	t.BaseType = baseType.BaseType
	if t.BaseType != "" {
		t = t.Fold(types, replacements)
	}
	if t != nil {
		t.Replacements = replacements
	}
	return t
}

func (t *Type) Node(types map[string]*Type) []ast.Node {
	if len(t.Properties) != 0 {
		// This is a struct
		return t.structNode(types)
	}
	if len(t.Members) != 0 {
		// This is an enum
		return t.enumNode(types)
	}
	switch {
	case strings.HasSuffix(t.Name, "OemActions") || t.Name == "ItemOrCollection" || t.Wildcard:
		// Just skip this type, it's an empty complex type
		return []ast.Node{}
	}
	if !strings.Contains(t.Namespace, ".") {
		// We should check for other versions of this type...
		for name, typeData := range types {
			if strings.HasPrefix(name, t.Namespace) && t.Name == typeData.Name && len(typeData.Properties) > 0 {
				return typeData.structNode(types)
			}
		}
	}
	panic("unknown type: " + fmt.Sprintf("%#v", t))
}

func (t *Type) structNode(types map[string]*Type) []ast.Node {
	structType := &ast.StructType{
		Fields: &ast.FieldList{
			List: make([]*ast.Field, 0, len(t.Properties)),
		},
	}
	// do these in order...
	if !t.ComplexType {
		// Redfish uses ComplexType for types that are not individually addressable,
		// so skip the ID fields in these types
		structType.Fields.List = append(structType.Fields.List, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent("ID")},
			Type:  &ast.Ident{Name: "string"},
			Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"@odata.id\"`"},
		})
		idProp, ok := t.Properties["Id"]
		if ok {
			field := idProp.ToField("Id", types, t.Replacements)
			structType.Fields.List = append(structType.Fields.List, field)
			delete(t.Properties, "Id")
		}
	}
	structType.Fields.List = append(structType.Fields.List, &ast.Field{
		Names: []*ast.Ident{ast.NewIdent("Type")},
		Type:  &ast.Ident{Name: "string"},
		Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"@odata.type,omitempty\"`"},
	},
		&ast.Field{
			Names: []*ast.Ident{ast.NewIdent("Context")},
			Type:  &ast.Ident{Name: "string"},
			Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"@odata.context,omitempty\"`"},
		})
	nameProp, ok := t.Properties["Name"]
	if ok {
		field := nameProp.ToField("Name", types, t.Replacements)
		structType.Fields.List = append(structType.Fields.List, field)
		delete(t.Properties, "Name")
	}
	descriptionProp, ok := t.Properties["Description"]
	if ok {
		field := descriptionProp.ToField("Description", types, t.Replacements)
		field.Tag = &ast.BasicLit{
			Kind:  token.STRING,
			Value: "`json:\",omitempty\"`",
		}
		structType.Fields.List = append(structType.Fields.List, field)
		delete(t.Properties, "Description")
	}
	fieldNames := slices.Sorted(maps.Keys(t.Properties))
	for _, name := range fieldNames {
		prop := t.Properties[name]
		field := prop.ToField(name, types, t.Replacements)
		structType.Fields.List = append(structType.Fields.List, field)
	}
	ret := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(t.GoTypeName()),
				Type: structType,
			},
		},
	}
	return []ast.Node{ret}
}

func (t *Type) underLyingEnumType() string {
	// TODO handle integer backed enums
	return "string"
}

func (t *Type) enumNode(map[string]*Type) []ast.Node {
	ret := []ast.Node{
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: ast.NewIdent(t.GoTypeName()),
					Type: &ast.Ident{
						Name: t.underLyingEnumType(),
					},
				},
			},
		},
	}
	constNode := &ast.GenDecl{
		Tok:   token.CONST,
		Specs: make([]ast.Spec, 0, len(t.Members)),
	}
	for _, member := range t.Members {
		valueSpec := &ast.ValueSpec{
			Names: []*ast.Ident{ast.NewIdent(t.GoTypeName() + "_" + member.Name)},
			Type:  &ast.Ident{Name: t.GoTypeName()},
		}
		if member.Value != "" {
			valueSpec.Values = []ast.Expr{
				ast.NewIdent(member.Value),
			}
		} else {
			valueSpec.Values = []ast.Expr{
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote(member.Name),
				},
			}
		}
		constNode.Specs = append(constNode.Specs, valueSpec)
	}
	return append(ret, constNode)
}

func (t *Type) GoTypeName() string {
	if strings.HasPrefix(t.Namespace, t.Name) {
		return t.Name
	}
	nameSpace := t.Namespace
	index := strings.Index(nameSpace, ".")
	if index != -1 {
		nameSpace = nameSpace[:index]
	}
	return strings.ReplaceAll(nameSpace+"_"+t.Name, ".", "_")
}

type PropType struct {
	Navigation bool
	Type       string
	CanBeNull  bool
	JsonName   string
}

func (p *PropType) ToField(name string, types map[string]*Type, replacements map[string]string) *ast.Field {
	switch name {
	case "Actions":
		return &ast.Field{
			Names: []*ast.Ident{ast.NewIdent("Actions")},
			Type:  &ast.Ident{Name: "map[string]Action"},
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "`json:\",omitempty\"`",
			},
		}
	case "Oem":
		return &ast.Field{
			Names: []*ast.Ident{ast.NewIdent("Oem")},
			Type:  &ast.Ident{Name: "map[string]any"},
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "`json:\",omitempty\"`",
			},
		}
	}
	rep, ok := replacements[p.Type]
	if ok {
		p.Type = rep
	}
	field := &ast.Field{
		Names: []*ast.Ident{ast.NewIdent(name)},
		Type:  p.Node(types).(ast.Expr),
	}
	if p.JsonName != "" {
		field.Tag = &ast.BasicLit{
			Kind:  token.STRING,
			Value: "`json:\"" + p.JsonName + "\"`",
		}
	}
	if p.Navigation || p.CanBeNull {
		field.Tag = &ast.BasicLit{
			Kind:  token.STRING,
			Value: "`json:\",omitempty\"`",
		}
	}
	ident, ok := field.Type.(*ast.Ident)
	if ok && ident.Name == "any" {
		field.Tag = &ast.BasicLit{
			Kind:  token.STRING,
			Value: "`json:\",omitempty\"`",
		}
	}
	return field
}

func (p *PropType) Node(types map[string]*Type) ast.Node {
	if p.Navigation {
		if p.CanBeNull {
			return &ast.Ident{Name: "*OdataID"}
		}
		return &ast.Ident{Name: "OdataID"}
	}
	typeName := p.Type
	prefix := ""
	switch {
	case strings.HasPrefix(typeName, "Collection("):
		prefix = "[]"
		typeName = strings.TrimPrefix(typeName, "Collection(")
		typeName = strings.TrimSuffix(typeName, ")")
	case p.CanBeNull:
		prefix = "*"
	}
	if strings.HasSuffix(typeName, "OemActions") {
		return &ast.Ident{Name: "map[string]any"}
	}
	switch typeName {
	case "Edm.Boolean":
		return &ast.Ident{Name: prefix + "bool"}
	case "Edm.Byte":
		return &ast.Ident{Name: prefix + "byte"}
	case "Edm.Date":
		return &ast.Ident{Name: prefix + "Date"}
	case "Edm.DateTimeOffset":
		return &ast.Ident{Name: prefix + "DateTimeOffset"}
	case "Edm.Decimal":
		return &ast.Ident{Name: prefix + "float64"}
	case "Edm.Double":
		return &ast.Ident{Name: prefix + "float64"}
	case "Edm.Duration":
		return &ast.Ident{Name: prefix + "Duration"}
	case "Edm.Int16":
		return &ast.Ident{Name: prefix + "int16"}
	case "Edm.Int32":
		return &ast.Ident{Name: prefix + "int32"}
	case "Edm.Int64":
		return &ast.Ident{Name: prefix + "int64"}
	case "Edm.PrimitiveType":
		return &ast.Ident{Name: prefix + "any"}
	case "Edm.SByte":
		return &ast.Ident{Name: prefix + "int8"}
	case "Edm.Single":
		return &ast.Ident{Name: prefix + "float32"}
	case "Edm.String":
		return &ast.Ident{Name: prefix + "string"}
	case "Edm.Guid":
		fallthrough
	case "Resource.UUID":
		return &ast.Ident{Name: prefix + "UUID"}
	case "Resource.Oem":
		return &ast.Ident{Name: "map[string]any"}
	case "Resource.Description":
		return &ast.Ident{Name: prefix + "string"}
	case "Resource.Name":
		return &ast.Ident{Name: prefix + "string"}
	case "Resource.Status":
		return &ast.Ident{Name: "Resource_Status"}
	case "Resource.PowerState":
		return &ast.Ident{Name: "Resource_PowerState"}
	default:
		typeData, ok := doTypeSearch(typeName, types)
		if !ok {
			fmt.Printf("Unknown type: %s\n", typeName)
			return &ast.Ident{Name: prefix + "any"}
		}
		return &ast.Ident{Name: prefix + typeData.GoTypeName()}
	}
}

func doTypeSearch(typeName string, types map[string]*Type) (*Type, bool) {
	typeData, ok := types[typeName]
	if ok {
		// Easiest case!
		return typeData, true
	}
	prefix, _, name := splitNamespace(typeName)
	if name == "" {
		prefix, name, _ = splitNamespace(typeName)
	}
	// This may have been folded into a a newer version, go and find it...
	for mapName, t := range types {
		myPrefix, _, myName := splitNamespace(mapName)
		if myName == "" {
			myPrefix, myName, _ = splitNamespace(mapName)
		}
		if myPrefix == prefix && myName == name {
			return t, true
		}
	}
	return nil, false
}

type MemberType struct {
	Name  string
	Value string
}
