package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gkamil110/pos-go/internal/domain"
)

func TestJSONRoundTripAndValidation(t *testing.T) {
	jsonFile := filepath.Join(t.TempDir(), "nested", "pos.json")
	repository := NewJSONRepository(jsonFile)

	course, err := repository.Load()
	if err != nil {
		t.Fatal(err)
	}

	student, err := course.AddStudent("Анна")
	if err != nil {
		t.Fatal(err)
	}
	criterion, err := course.AddCriterion("КПО", 10, 1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = course.SetEvaluation(domain.EvaluationFactory{}, student.ID(), criterion.ID(), 8.5, "Первое ДЗ")
	if err != nil {
		t.Fatal(err)
	}

	if err := repository.Save(course); err != nil {
		t.Fatal(err)
	}

	loadedCourse, err := repository.Load()
	if err != nil {
		t.Fatal(err)
	}

	evaluations, err := loadedCourse.Evaluations(student.ID())
	if err != nil || len(evaluations) != 1 || evaluations[0].Score().Value() != 8.5 || evaluations[0].Comment() != "Первое ДЗ" {
		t.Fatalf("Оценка неправильно загрузилась: %v, %v", evaluations, err)
	}

	if newStudent, err := loadedCourse.AddStudent("Борис"); err != nil || newStudent.ID() != "S2" {
		t.Fatalf("Ошибка с ID после загрузки: %v, %v", newStudent, err)
	}
}

func TestCorruptStorageIsNotSilentlyReplaced(t *testing.T) {
	jsonFile := filepath.Join(t.TempDir(), "corrupt.json")

	if err := os.WriteFile(jsonFile, []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewJSONRepository(jsonFile).Load(); err == nil {
		t.Fatal("Поврежденный JSON не вызвал ошибку")
	}

	invalidData := `{"version":1,"students":[{"id":"S1","name":"Анна"}],"criteria":[{"id":"C1","name":"КПО","max_score":10,"weight":1}],"evaluations":[{"student_id":"S1","criterion_id":"C1","score":100}]}`

	if err := os.WriteFile(jsonFile, []byte(invalidData), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewJSONRepository(jsonFile).Load(); !errors.Is(err, domain.ErrInvalidScore) {
		t.Fatalf("Не обнаружена неправильная оценка: %v", err)
	}
}
