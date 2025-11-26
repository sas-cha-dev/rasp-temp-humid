package data_clean

import (
	"BeRoHuTe/internal/contracts"
	"context"
	"fmt"
	"time"
)

type ButtonRepository interface {
	GetAll(offset int, limit int) ([]*contracts.ButtonReading, error)
}

type SensorRepository interface {
	GetInBetween(start time.Time, end time.Time) ([]*contracts.SensorReading, error)
	Delete(id int64) error
}

type AppOption func(*App) error

func WithBeforeCleanUp(fn func() error) AppOption {
	return func(app *App) error {
		app.beforeCleanUp = fn
		return nil
	}
}

func WithAfterCleanUp(fn func() error) AppOption {
	return func(app *App) error {
		app.afterCleanUp = fn
		return nil
	}
}

type App struct {
	btnRepo    ButtonRepository
	sensorRepo SensorRepository
	start      bool
	stop       chan bool

	beforeCleanUp func() error
	afterCleanUp  func() error
}

func NewApp(btnRepo ButtonRepository, sensorRepo SensorRepository, options ...AppOption) (*App, error) {
	app := &App{
		btnRepo:    btnRepo,
		sensorRepo: sensorRepo,
	}

	for _, option := range options {
		err := option(app)
		if err != nil {
			return nil, err
		}
	}

	return app, nil
}

func (a *App) Start(ctx context.Context, interval time.Duration, execDirectly bool) {
	if a.start {
		return
	}

	a.start = true
	a.stop = make(chan bool)

	startInterval := interval
	if execDirectly {
		startInterval = time.Millisecond
	}
	ticker := time.NewTicker(startInterval)

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			select {
			case <-ctx.Done():
			case <-a.stop:
				return
			default:

			}

			err := a.cleanUp()
			if err != nil {
				fmt.Printf("Error cleaning up: %v\n", err)
			}

			ticker.Reset(interval)
		}
	}()
}

func (a *App) cleanUp() error {
	if a.beforeCleanUp != nil {
		err := a.beforeCleanUp()
		if err != nil {
			return fmt.Errorf("aborted cleanup caused by error in beforeCleanUp(): %v", err)
		}
	}

	if err := a.dataCleanUp(); err != nil {
		return fmt.Errorf("aborted cleanup caused by error in dataCleanUp(): %v", err)
	}

	if a.afterCleanUp != nil {
		err := a.afterCleanUp()
		if err != nil {
			return fmt.Errorf("aborted cleanup caused by error in afterCleanUp(): %v", err)
		}
	}

	return nil
}

func (a *App) dataCleanUp() error {
	perPage := 10 // add pagination to reduce memory usage
	currPage := 1

	counter := 0

	fmt.Println("[DataCleanUp] Start")
	for {
		allButtonReadings, err := a.btnRepo.GetAll(
			(currPage-1)*perPage,
			perPage,
		)
		if err != nil {
			return err
		}

		if len(allButtonReadings) == 0 {
			break
		}

		for _, reading := range allButtonReadings {
			count, err := a.cleanUpForSensor(reading)
			if err != nil {
				return err
			}
			counter += count
		}

		currPage++
	}

	fmt.Printf("[DataCleanUp] Cleaned %d sensor entries\n", counter)
	fmt.Println("[DataCleanUp] END")

	return nil
}

func (a *App) cleanUpForSensor(reading *contracts.ButtonReading) (int, error) {
	// get all sensor readings in between [buttonStart, buttonEnd+10min] and delete them
	sensorEndTime := reading.EndAt.Add(time.Minute * 10)

	allSensorData, err := a.sensorRepo.GetInBetween(reading.StartAt, sensorEndTime)
	if err != nil {
		return 0, err
	}

	counter := 0
	for _, sensor := range allSensorData {
		if err := a.sensorRepo.Delete(sensor.ID); err != nil {
			return counter, err
		}
		counter++
	}

	return counter, nil
}

func (a *App) Stop() {
	a.stop <- true
}
