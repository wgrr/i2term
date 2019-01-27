// +build linux

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	_ "golang.org/x/image/vp8l"
	"golang.org/x/image/webp"
	"golang.org/x/sys/unix"
)

func i2term(img io.Reader, name string, wscale, hscale float64) (row, col int, err error) {
	var cfg image.Config
	switch filepath.Ext(name) {
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
		err = errors.New("dummy")
		break
	}
	if err != nil {
		// slow path, try to guess
		cfg, _, err = image.DecodeConfig(img)
		if err != nil {
			return 0, 0, err
		}
	}

	win, err := unix.IoctlGetWinsize(0, unix.TIOCGWINSZ)
	if err != nil {
		// it might be a pipe writing to our stdin, so try stdout
		win, err = unix.IoctlGetWinsize(1, unix.TIOCGWINSZ)
		if err != nil {
			// so maybe process substution? try stderr
			win, err = unix.IoctlGetWinsize(2, unix.TIOCGWINSZ)
			if err != nil {
				// we tried out best
				return 0, 0, err
			}
		}
	}

	// just being paranoid about kernel input
	winrow, winpxrow := math.Max(float64(win.Row), 1.0), math.Max(float64(win.Ypixel-2), 1.0)
	wincol, winpxcol := math.Max(float64(win.Col), 1.0), math.Max(float64(win.Xpixel-2), 1.0)

	// user input, avoid div by 0
	fontsclw := math.Max(wscale, 0.01)
	fontsclh := math.Max(hscale, 0.01)

	col = int(math.Ceil((winpxcol * fontsclh) / wincol))
	row = int(math.Ceil((winpxrow * fontsclw) / winrow))
	return cfg.Height / row, cfg.Width / col, nil
}

func main() {
	wscaleFactor := flag.Float64("w", 1.0, "font width scaling factor")
	hscaleFactor := flag.Float64("h", 1.0, "font height scaling factor")
	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) < 1 {
		// can we do this os the fly?
		// atm it's a hack to close shell pipe writer.
		tmp, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fatal(err.Error())
		}
		row, col, err := i2term(bytes.NewReader(tmp), "<stdin>", *wscaleFactor, *hscaleFactor)
		if err != nil {
			fatal(err.Error())
		}
		fmt.Printf("%d %d\n", row, col)
		return
	}
	for _, file := range flag.Args() {
		img, err := os.Open(file)
		if err != nil {
			fatal(err.Error())
		}
		defer img.Close()
		row, col, err := i2term(img, file, *wscaleFactor, *hscaleFactor)
		if err != nil {
			fatal(err.Error())
		}
		fmt.Fprintf(os.Stdout, "%d %d\n", row, col)
	}
}

// noreturn
func fatal(err string) {
	fmt.Fprintf(os.Stderr, "i2term: %v\n", err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintf(os.Stderr, "i2term: usage: [ flags ] [ file... ]\n")
	flag.PrintDefaults()
	os.Exit(2)
}
