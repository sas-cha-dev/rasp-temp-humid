package rpi

import (
	"BeRoHuTe/internal/sensor"
	"context"
	"fmt"
	"github.com/stianeikeland/go-rpio/v4"
	"log"
	"time"
)

type RealButtonService struct {
	pin      rpio.Pin
	start    bool
	stop     chan bool
	pinState sensor.ButtonState

	onPushFns    []func(state sensor.ButtonState) error
	onReleaseFns []func(state sensor.ButtonState) error
}

func NewRealButtonService(pin int) (sensor.ButtonService, error) {
	if err := rpio.Open(); err != nil {
		return nil, err
	}

	if pin == 0 {
		pin = 24
	}

	rpin := rpio.Pin(pin)
	rpin.Input()

	return &RealButtonService{
		pin:          rpin,
		stop:         make(chan bool),
		pinState:     sensor.ButtonStateUnknown,
		onPushFns:    make([]func(state sensor.ButtonState) error, 0),
		onReleaseFns: make([]func(state sensor.ButtonState) error, 0),
	}, nil
}

func (b RealButtonService) GetCurrentState() (sensor.ButtonState, error) {
	switch b.pin.Read() {
	case rpio.High:
		return sensor.ButtonStateOpen, nil
	case rpio.Low:
		return sensor.ButtonStateClosed, nil
	default:
		return sensor.ButtonStateUnknown, nil
	}
}

func (b RealButtonService) OnPush(fn func(state sensor.ButtonState) error) {
	b.onPushFns = append(b.onPushFns, fn)
}

func (b RealButtonService) OnRelease(fn func(state sensor.ButtonState) error) {
	b.onReleaseFns = append(b.onReleaseFns, fn)
}

func (b RealButtonService) Start(ctx context.Context, dur time.Duration) {
	if b.start == true {
		return
	}

	b.start = true
	ticker := time.NewTicker(dur)

	b.stop = make(chan bool)
	go func() {
		defer ticker.Stop()
		
		select {
		case <-b.stop:
		case <-ctx.Done():
			return
		default:
		}

		for range ticker.C {
			err := b.listenToEdge()
			if err != nil {
				log.Println("ButtonService error: ", err)
			}
		}
	}()
}

func (b RealButtonService) listenToEdge() error {
	var err error

	newState, err := b.GetCurrentState()
	if err != nil {
		return fmt.Errorf("cannot get current state: %v", err)
	}

	if newState == sensor.ButtonStateUnknown {
		return fmt.Errorf("new button state should be either 1 or 0")
	}

	if b.pinState == sensor.ButtonStateClosed && newState == sensor.ButtonStateOpen { // low to high edge = push
		b.executeFns(true, false)
	} else if b.pinState == sensor.ButtonStateOpen && newState == sensor.ButtonStateClosed { // high to low edge = release
		b.executeFns(false, true)
	}

	b.pinState = newState

	return nil
}

func (b RealButtonService) executeFns(push, release bool) {
	if push {
		for _, fn := range b.onPushFns {
			err := fn(sensor.ButtonStateOpen)
			if err != nil {
				log.Println("ButtonService error: ", err)
			}
		}
	}
	if release {
		for _, fn := range b.onReleaseFns {
			err := fn(sensor.ButtonStateClosed)
			if err != nil {
				log.Println("ButtonService error: ", err)
			}
		}
	}
}

func (b RealButtonService) Close() error {
	b.start = false
	return rpio.Close()
}
