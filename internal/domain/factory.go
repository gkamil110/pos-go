package domain

import "strings"

// EvaluationFactory создает оценки и проверяет их правильность
type EvaluationFactory struct{}

func (EvaluationFactory) Create(course *Course, studentID StudentID, criterionID CriterionID, points float64, comment string) (Evaluation, error) {
	if _, err := course.Student(studentID); err != nil {
		return Evaluation{}, err
	}

	criterion, err := course.Criterion(criterionID)
	if err != nil {
		return Evaluation{}, err
	}

	score, err := NewScore(points, criterion.maxScore)
	if err != nil {
		return Evaluation{}, err
	}

	comment = strings.TrimSpace(comment)
	return Evaluation{studentID: studentID, criterionID: criterionID, score: score, comment: comment}, nil
}
