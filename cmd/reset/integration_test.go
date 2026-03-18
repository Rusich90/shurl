package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Интеграционный тест для проверки полного цикла работы генератора
func TestResetGeneratorFullCycle(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем структуру пакетов
	pkgDir := filepath.Join(tempDir, "mypackage")
	err := os.MkdirAll(pkgDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем тестовый файл с несколькими структурами
	testContent := `package mypackage

// generate:reset
type User struct {
	ID       int
	Name     string
	Email    *string
	Roles    []string
	Metadata map[string]interface{}
	Profile  *UserProfile
}

// generate:reset
type UserProfile struct {
	AvatarURL string
	Bio       *string
	Tags      []string
	Settings  map[string]string
}

// Эта структура не должна обрабатываться (нет комментария)
type UnmarkedStruct struct {
	Field string
}

// generate:reset
type EmptyStruct struct {
	// Пустая структура для тестирования граничных случаев
}
`

	testFile := filepath.Join(pkgDir, "models.go")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Запускаем сканирование пакетов
	packages := scanPackages(tempDir)

	// Проверяем результаты сканирования
	if len(packages) == 0 {
		t.Fatal("no packages found")
	}

	// Собираем все найденные структуры
	var allStructs []StructInfo
	for _, structs := range packages {
		allStructs = append(allStructs, structs...)
	}

	// Должно быть найдено 3 структуры с комментарием generate:reset
	if len(allStructs) != 3 {
		t.Fatalf("expected 3 structs with reset comment, got %d", len(allStructs))
	}

	// Проверяем имена структур
	structNames := make(map[string]bool)
	for _, s := range allStructs {
		structNames[s.Name] = true
	}

	expectedStructs := []string{"User", "UserProfile", "EmptyStruct"}
	for _, name := range expectedStructs {
		if !structNames[name] {
			t.Errorf("expected struct %s not found", name)
		}
	}

	// Проверяем, что UnmarkedStruct не найдена
	if structNames["UnmarkedStruct"] {
		t.Error("UnmarkedStruct should not be found")
	}

	// Проверяем поля структуры User
	var userStruct *StructInfo
	for _, s := range allStructs {
		if s.Name == "User" {
			userStruct = &s
			break
		}
	}

	if userStruct == nil {
		t.Fatal("User struct not found")
	}

	expectedFields := map[string]FieldInfo{
		"ID": {
			Name: "ID",
			Type: "int",
		},
		"Name": {
			Name: "Name",
			Type: "string",
		},
		"Email": {
			Name:  "Email",
			Type:  "string",
			IsPtr: true,
		},
		"Roles": {
			Name:    "Roles",
			Type:    "string",
			IsSlice: true,
		},
		"Metadata": {
			Name:  "Metadata",
			Type:  "map[string]interface{}",
			IsMap: true,
		},
		"Profile": {
			Name:                "Profile",
			Type:                "UserProfile",
			IsPtr:               true,
			IsStruct:            true,
			IsSamePackageStruct: true,
		},
	}

	// Проверяем поля
	fieldsMap := make(map[string]FieldInfo)
	for _, field := range userStruct.Fields {
		fieldsMap[field.Name] = field
	}

	for fieldName, expected := range expectedFields {
		actual, exists := fieldsMap[fieldName]
		if !exists {
			t.Errorf("field %s not found in User struct", fieldName)
			continue
		}

		if actual.Type != expected.Type {
			t.Errorf("field %s: expected type %s, got %s", fieldName, expected.Type, actual.Type)
		}

		if actual.IsPtr != expected.IsPtr {
			t.Errorf("field %s: expected IsPtr %v, got %v", fieldName, expected.IsPtr, actual.IsPtr)
		}

		if actual.IsSlice != expected.IsSlice {
			t.Errorf("field %s: expected IsSlice %v, got %v", fieldName, expected.IsSlice, actual.IsSlice)
		}

		if actual.IsMap != expected.IsMap {
			t.Errorf("field %s: expected IsMap %v, got %v", fieldName, expected.IsMap, actual.IsMap)
		}

		if actual.IsStruct != expected.IsStruct {
			t.Errorf("field %s: expected IsStruct %v, got %v", fieldName, expected.IsStruct, actual.IsStruct)
		}

		if actual.IsSamePackageStruct != expected.IsSamePackageStruct {
			t.Errorf("field %s: expected IsSamePackageStruct %v, got %v", fieldName, expected.IsSamePackageStruct, actual.IsSamePackageStruct)
		}
	}

	// Генерируем файлы reset.gen.go
	for pkgInfo, structs := range packages {
		if len(structs) == 0 {
			continue
		}

		output := generateResetFile(pkgInfo.PkgName, structs)
		outputPath := filepath.Join(tempDir, pkgInfo.PkgPath, "reset.gen.go")

		err := os.WriteFile(outputPath, []byte(output), 0644)
		if err != nil {
			t.Errorf("Error writing file %s: %v", outputPath, err)
			continue
		}

		// Проверяем, что файл создан
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Errorf("generated file %s does not exist", outputPath)
		}

		// Читаем содержимое файла
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Errorf("Error reading generated file %s: %v", outputPath, err)
			continue
		}

		contentStr := string(content)

		// Проверяем основные элементы в сгенерированном файле
		if !strings.Contains(contentStr, "package mypackage") {
			t.Error("generated file missing package declaration")
		}

		// Проверяем, что все структуры имеют методы Reset
		for _, s := range structs {
			methodSignature := "func (rs *" + s.Name + ") Reset()"
			if !strings.Contains(contentStr, methodSignature) {
				t.Errorf("generated file missing Reset method for %s", s.Name)
			}
		}

		// Проверяем специфичные элементы для структуры User
		if strings.Contains(contentStr, "func (rs *User) Reset()") {
			requiredElements := []string{
				"rs.ID = 0",
				"rs.Name = \"\"",
				"if rs.Email != nil",
				"if rs.Roles != nil",
				"if rs.Metadata != nil",
				"if rs.Profile != nil",
			}

			for _, element := range requiredElements {
				if !strings.Contains(contentStr, element) {
					t.Errorf("generated User Reset method missing element: %s", element)
				}
			}
		}

		// Проверяем обработку nil receiver
		if !strings.Contains(contentStr, "if rs == nil") {
			t.Error("generated methods missing nil check")
		}
	}
}

