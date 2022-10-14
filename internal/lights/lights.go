package lights

import (
	"strconv"
	"sync"
	"time"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

type Flare struct {
	Position int
	Color    uint32
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

// helper function to convert a rgb uint8s to a uint32 for colors
func rgbToColor(r uint8, g uint8, b uint8) uint32 {
	return uint32(uint32(r)<<16 | uint32(g)<<8 | uint32(b))
}

// helper function to convert a hex string to a uint32 for colors
func HexToColor(hex string) uint32 {
	values, err := strconv.ParseUint(string(hex), 16, 32)
	if err != nil {
		return 0
	}
	return rgbToColor(uint8(values>>16), uint8((values>>8)&0xFF), uint8(values&0xFF))

}

func Setup(brightness, ledCount int, stripType string) (*LED, error) {
	opt := ws2811.DefaultOptions
	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledCount

	// check the LED strip type to determine the correct RGB order for the strip
	switch stripType {
	case "RBG":
		opt.Channels[0].StripeType = ws2811.WS2811StripRBG
	case "GRB":
		opt.Channels[0].StripeType = ws2811.WS2811StripGRB
	case "GBR":
		opt.Channels[0].StripeType = ws2811.WS2811StripGBR
	case "BRG":
		opt.Channels[0].StripeType = ws2811.WS2811StripBRG
	case "BGR":
		opt.Channels[0].StripeType = ws2811.WS2811StripBGR
	default:
		opt.Channels[0].StripeType = ws2811.WS2811StripRGB
	}

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
	// startup animation, verify all leds work
	ls.Display(HexToColor("06BA90"))
	time.Sleep(2 * time.Second)
	ls.Display(HexToColor("000000"))
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

func (ls *LED) Display(color uint32) error {
	for i := 0; i < len(ls.WS.Leds(0)); i++ {
		ls.WS.Leds(0)[i] = color
	}
	if err := ls.WS.Render(); err != nil {
		return err
	}
	return nil
}

func (ls *LED) DisplayFlares(delayms int, ledValues *sync.Map) error {
	// clear out the led strip to start
	for i := 0; i < len(ls.WS.Leds(0)); i++ {
		ls.WS.Leds(0)[i] = HexToColor("000000")
	}

	// iterate over the flare map to set the lights up
	ledValues.Range(func(key, value interface{}) bool {
		if value.(Flare).Position == len(ls.WS.Leds(0)) {
			ls.WS.Leds(0)[value.(Flare).Position-1] = value.(Flare).Color
			ls.WS.Leds(0)[value.(Flare).Position-2] = value.(Flare).Color
			ls.WS.Leds(0)[value.(Flare).Position-3] = value.(Flare).Color
			ledValues.Delete(key)
		} else if value.(Flare).Position < len(ls.WS.Leds(0)) {
			ls.WS.Leds(0)[value.(Flare).Position] = value.(Flare).Color
			if value.(Flare).Position > 0 {
				ls.WS.Leds(0)[value.(Flare).Position-1] = value.(Flare).Color
			}
			if value.(Flare).Position > 1 {
				ls.WS.Leds(0)[value.(Flare).Position-2] = value.(Flare).Color
			}
			if value.(Flare).Position > 2 {
				ls.WS.Leds(0)[value.(Flare).Position-3] = value.(Flare).Color
			}
		}
		return true
	})
	// render the light strip
	if err := ls.WS.Render(); err != nil {
		return err
	}
	// once the strip has rendered, increment all flares by 1
	ledValues.Range(func(key, value interface{}) bool {
		ledValues.Store(key, Flare{
			Position: value.(Flare).Position + 1,
			Color:    value.(Flare).Color,
		})
		return true
	})

	// add a delay if required (shorter led strips will move flares faster than longer ones will)
	time.Sleep(time.Duration(delayms) * time.Millisecond)
	return nil
}
