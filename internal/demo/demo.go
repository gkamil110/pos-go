package demo

import (
	"fmt"
	"io"

	"github.com/gkamil110/pos-go/internal/application"
)

func Run(facade *application.Facade, output io.Writer, csvFile string) error {
	anna, err := facade.AddStudent("Анна Иванова")
	if err != nil {
		return err
	}
	boris, err := facade.AddStudent("Борис Петров")
	if err != nil {
		return err
	}

	labs, err := facade.AddCriterion("Лабораторные", 10, 0.5)
	if err != nil {
		return err
	}
	homework, err := facade.AddCriterion("Домашние задания", 20, 0.3)
	if err != nil {
		return err
	}
	activity, err := facade.AddCriterion("Активность", 5, 0.2)
	if err != nil {
		return err
	}

	evaluations := []struct {
		studentID   string
		criterionID string
		score       float64
		comment     string
	}{
		{string(anna.ID()), string(labs.ID()), 9, "Отличная работа"},
		{string(anna.ID()), string(homework.ID()), 18, ""},
		{string(anna.ID()), string(activity.ID()), 5, ""},
		{string(boris.ID()), string(labs.ID()), 7, ""},
		{string(boris.ID()), string(homework.ID()), 10, ""},
	}

	for _, evaluation := range evaluations {
		err := facade.SetEvaluation(evaluation.studentID, evaluation.criterionID, evaluation.score, evaluation.comment)
		if err != nil {
			return err
		}
	}

	fmt.Fprintln(output, "Демо: созданы 2 студента, 3 критерия, выставлены 5 оценок.")
	fmt.Fprintln(output, "Отсутствующая оценка по критерию даёт 0 за этот критерий.")

	results, err := facade.Results()
	if err != nil {
		return err
	}
	for _, result := range results {
		fmt.Fprintf(output, "%s (%s): %.2f%%\n", result.StudentName, result.StudentID, result.Percentage)
	}

	count, err := facade.Export(csvFile)
	if err != nil {
		return err
	}
	fmt.Fprintf(output, "CSV экспортирован: %s (строк: %d)\n", csvFile, count)
	return nil
}
