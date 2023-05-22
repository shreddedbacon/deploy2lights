package oled

import (
	"fmt"
	"image"
	"log"
	"strings"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/devices/v3/ssd1306/image1bit"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var (
	TextRow1 fixed.Point26_6 = fixed.P(2, 12)
	TextRow2 fixed.Point26_6 = fixed.P(2, 24)
	TextRow3 fixed.Point26_6 = fixed.P(2, 36)
	TextRow4 fixed.Point26_6 = fixed.P(2, 48)
	TextRow5 fixed.Point26_6 = fixed.P(2, 50)
	TextRow6 fixed.Point26_6 = fixed.P(2, 62)
)

type TextAlign string

const (
	AlignLeft   TextAlign = "left"
	AlignCenter TextAlign = "center"
	AlignRight  TextAlign = "right"
)

var logo = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x80, 0xc0, 0xc0, 0xe0, 0xf0, 0xf0, 0xf8, 0xf8,
	0xfc, 0xfe, 0xfe, 0xff, 0xff, 0x00, 0x00, 0x00, 0xfc, 0xf8, 0xf8, 0xf0, 0xf0, 0xe0, 0xc0, 0xc0,
	0x80, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xf8, 0xf8, 0xfc, 0xfc, 0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xfe, 0xfe, 0xfc, 0xf8, 0xf8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0xe0, 0xe0, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x3f, 0x7f, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00,
	0xc0, 0xe0, 0xf0, 0x70, 0x70, 0x30, 0x70, 0x60, 0xe0, 0xf0, 0xf0, 0x00, 0x00, 0x80, 0xe0, 0xe0,
	0x70, 0x70, 0x30, 0x70, 0x70, 0xe0, 0xf0, 0xf0, 0xf0, 0x00, 0x00, 0xc0, 0xe0, 0xe0, 0x70, 0x70,
	0x30, 0x70, 0x70, 0xe0, 0xe0, 0xc0, 0x00, 0x00, 0x80, 0xc0, 0xe0, 0x70, 0x70, 0x30, 0x30, 0x70,
	0x70, 0xe0, 0xc0, 0x80, 0x00, 0x00, 0xf0, 0xf0, 0xf0, 0x60, 0x30, 0x30, 0x70, 0xf0, 0xe0, 0xc0,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xfe, 0xfc, 0xfc, 0xf8, 0xf8, 0xf1, 0xe1, 0xe3, 0xc3,
	0xc7, 0x8f, 0x8f, 0x1f, 0x1f, 0x3f, 0x3f, 0x7f, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x0f,
	0x3f, 0x7f, 0xf0, 0xe0, 0xc0, 0xc0, 0xc0, 0x60, 0x79, 0xff, 0xff, 0x00, 0x00, 0x1f, 0x7f, 0xff,
	0xe0, 0xc0, 0xc0, 0xc0, 0xe0, 0x70, 0xff, 0xff, 0xff, 0x00, 0x0f, 0x3f, 0x7f, 0xf0, 0xe0, 0xc0,
	0xc0, 0xc0, 0xe0, 0xf0, 0x7f, 0x3f, 0x0f, 0x00, 0x1f, 0x3f, 0x7f, 0xe0, 0xe0, 0xc0, 0xc0, 0xe0,
	0xe0, 0x7f, 0x7f, 0x1f, 0x00, 0x00, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x00, 0x00, 0x7f, 0xff, 0xff,
	0x0f, 0x0f, 0x1f, 0x1f, 0x3f, 0x7f, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0x7f, 0x7f, 0x3e, 0x3e, 0x1c, 0x0c, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0e, 0x1e,
	0x38, 0x30, 0x30, 0x30, 0x30, 0x3c, 0x1f, 0x0f, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x03, 0x07, 0x07, 0x0f, 0x0f,
	0x1f, 0x3f, 0x3f, 0x7f, 0x7f, 0x7f, 0x3f, 0x3f, 0x1f, 0x0f, 0x0f, 0x07, 0x07, 0x03, 0x03, 0x01,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xe0,
	0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xfc, 0xfe,
	0xfe, 0x06, 0x06, 0x06, 0x06, 0x00, 0x00, 0xfc, 0xfc, 0xfe, 0x26, 0x26, 0x3e, 0x3c, 0x38, 0x00,
	0xc0, 0xe2, 0xf6, 0x36, 0x36, 0x36, 0xfe, 0xfc, 0x00, 0x00, 0xfc, 0xfe, 0x86, 0x06, 0x06, 0xff,
	0xff, 0x00, 0x00, 0x02, 0x1e, 0x7e, 0xf0, 0xc0, 0xf8, 0xfe, 0x1e, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x03,
	0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x03, 0x03, 0x03, 0x03, 0x03, 0x00, 0x00,
	0x00, 0x01, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x00, 0x00, 0x00, 0x01, 0x03, 0x03, 0x03, 0x03,
	0x03, 0x00, 0x00, 0x18, 0x18, 0x18, 0x1f, 0x0f, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

type Display struct {
	Opts *ssd1306.Opts
	Dev  *ssd1306.Dev
	Img  *image1bit.VerticalLSB
	Text font.Drawer
}

func NewDisplay(w, h int, port spi.Port, dc, rst gpio.PinIO) *Display {
	rst.In(gpio.PullUp, gpio.BothEdges) // need to pull rst pin high to control display
	opts := ssd1306.Opts{W: w, H: h}    // + rotated etc
	if p, ok := port.(spi.Pins); ok {
		log.Printf("Using pins CLK: %s  MOSI: %s  CS: %s", p.CLK(), p.MOSI(), p.CS())
	}

	dev, err := ssd1306.NewSPI(port, dc, &opts)
	if err != nil {
		log.Fatalf("failed to initialize ssd1306: %v", err)
	}
	log.Printf("Display Bounds: %#+v", dev.Bounds())
	img := image1bit.NewVerticalLSB(dev.Bounds())

	face := basicfont.Face7x13
	text := font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{image1bit.On},
		Face: face,
		Dot:  fixed.P(0, 64), // start bottom left
	}
	return &Display{
		Opts: &opts,
		Dev:  dev,
		Img:  img,
		Text: text,
	}
}

func (d *Display) DrawText(txt string, dot fixed.Point26_6, align TextAlign) {
	d.Text.Dot = dot
	switch align {
	case "center":
		d.Text.DrawString(centerString(txt, len("==================")))
	case "left":
		d.Text.DrawString(padLeft(txt, len("==================")))
	case "right":
		fallthrough
	default:
		d.Text.DrawString(txt)
	}
	if err := d.Dev.Draw(d.Dev.Bounds(), d.Img, image.Point{}); err != nil {
		log.Println(err)
	}
}

func centerString(str string, width int) string {
	spaces := int(float64(width-len(str)) / 2)
	return strings.Repeat(" ", spaces) + str + strings.Repeat(" ", width-(spaces+len(str)))
}

func padLeft(str string, width int) string {
	return fmt.Sprintf(`%*s`, width, str)
}

func (d *Display) PrintLogo() {
	if _, err := d.Dev.Write(logo); err != nil {
		log.Println(err)
	}
}

func (d *Display) Clear() {
	img := image1bit.NewVerticalLSB(d.Dev.Bounds())
	text := font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{image1bit.On},
		Face: basicfont.Face7x13,
		Dot:  fixed.P(0, 64), // start bottom left
	}
	d.Text = text
	d.Img = img
	c := make([]byte, d.Opts.W*d.Opts.H/8)
	if _, err := d.Dev.Write(c); err != nil {
		log.Println(err)
	}
}
