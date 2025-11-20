// Program wavy generates images of waveforms for documentation
// purposes.
//
// The input file is a plain text file:
//
// <chars><space><signal><space><commad-values>
// <blank lines = half line skip>
//
// The size of the generated image is the smallest size that can render
// the image.
package main

import (
	"flag"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	in    = flag.String("input", "", "input file in wvy format")
	out   = flag.String("output", "", "output file in png format")
	fSize = flag.Float64("fs", 16, "font size")
	debug = flag.Bool("debug", false, "use to generate vertical lines")
)

type signal struct {
	name                  string
	clk                   bool
	states                string
	phase                 float64
	clkHalfPeriodMinusOne int
	labels                []string
}

// parse digests a line
func parse(i int, line string) (sig signal) {
	if len(line) == 0 {
		return
	}
	ts := strings.Split(line, " ")
	if c := len(ts); c < 2 {
		log.Fatalf("%s:%d: need two or more fields: got %d", *in, i, c)
	}
	if strings.HasPrefix(ts[0], "+") {
		sig.clk = true
		ns := strings.Split(ts[0], ",")
		n := len(ns)
		if n == 0 || n > 2 {
			log.Fatalf("%s:%d: clock signal requires number[,phase]: %q", *in, i, line)
		}
		sig.clkHalfPeriodMinusOne, _ = strconv.Atoi(ns[0][1:])
		if n == 2 {
			var err error
			sig.phase, err = strconv.ParseFloat(ns[1], 64)
			if err != nil {
				log.Fatalf("%s:%d: clock phase parse err: %v", err)
			}
		}
	} else {
		sig.states = ts[0]
	}
	sig.name = ts[1]
	if len(ts) > 2 {
		sig.labels = strings.Split(ts[2], ",")
	}
	return
}

// mult determines the maximum size of a rendered item (step) and the
// total number of steps required to render the wave. This is the
// maximum of the length of a label, or the minimum distance between
// "[" and "]", or the clk period width.
func mult(sig signal) (step, width int) {
	if sig.name == "" {
		return 0, 0
	}
	width = len(sig.states)
	if width == 0 {
		step = 2 * (1 + sig.clkHalfPeriodMinusOne)
		return
	}
	if len(sig.labels) != 0 {
		i := 0
		for _, w := range sig.labels {
			for ; i < len(sig.states) && sig.states[i] != '<'; i++ {
			}
			found := i < len(sig.states) && sig.states[i] == '<'
			if !found {
				break
			}
			from := i
			for ; i < len(sig.states) && sig.states[i] != '>'; i++ {
			}
			if i != len(sig.states) {
				i++
			}
			c := 3 + len(w)
			d := 1
			for c > d*(i-from) {
				d++
			}
			if d > step {
				step = d
			}
		}
	}
	return
}

