package lights

import (
	"strconv"
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

func rgbToColor(r uint8, g uint8, b uint8) uint32 {
	return uint32(uint32(r)<<16 | uint32(g)<<8 | uint32(b))
}

func HexToColor(hex string) uint32 {
	values, err := strconv.ParseUint(string(hex), 16, 32)

	if err != nil {
		return 0
	}

	return rgbToColor(uint8(values>>16), uint8(values&0xFF), uint8((values>>8)&0xFF))

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
}

func (ls *LED) Wipe(color uint32) error {
	for i := 0; i < len(ls.WS.Leds(0)); i++ {
		ls.WS.Leds(0)[i] = color
		if err := ls.WS.Render(); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
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
