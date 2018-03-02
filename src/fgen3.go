package main

import (
  "flag"
  "fmt"
  "image"
  "image/color"
  "image/png"
  "math"
  "math/cmplx"
  "os"
  "sync"
)

// All the parameters needed to describe a fractal.
type fractalParams struct {
  function chaoticFunction
  start    complex128
}

// All the parameters needed to describe the quality and style of the rendering.
type renderingParams struct {
  max_iterations   int
  escape_threshold float64
  viewing_window   rectangle
  display_quality  image.Rectangle
  colorSchemeFunc  colorScheme
}

// A colorScheme returns a color based on a float64 value in the range [0.00, 1.00]
type colorScheme func(val float64) color.RGBA

// A chaotic function is just a function from complex128 x complex128 --> complex 128
// Certain chaotic functions are well known for creating cool / popular fractals, but
// you can use any function you want.
// For example: The function that describes the mandelbrot set is a^2 + b.
type chaoticFunction func(a, b complex128) complex128

type point struct {
  X, Y float64
}

type rectangle struct {
  Min, Max point
}

func (r rectangle) Dx() float64 {
  return math.Abs(r.Min.X - r.Max.X)
}

func (r rectangle) Dy() float64 {
  return math.Abs(r.Min.Y - r.Max.Y)
}

// -----------------------------
// Implemented chaotic functions
// -----------------------------

func Man(a, b complex128) complex128 {
  return a*a + b
}

func test0(a, b complex128) complex128 {
  return a * cmplx.Sqrt(cmplx.Cosh(a*a*a)) * b
}

func test1(a, b complex128) complex128 {
  return cmplx.Sqrt(cmplx.Sinh(a)) + b
}

func test2(a, b complex128) complex128 {
  return a*a*cmplx.Exp(a) + b
}

func test3(a, b complex128) complex128 {
  return a*a*a*a*a + b
}

func test4(a, b complex128) complex128 {
  return (a*a+a)/cmplx.Log(a) + b
}

// ----------------------
// End chaotic functions
// ----------------------

// -------------
// Color schemes
// -------------

func BlackAndGreen(val float64) color.RGBA {
  if val == 1.0 {
    return color.RGBA{0, 0, 0, 255}
  }
  // Green
  r := uint8(0)                               //uint8(math.Pow(val, 1.5) * 255)
  g := 255 - uint8(math.Pow(val, .5)*255)     //uint8(math.Pow(math.Atan(val), 2) * 12000)
  b := 255/2 - uint8(math.Pow(val, .5)*255)/2 //uint8(math.Pow(math.Atan(val), 2) * 12000) / 2
  a := uint8(255)
  // r := uint8(math.Pow(val, 1.5) * 255)
  // g := uint8(math.Pow(val, 3) * 200)
  // b := uint8(0)
  return color.RGBA{r, g, b, a}
}

// -----------------
// End color schemes
// -----------------

// -----------
// Core logic
// -----------

// Runs the chaotic function on a coordinate and returns the number of times iteration was needed in order to escape.
func GetEscapeIterations(x, y float64, fparams fractalParams, rparams renderingParams) int {
  a := complex(x, y)
  b := fparams.start
  iterations := 1

  for cmplx.Abs(a) < rparams.escape_threshold && iterations < rparams.max_iterations {
    a = fparams.function(a, b)
    iterations++
  }
  return iterations
}

// Creates a fractal image and writes it to a png file.
// NOTE: file_name takes only the first part of the file name (without the .png)
func CreateFractalImage(fparams fractalParams, rparams renderingParams, file_name string) {

  // Calculate the escape values histogram
  im := image.NewRGBA(rparams.display_quality)
  hist := make([]int, rparams.max_iterations)

  dx := rparams.viewing_window.Dx() / float64(rparams.display_quality.Dx())
  dy := rparams.viewing_window.Dy() / float64(rparams.display_quality.Dy())
  tmp_img := make([][]int, rparams.display_quality.Dx())
  for i := 0; i < len(tmp_img); i++ {
    tmp_img[i] = make([]int, rparams.display_quality.Dy())
  }

  var wg sync.WaitGroup
  for x := 0; x < rparams.display_quality.Dx(); x++ {
    for y := 0; y < rparams.display_quality.Dy(); y++ {
      wg.Add(1)
      // Calculating the escape iterations is expensive, so do it in parallel
      go func(ax, ay int) {
        defer wg.Done()
        adjustedX := float64(ax)*dx + rparams.viewing_window.Min.X
        adjustedY := float64(ay)*dy + rparams.viewing_window.Min.Y
        tmp_img[ax][ay] = GetEscapeIterations(adjustedX, adjustedY, fparams, rparams)
      }(x, y)
    }
  }
  wg.Wait()
  for x := 0; x < rparams.display_quality.Dx(); x++ {
    for y := 0; y < rparams.display_quality.Dy(); y++ {
      hist[tmp_img[x][y]-1]++
    }
  }

  // Normalize the histogram counts (to get a CDF)
  // NOTE: bins is a CDF (cumulative distribution function)
  //       A.K.A: it tracks the percentage of pixels below each value.
  // TODO: bins may be sparse, consider using a hash table / map or other structure.
  total := float64(rparams.display_quality.Dx() * rparams.display_quality.Dy())

  bins := make([]float64, rparams.max_iterations)
  running_total := 0.0

  for v := 0; v < len(bins)-1; v++ {
    running_total += float64(hist[v])
    bins[v] = running_total / total
  }
  // The last bin is for pixels that did not escape in max_iterations.
  bins[len(bins)-1] = 1.0

  // Get the actual pixel colors
  for x := 0; x < rparams.display_quality.Dx(); x++ {
    for y := 0; y < rparams.display_quality.Dy(); y++ {
      color := rparams.colorSchemeFunc(bins[tmp_img[x][y]-1])
      im.SetRGBA(x, rparams.display_quality.Dy()-y-1, color)
    }
  }

  imageFile, err := os.Create(file_name)
  if err == nil {
    defer imageFile.Close()
    png.Encode(imageFile, im)
  } else {
    fmt.Println("failed")
  }
}

func main() {
  outfile := flag.String("outfile", "fractal.png", "the file to save the image to, this should be a .png")
  // TODO: Convert most params to use flags
  flag.Parse()

  fparams := fractalParams{
    function: test0,
    start:    complex(0.8, 0.6),
  }

  center_point := point{-0.8, 0.425}
  zoom := 7.25

  rparams := renderingParams{
    max_iterations:   100,
    escape_threshold: 2.0,
    viewing_window: rectangle{
      point{
        X: -1.080/zoom + center_point.X,
        Y: -1.920/zoom + center_point.Y,
      }, point{
        X: 1.080/zoom + center_point.X,
        Y: 1.920/zoom + center_point.Y,
      },
    },
    display_quality: image.Rect(0, 0, 1080, 1920),
    colorSchemeFunc: BlackAndGreen,
  }

  CreateFractalImage(fparams, rparams, *outfile)
}