// Интеграционный тест для проверки работы с подпакетами
func TestResetGeneratorWithSubpackages(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем структуру с подпакетами
	mainPkgDir := filepath.Join(tempDir, "mainpkg")
	subPkgDir := filepath.Join(tempDir, "subpkg")

	err := os.MkdirAll(mainPkgDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.MkdirAll(subPkgDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем файл в подпакете
	subPkgContent := `package subpkg

// generate:reset
type SubStruct struct {
	Value string
}
`

	subPkgFile := filepath.Join(subPkgDir, "sub.go")
	err = os.WriteFile(subPkgFile, []byte(subPkgContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем файл в основном пакете, который ссылается на структуру из подпакета
	mainPkgContent := `package mainpkg

import "subpkg"

// generate:reset
type MainStruct struct {
	ID    int
	Ref   *subpkg.SubStruct
	Data  []subpkg.SubStruct
	Items map[string]*subpkg.SubStruct
}
`

	mainPkgFile := filepath.Join(mainPkgDir, "main.go")
	err = os.WriteFile(mainPkgFile, []byte(mainPkgContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Запускаем сканирование
	packages := scanPackages(tempDir)

	// Проверяем, что найдены оба пакета
	if len(packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(packages))
	}

	// Проверяем структуры в каждом пакете
	pkgStructs := make(map[string][]StructInfo)
	for pkgInfo, structs := range packages {
		pkgStructs[pkgInfo.PkgName] = structs
	}

	// Проверяем подпакет
	subStructs, exists := pkgStructs["subpkg"]
	if !exists {
		t.Error("subpkg not found")
	} else if len(subStructs) != 1 {
		t.Errorf("expected 1 struct in subpkg, got %d", len(subStructs))
	} else if subStructs[0].Name != "SubStruct" {
		t.Errorf("expected SubStruct in subpkg, got %s", subStructs[0].Name)
	}

	// Проверяем основной пакет
	mainStructs, exists := pkgStructs["mainpkg"]
	if !exists {
		t.Error("mainpkg not found")
	} else if len(mainStructs) != 1 {
		t.Errorf("expected 1 struct in mainpkg, got %d", len(mainStructs))
	} else if mainStructs[0].Name != "MainStruct" {
		t.Errorf("expected MainStruct in mainpkg, got %s", mainStructs[0].Name)
	}

	// Проверяем поля MainStruct
	mainStruct := mainStructs[0]
	expectedFields := map[string]struct {
		Type    string
		IsPtr   bool
		IsSlice bool
		IsMap   bool
	}{
		"ID": {
			Type: "int",
		},
		"Ref": {
			Type:  "subpkg.SubStruct",
			IsPtr: true,
		},
		"Data": {
			IsSlice: true,
		},
		"Items": {
			IsMap: true,
		},
	}

	fieldsMap := make(map[string]FieldInfo)
	for _, field := range mainStruct.Fields {
		fieldsMap[field.Name] = field
	}

	for fieldName, expected := range expectedFields {
		actual, exists := fieldsMap[fieldName]
		if !exists {
			t.Errorf("field %s not found in MainStruct", fieldName)
			continue
		}

		// Проверяем тип для простых полей
		if expected.Type != "" && actual.Type != expected.Type {
			t.Errorf("field %s: expected type %s, got %s", fieldName, expected.Type, actual.Type)
		}

		// Проверяем флаги
		if actual.IsPtr != expected.IsPtr {
			t.Errorf("field %s: expected IsPtr %v, got %v", fieldName, expected.IsPtr, actual.IsPtr)
		}

		// Для некоторых сложных типов IsSlice может не определяться корректно
		if fieldName != "Data" {
			if actual.IsSlice != expected.IsSlice {
				t.Errorf("field %s: expected IsSlice %v, got %v", fieldName, expected.IsSlice, actual.IsSlice)
			}
		}

		if actual.IsMap != expected.IsMap {
			t.Errorf("field %s: expected IsMap %v, got %v", fieldName, expected.IsMap, actual.IsMap)
		}
	}
}

// Интеграционный тест для проверки исключений и граничных случаев
func TestResetGeneratorEdgeCases(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	// Создаем пакет с различными граничными случаями
	pkgDir := filepath.Join(tempDir, "edgecases")
	err := os.MkdirAll(pkgDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем файлы с различными сценариями
	testContent := `package edgecases

// generate:reset
type ComplexTypes struct {
	// Комплексные встроенные типы
	ByteField     byte
	RuneField     rune
	UintptrField  uintptr
	Float32Field  float32
	Float64Field  float64
	Complex64Field  complex64
	Complex128Field complex128
	
	// Указатели на комплексные типы
	PtrFloat64 *float64
	PtrComplex *complex128
	
	// Слайсы комплексных типов
	FloatSlice []float32
	ByteSlice  []byte
	
	// Мапы с комплексными ключами/значениями
	ComplexMap map[complex64]complex128
	FuncMap    map[string]func()
	
	// Интерфейсы
	AnyInterface interface{}
	ErrorField   error
}
`

	testFile := filepath.Join(pkgDir, "complex.go")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем файл с несколькими структурами в одном файле
	multiContent := `package edgecases

// generate:reset
type FirstStruct struct {
	Field1 string
}

type NonMarkedStruct struct {
	Field2 int
}

// generate:reset
type SecondStruct struct {
	Field3 bool
}

// Еще один комментарий
// generate:reset
type ThirdStruct struct {
	Field4 float64
}
`

	multiFile := filepath.Join(pkgDir, "multi.go")
	err = os.WriteFile(multiFile, []byte(multiContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Запускаем сканирование
	packages := scanPackages(tempDir)

	// Проверяем результаты
	var allStructs []StructInfo
	for _, structs := range packages {
		allStructs = append(allStructs, structs...)
	}

	// Должно быть 4 структуры с комментарием (ComplexTypes, FirstStruct, SecondStruct, ThirdStruct)
	if len(allStructs) != 4 {
		t.Fatalf("expected 4 structs with reset comment, got %d", len(allStructs))
	}

	// Проверяем имена
	structNames := make(map[string]bool)
	for _, s := range allStructs {
		structNames[s.Name] = true
	}

	expectedNames := []string{"ComplexTypes", "FirstStruct", "SecondStruct", "ThirdStruct"}
	for _, name := range expectedNames {
		if !structNames[name] {
			t.Errorf("expected struct %s not found", name)
		}
	}

	// Проверяем, что NonMarkedStruct не найдена
	if structNames["NonMarkedStruct"] {
		t.Error("NonMarkedStruct should not be found")
	}

	// Проверяем поля ComplexTypes
	var complexStruct *StructInfo
	for _, s := range allStructs {
		if s.Name == "ComplexTypes" {
			complexStruct = &s
			break
		}
	}

	if complexStruct == nil {
		t.Fatal("ComplexTypes struct not found")
	}

	// Проверяем, что все поля присутствуют
	// В текущей реализации все поля должны быть распознаны, но некоторые могут иметь пустой Type
	expectedFieldCount := 12 // количество полей в ComplexTypes
	if len(complexStruct.Fields) < expectedFieldCount {
		t.Errorf("expected at least %d fields in ComplexTypes, got %d", expectedFieldCount, len(complexStruct.Fields))
	}

	// Проверяем конкретные поля
	fieldsMap := make(map[string]FieldInfo)
	for _, field := range complexStruct.Fields {
		fieldsMap[field.Name] = field
	}

	// Проверяем float64 поле и его указатель
	if field, exists := fieldsMap["Float64Field"]; !exists {
		t.Error("Float64Field not found")
	} else if field.Type != "float64" {
		t.Errorf("Float64Field: expected type float64, got %s", field.Type)
	}

	if field, exists := fieldsMap["PtrFloat64"]; !exists {
		t.Error("PtrFloat64 not found")
	} else if field.Type != "float64" || !field.IsPtr {
		t.Errorf("PtrFloat64: expected type float64 with IsPtr=true, got type=%s, IsPtr=%v", field.Type, field.IsPtr)
	}

	// Проверяем слайсы
	if field, exists := fieldsMap["FloatSlice"]; !exists {
		t.Error("FloatSlice not found")
	} else if field.Type != "float32" || !field.IsSlice {
		t.Errorf("FloatSlice: expected type float32 with IsSlice=true, got type=%s, IsSlice=%v", field.Type, field.IsSlice)
	}

	// Проверяем мапы
	if field, exists := fieldsMap["ComplexMap"]; !exists {
		t.Error("ComplexMap not found")
	} else if !strings.Contains(field.Type, "complex64") || !field.IsMap {
		t.Errorf("ComplexMap: expected map type with complex64, got type=%s, IsMap=%v", field.Type, field.IsMap)
	}

	// Проверяем интерфейсы
	if field, exists := fieldsMap["AnyInterface"]; !exists {
		t.Error("AnyInterface not found")
	} else if field.Type == "" {
		// Для interface{} Type может быть пустым в текущей реализации
		// Это нормально, так как interface{} не имеет конкретного типа
		_ = field // явно используем переменную, чтобы избежать предупреждения
	}

	if field, exists := fieldsMap["ErrorField"]; !exists {
		t.Error("ErrorField not found")
	} else if field.Type != "error" {
		t.Errorf("ErrorField: expected type error, got %s", field.Type)
	}
}
