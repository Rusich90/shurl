package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

// Тест для функции buildVersionOrDefault
func TestBuildVersionOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "valid version",
			version:  "v1.2.3",
			expected: "v1.2.3",
		},
		{
			name:     "empty version",
			version:  "",
			expected: "N/A",
		},
		{
			name:     "version with commit hash",
			version:  "v1.0.0-abc123",
			expected: "v1.0.0-abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildVersionOrDefault(tt.version)
			if result != tt.expected {
				t.Errorf("buildVersionOrDefault(%q) = %q, want %q", tt.version, result, tt.expected)
			}
		})
	}
}

// Тест для функции buildDateOrDefault
func TestBuildDateOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		expected string
	}{
		{
			name:     "valid date",
			date:     "2024-01-15",
			expected: "2024-01-15",
		},
		{
			name:     "empty date",
			date:     "",
			expected: "N/A",
		},
		{
			name:     "date with time",
			date:     "2024-01-15T10:30:00Z",
			expected: "2024-01-15T10:30:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDateOrDefault(tt.date)
			if result != tt.expected {
				t.Errorf("buildDateOrDefault(%q) = %q, want %q", tt.date, result, tt.expected)
			}
		})
	}
}

// Тест для функции buildCommitOrDefault
func TestBuildCommitOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		commit   string
		expected string
	}{
		{
			name:     "valid commit",
			commit:   "abc123def456",
			expected: "abc123def456",
		},
		{
			name:     "empty commit",
			commit:   "",
			expected: "N/A",
		},
		{
			name:     "full commit hash",
			commit:   "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			expected: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCommitOrDefault(tt.commit)
			if result != tt.expected {
				t.Errorf("buildCommitOrDefault(%q) = %q, want %q", tt.commit, result, tt.expected)
			}
		})
	}
}

// Тест для функции printBuildInfo
func TestPrintBuildInfo(t *testing.T) {
	// Перехватываем вывод через os.Stdout
	// Используем os.Pipe для перехвата
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	// Вызываем функцию
	printBuildInfo()

	// Восстанавливаем stdout
	w.Close()
	os.Stdout = oldStdout

	// Считываем вывод
	var outputBuf bytes.Buffer
	outputBuf.ReadFrom(r)
	r.Close()
	output := outputBuf.String()

	// Проверяем, что вывод содержит ожидаемые заголовки
	expectedOutputs := []string{
		"Build version:",
		"Build date:",
		"Build commit:",
	}

	for _, expected := range expectedOutputs {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("output does not contain expected string %q", expected)
		}
	}

	// Проверяем, что вывод содержит "N/A" для всех полей (так как переменные build пусты)
	if !bytes.Contains([]byte(output), []byte("N/A")) {
		t.Errorf("output should contain 'N/A' for empty build variables")
	}
}

// Тест для проверки вывода printBuildInfo с непустыми значениями
func TestPrintBuildInfoWithValues(t *testing.T) {
	// Временно устанавливаем значения переменных сборки
	originalVersion := buildVersion
	originalDate := buildDate
	originalCommit := buildCommit

	buildVersion = "v2.0.0"
	buildDate = "2024-03-20"
	buildCommit = "deadbeef"

	// Создаем буфер для перехвата вывода
	var outputBuf bytes.Buffer

	// Перехватываем вывод через os.Stdout
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	// Вызываем функцию
	printBuildInfo()

	// Восстанавливаем stdout
	w.Close()
	os.Stdout = oldStdout

	// Считываем вывод
	outputBuf.ReadFrom(r)
	r.Close()
	output := outputBuf.String()

	// Восстанавливаем оригинальные значения
	buildVersion = originalVersion
	buildDate = originalDate
	buildCommit = originalCommit

	// Проверяем, что вывод содержит непустые значения
	expectedOutputs := []string{
		"Build version: v2.0.0",
		"Build date: 2024-03-20",
		"Build commit: deadbeef",
	}

	for _, expected := range expectedOutputs {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("output does not contain expected string %q", expected)
		}
	}
}

// Тест для проверки, что printBuildInfo выводит информацию в правильном формате
func TestPrintBuildInfoFormat(t *testing.T) {
	// Создаем буфер для перехвата вывода
	var outputBuf bytes.Buffer

	// Перехватываем вывод через os.Stdout
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	// Вызываем функцию
	printBuildInfo()

	// Восстанавливаем stdout
	w.Close()
	os.Stdout = oldStdout

	// Считываем вывод
	outputBuf.ReadFrom(r)
	r.Close()
	output := outputBuf.String()

	// Проверяем формат вывода (каждая строка должна заканчиваться \n)
	lines := bytes.Split(bytes.TrimSpace([]byte(output)), []byte("\n"))
	if len(lines) != 3 {
		t.Errorf("expected 3 lines of output, got %d", len(lines))
	}

	// Проверяем, что каждая строка содержит двоеточие и пробел после него
	for i, line := range lines {
		lineStr := string(line)
		if !bytes.Contains(line, []byte(": ")) {
			t.Errorf("line %d does not contain ': ': %q", i+1, lineStr)
		}
	}
}

// Тест для проверки поведения функций с граничными случаями
func TestBuildInfoFunctionsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "whitespace only",
			value:    "   ",
			expected: "   ",
		},
		{
			name:     "special characters",
			value:    "v1.0.0-beta+build.123",
			expected: "v1.0.0-beta+build.123",
		},
		{
			name:     "unicode",
			value:    "версия1.0",
			expected: "версия1.0",
		},
		{
			name:     "very long string",
			value:    fmt.Sprintf("%01000d", 0),
			expected: fmt.Sprintf("%01000d", 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildVersionOrDefault(tt.value)
			if result != tt.expected {
				t.Errorf("buildVersionOrDefault(%q) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}