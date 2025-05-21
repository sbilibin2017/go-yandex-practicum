// Package noexit предоставляет статический анализатор для Go,
// запрещающий прямые вызовы os.Exit в функции main пакета main.
//
// Это позволяет повысить тестируемость и читаемость кода, а также
// избежать неожиданного завершения приложения.
//
// Анализатор предназначен для использования в составе multichecker.
//
// Пример:
//
//	func main() {
//	    if err := run(); err != nil {
//	        fmt.Fprintln(os.Stderr, err)
//	        os.Exit(1) // ❌ Этот вызов будет запрещён анализатором noexit
//	    }
//	}
//
// Чтобы избежать ошибки, перенесите os.Exit из main в другую функцию.
package noexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer проверяет, что в функции main пакета main отсутствует
// прямой вызов os.Exit.
//
// Это помогает соблюдать практику чистого выхода из main через возврат
// кода завершения или делегирование ошибки другим функциям.
var Analyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "forbids direct calls to os.Exit in function main of package main",
	Run:  run,
}

// run выполняет анализ кода для поиска вызова os.Exit в функции main.
//
// Функция проходит по каждому AST-файлу пакета, проверяет, что
// пакет называется "main", а затем ищет вызовы os.Exit в теле функции main.
func run(pass *analysis.Pass) (interface{}, error) {
	// Проверяем, что пакет называется main
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// Проходим по каждому файлу пакета
	for _, file := range pass.Files {
		// Ищем функцию main
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				return true
			}

			// В теле main ищем вызовы os.Exit
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if ident.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "direct call to os.Exit in main function is forbidden")
				}
				return true
			})

			return false // main функция одна — дальше не ищем
		})
	}
	return nil, nil
}
