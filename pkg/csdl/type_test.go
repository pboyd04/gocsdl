package csdl_test

import (
	"go/ast"
	"testing"

	"github.com/pboyd04/gocsdl/pkg/csdl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTypeFromEntityType(t *testing.T) {
	tests := []struct {
		name        string
		entityType  *csdl.EntityType
		namespace   string
		wantName    string
		wantProps   int
		wantExcerpt map[string]bool
		wantTypeAnn int
		wantPropAnn map[string]int
	}{
		{
			name: "basic entity type",
			entityType: &csdl.EntityType{
				Name:     "TestEntity",
				BaseType: "Resource.v1_0_0.Resource",
				Annotation: []*csdl.Annotation{
					{Term: "Redfish.TestTypeAnnotation"},
				},
				Property: []*csdl.Property{
					{Name: "Prop1", Type: "Edm.String", Annotation: []*csdl.Annotation{{Term: "Validation.Pattern"}}},
					{Name: "Prop2", Type: "Edm.Int32"},
				},
			},
			namespace:   "Test.v1_0_0",
			wantName:    "TestEntity",
			wantProps:   2,
			wantTypeAnn: 1,
			wantPropAnn: map[string]int{
				"Prop1": 1,
				"Prop2": 0,
			},
		},
		{
			name: "entity with nullable property",
			entityType: &csdl.EntityType{
				Name: "NullableEntity",
				Property: []*csdl.Property{
					{Name: "Prop1", Type: "Edm.String", Nullable: boolPtr(true)},
					{Name: "Prop2", Type: "Edm.String", Nullable: boolPtr(false)},
				},
			},
			namespace: "Test.v1_0_0",
			wantName:  "NullableEntity",
			wantProps: 2,
		},
		{
			name: "entity with navigation properties",
			entityType: &csdl.EntityType{
				Name: "NavEntity",
				NavigationProperty: []*csdl.NavigationProperty{
					{Name: "Nav1", Type: "Test.Related", Annotation: []*csdl.Annotation{{Term: "Redfish.ExcerptCopy"}}},
					{Name: "Nav2", Type: "Test.Other", Nullable: boolPtr(false)},
				},
			},
			namespace: "Test.v1_0_0",
			wantName:  "NavEntity",
			wantProps: 2,
			wantExcerpt: map[string]bool{
				"Nav1": true,
				"Nav2": false,
			},
			wantPropAnn: map[string]int{
				"Nav1": 1,
				"Nav2": 0,
			},
		},
		{
			name: "entity with mixed properties",
			entityType: &csdl.EntityType{
				Name: "MixedEntity",
				Property: []*csdl.Property{
					{Name: "Prop1", Type: "Edm.String"},
				},
				NavigationProperty: []*csdl.NavigationProperty{
					{Name: "Nav1", Type: "Test.Related", Annotation: []*csdl.Annotation{{Term: "Redfish.ExcerptCopy"}}},
				},
			},
			namespace: "Test.v1_0_0",
			wantName:  "MixedEntity",
			wantProps: 2,
			wantExcerpt: map[string]bool{
				"Nav1": true,
			},
			wantPropAnn: map[string]int{
				"Nav1":  1,
				"Prop1": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csdl.NewTypeFromEntityType(tt.entityType, tt.namespace)
			assert.Equal(t, tt.wantName, result.Name)
			assert.Equal(t, tt.namespace, result.Namespace)
			assert.Equal(t, tt.wantProps, len(result.Properties))
			assert.Len(t, result.Annotations, tt.wantTypeAnn)
			assert.False(t, result.ComplexType)
			for propName, wantExcerpt := range tt.wantExcerpt {
				prop, ok := result.Properties[propName]
				if !ok {
					t.Errorf("missing property %q", propName)
					continue
				}
				if prop.Excerpt != wantExcerpt {
					t.Errorf("property %q Excerpt = %v, want %v", propName, prop.Excerpt, wantExcerpt)
				}
				assert.Len(t, prop.Annotations, tt.wantPropAnn[propName])
			}
			for propName, wantAnn := range tt.wantPropAnn {
				prop, ok := result.Properties[propName]
				if !ok {
					t.Errorf("missing property %q", propName)
					continue
				}
				assert.Len(t, prop.Annotations, wantAnn)
			}
		})
	}
}

