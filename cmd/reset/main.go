package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// StructInfo содержит информацию о структуре
type StructInfo struct {
	Name    string
	Fields  []FieldInfo
	Comment string
	PkgName string
	PkgPath string
}

// FieldInfo содержит информацию о поле структуры
type FieldInfo struct {
	Name                string
	Type                string
	IsPtr               bool
	IsSlice             bool
	IsMap               bool
	IsStruct            bool
	IsSamePackageStruct bool
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: reset <directory>")
		os.Exit(1)
	}

	rootDir := os.Args[1]

	// Сканируем все пакеты
	packages := scanPackages(rootDir)

	// Генерируем файлы reset.gen.go для каждого пакета
	for pkgInfo, structs := range packages {
		if len(structs) == 0 {
			continue
		}

		output := generateResetFile(pkgInfo.PkgName, structs)
		outputPath := filepath.Join(rootDir, pkgInfo.PkgPath, "reset.gen.go")

		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			fmt.Printf("Error writing file %s: %v\n", outputPath, err)
			continue
		}

		fmt.Printf("Generated: %s\n", outputPath)
	}
}

// PackageInfo содержит информацию о пакете
type PackageInfo struct {
	PkgName string
	PkgPath string
}

// scanPackages сканирует все пакеты в директории
func scanPackages(rootDir string) map[PackageInfo][]StructInfo {
	packages := make(map[PackageInfo][]StructInfo)

	fmt.Printf("Scanning directory: %s\n", rootDir)

	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		fmt.Printf("  Visiting: %s (dir: %v)\n", path, info.IsDir())
		if err != nil {
			return err
		}

		// Пропускаем директории
		if info.IsDir() {
			// Пропускаем vendor, go.mod, и другие служебные директории
			if info.Name() == "vendor" || info.Name() == "profiles" {
				return filepath.SkipDir
			}
			return nil
		}

		// Обрабатываем только .go файлы
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Пропускаем сгенерированные файлы и тесты
		if strings.Contains(path, "_test.go") || strings.Contains(path, "reset.gen.go") {
			return nil
		}

		// Определяем пакет
		pkgPath := getPackagePath(rootDir, path)
		if pkgPath == "" {
			return nil
		}

		pkgName, structs := parseFile(path)
		if len(structs) > 0 {
			fmt.Printf("Found %d structs in %s (pkg: %s)\n", len(structs), path, pkgName)
		}
		for i := range structs {
			structs[i].PkgPath = pkgPath
		}
		key := PackageInfo{PkgName: pkgName, PkgPath: pkgPath}
		packages[key] = append(packages[key], structs...)

		return nil
	})

	return packages
}

// getPackagePath получает путь к пакету относительно rootDir
func getPackagePath(rootDir, filePath string) string {
	relPath, err := filepath.Rel(rootDir, filePath)
	if err != nil {
		return ""
	}

	// Проверяем, что путь находится внутри rootDir
	if strings.HasPrefix(relPath, "..") {
		return ""
	}

	// Удаляем имя файла
	dir := filepath.Dir(relPath)
	if dir == "." {
		return "."
	}

	// Заменяем разделители на точки
	return strings.ReplaceAll(dir, string(filepath.Separator), "/")
}