func main() {
	flag.Parse()

	b, err := ioutil.ReadFile(*in)
	if err != nil {
		log.Fatalf("failed to read %q: %v", *in, err)
	}
	text := 1
	width := 1
	step := 1

	var sigs []signal
	for i, line := range strings.Split(string(b), "\n") {
		if i == 0 && line == "" {
			continue
		}
		sig := parse(i, line)
		n, w := mult(sig)
		if n > step {
			step = n
		}
		if w > width {
			width = w
		}
		if t := len(sig.name); t > text {
			text = t
		}
		sigs = append(sigs, sig)
	}
	if len(sigs) > 1 && sigs[len(sigs)-1].name == "" {
		sigs = sigs[:len(sigs)-1]
	}

	const spacer = 1.8

	right := float64(text) * *fSize
	wide := right + *fSize*0.5*float64(step*(-1+width))
	high := math.Ceil(spacer * *fSize * float64(2+len(sigs)))

	dest := image.NewRGBA(image.Rect(0, 0, int(wide), int(high)))
	gc := draw2dimg.NewGraphicContext(dest)

	gc.SetFillColor(color.White)
	gc.SetStrokeColor(color.Black)

	gc.MoveTo(0, 0)
	gc.LineTo(wide, 0)
	gc.LineTo(wide, high)
	gc.LineTo(0, high)
	gc.Close()
	gc.Fill()

	mp := "^_"
	for i, sig := range sigs {
		if !sig.clk {
			continue
		}
		var parts []string
		outer := 0
		for k := width + 2*(sig.clkHalfPeriodMinusOne+1); k > 0; k-- {
			c := mp[outer : outer+1]
			k++
			for j := 0; k > 0; j++ {
				k--
				parts = append(parts, c)
				if j == sig.clkHalfPeriodMinusOne {
					break
				}
			}
			outer = 1 - outer
		}
		sigs[i].states = strings.Join(parts, "")
	}

	full := 0.5 * *fSize * float64(step)

	if *debug {
		for i := 0; i < width; i++ {
			gc.MoveTo(right+full*(0.5+float64(i)), 0)
			gc.LineTo(right+full*(0.5+float64(i)), high)
			gc.Stroke()
		}
	}

	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal(err)
	}
	fondata := draw2d.FontData{Name: "goregular", Family: draw2d.FontFamilyMono, Style: draw2d.FontStyleNormal}
	draw2d.RegisterFont(
		fondata,
		font,
	)
	gc.SetFontData(fondata)

	for i, sig := range sigs {
		if sig.name == "" {
			continue
		}

		phase := sig.phase * full * float64(1+sig.clkHalfPeriodMinusOne)
		half := 0.5 * full
		demi := 0.5 * *fSize
		bot := *fSize * (0.5 + float64(2+i)) * spacer
		top := bot - *fSize*spacer
		mid := 0.5 * (bot + top)

		gc.SetFillColor(color.RGBA{0xcc, 0xcc, 0xcc, 0xff})
		gc.MoveTo(0, bot-1)
		gc.LineTo(wide, bot-1)
		gc.LineTo(wide, top+1)
		gc.LineTo(0, top+1)
		gc.Close()
		gc.Fill()

		var oldC string
		var labNo int
		var lastStart, lastEnd float64
		soFar := make(map[string]bool)
		for i, c := range strings.Split(sig.states, "") {
			if oldC == "" {
				oldC = c
				continue
			}
			showLabel := false
			nextStart := lastStart
			start := right + full*(-1+float64(i)) - phase
			combo := oldC + c
			gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0xff, 0xff})
			switch combo {
			case "^^", "/^", "%^", "^%":
				gc.MoveTo(start, mid-demi)
				gc.LineTo(start+full, mid-demi)
				gc.Stroke()
			case "^_":
				gc.MoveTo(start, mid-demi)
				gc.LineTo(start+half, mid-demi)
				gc.LineTo(start+half, mid+demi)
				gc.LineTo(start+full, mid+demi)
				gc.Stroke()
			case "^\\":
				gc.MoveTo(start, mid-demi)
				gc.LineTo(start+half*0.7, mid-demi)
				gc.LineTo(start+half*1.3, mid+demi)
				gc.LineTo(start+full, mid+demi)
				gc.Stroke()
			case "__", "\\_", "%_", "_%":
				gc.MoveTo(start, mid+demi)
				gc.LineTo(start+full, mid+demi)
				gc.Stroke()
			case "_^":
				gc.MoveTo(start, mid+demi)
				gc.LineTo(start+half, mid+demi)
				gc.LineTo(start+half, mid-demi)
				gc.LineTo(start+full, mid-demi)
				gc.Stroke()
			case "_/":
				gc.MoveTo(start, mid+demi)
				gc.LineTo(start+half*0.7, mid+demi)
				gc.LineTo(start+half*1.3, mid-demi)
				gc.LineTo(start+full, mid-demi)
				gc.Stroke()
			case "><", ">>":
				gc.MoveTo(start, mid-demi)
				gc.LineTo(start+0.7*half, mid-demi)
				gc.LineTo(start+1.3*half, mid+demi)
				gc.LineTo(start+full, mid+demi)
				lastEnd = start + half
				showLabel = combo != ">>"
				gc.MoveTo(start, mid+demi)
				gc.LineTo(start+0.7*half, mid+demi)
				gc.LineTo(start+1.3*half, mid-demi)
				gc.LineTo(start+full, mid-demi)
				gc.Stroke()
				nextStart = start + half
			case "^x":
				gc.MoveTo(start, mid-demi)
				gc.LineTo(start+full, mid-demi)
				gc.MoveTo(start+half, mid-demi)
				gc.LineTo(start+full, mid+demi)
				gc.Stroke()
			case "_z":
				gc.MoveTo(start, mid+demi)
				gc.LineTo(start+half, mid+demi)
				gc.Stroke()
				gc.SetLineDash([]float64{*fSize * 0.3, *fSize * 0.2}, 0)
				gc.MoveTo(start+half, mid+demi)
				gc.LineTo(start+half, mid)
				gc.LineTo(start+full, mid)
				gc.Stroke()
				gc.SetLineDash(nil, 0)
			case "x^":
				gc.MoveTo(start, mid+demi)
				gc.LineTo(start+half, mid-demi)
				gc.MoveTo(start, mid-demi)
				gc.LineTo(start+full, mid-demi)
				gc.Stroke()
			case "x_":
				gc.MoveTo(start, mid-demi)
				gc.LineTo(start+half, mid+demi)
				gc.MoveTo(start, mid+demi)
				gc.LineTo(start+full, mid+demi)
				gc.Stroke()
			case "xx", "<-", "->", "--", "x%", "%x", "-%", "%-":
				gc.MoveTo(start, mid+demi)
				gc.LineTo(start+full, mid+demi)
				gc.MoveTo(start, mid-demi)
				gc.LineTo(start+full, mid-demi)
				gc.Stroke()
			case ">x":
				gc.MoveTo(start, mid-demi)
				gc.LineTo(start+0.7*half, mid-demi)
				gc.LineTo(start+half, mid)
				gc.LineTo(start+0.7*half, mid+demi)
				gc.LineTo(start, mid+demi)
				lastEnd = start + half
				showLabel = true
				gc.Stroke()
				gc.MoveTo(start+full, mid+demi)
				gc.LineTo(start+half, mid)
				gc.LineTo(start+full, mid-demi)
				gc.Stroke()
			case "x<":
				gc.MoveTo(start, mid+demi)
				gc.LineTo(start+half, mid)
				gc.LineTo(start, mid-demi)
				gc.Stroke()
				gc.MoveTo(start+full, mid-demi)
				gc.LineTo(start+1.1*half, mid-demi)
				gc.LineTo(start+half, mid)
				gc.LineTo(start+1.1*half, mid+demi)
				gc.LineTo(start+full, mid+demi)
				gc.Stroke()
				nextStart = start + half
			case "zx":
				gc.SetLineDash([]float64{*fSize * 0.3, *fSize * 0.2}, 0)
				gc.MoveTo(start, mid)
				gc.LineTo(start+half, mid)
				gc.Stroke()
				gc.SetLineDash(nil, 0)
				gc.MoveTo(start+full, mid+demi)
				gc.LineTo(start+half, mid)
				gc.LineTo(start+full, mid-demi)
				gc.Stroke()
			case "zz", "z%", "%z":
				gc.SetLineDash([]float64{*fSize * 0.3, *fSize * 0.2}, 0)
				gc.MoveTo(start, mid)
				gc.LineTo(start+full, mid)
				gc.Stroke()
				gc.SetLineDash(nil, 0)
			case "%%":
				vert := *fSize * spacer * 0.5
				stop := start + full
				gc.SetFillColor(image.White)
				gc.MoveTo(start, mid-vert)
				gc.LineTo(start-half*0.2, mid+vert/10)
				gc.LineTo(start+half*0.2, mid+vert/10)
				gc.LineTo(start, mid+vert)
				gc.LineTo(stop, mid+vert)
				gc.LineTo(stop+half*0.2, mid-vert/10)
				gc.LineTo(stop-half*0.2, mid-vert/10)
				gc.LineTo(stop, mid-vert)
				gc.Close()
				gc.Fill()
			default:
				if !soFar[combo] {
					soFar[combo] = true
					log.Printf("unrecognized signal pair %q:%d = %q", sig.name, i, combo)
				}
			}
			if showLabel && labNo < len(sig.labels) {
				lab := sig.labels[labNo]
				gc.SetFillColor(image.Black)
				gc.SetFontSize(0.8 * *fSize)

				_, _, w, _ := gc.GetStringBounds(lab)
				gc.FillStringAt(lab, 0.5*(lastStart+lastEnd-w), bot-0.5**fSize)
				labNo++
			}
			lastStart = nextStart
			oldC = c
		}

		gc.SetFillColor(color.White)
		gc.MoveTo(0, bot-1)
		gc.LineTo(right, bot-1)
		gc.LineTo(right, top+1)
		gc.LineTo(0, top+1)
		gc.Close()
		gc.Fill()

		gc.SetFillColor(image.Black)
		gc.SetFontSize(*fSize)

		_, _, w, _ := gc.GetStringBounds(sig.name)
		gc.FillStringAt(sig.name, right-(*fSize*0.5+w), bot-0.4**fSize)
	}

	// Save to file
	if err := draw2dimg.SaveToPngFile(*out, dest); err != nil {
		log.Fatalf("error saving to %q: %v", *out, err)
	}
}
