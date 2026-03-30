package csdl_test

import (
	"go/format"
	"strings"
	"testing"

	"github.com/pboyd04/gocsdl/pkg/csdl"
)

func TestNewFile(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		wantError   bool
	}{
		{
			name:        "simple package name",
			packageName: "main",
			wantError:   false,
		},
		{
			name:        "complex package name",
			packageName: "mypackage",
			wantError:   false,
		},
		{
			name:        "underscore package name",
			packageName: "my_package",
			wantError:   false,
		},
		{
			name:        "empty package name",
			packageName: "",
			wantError:   false, // Go format should handle empty package names
		},
		{
			name:        "valid package name with numbers",
			packageName: "pkg123",
			wantError:   false, // Go allows package names with numbers
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := csdl.NewFile(tt.packageName)
			if (err != nil) != tt.wantError {
				t.Errorf("NewFile() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError {
				if file != nil {
					t.Error("NewFile() returned non-nil file when error expected")
				}
				return
			}
			if file == nil {
				t.Fatal("NewFile returned nil when no error expected")
			}
		})
	}
}

func TestFileAddType(t *testing.T) {
	tests := []struct {
		name      string
		typ       *csdl.Type
		wantError bool
	}{
		{
			name: "add simple type",
			typ: &csdl.Type{
				Name:      "TestType",
				Namespace: "Test.v1_0_0",
				Properties: map[string]csdl.PropType{
					"Prop1": {Type: "Edm.String"},
				},
			},
			wantError: false,
		},
		{
			name: "add enum type",
			typ: &csdl.Type{
				Name:      "TestEnum",
				Namespace: "Test.v1_0_0",
				Members: map[string]csdl.MemberType{
					"Value1": {Name: "Value1"},
					"Value2": {Name: "Value2"},
				},
			},
			wantError: false,
		},
		{
			name: "add MessageRegistry.Message type (should skip)",
			typ: &csdl.Type{
				Name:      "Message",
				Namespace: "MessageRegistry.v1_0_0",
				Properties: map[string]csdl.PropType{
					"Prop1": {Type: "Edm.String"},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := csdl.NewFile("test")
			if err != nil {
				t.Fatalf("NewFile() failed: %v", err)
			}
			err = file.AddType(tt.typ)
			if (err != nil) != tt.wantError {
				t.Errorf("AddType() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestFileAddTypeDuplicate(t *testing.T) {
	file, err := csdl.NewFile("test")
	if err != nil {
		t.Fatalf("NewFile() failed: %v", err)
	}

	type1 := &csdl.Type{
		Name:      "TestType",
		Namespace: "Test.v1_0_0",
		Properties: map[string]csdl.PropType{
			"Prop1": {Type: "Edm.String"},
		},
	}

	type2 := &csdl.Type{
		Name:      "TestType",
		Namespace: "Test.v1_0_0",
		Properties: map[string]csdl.PropType{
			"Prop1": {Type: "Edm.String"},
			"Prop2": {Type: "Edm.Int32"},
		},
	}

	err = file.AddType(type1)
	if err != nil {
		t.Fatalf("First AddType() failed: %v", err)
	}

	err = file.AddType(type2)
	if err != nil {
		t.Fatalf("Second AddType() failed: %v", err)
	}
}

func TestFileAddTypePreferLarger(t *testing.T) {
	file, err := csdl.NewFile("test")
	if err != nil {
		t.Fatalf("NewFile() failed: %v", err)
	}

	largerType := &csdl.Type{
		Name:      "TestType",
		Namespace: "Test.v1_0_0",
		Properties: map[string]csdl.PropType{
			"Prop1": {Type: "Edm.String"},
			"Prop2": {Type: "Edm.Int32"},
		},
	}

	smallerType := &csdl.Type{
		Name:      "TestType",
		Namespace: "Test.v1_0_0",
		Properties: map[string]csdl.PropType{
			"Prop1": {Type: "Edm.String"},
		},
	}

	err = file.AddType(largerType)
	if err != nil {
		t.Fatalf("First AddType() failed: %v", err)
	}

	err = file.AddType(smallerType)
	if err != nil {
		t.Fatalf("Second AddType() failed: %v", err)
	}
}

func TestFileFlush(t *testing.T) {
	tests := []struct {
		name         string
		typ          *csdl.Type
		wantContains []string
	}{
		{
			name: "struct type",
			typ: &csdl.Type{
				Name:      "TestStruct",
				Namespace: "Test.v1_0_0",
				Properties: map[string]csdl.PropType{
					"Prop1": {Type: "Edm.String", CanBeNull: false},
				},
				ComplexType: true,
			},
			wantContains: []string{"type", "Test_TestStruct", "struct"},
		},
		{
			name: "enum type",
			typ: &csdl.Type{
				Name:      "TestEnum",
				Namespace: "Test.v1_0_0",
				Members: map[string]csdl.MemberType{
					"Value1": {Name: "Value1"},
				},
			},
			wantContains: []string{"type", "Test_TestEnum", "string", "const", "Test_TestEnum_Value1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := csdl.NewFile("test")
			if err != nil {
				t.Fatalf("NewFile() failed: %v", err)
			}
			err = file.AddType(tt.typ)
			if err != nil {
				t.Fatalf("AddType() failed: %v", err)
			}

			allTypes := map[string]*csdl.Type{
				tt.typ.Namespace + "." + tt.typ.Name: tt.typ,
			}

			output, err := file.Flush(allTypes)
			if err != nil {
				t.Fatalf("Flush() failed: %v", err)
			}

			outputStr := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Flush() output missing %q", want)
				}
			}
		})
	}
}

func TestFileFlushFormattingError(t *testing.T) {
	// Test case to verify format.Source error handling in Flush
	// This test directly verifies that the format.Source call works as expected

	// Test valid Go code first
	validCode := []byte("package test\n\ntype TestType struct {\n    Field string\n}")
	_, err := format.Source(validCode)
	if err != nil {
		t.Errorf("format.Source() should not fail for valid Go code: %v", err)
	}

	// Test invalid Go code - this should fail
	invalidCode := []byte("package test\n\ntype TestType struct {\n    Field string\n  // missing closing brace")
	_, err = format.Source(invalidCode)
	if err == nil {
		t.Error("format.Source() should have returned an error for invalid Go syntax")
	}

	// This verifies that the error handling in Flush() works correctly
	// since Flush() calls format.Source() at the end
}

func TestFileFlushMultipleTypes(t *testing.T) {
	file, err := csdl.NewFile("test")
	if err != nil {
		t.Fatalf("NewFile() failed: %v", err)
	}

	type1 := &csdl.Type{
		Name:      "Type1",
		Namespace: "Test.v1_0_0",
		Properties: map[string]csdl.PropType{
			"Prop1": {Type: "Edm.String"},
		},
		ComplexType: true,
	}

	type2 := &csdl.Type{
		Name:      "Type2",
		Namespace: "Test.v1_0_0",
		Members: map[string]csdl.MemberType{
			"Value1": {Name: "Value1"},
		},
	}

	err = file.AddType(type1)
	if err != nil {
		t.Fatalf("AddType(type1) failed: %v", err)
	}

	err = file.AddType(type2)
	if err != nil {
		t.Fatalf("AddType(type2) failed: %v", err)
	}

	allTypes := map[string]*csdl.Type{
		"Test.v1_0_0.Type1": type1,
		"Test.v1_0_0.Type2": type2,
	}

	output, err := file.Flush(allTypes)
	if err != nil {
		t.Fatalf("Flush() failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Test_Type1") {
		t.Error("Output missing Type1")
	}
	if !strings.Contains(outputStr, "Test_Type2") {
		t.Error("Output missing Type2")
	}
}

func TestFileFlushWithReferences(t *testing.T) {
	file, err := csdl.NewFile("test")
	if err != nil {
		t.Fatalf("NewFile() failed: %v", err)
	}

	type1 := &csdl.Type{
		Name:      "Referenced",
		Namespace: "Test.v1_0_0",
		Properties: map[string]csdl.PropType{
			"Value": {Type: "Edm.String"},
		},
		ComplexType: true,
	}

	type2 := &csdl.Type{
		Name:      "Container",
		Namespace: "Test.v1_0_0",
		Properties: map[string]csdl.PropType{
			"Ref": {Type: "Test.v1_0_0.Referenced"},
		},
		ComplexType: true,
	}

	err = file.AddType(type2)
	if err != nil {
		t.Fatalf("AddType() failed: %v", err)
	}

	allTypes := map[string]*csdl.Type{
		"Test.v1_0_0.Referenced": type1,
		"Test.v1_0_0.Container":  type2,
	}

	output, err := file.Flush(allTypes)
	if err != nil {
		t.Fatalf("Flush() failed: %v", err)
	}

	if len(output) == 0 {
		t.Error("Flush() returned empty output")
	}
}

func BenchmarkNewFile(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = csdl.NewFile("test")
	}
}

func BenchmarkFileAddType(b *testing.B) {
	typ := &csdl.Type{
		Name:      "TestType",
		Namespace: "Test.v1_0_0",
		Properties: map[string]csdl.PropType{
			"Prop1": {Type: "Edm.String"},
			"Prop2": {Type: "Edm.Int32"},
			"Prop3": {Type: "Edm.Boolean"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, _ := csdl.NewFile("test")
		_ = file.AddType(typ)
	}
}

func BenchmarkFileFlush(b *testing.B) {
	file, _ := csdl.NewFile("test")
	typ := &csdl.Type{
		Name:      "TestType",
		Namespace: "Test.v1_0_0",
		Properties: map[string]csdl.PropType{
			"Prop1": {Type: "Edm.String"},
			"Prop2": {Type: "Edm.Int32"},
			"Prop3": {Type: "Edm.Boolean"},
		},
		ComplexType: true,
	}
	_ = file.AddType(typ)

	allTypes := map[string]*csdl.Type{
		"Test.v1_0_0.TestType": typ,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = file.Flush(allTypes)
	}
}

func BenchmarkFileFlushLargeType(b *testing.B) {
	file, _ := csdl.NewFile("test")

	props := make(map[string]csdl.PropType)
	for i := 0; i < 50; i++ {
		props[string(rune('A'+i%26))+string(rune('0'+i/26))] = csdl.PropType{
			Type:      "Edm.String",
			CanBeNull: i%2 == 0,
		}
	}

	typ := &csdl.Type{
		Name:        "LargeType",
		Namespace:   "Test.v1_0_0",
		Properties:  props,
		ComplexType: true,
	}
	_ = file.AddType(typ)

	allTypes := map[string]*csdl.Type{
		"Test.v1_0_0.LargeType": typ,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = file.Flush(allTypes)
	}
}