func TestNewTypeFromComplexType(t *testing.T) {
	tests := []struct {
		name         string
		complexType  *csdl.ComplexType
		namespace    string
		wantName     string
		wantProps    int
		wantWildcard bool
		wantExcerpt  map[string]bool
		wantNull     map[string]bool
		wantTypeAnn  int
		wantPropAnn  map[string]int
	}{
		{
			name: "basic complex type",
			complexType: &csdl.ComplexType{
				Name:     "TestComplex",
				BaseType: "Resource.Links",
				Annotation: []*csdl.Annotation{
					{Term: "Redfish.TestTypeAnnotation"},
				},
				Property: []*csdl.Property{
					{Name: "Prop1", Type: "Edm.String", Annotation: []*csdl.Annotation{{Term: "Validation.Pattern"}}},
				},
			},
			namespace:   "Test.v1_0_0",
			wantName:    "TestComplex",
			wantProps:   1,
			wantTypeAnn: 1,
			wantPropAnn: map[string]int{
				"Prop1": 1,
			},
			wantNull: map[string]bool{
				"Prop1": true,
			},
		},
		{
			name: "wildcard with Redfish.DynamicPropertyPatterns",
			complexType: &csdl.ComplexType{
				Name: "WildcardComplex",
				Annotation: []*csdl.Annotation{
					{Term: "Redfish.DynamicPropertyPatterns"},
				},
			},
			namespace:    "Test.v1_0_0",
			wantName:     "WildcardComplex",
			wantProps:    0,
			wantWildcard: true,
			wantTypeAnn:  1,
		},
		{
			name: "wildcard with OData.AdditionalProperties",
			complexType: &csdl.ComplexType{
				Name: "ODataWildcard",
				Annotation: []*csdl.Annotation{
					{Term: "OData.AdditionalProperties", Bool: true},
				},
			},
			namespace:    "Test.v1_0_0",
			wantName:     "ODataWildcard",
			wantProps:    0,
			wantWildcard: true,
			wantTypeAnn:  1,
		},
		{
			name: "complex with navigation properties",
			complexType: &csdl.ComplexType{
				Name: "NavComplex",
				Property: []*csdl.Property{
					{Name: "Prop1", Type: "Edm.String", Nullable: boolPtr(false)},
				},
				NavigationProperty: []*csdl.NavigationProperty{
					{Name: "Nav1", Type: "Test.Related", Annotation: []*csdl.Annotation{{Term: "Redfish.ExcerptCopy"}}},
					{Name: "Nav2", Type: "Test.Related2", Nullable: boolPtr(false)},
				},
			},
			namespace: "Test.v1_0_0",
			wantName:  "NavComplex",
			wantProps: 3,
			wantExcerpt: map[string]bool{
				"Nav1": true,
				"Nav2": false,
			},
			wantPropAnn: map[string]int{
				"Prop1": 0,
				"Nav1":  1,
				"Nav2":  0,
			},
			wantNull: map[string]bool{
				"Prop1": false,
				"Nav1":  true,
				"Nav2":  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csdl.NewTypeFromComplexType(tt.complexType, tt.namespace)
			assert.Equal(t, tt.wantName, result.Name)
			assert.Equal(t, tt.namespace, result.Namespace)
			assert.Equal(t, tt.wantProps, len(result.Properties))
			assert.Equal(t, tt.wantWildcard, result.Wildcard)
			assert.Len(t, result.Annotations, tt.wantTypeAnn)
			assert.True(t, result.ComplexType)
			for propName, wantExcerpt := range tt.wantExcerpt {
				prop, ok := result.Properties[propName]
				if !ok {
					t.Errorf("missing property %q", propName)
					continue
				}
				if prop.Excerpt != wantExcerpt {
					t.Errorf("property %q Excerpt = %v, want %v", propName, prop.Excerpt, wantExcerpt)
				}
				assert.Len(t, prop.Annotations, tt.wantPropAnn[propName])
			}
			for propName, wantNull := range tt.wantNull {
				prop, ok := result.Properties[propName]
				if !ok {
					t.Errorf("missing property %q", propName)
					continue
				}
				if prop.CanBeNull != wantNull {
					t.Errorf("property %q CanBeNull = %v, want %v", propName, prop.CanBeNull, wantNull)
				}
			}
			for propName, wantAnn := range tt.wantPropAnn {
				prop, ok := result.Properties[propName]
				if !ok {
					t.Errorf("missing property %q", propName)
					continue
				}
				assert.Len(t, prop.Annotations, wantAnn)
			}
		})
	}
}

