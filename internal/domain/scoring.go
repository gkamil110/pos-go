package domain

// Интерфейс для разных способов подсчета итоговой оценки
type ScoringStrategy interface {
	Name() string
	Calculate(criteria []Criterion, evaluations map[CriterionID]Evaluation) float64
}

// Подсчет оценки с учетом весов критериев
type WeightedStrategy struct{}

func (WeightedStrategy) Name() string { return "weighted" }

func (WeightedStrategy) Calculate(criteria []Criterion, evaluations map[CriterionID]Evaluation) float64 {
	var weightedSum, totalWeight float64

	for _, criterion := range criteria {
		totalWeight += criterion.weight
		if evaluation, exists := evaluations[criterion.id]; exists {
			weightedSum += criterion.weight * evaluation.score.value / criterion.maxScore
		}
	}
	return 100 * weightedSum / totalWeight
}

// Подсчет оценки, когда все критерии одинаково важны
type EqualStrategy struct{}

func (EqualStrategy) Name() string { return "equal" }

func (EqualStrategy) Calculate(criteria []Criterion, evaluations map[CriterionID]Evaluation) float64 {
	var totalScore float64

	for _, criterion := range criteria {
		if evaluation, exists := evaluations[criterion.id]; exists {
			totalScore += evaluation.score.value / criterion.maxScore
		}
	}
	return 100 * totalScore / float64(len(criteria))
}

// ScoreService рассчитывает итог студента
type ScoreService struct {
	strategy ScoringStrategy
}

func NewScoreService(strategy ScoringStrategy) ScoreService {
	return ScoreService{strategy: strategy}
}

func (s ScoreService) Policy() string { return s.strategy.Name() }

func (s ScoreService) Final(course *Course, studentID StudentID) (float64, error) {
	if _, err := course.Student(studentID); err != nil {
		return 0, err
	}

	criteria := course.Criteria()
	if len(criteria) == 0 {
		return 0, ErrEmptyRubric
	}

	return s.strategy.Calculate(criteria, course.evaluations[studentID]), nil
}
