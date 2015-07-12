package main

import(
    "image"
    "image/png"
    "image/color"
    "fmt"
    "os"
    "math"
    "math/cmplx"
)

type chaoticIterator func(accum, start complex128) complex128

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

func CalcColor(val float64) color.RGBA {
  if (val == 1.0) {
    return color.RGBA{0,0,0,255}
  }
  val = val * val * val
  // Green
  r := 255 - uint8(255)//uint8(math.Pow(val, 1.5) * 255)
  g := 255 - uint8(math.Pow(val, .5) * 255)//uint8(math.Pow(math.Atan(val), 2) * 12000)
  b := 255 / 2 - uint8(math.Pow(val, .5) * 255) / 2//uint8(math.Pow(math.Atan(val), 2) * 12000) / 2
  a := uint8(255)
  // r := uint8(math.Pow(val, 1.5) * 255)
  // g := uint8(math.Pow(val, 3) * 200)
  // b := uint8(0)
  return color.RGBA{r,g,b,a}
}

func GetEscapeIterations(x, y float64, maxIteration int, givenFunc chaoticIterator, error float64, start complex128) int {
  iterated := complex(x, y)
  iteration := 1

  for (cmplx.Abs(iterated) < error && iteration < maxIteration) {
    iterated = givenFunc(iterated, start)
    iteration++
  }
  return iteration
}

func Man(val, val0 complex128) complex128 {
  return val * val + val0
}

func test0(val, val0 complex128) complex128 {
  return val * cmplx.Sqrt(cmplx.Cosh(val * val * val)) * val0
}

func test1(val, val0 complex128) complex128 {
  return cmplx.Sqrt(cmplx.Sinh(val)) + val0
}

func test2(val, val0 complex128) complex128 {
  return val * val * cmplx.Exp(val) + val0
}

func test3(val, val0 complex128) complex128 {
  return val * val * val * val * val + val0
}

func test4(val, val0 complex128) complex128 {
  return (val * val + val) / cmplx.Log(val) + val0
}

func CreateFractalImage(function chaoticIterator, max_iterations int, start complex128, max_value float64, viewing_window rectangle, display_quality image.Rectangle, file_name string) {

  // Calculate Color Values
  im := image.NewRGBA(display_quality)
  hist := make([]int, max_iterations)

  dx := viewing_window.Dx() / float64(display_quality.Dx())
  dy := viewing_window.Dy() / float64(display_quality.Dy())

  for x := 0; x < display_quality.Dx(); x++ {
    for y := 0; y < display_quality.Dy(); y++ {
      adjustedX := float64(x) * dx + viewing_window.Min.X
      adjustedY := float64(y) * dy + viewing_window.Min.Y
      hist[GetEscapeIterations(adjustedX, adjustedY, max_iterations, function, max_value, start) - 1] += 1
    }
  }

  // Setup and calculate the histogram 
  // --> vals = percentage of pixels below the current val in terms of iterations.
  vals := make([]float64, max_iterations)
  total := 0
  last_val := 0.0
  for i, h := range hist {
      total += h
      fmt.Printf("%d: %d\n", i, h)
      last_val = float64(h)
  }

  vals[0] = float64(hist[0]) / float64(total)
  for v := 1; v < len(vals) - 1; v++ {
      vals[v] = vals[v - 1] + float64(hist[v]) / (float64(total) - last_val)
      fmt.Printf("%d: %.5f\n", v, vals[v])
  }
  vals[len(vals) - 1] = 1.0

  fmt.Printf("First Pass Done, percentage captured : %d%%\n", (int)(vals[max_iterations - 1]  * 100))

  // Get the actual pixel values and assign them to the image
  for x := 0; x < display_quality.Dx(); x++ {
      for y := 0; y < display_quality.Dy(); y++ {
          adjustedX := float64(x) * dx + viewing_window.Min.X
          adjustedY := float64(y) * dy + viewing_window.Min.Y
          val := vals[GetEscapeIterations(adjustedX, adjustedY, max_iterations, function, max_value, start) - 1]
          col := CalcColor(val)
          im.SetRGBA(x,display_quality.Dy() - y - 1,col)
      }
  }
  imageFile, err := os.Create(file_name)
  if (err == nil){
      defer imageFile.Close()
      png.Encode(imageFile, im)
  } else {
      fmt.Println("failed")
  }
}

func Test0() {
  function := test0
  max_value := 2.0
  start := complex(0.8, 0.6)
  zoom := 7.25
  display_quality := image.Rect(0, 0, 1080, 1920)
  file_name := "REVERSE_OVERGROWTH.png"
  center_point := point{-0.8, 0.425}
  viewing_window := rectangle{point{-1.080 / zoom + center_point.X, -1.920 / zoom + center_point.Y}, point{ 1.080 / zoom + center_point.X, 1.920 / zoom + center_point.Y}}
  max_iterations := 100
  seconds := float64(max_iterations * display_quality.Dx() * display_quality.Dy() * 28) / (10 * 1080 * 1920 * 4)
  fmt.Printf("Expected creation time: %.3f seconds (%.5f minutes)\n", seconds, seconds / 60)
  CreateFractalImage(function, max_iterations, start, max_value, viewing_window, display_quality, file_name)
}

func main() {
  Test0()
}