func TestNewTypeFromEnumType(t *testing.T) {
	tests := []struct {
		name             string
		enumType         *csdl.EnumType
		namespace        string
		wantName         string
		wantMembers      int
		wantMemberValues map[string]string
		wantTypeAnn      int
		wantMemberAnn    map[string]int
	}{
		{
			name: "basic enum",
			enumType: &csdl.EnumType{
				Name: "TestEnum",
				Member: []*csdl.Member{
					{Name: "Value1"},
					{Name: "Value2"},
				},
			},
			namespace:   "Test.v1_0_0",
			wantName:    "TestEnum",
			wantMembers: 2,
		},
		{
			name: "enum with values",
			enumType: &csdl.EnumType{
				Name: "ValuedEnum",
				Annotation: []*csdl.Annotation{
					{Term: "Redfish.TestTypeAnnotation"},
				},
				Member: []*csdl.Member{
					{Name: "First", Value: stringPtr("1"), Annotation: []*csdl.Annotation{{Term: "Enum.MemberAnnotation"}}},
					{Name: "Second", Value: stringPtr("2")},
				},
			},
			namespace:   "Test.v1_0_0",
			wantName:    "ValuedEnum",
			wantMembers: 2,
			wantTypeAnn: 1,
			wantMemberAnn: map[string]int{
				"First":  1,
				"Second": 0,
			},
		},
		{
			name: "enum member values defined correctly",
			enumType: &csdl.EnumType{
				Name: "ValueCheckEnum",
				Member: []*csdl.Member{
					{Name: "AutoValue"},
					{Name: "DefinedValue", Value: stringPtr("42")},
				},
			},
			namespace:   "Test.v1_0_0",
			wantName:    "ValueCheckEnum",
			wantMembers: 2,
			wantMemberValues: map[string]string{
				"AutoValue":    "",
				"DefinedValue": "42",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := csdl.NewTypeFromEnumType(tt.enumType, tt.namespace)
			if result.Name != tt.wantName {
				t.Errorf("Name = %v, want %v", result.Name, tt.wantName)
			}
			if result.Namespace != tt.namespace {
				t.Errorf("Namespace = %v, want %v", result.Namespace, tt.namespace)
			}
			assert.Len(t, result.Annotations, tt.wantTypeAnn)
			if len(result.Members) != tt.wantMembers {
				t.Errorf("Members count = %v, want %v", len(result.Members), tt.wantMembers)
			}
			for memberName, wantValue := range tt.wantMemberValues {
				member, ok := result.Members[memberName]
				if !ok {
					t.Errorf("member %q not found", memberName)
					continue
				}
				if member.Value != wantValue {
					t.Errorf("member %q value = %q, want %q", memberName, member.Value, wantValue)
				}
			}
			for memberName, wantAnn := range tt.wantMemberAnn {
				member, ok := result.Members[memberName]
				if !ok {
					t.Errorf("member %q not found", memberName)
					continue
				}
				assert.Len(t, member.Annotations, wantAnn)
			}
		})
	}
}

