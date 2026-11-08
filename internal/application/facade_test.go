package application_test

import (
	"errors"
	"math"
	"path/filepath"
	"testing"

	"github.com/gkamil110/pos-go/internal/adapters/exportcsv"
	"github.com/gkamil110/pos-go/internal/adapters/storage"
	"github.com/gkamil110/pos-go/internal/application"
	"github.com/gkamil110/pos-go/internal/domain"
)

func TestFacadeEndToEnd(t *testing.T) {
	jsonFile := filepath.Join(t.TempDir(), "data.json")
	repository := storage.NewJSONRepository(jsonFile)
	scoreService := domain.NewScoreService(domain.WeightedStrategy{})
	facade := application.NewFacade(repository, exportcsv.New(), scoreService)

	student, err := facade.AddStudent("Анна")
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := facade.AddCriterion("КПО", 10, 2)
	if err != nil {
		t.Fatal(err)
	}

	studentID := string(student.ID())
	criterionID := string(criterion.ID())

	if err := facade.SetEvaluation(studentID, criterionID, 8, "хорошо"); err != nil {
		t.Fatal(err)
	}
	if err := facade.SetEvaluation(studentID, criterionID, 9, "исправлено"); err != nil {
		t.Fatal(err)
	}

	evaluations, err := facade.Evaluations(studentID)
	if err != nil || len(evaluations) != 1 || evaluations[0].Score != 9 {
		t.Fatalf("Ошибка проверки оценок: %+v, %v", evaluations, err)
	}

	result, err := facade.Result(studentID)
	if err != nil || math.Abs(result.Percentage-90) > 1e-8 {
		t.Fatalf("Неправильный итог: %+v, %v", result, err)
	}

	csvFile := filepath.Join(t.TempDir(), "results.csv")
	if count, err := facade.Export(csvFile); err != nil || count != 1 {
		t.Fatalf("Ошибка экспорта: %d, %v", count, err)
	}

	if err := facade.RemoveEvaluation(studentID, criterionID); err != nil {
		t.Fatal(err)
	}

	result, err = facade.Result(studentID)
	if err != nil || result.Percentage != 0 {
		t.Fatalf("Ошибка после удаления оценки: %+v, %v", result, err)
	}

	if err := facade.SetEvaluation(studentID, "C999", 5, ""); !errors.Is(err, domain.ErrCriterionNotFound) {
		t.Fatalf("Несуществующий критерий не вызвал нужную ошибку: %v", err)
	}
}