// parseFile парсит файл и возвращает структуры с комментарием // generate:reset
func parseFile(filePath string) (string, []StructInfo) {
	fmt.Printf("    Parsing file: %s\n", filePath)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("    Error parsing file: %v\n", err)
		return "", nil
	}

	// Получаем имя пакета из AST
	pkgName := file.Name.Name
	if pkgName == "" {
		pkgName = "main"
	}

	var structs []StructInfo

	// Создаем карту комментариев
	commentMap := make(map[*ast.TypeSpec]*ast.CommentGroup)
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if strings.TrimSpace(comment.Text) == "// generate:reset" {
				// Находим ближайший тип после комментария
				for _, decl := range file.Decls {
					if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
						for _, spec := range genDecl.Specs {
							if typeSpec, ok := spec.(*ast.TypeSpec); ok {
								// Проверяем, что комментарий находится перед типом
								if comment.Pos() < typeSpec.Pos() {
									// Проверяем, что нет других типов между комментарием и типом
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

			fmt.Printf("    Found type: %s\n", typeSpec.Name.Name)

			// Проверяем, есть ли комментарий // generate:reset
			if !hasGenerateResetComment(commentMap[typeSpec]) {
				fmt.Printf("    No reset comment for %s\n", typeSpec.Name.Name)
				continue
			}

			fmt.Printf("    Found reset comment for %s\n", typeSpec.Name.Name)

			// Проверяем, что это структура
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structInfo := StructInfo{
				Name:    typeSpec.Name.Name,
				Comment: typeSpec.Comment.Text(),
			}

			// Обрабатываем поля структуры
			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					fieldInfo := parseField(field)
					fieldInfo.Name = name.Name
					structInfo.Fields = append(structInfo.Fields, fieldInfo)
				}
			}

			structInfo.PkgName = pkgName
			structs = append(structs, structInfo)
		}
	}

	fmt.Printf("    Found %d structs in %s (pkg: %s)\n", len(structs), filePath, pkgName)
	return pkgName, structs
}

// hasGenerateResetComment проверяет наличие комментария // generate:reset
func hasGenerateResetComment(comment *ast.CommentGroup) bool {
	if comment == nil {
		return false
	}

	for _, c := range comment.List {
		if strings.TrimSpace(c.Text) == "// generate:reset" {
			return true
		}
	}

	return false
}

// parseField парсит поле структуры
func parseField(field *ast.Field) FieldInfo {
	info := FieldInfo{}

	// Получаем тип поля
	switch t := field.Type.(type) {
	case *ast.Ident:
		// Простой тип (int, string, bool и т.д.) или структура в том же пакете
		info.Type = t.Name
		info.IsPtr = false
		// Проверяем, является ли тип встроенным (примитивным)
		if !isBuiltinType(t.Name) {
			info.IsStruct = true
			info.IsSamePackageStruct = true
		}
	case *ast.StarExpr:
		// Указатель
		if ident, ok := t.X.(*ast.Ident); ok {
			info.Type = ident.Name
			info.IsPtr = true
			// Проверяем, является ли тип встроенным (примитивным)
			if !isBuiltinType(ident.Name) {
				info.IsStruct = true
				info.IsSamePackageStruct = true
			}
		} else if selExpr, ok := t.X.(*ast.SelectorExpr); ok {
			info.Type = fmt.Sprintf("%s.%s", selExpr.X, selExpr.Sel.Name)
			info.IsPtr = true
			info.IsStruct = true
		}
	case *ast.ArrayType:
		// Слайс
		if ident, ok := t.Elt.(*ast.Ident); ok {
			info.Type = ident.Name
			info.IsSlice = true
		} else if starExpr, ok := t.Elt.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				info.Type = ident.Name
				info.IsSlice = true
				info.IsPtr = true
			}
		}
	case *ast.MapType:
		// Мапа
		info.IsMap = true
		if keyIdent, ok := t.Key.(*ast.Ident); ok {
			if valIdent, ok := t.Value.(*ast.Ident); ok {
				info.Type = fmt.Sprintf("map[%s]%s", keyIdent.Name, valIdent.Name)
			} else {
				info.Type = fmt.Sprintf("map[%s]interface{}", keyIdent.Name)
			}
		}
	case *ast.SelectorExpr:
		// Структура из другого пакета
		info.Type = fmt.Sprintf("%s.%s", t.X, t.Sel.Name)
		info.IsStruct = true
	}

	return info
}

// isBuiltinType проверяет, является ли тип встроенным (примитивным)
func isBuiltinType(typ string) bool {
	switch typ {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "complex64", "complex128", "string", "bool", "byte", "rune", "uintptr", "error":
		return true
	}
	return false
}