func TestTypeFold(t *testing.T) {
	tests := []struct {
		name         string
		baseType     *csdl.Type
		derivedType  *csdl.Type
		replacements map[string]string
		wantProps    int
	}{
		{
			name: "simple inheritance",
			baseType: &csdl.Type{
				Name:      "Base",
				Namespace: "Test.v1_0_0",
				Properties: map[string]csdl.PropType{
					"BaseProp": {Type: "Edm.String"},
				},
			},
			derivedType: &csdl.Type{
				Name:      "Derived",
				Namespace: "Test.v1_0_0",
				BaseType:  "Test.v1_0_0.Base",
				Properties: map[string]csdl.PropType{
					"DerivedProp": {Type: "Edm.Int32"},
				},
			},
			replacements: map[string]string{},
			wantProps:    2,
		},
		{
			name: "property override",
			baseType: &csdl.Type{
				Name:      "Base",
				Namespace: "Test.v1_0_0",
				Properties: map[string]csdl.PropType{
					"Prop": {Type: "Edm.String"},
				},
			},
			derivedType: &csdl.Type{
				Name:      "Derived",
				Namespace: "Test.v1_0_0",
				BaseType:  "Test.v1_0_0.Base",
				Properties: map[string]csdl.PropType{
					"Prop": {Type: "Edm.Int32"},
				},
			},
			replacements: map[string]string{},
			wantProps:    1,
		},
		{
			name: "base type replaced",
			baseType: &csdl.Type{
				Name:      "ReplacementBase",
				Namespace: "Other.v1_0_0",
				Properties: map[string]csdl.PropType{
					"BaseProp": {Type: "Edm.String"},
				},
			},
			derivedType: &csdl.Type{
				Name:      "Derived",
				Namespace: "Test.v1_0_0",
				BaseType:  "Legacy.v1_0_0.Base",
				Properties: map[string]csdl.PropType{
					"DerivedProp": {Type: "Edm.Int32"},
				},
			},
			replacements: map[string]string{
				"Legacy.v1_0_0.Base": "Other.v1_0_0.ReplacementBase",
			},
			wantProps: 2,
		},
		{
			name:     "resource base type fallback",
			baseType: nil,
			derivedType: &csdl.Type{
				Name:       "Derived",
				Namespace:  "Test.v1_0_0",
				BaseType:   "Resource.v1_0_0.Resource",
				Properties: map[string]csdl.PropType{},
			},
			replacements: map[string]string{},
			wantProps:    4,
		},
		{
			name:     "resource links base type fallback",
			baseType: nil,
			derivedType: &csdl.Type{
				Name:       "DerivedLinks",
				Namespace:  "Test.v1_0_0",
				BaseType:   "Resource.Links",
				Properties: map[string]csdl.PropType{},
			},
			replacements: map[string]string{},
			wantProps:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			types := map[string]*csdl.Type{
				tt.derivedType.Namespace + "." + tt.derivedType.Name: tt.derivedType,
			}
			if tt.baseType != nil {
				types[tt.baseType.Namespace+"."+tt.baseType.Name] = tt.baseType
			}
			result := tt.derivedType.Fold(types, tt.replacements)
			if result == nil {
				t.Fatal("Fold returned nil")
			}
			if len(result.Properties) != tt.wantProps {
				t.Errorf("Properties count = %v, want %v", len(result.Properties), tt.wantProps)
			}
		})
	}
}

func TestTypeGoTypeName(t *testing.T) {
	tests := []struct {
		name     string
		typ      *csdl.Type
		wantName string
	}{
		{
			name: "simple type",
			typ: &csdl.Type{
				Name:      "TestType",
				Namespace: "Test.v1_0_0",
			},
			wantName: "Test_TestType",
		},
		{
			name: "namespace matches name prefix",
			typ: &csdl.Type{
				Name:      "TestType",
				Namespace: "TestType.v1_0_0",
			},
			wantName: "TestType",
		},
		{
			name: "simple namespace",
			typ: &csdl.Type{
				Name:      "Status",
				Namespace: "Resource",
			},
			wantName: "Resource_Status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.typ.GoTypeName()
			if result != tt.wantName {
				t.Errorf("GoTypeName() = %v, want %v", result, tt.wantName)
			}
		})
	}
}

