// +build linux

package main

import (
	"bufio"
	"bytes"
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
	win, err := unix.IoctlGetWinsize(1, unix.TIOCGWINSZ)
	if err != nil {
		fatal(err.Error())
	}

	// just being paranoid about kernel input
	winrow, winpxrow := math.Max(float64(win.Row), 1.0), math.Max(float64(win.Ypixel), 1.0)
	wincol, winpxcol := math.Max(float64(win.Col), 1.0), math.Max(float64(win.Xpixel), 1.0)

	fontsclw := math.Max(*charwScale, 0.01)
	fontsclh := math.Max(*charhScale, 0.01)

	charw := int(math.Ceil((winpxrow * fontsclw) / winrow))
	charh := int(math.Ceil((winpxcol * fontsclh) / wincol))

	var term unix.Termios
	tmp, err := unix.IoctlGetTermios(0, unix.TCGETS)
	if err != nil {
		fatal("couldn't setup termio to listen to terminal input: " + err.Error())
	}

	term = *tmp
	term.Lflag &= (^uint32(unix.ICANON) & ^uint32(unix.ECHO))
	err = unix.IoctlSetTermios(0, unix.TCSETS, &term)
	if err != nil {
		fatal("couldn't setup termio to listen to terminal input: " + err.Error())
	}

	os.Stdout.WriteString("\x1b[6n")
	r := bufio.NewReader(os.Stdin)
	termin, err := r.ReadBytes(byte('R'))
	if err != nil {
		// try restore user terminal
		unix.IoctlSetTermios(0, unix.TCSETS, tmp)
		fatal("unexpected input from terminal")
	}

	tmpbuf := bytes.NewBuffer(termin[2:])
	row, err := tmpbuf.ReadBytes(byte(';'))
	if err != nil {
		fatal("unexpected input from terminal")
	}
	col, err := tmpbuf.ReadBytes(byte('R'))
	if err != nil {
		fatal("unexpected input from terminal")
	}

	err = unix.IoctlSetTermios(0, unix.TCSETS, tmp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lukeidraw: %syour terminal is broken, fix it manually by typing reset\n", err)
	}
	fmt.Printf("%d %d %s %s\n", cfg.Width/charw, cfg.Height/charh, string(row[:len(row)-1]), string(col[:len(col)-1]))
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
