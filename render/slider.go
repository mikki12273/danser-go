package render

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath/math32"
	"github.com/wieku/danser-go/render/framebuffer"
	"github.com/wieku/danser-go/settings"
	"github.com/wieku/danser-go/utils"
	"github.com/wieku/glhf"
	_ "image/png"
	"io/ioutil"
	"log"
)

var sliderShader *glhf.Shader = nil
var fboShader *glhf.Shader
var fboSlice *glhf.VertexSlice
var sliderVertexFormat glhf.AttrFormat
var cam mgl32.Mat4
var fbo *framebuffer.Framebuffer
var fboUnit int32
var CS float64
var unitCircle []float32

func SetupSlider() {

	sliderVertexFormat = glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
		{Name: "center", Type: glhf.Vec3},
		//{Name: "in_tex_coord", Type: glhf.Vec2},
	}
	var err error

	svert, _ := ioutil.ReadFile("assets/shaders/slider.vsh")
	sfrag, _ := ioutil.ReadFile("assets/shaders/slider.fsh")
	sliderShader, err = glhf.NewShader(sliderVertexFormat, glhf.AttrFormat{{Name: "col_border", Type: glhf.Vec4}, {Name: "proj", Type: glhf.Mat4}, {Name: "trans", Type: glhf.Mat4}, {Name: "col_border1", Type: glhf.Vec4}}, string(svert), string(sfrag))
	if err != nil {
		log.Println(err)
	}

	fvert, _ := ioutil.ReadFile("assets/shaders/fbopass.vsh")
	ffrag, _ := ioutil.ReadFile("assets/shaders/fbopass.fsh")
	fboShader, err = glhf.NewShader(glhf.AttrFormat{
		{Name: "in_position", Type: glhf.Vec3},
		{Name: "in_tex_coord", Type: glhf.Vec2},
	}, glhf.AttrFormat{{Name: "tex", Type: glhf.Int}}, string(fvert), string(ffrag))

	if err != nil {
		log.Println("FboPass: " + err.Error())
	}

	fbo = framebuffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, true)
	fbo.Texture().Bind(29)
	fboUnit = 29

	fboSlice = glhf.MakeVertexSlice(fboShader, 6, 6)
	fboSlice.Begin()
	fboSlice.SetVertexData([]float32{
		-1, -1, 0, 0, 0,
		1, -1, 0, 1, 0,
		-1, 1, 0, 0, 1,
		1, -1, 0, 1, 0,
		1, 1, 0, 1, 1,
		-1, 1, 0, 0, 1,
	})
	fboSlice.End()

}

type SliderRenderer struct{}

func NewSliderRenderer() *SliderRenderer {
	return &SliderRenderer{}
}

func (sr *SliderRenderer) Begin() {

	gl.Disable(gl.BLEND)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthMask(true)
	gl.DepthFunc(gl.LESS)

	fbo.Begin()

	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	sliderShader.Begin()

	sliderShader.SetUniformAttr(1, cam)
}

func (sr *SliderRenderer) SetColor(color mgl32.Vec4, prev mgl32.Vec4) {
	sliderShader.SetUniformAttr(0, color)
	if settings.Objects.EnableCustomSliderBorderGradientOffset {
		sliderShader.SetUniformAttr(3, utils.GetColorShifted(color, settings.Objects.SliderBorderGradientOffset))
	} else {
		sliderShader.SetUniformAttr(3, prev)
	}
}

func (sr *SliderRenderer) SetScale(scale float64) {
	sliderShader.SetUniformAttr(2, mgl32.Scale3D(float32(scale), float32(scale), 1))
}

func (sr *SliderRenderer) EndAndRender() {

	sliderShader.End()
	fbo.End()
	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)
	gl.Enable(gl.BLEND)

	gl.BlendEquation(gl.FUNC_ADD)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	fboShader.Begin()
	fboShader.SetUniformAttr(0, int32(fboUnit))
	fboSlice.BeginDraw()
	fboSlice.Draw()
	fboSlice.EndDraw()
	fboShader.End()
	gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
}

func (self *SliderRenderer) SetCamera(camera mgl32.Mat4) {
	cam = camera
	sliderShader.SetUniformAttr(1, cam)
}

func (self *SliderRenderer) GetShape(curve []bmath.Vector2f) ([]float32, int) {
	return createMesh(curve), int(settings.Objects.SliderLOD)
}

func (self *SliderRenderer) UploadMesh(mesh []float32) *glhf.VertexSlice {
	slice := glhf.MakeVertexSlice(sliderShader, len(mesh)/6, len(mesh)/6)
	slice.Begin()
	slice.SetVertexData(mesh)
	slice.End()
	return slice
}

func createMesh(curve []bmath.Vector2f) []float32 {
	if len(unitCircle) == 0 {
		createCircle(int(settings.Objects.SliderLOD))
	}

	vertices := make([]float32, 6*3*int(settings.Objects.SliderLOD)*len(curve))

	for i, v := range curve {
		for j, s := range unitCircle {
			vertices[i*len(unitCircle)+j] = s
			if j%3 == 0 {
				vertices[i*len(unitCircle)+j] *= float32(CS)
				vertices[i*len(unitCircle)+j] += v.X
			} else if j%3 == 1 {
				vertices[i*len(unitCircle)+j] *= float32(CS)
				vertices[i*len(unitCircle)+j] += v.Y
			}
		}
	}

	return vertices
}

func set(array []float32, index int, data ...float32) {
	copy(array[index:index+len(data)], data)
}

func createCircle(segments int) {
	points := make([]bmath.Vector2f, segments+2)
	points[0] = bmath.NewVec2f(0, 0)

	for i := 0; i < segments; i++ {
		points[i+1] = bmath.NewVec2fRad(float32(i)/float32(segments)*2*math32.Pi, 1)
	}

	points[segments+1] = points[1]

	unitCircle = make([]float32, 6*3*segments)

	for j := range points {
		if j >= 2 {
			p1, p2, p3 := points[j-1], points[j], points[0]
			set(unitCircle, (j-2)*6*3, p1.X, p1.Y, 1.0, p3.X, p3.Y, 0.0, p2.X, p2.Y, 1.0, p3.X, p3.Y, 0.0, p3.X, p3.Y, 0.0, p3.X, p3.Y, 0.0)
		}
	}
}
