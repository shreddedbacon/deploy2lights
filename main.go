package main

import (
	"fmt"
	"log"
	"time"

	"github.com/stianeikeland/go-rpio"
	"github.com/uselagoon/deploy2lights/internal/lights"
)

func main() {

	ls, err := lights.Setup(ledBrightness, ledCount)
	if err != nil {
		log.Fatal(err)
	}
	// startup animation, once this is complete, builds can start
	ls.Startup()

	fmt.Println("opening gpio")
	err := rpio.Open()
	if err != nil {
		panic(fmt.Sprint("unable to open gpio", err.Error()))
	}

	defer rpio.Close()

	pin := rpio.Pin(17)
	pin.Input()
	pin.PullUp()
	pin.Detect(rpio.FallEdge)

	defer pin.Detect(rpio.NoEdge)

	for {
		if pin.EdgeDetected() {
			fmt.Println("Button")
			ls.Wipe(0x00FF00)
			ls.Wipe(0x00FF88)
			ls.Wipe(0x00FF00)
			ls.Wipe(0x00FF88)
		}
		time.Sleep(time.Second)
	}
}
