package exportcsv

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gkamil110/pos-go/internal/application"
)

func TestExportCSV(t *testing.T) {
	csvFile := filepath.Join(t.TempDir(), "results", "students.csv")
	results := []application.ResultRow{
		{StudentID: "S1", StudentName: "Петров, Алексей", Percentage: 86.537},
	}

	if err := New().Export(csvFile, results); err != nil {
		t.Fatalf("Не удалось сохранить CSV: %v", err)
	}

	file, err := os.Open(csvFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	actual, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Ошибка чтения CSV: %v", err)
	}

	expected := [][]string{
		{"student_id", "student_name", "final_percent"},
		{"S1", "Петров, Алексей", "86.54"},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Получили %v, ожидали %v", actual, expected)
	}
}