func TestTypeNodeStructWhenNotComplexType(t *testing.T) {
	typ := &csdl.Type{
		Name:        "TestEntity",
		Namespace:   "Test.v1_0_0",
		ComplexType: false,
		Properties: map[string]csdl.PropType{
			"Id":          {Type: "Edm.String"},
			"Name":        {Type: "Edm.String"},
			"Description": {Type: "Edm.String"},
			"Alpha":       {Type: "Edm.String"},
		},
	}

	nodes := typ.Node(map[string]*csdl.Type{})
	require.Len(t, nodes, 1)

	decl, ok := nodes[0].(*ast.GenDecl)
	require.True(t, ok)
	require.Len(t, decl.Specs, 1)

	typeSpec, ok := decl.Specs[0].(*ast.TypeSpec)
	require.True(t, ok)

	structType, ok := typeSpec.Type.(*ast.StructType)
	require.True(t, ok)

	fieldNames := make([]string, 0, len(structType.Fields.List))
	for _, field := range structType.Fields.List {
		require.NotEmpty(t, field.Names)
		fieldNames = append(fieldNames, field.Names[0].Name)
	}
	assert.Equal(t, []string{"ID", "Id", "Type", "Context", "Name", "Description", "Alpha"}, fieldNames)

	assert.Equal(t, "`json:\"@odata.id\"`", structType.Fields.List[0].Tag.Value)
	assert.Equal(t, "`json:\"@odata.type,omitempty\"`", structType.Fields.List[2].Tag.Value)
	assert.Equal(t, "`json:\"@odata.context,omitempty\"`", structType.Fields.List[3].Tag.Value)
	assert.Equal(t, "`json:\",omitempty\"`", structType.Fields.List[5].Tag.Value)
}

