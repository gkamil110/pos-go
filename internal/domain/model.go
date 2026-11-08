package domain

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type StudentID string
type CriterionID string

// Студент и его уникальный ID
type Student struct {
	id   StudentID
	name string
}

func (s Student) ID() StudentID { return s.id }
func (s Student) Name() string  { return s.name }

// Критерий оценивания
type Criterion struct {
	id       CriterionID
	name     string
	maxScore float64
	weight   float64
}

func (c Criterion) ID() CriterionID   { return c.id }
func (c Criterion) Name() string      { return c.name }
func (c Criterion) MaxScore() float64 { return c.maxScore }
func (c Criterion) Weight() float64   { return c.weight }

// Score хранит проверенный балл
type Score struct{ value float64 }

func NewScore(points, maxScore float64) (Score, error) {
	if !finite(points) || !finite(maxScore) || maxScore <= 0 || points < 0 || points > maxScore {
		return Score{}, ErrInvalidScore
	}
	return Score{value: points}, nil
}

func (s Score) Value() float64 { return s.value }

// Оценка студента по определенному критерию
type Evaluation struct {
	studentID   StudentID
	criterionID CriterionID
	score       Score
	comment     string
}

func (e Evaluation) StudentID() StudentID     { return e.studentID }
func (e Evaluation) CriterionID() CriterionID { return e.criterionID }
func (e Evaluation) Score() Score             { return e.score }
func (e Evaluation) Comment() string          { return e.comment }

// Course хранит студентов, критерии и их оценки
type Course struct {
	students        map[StudentID]Student
	criteria        map[CriterionID]Criterion
	evaluations     map[StudentID]map[CriterionID]Evaluation
	nextStudentID   int
	nextCriterionID int
}

func NewCourse() *Course {
	return &Course{
		students:    make(map[StudentID]Student),
		criteria:    make(map[CriterionID]Criterion),
		evaluations: make(map[StudentID]map[CriterionID]Evaluation),
	}
}

func (c *Course) AddStudent(name string) (Student, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Student{}, ErrInvalidName
	}

	c.nextStudentID++
	student := Student{id: StudentID(fmt.Sprintf("S%d", c.nextStudentID)), name: name}
	c.students[student.id] = student
	return student, nil
}

func (c *Course) AddCriterion(name string, maxScore, weight float64) (Criterion, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Criterion{}, ErrInvalidName
	}
	if !finite(maxScore) || maxScore <= 0 {
		return Criterion{}, ErrInvalidMaximum
	}
	if !finite(weight) || weight <= 0 {
		return Criterion{}, ErrInvalidWeight
	}

	totalWeight := weight
	for _, criterion := range c.criteria {
		if strings.EqualFold(criterion.name, name) {
			return Criterion{}, ErrDuplicateCriterion
		}
		totalWeight += criterion.weight
		if !finite(totalWeight) {
			return Criterion{}, ErrInvalidWeight
		}
	}

	c.nextCriterionID++
	newCriterion := Criterion{
		id:       CriterionID(fmt.Sprintf("C%d", c.nextCriterionID)),
		name:     name,
		maxScore: maxScore,
		weight:   weight,
	}
	c.criteria[newCriterion.id] = newCriterion
	return newCriterion, nil
}

func (c *Course) Student(id StudentID) (Student, error) {
	student, exists := c.students[id]
	if !exists {
		return Student{}, ErrStudentNotFound
	}
	return student, nil
}

func (c *Course) Criterion(id CriterionID) (Criterion, error) {
	criterion, exists := c.criteria[id]
	if !exists {
		return Criterion{}, ErrCriterionNotFound
	}
	return criterion, nil
}

func (c *Course) Students() []Student {
	students := make([]Student, 0, len(c.students))
	for _, student := range c.students {
		students = append(students, student)
	}
	sort.Slice(students, func(i, j int) bool {
		return number(string(students[i].id)) < number(string(students[j].id))
	})
	return students
}

func (c *Course) Criteria() []Criterion {
	criteria := make([]Criterion, 0, len(c.criteria))
	for _, criterion := range c.criteria {
		criteria = append(criteria, criterion)
	}
	sort.Slice(criteria, func(i, j int) bool {
		return number(string(criteria[i].id)) < number(string(criteria[j].id))
	})
	return criteria
}

func (c *Course) SetEvaluation(factory EvaluationFactory, studentID StudentID, criterionID CriterionID, points float64, comment string) (Evaluation, error) {
	evaluation, err := factory.Create(c, studentID, criterionID, points, comment)
	if err != nil {
		return Evaluation{}, err
	}

	if c.evaluations[studentID] == nil {
		c.evaluations[studentID] = make(map[CriterionID]Evaluation)
	}
	c.evaluations[studentID][criterionID] = evaluation
	return evaluation, nil
}

func (c *Course) RemoveEvaluation(studentID StudentID, criterionID CriterionID) error {
	if _, err := c.Student(studentID); err != nil {
		return err
	}
	if _, err := c.Criterion(criterionID); err != nil {
		return err
	}
	if _, exists := c.evaluations[studentID][criterionID]; !exists {
		return ErrEvaluationNotFound
	}

	delete(c.evaluations[studentID], criterionID)
	return nil
}

func (c *Course) Evaluation(studentID StudentID, criterionID CriterionID) (Evaluation, bool) {
	evaluation, exists := c.evaluations[studentID][criterionID]
	return evaluation, exists
}

func (c *Course) Evaluations(studentID StudentID) ([]Evaluation, error) {
	if _, err := c.Student(studentID); err != nil {
		return nil, err
	}

	evaluations := make([]Evaluation, 0, len(c.evaluations[studentID]))
	for _, evaluation := range c.evaluations[studentID] {
		evaluations = append(evaluations, evaluation)
	}
	sort.Slice(evaluations, func(i, j int) bool {
		return number(string(evaluations[i].criterionID)) < number(string(evaluations[j].criterionID))
	})
	return evaluations, nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func number(id string) int {
	num, err := strconv.Atoi(id[1:])
	if err != nil {
		return 0
	}
	return num
}
