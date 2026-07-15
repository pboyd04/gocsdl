package csdl_test

import (
	"testing"

	"github.com/pboyd04/gocsdl/pkg/csdl"
	"github.com/stretchr/testify/assert"
)

type NamedElementTestCase struct {
	Name          string
	Type          csdl.Searchable
	ExpectedName  string
	ExpectedXPath string
}

func GetNamedElementTestCases() []NamedElementTestCase {
	return []NamedElementTestCase{
		{Name: "Schema", Type: &csdl.Schema{Namespace: "Test.v1_0_0"}, ExpectedName: "Test.v1_0_0", ExpectedXPath: "Test.v1_0_0"},
		{Name: "Action", Type: csdl.Action{Name: "TestAction"}, ExpectedName: "TestAction", ExpectedXPath: "Action[@Name='TestAction']"},
		{Name: "Annotation", Type: csdl.Annotation{Term: "Test.Term"}, ExpectedName: "Test.Term", ExpectedXPath: "Annotation[@Term='Test.Term']"},
		{Name: "Annotations", Type: csdl.Annotations{Target: "Test.Target"}, ExpectedName: "Test.Target", ExpectedXPath: "Annotations[@Target='Test.Target']"},
		{Name: "ComplexType", Type: csdl.ComplexType{Name: "TestComplex"}, ExpectedName: "TestComplex", ExpectedXPath: "ComplexType[@Name='TestComplex']"},
		{Name: "EntityContainer", Type: &csdl.EntityContainer{Name: "ServiceRoot"}, ExpectedName: "ServiceRoot", ExpectedXPath: "EntityContainer[@Name='ServiceRoot']"},
		{Name: "EntityType", Type: &csdl.EntityType{Name: "TestEntity"}, ExpectedName: "TestEntity", ExpectedXPath: "EntityType[@Name='TestEntity']"},
		{Name: "EnumType", Type: csdl.EnumType{Name: "TestEnum"}, ExpectedName: "TestEnum", ExpectedXPath: "EnumType[@Name='TestEnum']"},
		{Name: "Member", Type: csdl.Member{Name: "TestMember"}, ExpectedName: "TestMember", ExpectedXPath: "Member[@Name='TestMember']"},
		{Name: "NavigationProperty", Type: csdl.NavigationProperty{Name: "TestNav"}, ExpectedName: "TestNav", ExpectedXPath: "NavigationProperty[@Name='TestNav']"},
		{Name: "Parameter", Type: csdl.Parameter{Name: "TestParam"}, ExpectedName: "TestParam", ExpectedXPath: "Parameter[@Name='TestParam']"},
		{Name: "Property", Type: csdl.Property{Name: "TestProp"}, ExpectedName: "TestProp", ExpectedXPath: "Property[@Name='TestProp']"},
		{Name: "Term", Type: csdl.Term{Name: "TestTerm"}, ExpectedName: "TestTerm", ExpectedXPath: "Term[@Name='TestTerm']"},
		{Name: "TypeDefinition", Type: csdl.TypeDefinition{Name: "TestTypeDef"}, ExpectedName: "TestTypeDef", ExpectedXPath: "TestTypeDef"},
		{Name: "ReturnType", Type: csdl.ReturnType{Type: "Collection(Test.Widget)"}, ExpectedName: "Collection(Test.Widget)", ExpectedXPath: "ReturnType[@Type='Collection(Test.Widget)']"},
	}
}

func TestGetName(t *testing.T) {
	t.Parallel()
	testCases := GetNamedElementTestCases()
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.ExpectedName, tc.Type.GetName())
		})
	}
}

func BenchmarkGetName(b *testing.B) {
	testCases := GetNamedElementTestCases()
	for _, tc := range testCases {
		b.Run(tc.Name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = tc.Type.GetName()
			}
		})
	}
}

func TestGetXPath(t *testing.T) {
	t.Parallel()
	for _, tc := range GetNamedElementTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.ExpectedXPath, tc.Type.GetXPath())
		})
	}
}

func BenchmarkGetXPath(b *testing.B) {
	testCases := GetNamedElementTestCases()
	for _, tc := range testCases {
		b.Run(tc.Name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = tc.Type.GetXPath()
			}
		})
	}
}

type GetChildrenTestCase struct {
	Name          string
	Type          csdl.Searchable
	ExpectedCount int
	ExpectedNil   bool
}

