// Пакет для статистического анализа кода.
// Для использования нужно скомипилировать/установить.
// Запустить staticlint <path_to_files>, где path_to_files путь до файлов,
// которые нужно проверить.
// В данный пакет включены анализаторы покета staticcheck.io,
// стандартные статические анализаторы пакета golang.org/x/tools/go/analysis/passes,
// анализатор osexitcheck.
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"staticlint/analisys"
)

// ConfigData описывает структуру конфигурации.
type ConfigData struct {
	StaticCheck []string
}

func main() {
	cfg := ConfigData{StaticCheck: []string{"SA4006", "SA5000", "SA6000", "SA9004"}}
	// определяем изначальные проверки.
	myChecks := []*analysis.Analyzer{
		analisys.OsExitCheckAnalyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
	}
	// определим map подключаемых правил.
	checks := make(map[string]bool)
	for _, v := range cfg.StaticCheck {
		checks[v] = true
	}

	// добавляем в массив нужные проверки.
	for _, v := range staticcheck.Analyzers {
		myChecks = append(myChecks, v.Analyzer)
	}

	multichecker.Main(myChecks...)
}
