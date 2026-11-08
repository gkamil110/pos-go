package domain

import (
	"errors"
	"math"
	"testing"
)

func TestCreateStudentAndCriterion(t *testing.T) {
	course := NewCourse()

	if _, err := course.AddStudent("   "); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("пустое имя студента: %v", err)
	}

	student, err := course.AddStudent("  Анна  ")
	if err != nil || student.Name() != "Анна" || student.ID() != "S1" {
		t.Fatalf("ошибка добавления студента: %+v, %v", student, err)
	}

	if _, err = course.AddCriterion("", 10, 1); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("пустое название критерия: %v", err)
	}

	for _, maxScore := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if _, err = course.AddCriterion("Критерий", maxScore, 1); !errors.Is(err, ErrInvalidMaximum) {
			t.Errorf("неправильный максимум %v: %v", maxScore, err)
		}
	}

	for _, weight := range []float64{0, -2, math.NaN(), math.Inf(-1)} {
		if _, err = course.AddCriterion("Критерий", 10, weight); !errors.Is(err, ErrInvalidWeight) {
			t.Errorf("неправильный вес %v: %v", weight, err)
		}
	}

	criterion, err := course.AddCriterion("  Домашка ", 20, 0.4)
	if err != nil || criterion.ID() != "C1" || criterion.Name() != "Домашка" {
		t.Fatalf("ошибка добавления критерия: %+v, %v", criterion, err)
	}

	if _, err = course.AddCriterion("домашка", 10, 1); !errors.Is(err, ErrDuplicateCriterion) {
		t.Fatalf("повторное название критерия: %v", err)
	}

	if _, err = course.AddCriterion("Очень большой вес", 10, 1e308); err != nil {
		t.Fatalf("ошибка с большим весом: %v", err)
	}
	if _, err = course.AddCriterion("Переполнение веса", 10, 1e308); !errors.Is(err, ErrInvalidWeight) {
		t.Fatalf("переполнение суммы весов: %v", err)
	}
}

func TestEvaluationFactoryAndAggregateInvariants(t *testing.T) {
	course := NewCourse()
	student, _ := course.AddStudent("Анна")
	criterion, _ := course.AddCriterion("Экзамен", 10, 1)
	factory := EvaluationFactory{}

	for _, points := range []float64{-1, 10.001, math.NaN(), math.Inf(1), math.Inf(-1)} {
		_, err := course.SetEvaluation(factory, student.ID(), criterion.ID(), points, "")
		if !errors.Is(err, ErrInvalidScore) {
			t.Errorf("неправильный балл %v: %v", points, err)
		}
	}

	if _, err := course.SetEvaluation(factory, "S999", criterion.ID(), 5, ""); !errors.Is(err, ErrStudentNotFound) {
		t.Fatalf("несуществующий студент: %v", err)
	}
	if _, err := course.SetEvaluation(factory, student.ID(), "C999", 5, ""); !errors.Is(err, ErrCriterionNotFound) {
		t.Fatalf("несуществующий критерий: %v", err)
	}

	if _, err := course.SetEvaluation(factory, student.ID(), criterion.ID(), 10, "  отлично "); err != nil {
		t.Fatal(err)
	}
	if _, err := course.SetEvaluation(factory, student.ID(), criterion.ID(), 7, "исправлено"); err != nil {
		t.Fatal(err)
	}

	evaluations, err := course.Evaluations(student.ID())
	if err != nil || len(evaluations) != 1 || evaluations[0].Score().Value() != 7 || evaluations[0].Comment() != "исправлено" {
		t.Fatalf("оценка не обновилась: %+v, %v", evaluations, err)
	}

	if err := course.RemoveEvaluation(student.ID(), criterion.ID()); err != nil {
		t.Fatal(err)
	}
	if err := course.RemoveEvaluation(student.ID(), criterion.ID()); !errors.Is(err, ErrEvaluationNotFound) {
		t.Fatalf("повторное удаление оценки: %v", err)
	}
}

func TestScoringStrategies(t *testing.T) {
	course := NewCourse()
	anna, _ := course.AddStudent("Анна")
	boris, _ := course.AddStudent("Борис")

	labs, _ := course.AddCriterion("Лабы", 10, 0.5)
	homework, _ := course.AddCriterion("Домашка", 20, 0.3)
	activity, _ := course.AddCriterion("Активность", 5, 0.2)

	factory := EvaluationFactory{}
	_, _ = course.SetEvaluation(factory, anna.ID(), labs.ID(), 9, "")
	_, _ = course.SetEvaluation(factory, anna.ID(), homework.ID(), 18, "")
	_, _ = course.SetEvaluation(factory, anna.ID(), activity.ID(), 5, "")
	_, _ = course.SetEvaluation(factory, boris.ID(), labs.ID(), 7, "")
	_, _ = course.SetEvaluation(factory, boris.ID(), homework.ID(), 10, "")

	weightedService := NewScoreService(WeightedStrategy{})

	for _, test := range []struct {
		id       StudentID
		expected float64
	}{{anna.ID(), 92}, {boris.ID(), 50}} {
		actual, err := weightedService.Final(course, test.id)
		if err != nil || math.Abs(actual-test.expected) > 1e-8 {
			t.Fatalf("итог студента %s: получили %g, ожидали %g, ошибка: %v", test.id, actual, test.expected, err)
		}
	}

	equalService := NewScoreService(EqualStrategy{})
	actual, err := equalService.Final(course, boris.ID())
	expected := 100 * (0.7 + 0.5 + 0) / 3

	if err != nil || math.Abs(actual-expected) > 1e-8 {
		t.Fatalf("итог с равными весами: получили %g, ожидали %g, ошибка: %v", actual, expected, err)
	}

	if _, err := weightedService.Final(course, "S999"); !errors.Is(err, ErrStudentNotFound) {
		t.Fatalf("студент не найден: %v", err)
	}
	if _, err := weightedService.Final(NewCourse(), "S1"); !errors.Is(err, ErrStudentNotFound) {
		t.Fatalf("студента нет в пустом курсе: %v", err)
	}

	emptyCourse := NewCourse()
	student, _ := emptyCourse.AddStudent("Анна")
	if _, err := weightedService.Final(emptyCourse, student.ID()); !errors.Is(err, ErrEmptyRubric) {
		t.Fatalf("нет критериев для расчета: %v", err)
	}
}

func TestSortIDsNumerically(t *testing.T) {
	course := NewCourse()

	for i := 0; i < 12; i++ {
		_, _ = course.AddStudent("Студент")
	}

	students := course.Students()
	if students[1].ID() != "S2" || students[9].ID() != "S10" {
		t.Fatalf("неправильный порядок студентов: %+v", students)
	}
}