func TestPropTypeNode(t *testing.T) {
	tests := []struct {
		name     string
		propType csdl.PropType
		wantType string
		types    map[string]*csdl.Type
	}{
		{
			name:     "navigation property",
			propType: csdl.PropType{Navigation: true, Type: "Test.Related", CanBeNull: false},
			wantType: "OdataID",
		},
		{
			name:     "nullable navigation property",
			propType: csdl.PropType{Navigation: true, Type: "Test.Related", CanBeNull: true},
			wantType: "*OdataID",
		},
		{
			name:     "Edm.Boolean",
			propType: csdl.PropType{Type: "Edm.Boolean", CanBeNull: false},
			wantType: "bool",
		},
		{
			name:     "nullable Edm.Boolean",
			propType: csdl.PropType{Type: "Edm.Boolean", CanBeNull: true},
			wantType: "*bool",
		},
		{
			name:     "Edm.String",
			propType: csdl.PropType{Type: "Edm.String", CanBeNull: false},
			wantType: "string",
		},
		{
			name:     "Edm.Byte",
			propType: csdl.PropType{Type: "Edm.Byte", CanBeNull: false},
			wantType: "byte",
		},
		{
			name:     "Edm.Date",
			propType: csdl.PropType{Type: "Edm.Date", CanBeNull: false},
			wantType: "Date",
		},
		{
			name:     "Edm.DateTimeOffset",
			propType: csdl.PropType{Type: "Edm.DateTimeOffset", CanBeNull: false},
			wantType: "DateTimeOffset",
		},
		{
			name:     "Edm.Decimal",
			propType: csdl.PropType{Type: "Edm.Decimal", CanBeNull: false},
			wantType: "float64",
		},
		{
			name:     "Edm.Duration",
			propType: csdl.PropType{Type: "Edm.Duration", CanBeNull: false},
			wantType: "Duration",
		},
		{
			name:     "Edm.Int16",
			propType: csdl.PropType{Type: "Edm.Int16", CanBeNull: false},
			wantType: "int16",
		},
		{
			name:     "Edm.Int32",
			propType: csdl.PropType{Type: "Edm.Int32", CanBeNull: false},
			wantType: "int32",
		},
		{
			name:     "Edm.Int64",
			propType: csdl.PropType{Type: "Edm.Int64", CanBeNull: false},
			wantType: "int64",
		},
		{
			name:     "Edm.Double",
			propType: csdl.PropType{Type: "Edm.Double", CanBeNull: false},
			wantType: "float64",
		},
		{
			name:     "Edm.PrimitiveType",
			propType: csdl.PropType{Type: "Edm.PrimitiveType", CanBeNull: false},
			wantType: "any",
		},
		{
			name:     "Edm.SByte",
			propType: csdl.PropType{Type: "Edm.SByte", CanBeNull: false},
			wantType: "int8",
		},
		{
			name:     "Edm.Single",
			propType: csdl.PropType{Type: "Edm.Single", CanBeNull: false},
			wantType: "float32",
		},
		{
			name:     "Edm.Guid",
			propType: csdl.PropType{Type: "Edm.Guid", CanBeNull: false},
			wantType: "UUID",
		},
		{
			name:     "Resource.UUID",
			propType: csdl.PropType{Type: "Resource.UUID", CanBeNull: false},
			wantType: "UUID",
		},
		{
			name:     "Collection type",
			propType: csdl.PropType{Type: "Collection(Edm.String)", CanBeNull: false},
			wantType: "[]string",
		},
		{
			name:     "Resource.Description",
			propType: csdl.PropType{Type: "Resource.Description", CanBeNull: false},
			wantType: "string",
		},
		{
			name:     "Resource.Name",
			propType: csdl.PropType{Type: "Resource.Name", CanBeNull: false},
			wantType: "string",
		},
		{
			name:     "Resource.Status",
			propType: csdl.PropType{Type: "Resource.Status", CanBeNull: false},
			wantType: "Resource_Status",
		},
		{
			name:     "Resource.PowerState",
			propType: csdl.PropType{Type: "Resource.PowerState", CanBeNull: false},
			wantType: "Resource_PowerState",
		},
		{
			name:     "Resource.Oem",
			propType: csdl.PropType{Type: "Resource.Oem", CanBeNull: false},
			wantType: "map[string]any",
		},
		{
			name:     "OemActions type",
			propType: csdl.PropType{Type: "ComputerSystem_OemActions", CanBeNull: false},
			wantType: "map[string]any",
		},
		{
			name:     "ItemOrCollection unknown type",
			propType: csdl.PropType{Type: "ItemOrCollection", CanBeNull: false},
			wantType: "any",
		},
		{
			name:     "custom known type",
			propType: csdl.PropType{Type: "Custom.v1_0_0.MyType", CanBeNull: false},
			wantType: "Custom_MyType",
			types: map[string]*csdl.Type{
				"Custom.v1_0_0.MyType": {Name: "MyType", Namespace: "Custom.v1_0_0"},
			},
		},
		{
			name:     "nullable custom known type",
			propType: csdl.PropType{Type: "Custom.v1_0_0.MyType", CanBeNull: true},
			wantType: "*Custom_MyType",
			types: map[string]*csdl.Type{
				"Custom.v1_0_0.MyType": {Name: "MyType", Namespace: "Custom.v1_0_0"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.propType.Node(tt.types)
			ident, ok := node.(*ast.Ident)
			if !ok {
				t.Fatalf("Expected *ast.Ident, got %T", node)
			}
			if ident.Name != tt.wantType {
				t.Errorf("Node type = %v, want %v", ident.Name, tt.wantType)
			}
		})
	}

	t.Run("Type.Node special empty complex types", func(t *testing.T) {
		types := map[string]*csdl.Type{}

		oemActionsType := &csdl.Type{Name: "ComputerSystemOemActions", Namespace: "Test.v1_0_0"}
		assert.Empty(t, oemActionsType.Node(types))

		itemOrCollectionType := &csdl.Type{Name: "ItemOrCollection", Namespace: "Test.v1_0_0"}
		assert.Empty(t, itemOrCollectionType.Node(types))

		wildcardType := &csdl.Type{Name: "Attributes", Namespace: "Test.v1_0_0", Wildcard: true}
		assert.Empty(t, wildcardType.Node(types))
	})
}

func TestPropTypeToField(t *testing.T) {
	tests := []struct {
		name      string
		propType  csdl.PropType
		fieldName string
		wantTag   string
	}{
		{
			name:      "Actions field",
			propType:  csdl.PropType{Type: "Test.Actions"},
			fieldName: "Actions",
			wantTag:   "`json:\",omitempty\"`",
		},
		{
			name:      "Oem field",
			propType:  csdl.PropType{Type: "Resource.Oem"},
			fieldName: "Oem",
			wantTag:   "`json:\",omitempty\"`",
		},
		{
			name:      "field with JsonName",
			propType:  csdl.PropType{Type: "Edm.String", JsonName: "@odata.id"},
			fieldName: "ID",
			wantTag:   "`json:\"@odata.id\"`",
		},
		{
			name:      "nullable field",
			propType:  csdl.PropType{Type: "Edm.String", CanBeNull: true},
			fieldName: "NullableField",
			wantTag:   "`json:\",omitempty\"`",
		},
		{
			name:      "navigation field",
			propType:  csdl.PropType{Type: "Test.Related", Navigation: true},
			fieldName: "NavField",
			wantTag:   "`json:\",omitempty\"`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			types := map[string]*csdl.Type{}
			replacements := map[string]string{}
			field := tt.propType.ToField(tt.fieldName, types, replacements)
			if field == nil {
				t.Fatal("ToField returned nil")
			}
			if len(field.Names) != 1 || field.Names[0].Name != tt.fieldName {
				t.Errorf("Field name = %v, want %v", field.Names[0].Name, tt.fieldName)
			}
			if field.Tag != nil && field.Tag.Value != tt.wantTag {
				t.Errorf("Field tag = %v, want %v", field.Tag.Value, tt.wantTag)
			}
		})
	}
}

