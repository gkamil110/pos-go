package application

import (
	"fmt"
	"strings"

	"github.com/gkamil110/pos-go/internal/domain"
)

type CourseRepository interface {
	Load() (*domain.Course, error)
	Save(*domain.Course) error
}

type ResultsExporter interface {
	Export(path string, rows []ResultRow) error
}

type ResultRow struct {
	StudentID   string
	StudentName string
	Percentage  float64
}

type EvaluationRow struct {
	CriterionID   string
	CriterionName string
	Score         float64
	Maximum       float64
	Comment       string
}

// Facade объединяет основные действия с оценками и студентами
type Facade struct {
	repository   CourseRepository
	exporter     ResultsExporter
	factory      domain.EvaluationFactory
	scoreService domain.ScoreService
}

func NewFacade(repository CourseRepository, exporter ResultsExporter, scoreService domain.ScoreService) *Facade {
	return &Facade{repository: repository, exporter: exporter, factory: domain.EvaluationFactory{}, scoreService: scoreService}
}

func (f *Facade) Policy() string {
	return f.scoreService.Policy()
}

func (f *Facade) AddStudent(name string) (domain.Student, error) {
	course, err := f.repository.Load()
	if err != nil {
		return domain.Student{}, err
	}

	student, err := course.AddStudent(name)
	if err != nil {
		return domain.Student{}, err
	}
	if err := f.repository.Save(course); err != nil {
		return domain.Student{}, err
	}
	return student, nil
}

func (f *Facade) AddCriterion(name string, maxScore, weight float64) (domain.Criterion, error) {
	course, err := f.repository.Load()
	if err != nil {
		return domain.Criterion{}, err
	}

	criterion, err := course.AddCriterion(name, maxScore, weight)
	if err != nil {
		return domain.Criterion{}, err
	}
	if err := f.repository.Save(course); err != nil {
		return domain.Criterion{}, err
	}
	return criterion, nil
}

func (f *Facade) Students() ([]domain.Student, error) {
	course, err := f.repository.Load()
	if err != nil {
		return nil, err
	}
	return course.Students(), nil
}

func (f *Facade) Criteria() ([]domain.Criterion, error) {
	course, err := f.repository.Load()
	if err != nil {
		return nil, err
	}
	return course.Criteria(), nil
}

func (f *Facade) SetEvaluation(studentID, criterionID string, score float64, comment string) error {
	course, err := f.repository.Load()
	if err != nil {
		return err
	}

	student := domain.StudentID(strings.TrimSpace(studentID))
	criterion := domain.CriterionID(strings.TrimSpace(criterionID))

	if _, err := course.SetEvaluation(f.factory, student, criterion, score, comment); err != nil {
		return err
	}
	return f.repository.Save(course)
}

func (f *Facade) RemoveEvaluation(studentID, criterionID string) error {
	course, err := f.repository.Load()
	if err != nil {
		return err
	}

	student := domain.StudentID(strings.TrimSpace(studentID))
	criterion := domain.CriterionID(strings.TrimSpace(criterionID))

	if err := course.RemoveEvaluation(student, criterion); err != nil {
		return err
	}
	return f.repository.Save(course)
}

func (f *Facade) Evaluations(studentID string) ([]EvaluationRow, error) {
	course, err := f.repository.Load()
	if err != nil {
		return nil, err
	}

	evaluations, err := course.Evaluations(domain.StudentID(strings.TrimSpace(studentID)))
	if err != nil {
		return nil, err
	}

	results := make([]EvaluationRow, 0, len(evaluations))
	for _, evaluation := range evaluations {
		criterion, err := course.Criterion(evaluation.CriterionID())
		if err != nil {
			return nil, fmt.Errorf("ошибка в данных оценок: %w", err)
		}

		results = append(results, EvaluationRow{
			CriterionID: string(criterion.ID()), CriterionName: criterion.Name(),
			Score: evaluation.Score().Value(), Maximum: criterion.MaxScore(), Comment: evaluation.Comment(),
		})
	}
	return results, nil
}

func (f *Facade) Result(studentID string) (ResultRow, error) {
	course, err := f.repository.Load()
	if err != nil {
		return ResultRow{}, err
	}

	student, err := course.Student(domain.StudentID(strings.TrimSpace(studentID)))
	if err != nil {
		return ResultRow{}, err
	}

	percent, err := f.scoreService.Final(course, student.ID())
	if err != nil {
		return ResultRow{}, err
	}
	return ResultRow{StudentID: string(student.ID()), StudentName: student.Name(), Percentage: percent}, nil
}

func (f *Facade) Results() ([]ResultRow, error) {
	course, err := f.repository.Load()
	if err != nil {
		return nil, err
	}

	results := make([]ResultRow, 0, len(course.Students()))
	for _, student := range course.Students() {
		percent, err := f.scoreService.Final(course, student.ID())
		if err != nil {
			return nil, err
		}
		results = append(results, ResultRow{StudentID: string(student.ID()), StudentName: student.Name(), Percentage: percent})
	}

	if len(course.Criteria()) == 0 {
		return nil, domain.ErrEmptyRubric
	}
	return results, nil
}

func (f *Facade) Export(csvFile string) (int, error) {
	results, err := f.Results()
	if err != nil {
		return 0, err
	}
	if err := f.exporter.Export(csvFile, results); err != nil {
		return 0, err
	}
	return len(results), nil
}
