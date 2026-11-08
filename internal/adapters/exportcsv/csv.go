package exportcsv

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gkamil110/pos-go/internal/application"
)

type Exporter struct{}

func New() *Exporter {
	return &Exporter{}
}

func (Exporter) Export(csvFile string, results []application.ResultRow) error {
	if csvFile == "" {
		return fmt.Errorf("не указан путь к CSV-файлу")
	}

	if err := os.MkdirAll(filepath.Dir(csvFile), 0755); err != nil {
		return fmt.Errorf("не удалось создать папку: %w", err)
	}

	file, err := os.Create(csvFile)
	if err != nil {
		return fmt.Errorf("не удалось создать CSV-файл: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	headers := []string{"student_id", "student_name", "final_percent"}

	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, result := range results {
		percent := strconv.FormatFloat(result.Percentage, 'f', 2, 64)
		record := []string{result.StudentID, result.StudentName, percent}

		if err := writer.Write(record); err != nil {
			return err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("ошибка записи CSV: %w", err)
	}

	if err := file.Sync(); err != nil {
		return err
	}
	return nil
}
