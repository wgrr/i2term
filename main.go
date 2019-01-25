// +build linux darwin freebsd openbsd netbsd dragonfly solaris

package main

import (
	"bufio"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/sys/unix"
)

const usagestr = `lukeidraw: usage: [ file ]
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, usagestr)
		os.Exit(1)
	}
	img, err := os.Open(os.Args[1])
	if err != nil {
		fatal(err.Error())
	}

	var cfg image.Config
	switch filepath.Ext(os.Args[1]) {
	case ".jpg":
		fallthrough
	case ".jpeg":
		cfg, err = jpeg.DecodeConfig(img)
		if err != nil {
			fatal(err.Error())
		}
	case ".gif":
		cfg, err = gif.DecodeConfig(img)
		if err != nil {
			fatal(err.Error())
		}
	case ".png":
		cfg, err = png.DecodeConfig(img)
		if err != nil {
			fatal(err.Error())
		}
	default:
		// try to guess, it might be a image without an filename extension
		cfg, _, err = image.DecodeConfig(img)
		if err != nil {
			fatal(err.Error())
		}
	}

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

	os.Stdout.Write([]byte{033, byte('['), byte('6'), byte('n')})
	r := bufio.NewReader(os.Stdin)
	termin, err := r.ReadBytes(byte('R'))
	if err != nil {
		// try to not break currenet terminal
		unix.IoctlSetTermios(0, unix.TCSETS, tmp)
		fatal("unexpected input from terminal")
	}

	err = unix.IoctlSetTermios(0, unix.TCSETS, tmp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reset: %s\nyour terminal is broken, fix it manually by typing reset", err)
	}

	pattern, err := regexp.Compile("[0-9]+;[0-9]+")
	if err != nil {
		// should never reach
		fatal(err.Error())
	}
	tpos := pattern.FindAll(termin, -1)
	if tpos == nil {
		fatal("unexpected input from terminal")
	}
	fmt.Printf("%d;%d-%s\n", cfg.Width, cfg.Height, tpos[0])
}

// noreturn
func fatal(err string) {
	fmt.Fprintf(os.Stderr, "lukeidraw: %v\n", err)
	os.Exit(1)
}