// generateResetFile генерирует содержимое файла reset.gen.go
func generateResetFile(pkgName string, structs []StructInfo) string {
	var sb strings.Builder

	sb.WriteString("// Code generated by reset generator. DO NOT EDIT.\n")
	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	for _, s := range structs {
		sb.WriteString(fmt.Sprintf("// Reset сбрасывает состояние структуры %s к начальным значениям\n", s.Name))
		sb.WriteString(fmt.Sprintf("func (rs *%s) Reset() {\n", s.Name))

		if len(s.Fields) == 0 {
			sb.WriteString("    if rs == nil {\n        return\n    }\n")
		} else {
			sb.WriteString("    if rs == nil {\n        return\n    }\n\n")

			for _, f := range s.Fields {
				sb.WriteString(generateFieldReset(f))
			}
		}

		sb.WriteString("}\n\n")
	}

	return sb.String()
}

// generateFieldReset генерирует код сброса для одного поля
func generateFieldReset(f FieldInfo) string {
	var sb strings.Builder

	// Для указателей на структуры из того же пакета (можно вызвать метод напрямую)
	if f.IsPtr && f.IsStruct && f.IsSamePackageStruct {
		sb.WriteString(fmt.Sprintf("    if rs.%s != nil {\n", f.Name))
		sb.WriteString(fmt.Sprintf("        rs.%s.Reset()\n", f.Name))
		sb.WriteString("    }\n\n")
		return sb.String()
	}

	// Для указателей на структуры из других пакетов (используем type assertion)
	if f.IsPtr && f.IsStruct {
		sb.WriteString(fmt.Sprintf("    if rs.%s != nil {\n", f.Name))
		sb.WriteString(fmt.Sprintf("        if resetter, ok := rs.%s.(interface{ Reset() }); ok {\n", f.Name))
		sb.WriteString(fmt.Sprintf("            resetter.Reset()\n"))
		sb.WriteString(fmt.Sprintf("        } else {\n"))
		sb.WriteString(fmt.Sprintf("            *rs.%s = %s{}\n", f.Name, f.Type))
		sb.WriteString(fmt.Sprintf("        }\n"))
		sb.WriteString(fmt.Sprintf("    }\n\n"))
		return sb.String()
	}

	// Для указателей на примитивы
	if f.IsPtr {
		sb.WriteString(fmt.Sprintf("    if rs.%s != nil {\n", f.Name))
		sb.WriteString(fmt.Sprintf("        *rs.%s = %s\n", f.Name, getZeroValue(f.Type)))
		sb.WriteString("    }\n\n")
		return sb.String()
	}

	// Для слайсов
	if f.IsSlice {
		sb.WriteString(fmt.Sprintf("    if rs.%s != nil {\n", f.Name))
		sb.WriteString("        rs." + f.Name + " = rs." + f.Name + "[:0]\n")
		sb.WriteString("    }\n\n")
		return sb.String()
	}

	// Для мап
	if f.IsMap {
		sb.WriteString(fmt.Sprintf("    if rs.%s != nil {\n", f.Name))
		sb.WriteString("        clear(rs." + f.Name + ")\n")
		sb.WriteString("    }\n\n")
		return sb.String()
	}

	// Для вложенных структур с методом Reset (не указатели)
	if f.IsStruct && !f.IsPtr {
		if f.IsSamePackageStruct {
			// Для структур из того же пакета вызываем метод напрямую
			sb.WriteString(fmt.Sprintf("    rs.%s.Reset()\n\n", f.Name))
		} else {
			// Для структур из других пакетов используем type assertion
			sb.WriteString(fmt.Sprintf("    if resetter, ok := rs.%s.(interface{ Reset() }); ok {\n", f.Name))
			sb.WriteString(fmt.Sprintf("        resetter.Reset()\n"))
			sb.WriteString(fmt.Sprintf("    } else {\n"))
			sb.WriteString(fmt.Sprintf("        rs.%s = %s{}\n", f.Name, f.Type))
			sb.WriteString(fmt.Sprintf("    }\n\n"))
		}
		return sb.String()
	}

	// Для примитивов
	sb.WriteString("    rs." + f.Name + " = " + getZeroValue(f.Type) + "\n\n")
	return sb.String()
}

// getZeroValue возвращает нулевое значение для типа
func getZeroValue(typ string) string {
	switch typ {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "complex64", "complex128":
		return "0"
	case "string":
		return `""`
	case "bool":
		return "false"
	default:
		return typ + "{}"
	}
}