func GetGetChildrenTestCases() []GetChildrenTestCase {
	annotation := []*csdl.Annotation{{Term: "Term1"}}
	return []GetChildrenTestCase{
		{
			Name: "Schema",
			Type: &csdl.Schema{
				Namespace:      "Test.v1_0_0",
				Action:         []*csdl.Action{{Name: "Action1"}},
				ComplexType:    []*csdl.ComplexType{{Name: "Complex1"}},
				EntityType:     []*csdl.EntityType{{Name: "Entity1"}},
				EnumType:       []*csdl.EnumType{{Name: "Enum1"}},
				Term:           []*csdl.Term{{Name: "Term1"}},
				TypeDefinition: []*csdl.TypeDefinition{{Name: "TypeDef1"}},
			},
			ExpectedCount: 6,
		},
		{
			Name: "Action",
			Type: csdl.Action{
				Name:       "TestAction",
				Annotation: annotation,
				Parameter:  []*csdl.Parameter{{Name: "Param1"}},
				ReturnType: &csdl.ReturnType{Type: "Edm.String"},
			},
			ExpectedCount: 3,
		},
		{Name: "ReturnType", Type: csdl.ReturnType{Type: "Collection(Test.Widget)"}, ExpectedNil: true},
		{Name: "Annotation", Type: csdl.Annotation{Term: "Test.Term"}, ExpectedNil: true},
		{Name: "Annotations", Type: csdl.Annotations{Target: "Test.Target", Annotation: []*csdl.Annotation{{Term: "Term1"}, {Term: "Term2"}}}, ExpectedCount: 2},
		{Name: "ComplexType", Type: csdl.ComplexType{Name: "TestComplex", Property: []*csdl.Property{{Name: "Prop1"}}, NavigationProperty: []*csdl.NavigationProperty{{Name: "Nav1"}}, Annotation: annotation}, ExpectedCount: 3},
		{Name: "EntityContainer", Type: &csdl.EntityContainer{Name: "ServiceRoot"}, ExpectedNil: true},
		{Name: "EntityType", Type: &csdl.EntityType{Name: "TestEntity", Property: []*csdl.Property{{Name: "Prop1"}}, NavigationProperty: []*csdl.NavigationProperty{{Name: "Nav1"}}, Annotation: annotation}, ExpectedCount: 3},
		{Name: "EnumType", Type: csdl.EnumType{Name: "TestEnum", Annotation: annotation, Member: []*csdl.Member{{Name: "Member1"}}}, ExpectedCount: 2},
		{Name: "Member", Type: csdl.Member{Name: "TestMember", Annotation: annotation}, ExpectedCount: 1},
		{Name: "NavigationProperty", Type: csdl.NavigationProperty{Name: "TestNav", Annotation: annotation}, ExpectedCount: 1},
		{Name: "Parameter", Type: csdl.Parameter{Name: "TestParam", Annotation: annotation}, ExpectedCount: 1},
		{Name: "Property", Type: csdl.Property{Name: "TestProp", Annotation: annotation}, ExpectedCount: 1},
		{Name: "Term", Type: csdl.Term{Name: "TestTerm", Annotation: annotation}, ExpectedCount: 1},
		{Name: "TypeDefinition", Type: csdl.TypeDefinition{Name: "TestTypeDef", Annotation: annotation}, ExpectedCount: 1},
	}
}

func TestGetChildren(t *testing.T) {
	t.Parallel()
	for _, tc := range GetGetChildrenTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			children := tc.Type.GetChildren()
			if tc.ExpectedNil {
				assert.Nil(t, children)
				return
			}
			assert.Len(t, children, tc.ExpectedCount)
		})
	}
}

func BenchmarkGetChildren(b *testing.B) {
	for _, tc := range GetGetChildrenTestCases() {
		b.Run(tc.Name, func(b *testing.B) {
			for b.Loop() {
				b.ReportAllocs()
				_ = tc.Type.GetChildren()
			}
		})
	}
}

func TestComplexTypeSetGetSchema(t *testing.T) {
	schema := &csdl.Schema{Namespace: "Test.v1_0_0"}
	complexType := &csdl.ComplexType{Name: "TestComplex"}

	complexType.SetSchema(schema)
	result := complexType.GetSchema()

	if result != schema {
		t.Error("SetSchema/GetSchema failed")
	}
}

func TestSchemaGetChildrenIncludesEntityContainer(t *testing.T) {
	schema := &csdl.Schema{
		Namespace: "Test.v1_0_0",
		EntityContainer: []*csdl.EntityContainer{
			{Name: "ServiceRoot"},
		},
	}

	children := schema.GetChildren()
	if len(children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(children))
	}

	container, ok := children[0].(*csdl.EntityContainer)
	if !ok {
		t.Fatalf("Expected child type *csdl.EntityContainer, got %T", children[0])
	}
	if container.Name != "ServiceRoot" {
		t.Errorf("EntityContainer.Name = %v, want %v", container.Name, "ServiceRoot")
	}
}

func TestEntityTypeSetGetSchema(t *testing.T) {
	schema := &csdl.Schema{Namespace: "Test.v1_0_0"}
	entityType := &csdl.EntityType{Name: "TestEntity"}

	entityType.SetSchema(schema)
	result := entityType.GetSchema()

	if result != schema {
		t.Error("SetSchema/GetSchema failed")
	}
}

func BenchmarkSchemaGetChildren(b *testing.B) {
	schema := &csdl.Schema{
		Namespace: "Test.v1_0_0",
		Action: []*csdl.Action{
			{Name: "Action1"},
			{Name: "Action2"},
		},
		ComplexType: []*csdl.ComplexType{
			{Name: "Complex1"},
			{Name: "Complex2"},
		},
		EntityType: []*csdl.EntityType{
			{Name: "Entity1"},
			{Name: "Entity2"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schema.GetChildren()
	}
}

func BenchmarkComplexTypeGetChildren(b *testing.B) {
	complexType := csdl.ComplexType{
		Name: "TestComplex",
		Property: []*csdl.Property{
			{Name: "Prop1"},
			{Name: "Prop2"},
			{Name: "Prop3"},
		},
		NavigationProperty: []*csdl.NavigationProperty{
			{Name: "Nav1"},
			{Name: "Nav2"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = complexType.GetChildren()
	}
}

func BenchmarkEntityTypeGetChildren(b *testing.B) {
	entityType := csdl.EntityType{
		Name: "TestEntity",
		Property: []*csdl.Property{
			{Name: "Prop1"},
			{Name: "Prop2"},
			{Name: "Prop3"},
		},
		NavigationProperty: []*csdl.NavigationProperty{
			{Name: "Nav1"},
			{Name: "Nav2"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = entityType.GetChildren()
	}
}
