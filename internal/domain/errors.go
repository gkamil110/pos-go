package domain

import "errors"

var (
	ErrInvalidName        = errors.New("имя или название не может быть пустым")
	ErrInvalidMaximum     = errors.New("максимальный балл должен быть больше нуля")
	ErrInvalidWeight      = errors.New("вес критерия должен быть больше нуля")
	ErrInvalidScore       = errors.New("балл должен быть от 0 до максимального")
	ErrStudentNotFound    = errors.New("такого студента нет")
	ErrCriterionNotFound  = errors.New("такой критерий не найден")
	ErrEvaluationNotFound = errors.New("оценка не найдена")
	ErrDuplicateCriterion = errors.New("критерий с таким названием уже добавлен")
	ErrEmptyRubric        = errors.New("для расчета итога нужно добавить хотя бы один критерий")
)
