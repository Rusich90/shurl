package checkers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// NewAnalyzer создает новый анализатор, объединяющий все проверки
func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "linter",
		Doc:  "Проверка на использование panic и вызовы log.Fatal/os.Exit вне main",
		Run:  run,
	}
}

func run(pass *analysis.Pass) (any, error) {
	// Проверка на использование panic
	panicChecker(pass)

	// Проверка на log.Fatal и os.Exit вне main
	exitChecker(pass)

	return nil, nil
}

func panicChecker(pass *analysis.Pass) {
	// Проверка вызовов panic
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if call, ok := n.(*ast.CallExpr); ok {
						if fun, ok := call.Fun.(*ast.Ident); ok && fun.Name == "panic" {
							pass.Reportf(call.Pos(), "использование panic в функции %s", fn.Name.Name)
						}
					}
					return true
				})
			}
		}
	}
}

func exitChecker(pass *analysis.Pass) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				// Пропускаем функцию main в пакете main
				if fn.Name.Name == "main" && pass.Pkg.Name() == "main" {
					continue
				}

				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if call, ok := n.(*ast.CallExpr); ok {
						if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
							// Проверка log.Fatal
							if ident, ok := fun.X.(*ast.Ident); ok {
								if ident.Name == "log" && fun.Sel.Name == "Fatal" {
									pass.Reportf(call.Pos(), "вызов log.Fatal вне функции main в пакете main")
								}
							}
							// Проверка os.Exit
							if ident, ok := fun.X.(*ast.Ident); ok {
								if ident.Name == "os" && fun.Sel.Name == "Exit" {
									pass.Reportf(call.Pos(), "вызов os.Exit вне функции main в пакете main")
								}
							}
						}
					}
					return true
				})
			}
		}
	}
}
