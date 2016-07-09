package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"io"
	"log"
	"os"
)

type FrameRect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type Frame struct {
	Filename string    `json:"filename"`
	Rect     FrameRect `json:"frame"`
}

type SpritesheetMeta struct {
	Image string `json:"image"`
}

type Spritesheet struct {
	Frames []Frame         `json:"frames"`
	Meta   SpritesheetMeta `json:"meta"`
}

type Bitmap struct {
	W          int
	H          int
	Number     int
	Pixels     []int
	PixelBytes []uint8
}

func NewBitmap(w int, h int, number int) *Bitmap {
	return &Bitmap{
		W:          w,
		H:          h,
		Number:     number,
		Pixels:     make([]int, w*h),
		PixelBytes: make([]uint8, h*w/8),
	}
}

func (bitmap *Bitmap) SetPixel(x int, y int) {
	byte := x/8 + y*bitmap.W/8
	bitmap.Pixels[y*bitmap.W+x] = 1
	bitmap.PixelBytes[byte] |= (1 << uint8(x%8))
}

func (bitmap *Bitmap) PrintByteArray(out io.Writer) {
	fmt.Fprint(out, "  {")
	for i, byte := range bitmap.PixelBytes {
		if i != 0 {
			fmt.Fprint(out, ",")
		}
		fmt.Fprintf(out, "%#02x", byte)
	}
	fmt.Fprint(out, "}")
}

func (bitmap *Bitmap) PrintAsciiDicks(out io.Writer) {
	fmt.Fprintf(out, "  /* Frame number %d\n", bitmap.Number)
	for y := 0; y < bitmap.H; y++ {
		if y != 0 {
			fmt.Fprint(out, "\n")
		}

		fmt.Fprint(out, "      ")
		for _, pixel := range bitmap.Pixels[y*bitmap.W : y*bitmap.W+bitmap.W] {
			if pixel == 0 {
				fmt.Fprintf(out, " ")
			} else {
				fmt.Fprintf(out, "#")
			}
		}
	}
	fmt.Fprint(out, "\n  */\n")
}

func main() {
	inPath := flag.String("in", "-", "input file or - for stdin")
	outPath := flag.String("out", "-", "output file or - for stdout")
	name := flag.String("name", "sprite", "name of outputted array variable")
	jsonPath := flag.String("json", "", "frame data json file")
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

	jsonFile, err := os.Open(*jsonPath)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	spritesheet := &Spritesheet{}
	json.NewDecoder(jsonFile).Decode(spritesheet)

	img, err := png.Decode(input)
	if err != nil {
		log.Fatal(err)
	}

	bitmaps := make([]*Bitmap, len(spritesheet.Frames))
	for number, frame := range spritesheet.Frames {
		bitmap := NewBitmap(frame.Rect.W, frame.Rect.H, number)
		for y := frame.Rect.Y; y < frame.Rect.Y+frame.Rect.H; y++ {
			for x := frame.Rect.X; x < frame.Rect.X+frame.Rect.W; x++ {
				pixel := img.At(x, y)
				r, g, b, a := pixel.RGBA()
				if r != 0 || g != 0 || b != 0 || a != 0 {
					bitmap.SetPixel(x-frame.Rect.X, y-frame.Rect.Y)
				}
			}
		}
		bitmaps[number] = bitmap
	}

	fmt.Fprintf(output, "const uint8_t %sFrameCount = %d;\n\n", *name, len(spritesheet.Frames))
	fmt.Fprintf(output, "const **uint8_t %s = {\n", *name)
	for n, bitmap := range bitmaps {
		if n != 0 {
			fmt.Fprint(output, ",\n")
		}
		bitmap.PrintAsciiDicks(output)
		bitmap.PrintByteArray(output)
	}
	fmt.Fprintln(output, "\n};")
}
