package button_test

import (
	"github.com/stianeikeland/go-rpio/v4"
	"log"
	"time"
)

func main() {
	if err := rpio.Open(); err != nil {
		log.Fatal(err)
	}
	defer rpio.Close()

	pin := rpio.Pin(24)
	pin.Input()

	for {
		res := pin.Read()
		println("Result on 24: ", res)
		time.Sleep(2000)
	}
}
