package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/gkamil110/pos-go/internal/domain"
)

type JSONRepository struct {
	filePath string
	mu       sync.Mutex
}

func NewJSONRepository(filePath string) *JSONRepository {
	return &JSONRepository{filePath: filePath}
}

type studentDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type criterionDTO struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	MaxScore float64 `json:"max_score"`
	Weight   float64 `json:"weight"`
}

type evaluationDTO struct {
	StudentID   string  `json:"student_id"`
	CriterionID string  `json:"criterion_id"`
	Score       float64 `json:"score"`
	Comment     string  `json:"comment"`
}

type courseDTO struct {
	Version     int             `json:"version"`
	Students    []studentDTO    `json:"students"`
	Criteria    []criterionDTO  `json:"criteria"`
	Evaluations []evaluationDTO `json:"evaluations"`
}

func (r *JSONRepository) Load() (*domain.Course, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.filePath)
	if errors.Is(err, os.ErrNotExist) {
		return domain.NewCourse(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл %q: %w", r.filePath, err)
	}

	var saved courseDTO
	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, fmt.Errorf("ошибка в JSON %q: %w", r.filePath, err)
	}
	if saved.Version != 1 {
		return nil, fmt.Errorf("неизвестная версия файла: %d", saved.Version)
	}

	course := domain.NewCourse()

	for _, item := range saved.Students {
		student, err := course.AddStudent(item.Name)
		if err != nil {
			return nil, fmt.Errorf("ошибка загрузки студента %q: %w", item.ID, err)
		}
		if string(student.ID()) != item.ID {
			return nil, fmt.Errorf("не совпадает ID студента: %q", item.ID)
		}
	}

	for _, item := range saved.Criteria {
		criterion, err := course.AddCriterion(item.Name, item.MaxScore, item.Weight)
		if err != nil {
			return nil, fmt.Errorf("ошибка загрузки критерия %q: %w", item.ID, err)
		}
		if string(criterion.ID()) != item.ID {
			return nil, fmt.Errorf("не совпадает ID критерия: %q", item.ID)
		}
	}

	factory := domain.EvaluationFactory{}
	used := make(map[string]bool)

	for _, item := range saved.Evaluations {
		key := item.StudentID + "\x00" + item.CriterionID
		if used[key] {
			return nil, fmt.Errorf("повторная оценка для %q и %q", item.StudentID, item.CriterionID)
		}
		used[key] = true

		_, err := course.SetEvaluation(factory, domain.StudentID(item.StudentID), domain.CriterionID(item.CriterionID), item.Score, item.Comment)
		if err != nil {
			return nil, fmt.Errorf("ошибка загрузки оценки %q / %q: %w", item.StudentID, item.CriterionID, err)
		}
	}
	return course, nil
}

func (r *JSONRepository) Save(course *domain.Course) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	saved := courseDTO{Version: 1, Students: []studentDTO{}, Criteria: []criterionDTO{}, Evaluations: []evaluationDTO{}}

	for _, student := range course.Students() {
		saved.Students = append(saved.Students, studentDTO{ID: string(student.ID()), Name: student.Name()})
	}

	for _, criterion := range course.Criteria() {
		saved.Criteria = append(saved.Criteria, criterionDTO{
			ID: string(criterion.ID()), Name: criterion.Name(), MaxScore: criterion.MaxScore(), Weight: criterion.Weight(),
		})
	}

	for _, student := range course.Students() {
		evaluations, err := course.Evaluations(student.ID())
		if err != nil {
			return err
		}
		for _, evaluation := range evaluations {
			saved.Evaluations = append(saved.Evaluations, evaluationDTO{
				StudentID: string(evaluation.StudentID()), CriterionID: string(evaluation.CriterionID()),
				Score: evaluation.Score().Value(), Comment: evaluation.Comment(),
			})
		}
	}

	folder := filepath.Dir(r.filePath)
	if err := os.MkdirAll(folder, 0755); err != nil {
		return fmt.Errorf("не удалось создать папку: %w", err)
	}

	tempFile, err := os.CreateTemp(folder, ".pos-*.tmp")
	if err != nil {
		return fmt.Errorf("ошибка создания временного файла: %w", err)
	}
	defer os.Remove(tempFile.Name())

	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(saved); err != nil {
		tempFile.Close()
		return fmt.Errorf("ошибка записи JSON: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempFile.Name(), r.filePath); err != nil {
		return fmt.Errorf("не удалось заменить файл %q: %w", r.filePath, err)
	}
	return nil
}

type MemoryRepository struct {
	course *domain.Course
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{course: domain.NewCourse()}
}

func (r *MemoryRepository) Load() (*domain.Course, error) {
	return r.course, nil
}

func (r *MemoryRepository) Save(course *domain.Course) error {
	r.course = course
	return nil
}
