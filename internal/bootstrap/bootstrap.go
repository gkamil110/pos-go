package bootstrap

import (
	"fmt"
	"io"

	"github.com/gkamil110/pos-go/internal/adapters/console"
	"github.com/gkamil110/pos-go/internal/adapters/exportcsv"
	"github.com/gkamil110/pos-go/internal/adapters/storage"
	"github.com/gkamil110/pos-go/internal/application"
	"github.com/gkamil110/pos-go/internal/domain"
)

// Container хранит основные компоненты приложения
type Container struct {
	Repository application.CourseRepository
	Exporter   application.ResultsExporter
	Facade     *application.Facade
	Console    *console.Console
}

func New(dataFile, csvFile, scoreMode string, input io.Reader, output io.Writer, demoEnabled bool) (*Container, error) {
	var strategy domain.ScoringStrategy

	switch scoreMode {
	case "weighted":
		strategy = domain.WeightedStrategy{}
	case "equal":
		strategy = domain.EqualStrategy{}
	default:
		return nil, fmt.Errorf("неизвестный способ подсчета: %q", scoreMode)
	}

	var repository application.CourseRepository
	if demoEnabled {
		repository = storage.NewMemoryRepository()
	} else {
		repository = storage.NewJSONRepository(dataFile)
	}

	exporter := exportcsv.New()
	scoreService := domain.NewScoreService(strategy)
	facade := application.NewFacade(repository, exporter, scoreService)
	consoleUI := console.New(facade, input, output, csvFile)

	return &Container{
		Repository: repository,
		Exporter:   exporter,
		Facade:     facade,
		Console:    consoleUI,
	}, nil
}
