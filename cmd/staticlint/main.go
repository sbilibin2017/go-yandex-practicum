// Package main реализует multichecker — консольное приложение для запуска статического анализа Go-кода.
//
// Multichecker объединяет несколько анализаторов:
//   - стандартные анализаторы из пакета golang.org/x/tools/go/analysis/passes;
//   - все анализаторы класса SA из пакета honnef.co/go/tools/staticcheck;
//   - все анализаторы из пакета honnef.co/go/tools/simple;
//   - собственный анализатор noexit, запрещающий вызовы os.Exit в функции main пакета main.
//
// Пример использования:
//
//	go run ./cmd/staticlint ./...
//
// Для подключения собственных анализаторов или отключения существующих, измените содержимое функции main.
package main

import (
	"github.com/sbilibin2017/go-yandex-practicum/cmd/staticlint/analyzers/noexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

// main инициализирует набор анализаторов и запускает multichecker.
//
// В состав анализаторов входят:
//   - SA-анализаторы staticcheck (например, SA1000 – неправильный синтаксис регулярного выражения);
//   - simple-анализаторы (например, использование более простого синтаксиса);
//   - пользовательский анализатор noexit (запрещает прямой вызов os.Exit в main функции).
func main() {
	var analyzers []*analysis.Analyzer

	// Добавляем все SA анализаторы staticcheck
	for _, a := range staticcheck.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	// Добавляем simple анализаторы (упрощения кода)
	for _, a := range simple.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	// Добавляем пользовательский анализатор запрета os.Exit в main
	analyzers = append(analyzers, noexit.Analyzer)

	// Запускаем multichecker с собранным списком анализаторов
	multichecker.Main(analyzers...)
}
