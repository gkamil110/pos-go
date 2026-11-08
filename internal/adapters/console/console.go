package console

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gkamil110/pos-go/internal/application"
)

type Console struct {
	facade  *application.Facade
	input   *bufio.Reader
	output  io.Writer
	csvFile string
}

func New(facade *application.Facade, in io.Reader, out io.Writer, csvFile string) *Console {
	return &Console{facade: facade, input: bufio.NewReader(in), output: out, csvFile: csvFile}
}

func (c *Console) print(format string, args ...any) {
	fmt.Fprintf(c.output, format, args...)
}

func (c *Console) readText(question string) (string, error) {
	c.print("%s", question)
	text, err := c.input.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	text = strings.TrimSpace(text)
	if errors.Is(err, io.EOF) && text == "" {
		return "", io.EOF
	}
	return text, nil
}

func (c *Console) readNumber(question string) (float64, error) {
	text, err := c.readText(question)
	if err != nil {
		return 0, err
	}
	text = strings.ReplaceAll(text, ",", ".")
	number, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("не получилось прочитать число: %q", text)
	}
	return number, nil
}

func (c *Console) Run() error {
	c.print("ПОС — Преподавательская Оценка Студентов\n")
	c.print("Режим подсчета: %s\n", c.facade.Policy())

	for {
		c.print("\n1. Список студентов       2. Добавить студента\n")
		c.print("3. Список критериев       4. Добавить критерий\n")
		c.print("5. Оценки студента        6. Поставить или изменить оценку\n")
		c.print("7. Удалить оценку         8. Итог студента\n")
		c.print("9. Итоги всех студентов  10. Экспорт в CSV\n")
		c.print("0. Выход\n")

		choice, err := c.readText("Выберите действие: ")
		if errors.Is(err, io.EOF) {
			c.print("\nЗавершение работы.\n")
			return nil
		}
		if err != nil {
			return err
		}
		if choice == "0" {
			c.print("До свидания!\n")
			return nil
		}
		if err := c.handleChoice(choice); err != nil {
			if errors.Is(err, io.EOF) {
				c.print("\nВвод завершен.\n")
				return nil
			}
			c.print("Ошибка: %v\n", err)
		}
	}
}

func (c *Console) handleChoice(choice string) error {
	switch choice {
	case "1":
		return c.printStudents()

	case "2":
		name, err := c.readText("Введите имя студента: ")
		if err != nil {
			return err
		}
		student, err := c.facade.AddStudent(name)
		if err != nil {
			return err
		}
		c.print("Добавлен студент %s (%s)\n", student.Name(), student.ID())

	case "3":
		return c.printCriteria()

	case "4":
		name, err := c.readText("Название критерия: ")
		if err != nil {
			return err
		}
		maxScore, err := c.readNumber("Максимальный балл: ")
		if err != nil {
			return err
		}
		weight, err := c.readNumber("Вес критерия (например 0.4): ")
		if err != nil {
			return err
		}
		criterion, err := c.facade.AddCriterion(name, maxScore, weight)
		if err != nil {
			return err
		}
		c.print("Добавлен критерий %s (%s)\n", criterion.Name(), criterion.ID())

	case "5":
		if err := c.printStudents(); err != nil {
			return err
		}
		studentID, err := c.readText("ID студента: ")
		if err != nil {
			return err
		}
		evaluations, err := c.facade.Evaluations(studentID)
		if err != nil {
			return err
		}
		if len(evaluations) == 0 {
			c.print("Оценок у студента пока нет.\n")
			return nil
		}
		for _, evaluation := range evaluations {
			c.print("%s | %s: %g / %g", evaluation.CriterionID, evaluation.CriterionName, evaluation.Score, evaluation.Maximum)
			if evaluation.Comment != "" {
				c.print(" | %s", evaluation.Comment)
			}
			c.print("\n")
		}

	case "6":
		if err := c.printStudents(); err != nil {
			return err
		}
		if err := c.printCriteria(); err != nil {
			return err
		}
		studentID, err := c.readText("ID студента: ")
		if err != nil {
			return err
		}
		criterionID, err := c.readText("ID критерия: ")
		if err != nil {
			return err
		}
		score, err := c.readNumber("Введите балл: ")
		if err != nil {
			return err
		}
		comment, err := c.readText("Комментарий (можно пустой): ")
		if err != nil {
			return err
		}
		if err := c.facade.SetEvaluation(studentID, criterionID, score, comment); err != nil {
			return err
		}
		c.print("Оценка сохранена. Если она уже была, то обновлена.\n")

	case "7":
		studentID, err := c.readText("ID студента: ")
		if err != nil {
			return err
		}
		criterionID, err := c.readText("ID критерия: ")
		if err != nil {
			return err
		}
		if err := c.facade.RemoveEvaluation(studentID, criterionID); err != nil {
			return err
		}
		c.print("Оценка удалена.\n")

	case "8":
		studentID, err := c.readText("ID студента: ")
		if err != nil {
			return err
		}
		result, err := c.facade.Result(studentID)
		if err != nil {
			return err
		}
		c.print("%s (%s): %.2f%%\n", result.StudentName, result.StudentID, result.Percentage)

	case "9":
		results, err := c.facade.Results()
		if err != nil {
			return err
		}
		if len(results) == 0 {
			c.print("Студентов пока нет.\n")
			return nil
		}
		for _, result := range results {
			c.print("%s | %s: %.2f%%\n", result.StudentID, result.StudentName, result.Percentage)
		}

	case "10":
		fileName, err := c.readText(fmt.Sprintf("Куда сохранить CSV? [по умолчанию %s]: ", c.csvFile))
		if err != nil {
			return err
		}
		if fileName == "" {
			fileName = c.csvFile
		}
		count, err := c.facade.Export(fileName)
		if err != nil {
			return err
		}
		c.print("Экспортировано студентов: %d. Файл: %s\n", count, fileName)

	default:
		c.print("Такого пункта меню нет.\n")
	}
	return nil
}

func (c *Console) printStudents() error {
	students, err := c.facade.Students()
	if err != nil {
		return err
	}
	if len(students) == 0 {
		c.print("Студентов пока нет.\n")
		return nil
	}
	c.print("Студенты:\n")
	for _, student := range students {
		c.print("  %s — %s\n", student.ID(), student.Name())
	}
	return nil
}

func (c *Console) printCriteria() error {
	criteria, err := c.facade.Criteria()
	if err != nil {
		return err
	}
	if len(criteria) == 0 {
		c.print("Критериев пока нет.\n")
		return nil
	}
	c.print("Критерии:\n")
	for _, criterion := range criteria {
		c.print("  %s — %s (макс: %g, вес: %g)\n", criterion.ID(), criterion.Name(), criterion.MaxScore(), criterion.Weight())
	}
	return nil
}
