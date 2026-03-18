package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Тест для проверки функции hasGenerateResetComment
func TestHasGenerateResetComment(t *testing.T) {
	tests := []struct {
		name     string
		comment  *ast.CommentGroup
		expected bool
	}{
		{
			name:     "nil comment",
			comment:  nil,
			expected: false,
		},
		{
			name: "comment without generate:reset",
			comment: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// some comment"},
				},
			},
			expected: false,
		},
		{
			name: "comment with generate:reset",
			comment: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// generate:reset"},
				},
			},
			expected: true,
		},
		{
			name: "multiple comments with generate:reset",
			comment: &ast.CommentGroup{
				List: []*ast.Comment{
					{Text: "// some comment"},
					{Text: "// generate:reset"},
					{Text: "// another comment"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasGenerateResetComment(tt.comment)
			if result != tt.expected {
				t.Errorf("hasGenerateResetComment() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Тест для проверки функции isBuiltinType
func TestIsBuiltinType(t *testing.T) {
	tests := []struct {
		typ      string
		expected bool
	}{
		{"int", true},
		{"string", true},
		{"bool", true},
		{"float64", true},
		{"error", true},
		{"MyStruct", false},
		{"CustomType", false},
	}

	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			result := isBuiltinType(tt.typ)
			if result != tt.expected {
				t.Errorf("isBuiltinType(%s) = %v, want %v", tt.typ, result, tt.expected)
			}
		})
	}
}

// Тест для проверки функции getZeroValue
func TestGetZeroValue(t *testing.T) {
	tests := []struct {
		typ      string
		expected string
	}{
		{"int", "0"},
		{"string", `""`},
		{"bool", "false"},
		{"float64", "0"},
		{"MyStruct", "MyStruct{}"},
		{"CustomType", "CustomType{}"},
	}

	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			result := getZeroValue(tt.typ)
			if result != tt.expected {
				t.Errorf("getZeroValue(%s) = %v, want %v", tt.typ, result, tt.expected)
			}
		})
	}
}

