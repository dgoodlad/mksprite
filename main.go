package main

import (
	"flag"
	"fmt"
	"image/png"
	"io"
	"log"
	"os"
)

const black = 0
const white = 1

func main() {
	inPath := flag.String("in", "-", "input file or - for stdin")
	outPath := flag.String("out", "-", "output file or - for stdout")
	flag.Parse()

	var input = io.Reader(os.Stdin)
	if *inPath != "-" {
		f, err := os.Open(*inPath)
		if err != nil {
			log.Fatal(err)
		}
		input = f
	}

	var output = io.Writer(os.Stdout)
	if *outPath != "-" {
		f, err := os.Create(*outPath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		output = f
	}

	img, err := png.Decode(input)
	if err != nil {
		log.Fatal(err)
	}

	bounds := img.Bounds()
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y
	pixels := make([]uint8, w*h/8)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := img.At(x, y)
			r, g, b, a := pixel.RGBA()
			// TODO LOL
			if r != 0 || g != 0 || b != 0 || a != 0 {
				byte := x/8 + y*w/8
				bit := x % 8
				pixels[byte] = pixels[byte] | (1 << uint8(bit))
			}
		}
	}

	fmt.Fprintln(output, "const uint8_t PROGMEM sprite[] = {")
	fmt.Fprint(output, "  ")
	for i, p := range pixels {
		if i != 0 {
			fmt.Fprint(output, ",")
		}
		fmt.Fprintf(output, "%#02x", p)
	}
	fmt.Fprintln(output, "\n};")
}
