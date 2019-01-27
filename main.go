// +build linux

package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/vp8l"
	"golang.org/x/image/webp"
	"golang.org/x/sys/unix"
)

var (
	charwScale = flag.Float64("w", 1.0, "font width scaling factor")
	charhScale = flag.Float64("h", 1.0, "font height scaling factor")
)

func main() {
	flag.Parse()
	flag.Usage = usage

	if len(flag.Args()) < 1 {
		usage()
	}
	file := flag.Arg(0)
	img, err := os.Open(file)
	if err != nil {
		fatal(err.Error())
	}
	var cfg image.Config
	switch filepath.Ext(file) {
	case ".jpg":
		fallthrough
	case ".jpeg":
		cfg, err = jpeg.DecodeConfig(img)
		break
	case ".png":
		cfg, err = png.DecodeConfig(img)
		break
	case ".gif":
		cfg, err = gif.DecodeConfig(img)
		break
	case ".bmp":
		cfg, err = bmp.DecodeConfig(img)
		break
	case ".tiff":
		cfg, err = tiff.DecodeConfig(img)
		break
	case ".webp":
		cfg, err = webp.DecodeConfig(img)
		break
	default:
		err = errors.New("stub")
		break
	}
	if err != nil {
		// slow path, try to guess
		cfg, _, err = image.DecodeConfig(img)
		if err != nil {
			fatal(err.Error())
		}
	}
	win, err := unix.IoctlGetWinsize(0, unix.TIOCGWINSZ)
	if err != nil {
		fatal(err.Error())
	}

	// just being paranoid about kernel input
	winrow, winpxrow := math.Max(float64(win.Row), 1.0), math.Max(float64(win.Ypixel), 1.0)
	wincol, winpxcol := math.Max(float64(win.Col), 1.0), math.Max(float64(win.Xpixel), 1.0)

	// user input, avoid div by 0
	fontsclw := math.Max(*charwScale, 0.01)
	fontsclh := math.Max(*charhScale, 0.01)

	charw := int(math.Ceil((winpxrow * fontsclw) / winrow))
	charh := int(math.Ceil((winpxcol * fontsclh) / wincol))

	fmt.Printf("%d %d\n", cfg.Width/charw, cfg.Height/charh)
}

// noreturn
func fatal(err string) {
	fmt.Fprintf(os.Stderr, "lukeidraw: %v\n", err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintf(os.Stderr, "lukeidraw: usage: [ flags ] [ file... ]\n")
	flag.PrintDefaults()
	os.Exit(2)
}