func BenchmarkNewTypeFromEntityType(b *testing.B) {
	entityType := &csdl.EntityType{
		Name:     "BenchEntity",
		BaseType: "Resource.v1_0_0.Resource",
		Property: []*csdl.Property{
			{Name: "Prop1", Type: "Edm.String"},
			{Name: "Prop2", Type: "Edm.Int32"},
			{Name: "Prop3", Type: "Edm.Boolean"},
		},
		NavigationProperty: []*csdl.NavigationProperty{
			{Name: "Nav1", Type: "Test.Related"},
		},
	}
	namespace := "Test.v1_0_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = csdl.NewTypeFromEntityType(entityType, namespace)
	}
}

func BenchmarkNewTypeFromComplexType(b *testing.B) {
	complexType := &csdl.ComplexType{
		Name:     "BenchComplex",
		BaseType: "Resource.Links",
		Property: []*csdl.Property{
			{Name: "Prop1", Type: "Edm.String"},
			{Name: "Prop2", Type: "Edm.Int32"},
		},
	}
	namespace := "Test.v1_0_0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = csdl.NewTypeFromComplexType(complexType, namespace)
	}
}

func BenchmarkTypeFold(b *testing.B) {
	baseType := &csdl.Type{
		Name:      "Base",
		Namespace: "Test.v1_0_0",
		Properties: map[string]csdl.PropType{
			"BaseProp1": {Type: "Edm.String"},
			"BaseProp2": {Type: "Edm.Int32"},
		},
	}
	derivedType := &csdl.Type{
		Name:      "Derived",
		Namespace: "Test.v1_0_0",
		BaseType:  "Test.v1_0_0.Base",
		Properties: map[string]csdl.PropType{
			"DerivedProp": {Type: "Edm.Boolean"},
		},
	}
	types := map[string]*csdl.Type{
		"Test.v1_0_0.Base":    baseType,
		"Test.v1_0_0.Derived": derivedType,
	}
	replacements := map[string]string{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testType := &csdl.Type{
			Name:      "Derived",
			Namespace: "Test.v1_0_0",
			BaseType:  "Test.v1_0_0.Base",
			Properties: map[string]csdl.PropType{
				"DerivedProp": {Type: "Edm.Boolean"},
			},
		}
		_ = testType.Fold(types, replacements)
	}
}

func BenchmarkGoTypeName(b *testing.B) {
	typ := &csdl.Type{
		Name:      "TestType",
		Namespace: "Test.v1_0_0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = typ.GoTypeName()
	}
}

func BenchmarkPropTypeNode(b *testing.B) {
	propType := csdl.PropType{
		Type:      "Edm.String",
		CanBeNull: true,
	}
	types := map[string]*csdl.Type{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = propType.Node(types)
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}