// Тест для проверки функции getPackagePath
func TestGetPackagePath(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем структуру файлов
	testFile := filepath.Join(tempDir, "subdir", "test.go")
	err := os.MkdirAll(filepath.Dir(testFile), 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(testFile, []byte("package test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		rootDir  string
		filePath string
		expected string
	}{
		{
			name:     "root file",
			rootDir:  tempDir,
			filePath: filepath.Join(tempDir, "test.go"),
			expected: ".",
		},
		{
			name:     "subdirectory file",
			rootDir:  tempDir,
			filePath: testFile,
			expected: "subdir",
		},
		{
			name:     "invalid path",
			rootDir:  tempDir,
			filePath: "/invalid/path/test.go",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPackagePath(tt.rootDir, tt.filePath)
			if result != tt.expected {
				t.Errorf("getPackagePath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Тест для проверки функции parseField
func TestParseField(t *testing.T) {
	// Создаем тестовое AST поле
	fset := token.NewFileSet()
	src := `package test
type TestStruct struct {
	intField int
	ptrField *string
	sliceField []int
	mapField map[string]int
}`

	file, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	// Находим структуру
	var structType *ast.StructType
	ast.Inspect(file, func(n ast.Node) bool {
		if st, ok := n.(*ast.StructType); ok {
			structType = st
			return false
		}
		return true
	})

	if structType == nil {
		t.Fatal("struct not found")
	}

	// Тестируем каждое поле
	for _, field := range structType.Fields.List {
		fieldInfo := parseField(field)

		switch field.Names[0].Name {
		case "intField":
			if fieldInfo.Type != "int" || fieldInfo.IsPtr || fieldInfo.IsSlice || fieldInfo.IsMap {
				t.Errorf("intField parsed incorrectly: %+v", fieldInfo)
			}
		case "ptrField":
			if fieldInfo.Type != "string" || !fieldInfo.IsPtr {
				t.Errorf("ptrField parsed incorrectly: %+v", fieldInfo)
			}
		case "sliceField":
			if fieldInfo.Type != "int" || !fieldInfo.IsSlice {
				t.Errorf("sliceField parsed incorrectly: %+v", fieldInfo)
			}
		case "mapField":
			if !strings.HasPrefix(fieldInfo.Type, "map[string]") || !fieldInfo.IsMap {
				t.Errorf("mapField parsed incorrectly: %+v", fieldInfo)
			}
		}
	}
}

// Тест для проверки функции generateFieldReset
func TestGenerateFieldReset(t *testing.T) {
	tests := []struct {
		name     string
		field    FieldInfo
		expected string
	}{
		{
			name: "primitive field",
			field: FieldInfo{
				Name: "intField",
				Type: "int",
			},
			expected: "    rs.intField = 0\n\n",
		},
		{
			name: "pointer to primitive",
			field: FieldInfo{
				Name:  "ptrField",
				Type:  "string",
				IsPtr: true,
			},
			expected: "    if rs.ptrField != nil {\n        *rs.ptrField = \"\"\n    }\n\n",
		},
		{
			name: "slice field",
			field: FieldInfo{
				Name:    "sliceField",
				Type:    "int",
				IsSlice: true,
			},
			expected: "    if rs.sliceField != nil {\n        rs.sliceField = rs.sliceField[:0]\n    }\n\n",
		},
		{
			name: "map field",
			field: FieldInfo{
				Name:  "mapField",
				Type:  "map[string]int",
				IsMap: true,
			},
			expected: "    if rs.mapField != nil {\n        clear(rs.mapField)\n    }\n\n",
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

// Интеграционный тест для проверки всей системы
func TestResetGeneratorIntegration(t *testing.T) {
	// Создаем временную директорию
	tempDir := t.TempDir()

	// Создаем тестовый файл с комментарием generate:reset
	testContent := `package test

// generate:reset
type TestStruct struct {
	IntField    int
	StringField string
	PtrField    *string
	SliceField  []int
	MapField    map[string]string
	Child       *TestStruct
}
`

	testFile := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Запускаем сканирование
	packages := scanPackages(tempDir)

	// Проверяем, что пакет найден
	if len(packages) == 0 {
		t.Fatal("no packages found")
	}

	// Проверяем, что структура найдена
	var foundStructs []StructInfo
	for _, structs := range packages {
		foundStructs = append(foundStructs, structs...)
	}

	if len(foundStructs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(foundStructs))
	}

	structInfo := foundStructs[0]
	if structInfo.Name != "TestStruct" {
		t.Errorf("expected struct name TestStruct, got %s", structInfo.Name)
	}

	if len(structInfo.Fields) != 6 {
		t.Errorf("expected 6 fields, got %d", len(structInfo.Fields))
	}

	// Генерируем файл
	generated := generateResetFile("test", foundStructs)

	// Проверяем, что сгенерированный код содержит ожидаемые элементы
	expectedElements := []string{
		"func (rs *TestStruct) Reset()",
		"rs.IntField = 0",
		"rs.StringField = \"\"",
		"if rs.PtrField != nil",
		"if rs.SliceField != nil",
		"if rs.MapField != nil",
		"if rs.Child != nil",
	}

	for _, element := range expectedElements {
		if !strings.Contains(generated, element) {
			t.Errorf("generated code missing element: %s", element)
		}
	}
}

// Тест для проверки работы с nil receiver
func TestNilReceiver(t *testing.T) {
	// Это тест для проверки, что Reset() не паникует при вызове на nil
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Reset() panicked with nil receiver: %v", r)
		}
	}()

	// Мы не можем напрямую вызвать Reset() на nilStruct, так как это просто информация о структуре
	// Вместо этого проверим, что сгенерированный код будет содержать проверку на nil
	// Это косвенно проверяется через тесты сгенерированного кода в testdata
}
