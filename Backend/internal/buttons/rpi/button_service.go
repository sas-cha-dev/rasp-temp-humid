package rpi

import (
	"BeRoHuTe/internal/buttons"
	"context"
	"fmt"
	"github.com/stianeikeland/go-rpio/v4"
	"log"
	"time"
)

// IMPORTANT must be in a separate package, so we can compile on a dev environment not having support on go-rpio library

type ButtonService struct {
	pin      rpio.Pin
	start    bool
	stop     chan bool
	pinState buttons.ButtonState

	onPushFns    []func(state buttons.ButtonState) error
	onReleaseFns []func(state buttons.ButtonState) error
}

func NewButtonService(pin int) (buttons.Service, error) {
	if err := rpio.Open(); err != nil {
		return nil, err
	}

	if pin == 0 {
		pin = 24
	}

	rpin := rpio.Pin(pin)
	rpin.Input()

	return &ButtonService{
		pin:          rpin,
		stop:         make(chan bool),
		pinState:     buttons.ButtonStateUnknown,
		onPushFns:    make([]func(state buttons.ButtonState) error, 0),
		onReleaseFns: make([]func(state buttons.ButtonState) error, 0),
	}, nil
}

func (b *ButtonService) GetCurrentState() (buttons.ButtonState, error) {
	switch b.pin.Read() {
	case rpio.High:
		return buttons.ButtonStateOpen, nil
	case rpio.Low:
		return buttons.ButtonStateClosed, nil
	default:
		return buttons.ButtonStateUnknown, nil
	}
}

func (b *ButtonService) OnPush(fn func(state buttons.ButtonState) error) {
	b.onPushFns = append(b.onPushFns, fn)
}

func (b *ButtonService) OnRelease(fn func(state buttons.ButtonState) error) {
	b.onReleaseFns = append(b.onReleaseFns, fn)
}

func (b *ButtonService) Start(ctx context.Context, dur time.Duration) {
	if b.start == true {
		return
	}

	b.start = true
	ticker := time.NewTicker(dur)

	b.stop = make(chan bool)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-b.stop:
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := b.listenToEdge()
				if err != nil {
					log.Println("ButtonService error: ", err)
				}
			default:
			}
		}
	}()
	return
}

func (b *ButtonService) listenToEdge() error {
	var err error

	newState, err := b.GetCurrentState()
	if err != nil {
		return fmt.Errorf("cannot get current state: %v", err)
	}

	if newState == buttons.ButtonStateUnknown {
		return fmt.Errorf("new button state should be either 1 or 0")
	}

	if b.pinState == buttons.ButtonStateClosed && newState == buttons.ButtonStateOpen { // low to high edge = push
		b.executeFns(true, false)
	} else if b.pinState == buttons.ButtonStateOpen && newState == buttons.ButtonStateClosed { // high to low edge = release
		b.executeFns(false, true)
	}

	b.pinState = newState

	return nil
}

func (b *ButtonService) executeFns(push, release bool) {
	if push {
		for _, fn := range b.onPushFns {
			err := fn(buttons.ButtonStateOpen)
			if err != nil {
				log.Println("ButtonService error: ", err)
			}
		}
	}
	if release {
		for _, fn := range b.onReleaseFns {
			err := fn(buttons.ButtonStateClosed)
			if err != nil {
				log.Println("ButtonService error: ", err)
			}
		}
	}
}

func (b *ButtonService) Stop() error {
	b.onPushFns = make([]func(state buttons.ButtonState) error, 0)
	b.onReleaseFns = make([]func(state buttons.ButtonState) error, 0)
	b.start = false
	return rpio.Close()
}
