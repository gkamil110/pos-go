package console

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gkamil110/pos-go/internal/adapters/exportcsv"
	"github.com/gkamil110/pos-go/internal/adapters/storage"
	"github.com/gkamil110/pos-go/internal/application"
	"github.com/gkamil110/pos-go/internal/domain"
)

func TestInteractiveScenario(t *testing.T) {
	tempDir := t.TempDir()
	jsonFile := filepath.Join(tempDir, "pos.json")
	csvFile := filepath.Join(tempDir, "result.csv")

	repository := storage.NewJSONRepository(jsonFile)
	scoreService := domain.NewScoreService(domain.WeightedStrategy{})
	facade := application.NewFacade(repository, exportcsv.New(), scoreService)

	// Проверяем добавление студента, критерия, оценки и экспорт результатов
	commands := "2\nАнна\n4\nКПО\n10\n1\n6\nS1\nC1\n9\nхорошо\n8\nS1\n10\n\n0\n"
	input := strings.NewReader(commands)

	var output bytes.Buffer

	console := New(facade, input, &output, csvFile)

	if err := console.Run(); err != nil {
		t.Fatalf("Ошибка при запуске консоли: %v", err)
	}

	expectedMessages := []string{
		"Добавлен студент Анна",
		"Добавлен критерий КПО",
		"Анна (S1): 90.00%",
		"Экспортировано студентов: 1",
	}

	for _, message := range expectedMessages {
		if !strings.Contains(output.String(), message) {
			t.Errorf("В выводе нет строки %q", message)
		}
	}

	if _, err := os.Stat(csvFile); err != nil {
		t.Errorf("Файл CSV не был создан: %v", err)
	}
}
