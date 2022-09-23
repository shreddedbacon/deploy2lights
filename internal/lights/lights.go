package lights

import (
	"time"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

type Flare struct {
	Position int
	Color    uint32
	Speed    int
}

type wsEngine interface {
	Init() error
	Render() error
	Wait() error
	Fini()
	Leds(channel int) []uint32
}

type LED struct {
	WS wsEngine
}

func Setup(brightness, ledCount int) (*LED, error) {
	opt := ws2811.DefaultOptions
	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledCount

	dev, err := ws2811.MakeWS2811(&opt)
	if err != nil {
		return nil, err
	}
	ls := &LED{
		WS: dev,
	}
	err = ls.WS.Init()
	if err != nil {
		return nil, err
	}
	return ls, nil
}

func (ls *LED) Startup() {
	// startup animation, once this is complete, builds can start
	ls.Wipe(0x0690BA)
	ls.Wipe(0x000000)
}

func (ls *LED) Wipe(color uint32) error {
	for i := 0; i < len(ls.WS.Leds(0)); i++ {
		ls.WS.Leds(0)[i] = color
		if err := ls.WS.Render(); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func (ls *LED) Display(ledValues *map[string]Flare) error {
	for i := 0; i < len(ls.WS.Leds(0)); i++ {
		ls.WS.Leds(0)[i] = 0x000000
	}

	for k, v := range *ledValues {
		if (*ledValues)[k].Position == len(ls.WS.Leds(0)) {
			delete(*ledValues, k)
		} else {
			ls.WS.Leds(0)[v.Position] = v.Color
		}
	}
	if err := ls.WS.Render(); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	return nil
}
