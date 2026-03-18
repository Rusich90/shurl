package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// Тесты для parseFile функции
func TestParseFile(t *testing.T) {
	// Тест с корректной структурой с комментарием
	validSrc := `package test

// generate:reset
type ValidStruct struct {
	Field int
}

type InvalidStruct struct {
	Field string
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", validSrc, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем карту комментариев как в оригинальной функции
	commentMap := make(map[*ast.TypeSpec]*ast.CommentGroup)
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if strings.TrimSpace(comment.Text) == "// generate:reset" {
				for _, decl := range file.Decls {
					if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
						for _, spec := range genDecl.Specs {
							if typeSpec, ok := spec.(*ast.TypeSpec); ok {
								if comment.Pos() < typeSpec.Pos() {
									hasOtherType := false
									for _, decl2 := range file.Decls {
										if genDecl2, ok := decl2.(*ast.GenDecl); ok && genDecl2.Tok == token.TYPE {
											for _, spec2 := range genDecl2.Specs {
												if typeSpec2, ok := spec2.(*ast.TypeSpec); ok {
													if typeSpec2.Pos() > comment.Pos() && typeSpec2.Pos() < typeSpec.Pos() {
														hasOtherType = true
														break
													}
												}
											}
										}
									}
									if !hasOtherType {
										commentMap[typeSpec] = group
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Проверяем количество деклараций
	typeDecls := 0
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			typeDecls++
		}
	}

	if typeDecls != 2 {
		t.Errorf("expected 2 type declarations, got %d", typeDecls)
	}

	// Проверяем структуры
	structsFound := 0
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if hasGenerateResetComment(commentMap[typeSpec]) {
				structsFound++

				// Проверяем, что это структура
				if _, ok := typeSpec.Type.(*ast.StructType); !ok {
					t.Error("expected struct type")
				}

				if typeSpec.Name.Name != "ValidStruct" {
					t.Errorf("expected ValidStruct, got %s", typeSpec.Name.Name)
				}
			}
		}
	}

	if structsFound != 1 {
		t.Errorf("expected 1 struct with reset comment, got %d", structsFound)
	}
}

// Тесты для различных типов полей
func TestParseFieldTypes(t *testing.T) {
	fset := token.NewFileSet()

	// Тест для разных типов полей
	src := `package test
	
type ComplexStruct struct {
	// Простые типы
	IntField int
	StringField string
	BoolField bool
	
	// Указатели
	PtrInt *int
	PtrString *string
	
	// Слайсы
	IntSlice []int
	StringSlice []string
	PtrSlice []*int
	
	// Мапы
	StringMap map[string]string
	IntMap map[int]int
	
	// Структуры из других пакетов
	ExtStruct other.PackageStruct
	PtrExtStruct *other.PackageStruct
	
	// Структуры из текущего пакета
	LocalStruct LocalType
	PtrLocalStruct *LocalType
}

type LocalType struct {
	Field int
}
`

	file, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Находим ComplexStruct
	var structType *ast.StructType
	ast.Inspect(file, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == "ComplexStruct" {
			if st, ok := ts.Type.(*ast.StructType); ok {
				structType = st
				return false
			}
		}
		return true
	})

	if structType == nil {
		t.Fatal("ComplexStruct not found")
	}

	// Проверяем каждое поле
	expectedFields := map[string]FieldInfo{
		"IntField": {
			Type: "int",
		},
		"StringField": {
			Type: "string",
		},
		"BoolField": {
			Type: "bool",
		},
		"PtrInt": {
			Type:  "int",
			IsPtr: true,
		},
		"PtrString": {
			Type:  "string",
			IsPtr: true,
		},
		"IntSlice": {
			Type:    "int",
			IsSlice: true,
		},
		"StringSlice": {
			Type:    "string",
			IsSlice: true,
		},
		"PtrSlice": {
			Type:    "int",
			IsSlice: true,
			IsPtr:   true,
		},
		"StringMap": {
			Type:  "map[string]string",
			IsMap: true,
		},
		"IntMap": {
			Type:  "map[int]int",
			IsMap: true,
		},
		"ExtStruct": {
			Type:     "other.PackageStruct",
			IsStruct: true,
		},
		"PtrExtStruct": {
			Type:     "other.PackageStruct",
			IsPtr:    true,
			IsStruct: true,
		},
		"LocalStruct": {
			Type:                "LocalType",
			IsStruct:            true,
			IsSamePackageStruct: true,
		},
		"PtrLocalStruct": {
			Type:                "LocalType",
			IsPtr:               true,
			IsStruct:            true,
			IsSamePackageStruct: true,
		},
	}

	// Парсим все поля
	parsedFields := make(map[string]FieldInfo)
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			fieldInfo := parseField(field)
			fieldInfo.Name = name.Name
			parsedFields[name.Name] = fieldInfo
		}
	}

	// Проверяем каждый ожидаемый тип
	for fieldName, expected := range expectedFields {
		parsed, exists := parsedFields[fieldName]
		if !exists {
			t.Errorf("field %s not found", fieldName)
			continue
		}

		if parsed.Type != expected.Type {
			t.Errorf("field %s: expected type %s, got %s", fieldName, expected.Type, parsed.Type)
		}

		if parsed.IsPtr != expected.IsPtr {
			t.Errorf("field %s: expected IsPtr %v, got %v", fieldName, expected.IsPtr, parsed.IsPtr)
		}

		if parsed.IsSlice != expected.IsSlice {
			t.Errorf("field %s: expected IsSlice %v, got %v", fieldName, expected.IsSlice, parsed.IsSlice)
		}

		if parsed.IsMap != expected.IsMap {
			t.Errorf("field %s: expected IsMap %v, got %v", fieldName, expected.IsMap, parsed.IsMap)
		}

		if parsed.IsStruct != expected.IsStruct {
			t.Errorf("field %s: expected IsStruct %v, got %v", fieldName, expected.IsStruct, parsed.IsStruct)
		}

		if parsed.IsSamePackageStruct != expected.IsSamePackageStruct {
			t.Errorf("field %s: expected IsSamePackageStruct %v, got %v", fieldName, expected.IsSamePackageStruct, parsed.IsSamePackageStruct)
		}
	}
}

// Тесты для generateResetFile
func TestGenerateResetFile(t *testing.T) {
	structs := []StructInfo{
		{
			Name: "SimpleStruct",
			Fields: []FieldInfo{
				{Name: "IntField", Type: "int"},
				{Name: "StringField", Type: "string"},
			},
		},
		{
			Name:   "EmptyStruct",
			Fields: []FieldInfo{},
		},
	}

	generated := generateResetFile("test", structs)

	// Проверяем, что сгенерированный код содержит правильные элементы
	if !strings.Contains(generated, "package test") {
		t.Error("generated code missing package declaration")
	}

	if !strings.Contains(generated, "func (rs *SimpleStruct) Reset()") {
		t.Error("generated code missing SimpleStruct Reset method")
	}

	if !strings.Contains(generated, "func (rs *EmptyStruct) Reset()") {
		t.Error("generated code missing EmptyStruct Reset method")
	}

	// Проверяем обработку nil receiver
	if !strings.Contains(generated, "if rs == nil") {
		t.Error("generated code missing nil check")
	}

	// Проверяем, что для пустой структуры генерируется только nil check
	if !strings.Contains(generated, "func (rs *EmptyStruct) Reset() {\n    if rs == nil {\n        return\n    }\n}") {
		t.Error("empty struct not generated correctly")
	}
}

// Тесты для generateFieldReset с различными комбинациями
func TestGenerateFieldResetCombinations(t *testing.T) {
	tests := []struct {
		name     string
		field    FieldInfo
		expected string
	}{
		{
			name: "pointer to same package struct",
			field: FieldInfo{
				Name:                "field",
				Type:                "MyStruct",
				IsPtr:               true,
				IsStruct:            true,
				IsSamePackageStruct: true,
			},
			expected: "    if rs.field != nil {\n        rs.field.Reset()\n    }\n\n",
		},
		{
			name: "pointer to external struct",
			field: FieldInfo{
				Name:     "field",
				Type:     "external.Struct",
				IsPtr:    true,
				IsStruct: true,
			},
			expected: "    if rs.field != nil {\n        if resetter, ok := rs.field.(interface{ Reset() }); ok {\n            resetter.Reset()\n        } else {\n            *rs.field = external.Struct{}\n        }\n    }\n\n",
		},
		{
			name: "non-pointer struct from same package",
			field: FieldInfo{
				Name:                "field",
				Type:                "MyStruct",
				IsStruct:            true,
				IsSamePackageStruct: true,
			},
			expected: "    rs.field.Reset()\n\n",
		},
		{
			name: "non-pointer external struct",
			field: FieldInfo{
				Name:     "field",
				Type:     "external.Struct",
				IsStruct: true,
			},
			expected: "    if resetter, ok := rs.field.(interface{ Reset() }); ok {\n        resetter.Reset()\n    } else {\n        rs.field = external.Struct{}\n    }\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFieldReset(tt.field)
			if result != tt.expected {
				t.Errorf("generateFieldReset() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Тест для проверки работы с вложенными структурами
func TestNestedStructures(t *testing.T) {
	// Создаем тестовые данные для вложенных структур
	nestedStructs := []StructInfo{
		{
			Name: "Parent",
			Fields: []FieldInfo{
				{
					Name:                "child",
					Type:                "Child",
					IsPtr:               true,
					IsStruct:            true,
					IsSamePackageStruct: true,
				},
			},
		},
		{
			Name: "Child",
			Fields: []FieldInfo{
				{Name: "value", Type: "int"},
			},
		},
	}

	generated := generateResetFile("test", nestedStructs)

	// Проверяем, что обе структуры имеют методы Reset
	if !strings.Contains(generated, "func (rs *Parent) Reset()") {
		t.Error("Parent Reset method not generated")
	}

	if !strings.Contains(generated, "func (rs *Child) Reset()") {
		t.Error("Child Reset method not generated")
	}

	// Проверяем, что Parent вызывает Reset у child
	if !strings.Contains(generated, "rs.child.Reset()") {
		t.Error("Parent does not call child.Reset()")
	}
}
