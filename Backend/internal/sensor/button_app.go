package sensor

import (
	"BeRoHuTe/internal/repository"
	"context"
	"fmt"
	"log"
	"time"
)

type ButtonApp struct {
	service ButtonService
	repo    repository.ButtonRepository

	startsAt time.Time
	endsAt   time.Time
}

func NewButtonApp(service ButtonService, repo repository.ButtonRepository) (*ButtonApp, error) {
	return &ButtonApp{
		service: service,
		repo:    repo,

		startsAt: time.Time{},
		endsAt:   time.Time{},
	}, nil
}

func (b *ButtonApp) Start(ctx context.Context) error {
	b.service.OnPush(b.buttonPushed)
	b.service.OnRelease(b.buttonReleased)

	b.service.Start(ctx, 10*time.Second)
	return nil
}

func (b *ButtonApp) buttonPushed(_ ButtonState) error {
	b.startsAt = time.Now()
	log.Printf("button pushed at %v", b.startsAt)
	return nil
}

func (b *ButtonApp) buttonReleased(_ ButtonState) error {
	b.endsAt = time.Now()
	if b.startsAt.IsZero() {
		return fmt.Errorf("button released but never pushed at %v", b.endsAt)
	}
	if b.endsAt.Before(b.startsAt) {
		return fmt.Errorf("button release cannot be before pushing it")
	}
	if b.startsAt.Sub(b.endsAt).Seconds() > 5 {
		return fmt.Errorf("button release too frequent (5 seconds between start and end)")
	}

	err := b.repo.Save(1, b.startsAt, b.endsAt)
	if err != nil {
		return err
	}

	log.Println("Button pushed and released [", b.startsAt, ",", b.endsAt, "]")

	b.startsAt = time.Time{}
	b.endsAt = time.Time{}
	return nil
}
