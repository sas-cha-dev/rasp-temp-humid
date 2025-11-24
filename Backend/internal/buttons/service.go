package buttons

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

type ButtonState string

const (
	ButtonStateUnknown ButtonState = "unknown"
	ButtonStateOpen    ButtonState = "open"
	ButtonStateClosed  ButtonState = "closed"
)

var allBtnStates = []ButtonState{ButtonStateOpen, ButtonStateClosed}

type Service interface {
	Start(ctx context.Context, dur time.Duration)
	GetCurrentState() (ButtonState, error)
	OnPush(fn func(state ButtonState) error)
	OnRelease(fn func(state ButtonState) error)
	Close() error
}

type DummyService struct {
	gpioPin      int
	onPush       []func(state ButtonState) error
	onRelease    []func(state ButtonState) error
	currentState ButtonState

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewDummyService(gpioPin int) Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &DummyService{
		gpioPin:      gpioPin,
		onPush:       make([]func(state ButtonState) error, 0),
		onRelease:    make([]func(state ButtonState) error, 0),
		currentState: ButtonStateUnknown,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start begins listening to button state changes
func (s *DummyService) Start(ctx context.Context, dur time.Duration) {
	s.wg.Add(1)
	go s.listenToButton(dur)
}

func (s *DummyService) listenToButton(rate time.Duration) {
	defer s.wg.Done()

	ticker := time.NewTicker(rate)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			newState, err := s.readState()
			if err != nil {
				continue
			}

			s.mu.Lock()
			previousState := s.currentState
			s.currentState = newState
			s.mu.Unlock()

			// Detect state changes and trigger callbacks
			if previousState != newState {
				s.handleStateChange(previousState, newState)
			}
		}
	}
}

func (s *DummyService) readState() (ButtonState, error) {
	// Simulate reading from GPIO pin
	ran := rand.IntN(2)
	return allBtnStates[ran], nil
}

func (s *DummyService) handleStateChange(oldState, newState ButtonState) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Button was pushed (transitioned to closed)
	if oldState == ButtonStateOpen && newState == ButtonStateClosed {
		for _, fn := range s.onPush {
			if err := fn(newState); err != nil {
				fmt.Printf("Error in onPush callback: %v\n", err)
			}
		}
	}

	// Button was released (transitioned to open)
	if oldState == ButtonStateClosed && newState == ButtonStateOpen {
		for _, fn := range s.onRelease {
			if err := fn(newState); err != nil {
				fmt.Printf("Error in onRelease callback: %v\n", err)
			}
		}
	}
}

func (s *DummyService) Close() error {
	s.cancel()  // Signal goroutine to stop
	s.wg.Wait() // Wait for goroutine to finish
	return nil
}

func (s *DummyService) GetCurrentState() (ButtonState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.currentState, nil
}

func (s *DummyService) OnPush(fn func(state ButtonState) error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.onPush = append(s.onPush, fn)
	return
}

func (s *DummyService) OnRelease(fn func(state ButtonState) error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.onRelease = append(s.onRelease, fn)
	return
}
