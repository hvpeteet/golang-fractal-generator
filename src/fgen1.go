package main

import(
    "image"
    "image/png"
    "image/color"
    "fmt"
    "os"
    "math"
)
type rectangle struct {
    Min, Max point
}
func (r rectangle) Dx() float64 {
    return math.Abs(r.Min.X - r.Max.X)
}
func (r rectangle) Dy() float64 {
    return math.Abs(r.Min.Y - r.Max.Y)
}
type point struct {
    X, Y float64
}
func main() {
    width := 1020
    height := 660

    scale := 67.0
    bottomLeft := point{-0.435,0.586}

    // scale := 1.0
    // bottomLeft := point{-2,-1}

    window := rectangle{bottomLeft, point{bottomLeft.X + 3.0 / scale, bottomLeft.Y + 2.0 / scale}}
    maxIterations := 100
    im := image.NewRGBA(image.Rect(0,0,width,height))
    hist := make([]int, maxIterations)
    for x := 0; x < width; x++ {
        for y := 0; y < height; y++ {
            adjustedX := float64(window.Min.X) + float64(x) / float64(width) * float64(window.Dx())
            adjustedY := float64(window.Min.Y) + float64(y) / float64(height) * float64(window.Dy())
            hist[Mandelbrot(adjustedX, adjustedY, maxIterations) - 1] += 1
        }
    }
    vals := make([]float64, maxIterations)
    total := 0
    for _,h := range hist {
        total += h
    }
    vals[0] = float64(hist[0]) / float64(total)
    for v := 1; v < len(vals) ; v++ {
        vals[v] = vals[v - 1] + float64(hist[v]) / float64(total)
    }
    fmt.Println(vals[maxIterations - 1])
    for x := 0; x < width; x++ {
        for y := 0; y < height; y++ {
            adjustedX := float64(window.Min.X) + float64(x) / float64(width) * float64(window.Dx())
            adjustedY := float64(window.Min.Y) + float64(y) / float64(height) * float64(window.Dy())
            val := 255 * vals[Mandelbrot(adjustedX, adjustedY, maxIterations) - 1]
            r := uint8(float64(uint8(val * 255)) * 0.)
            g := uint8(val * 255 * 2.0)
            b := uint8(val * 255 * 5.0)//uint8(math.Min(val * 255 + 40, 255))
            a := uint8(255)
            if val >= 255 * vals[len(vals) - 1] {
                r = 255
                g = 255
                b = 255
                a = 0
            }
            col := color.RGBA{r,g,b,a}
            im.SetRGBA(x,y,col)
            
        }
    }
    imageFile, err := os.Create("Creation2.png")
    if (err == nil){
        defer imageFile.Close()
        png.Encode(imageFile, im)
    } else {
        fmt.Println("failed")
    }
}
// Returns the number of iterations
func Mandelbrot(x, y float64, maxIteration int) int {
    x0 := x
    y0 := y
    x = 0.0
    y = 0.0

    iteration := 0

    for (x*x + y*y < 2*2 && iteration < maxIteration) {
        xtemp := x*x - y*y + x0
        y = 2*x*y+y0
        x = xtemp
        iteration += 1
    }
    return iteration
}