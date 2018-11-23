package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"

	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/camera/control"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
)

//Główne
var scene = core.NewNode()

//Zmienne
var scale_x = float32(0)
var scale_y = float32(0)
var cylinder_x = float32(-0.05)
var cylinder_y = float32(-9 * 0.1)
var mnożnik = float64(2000)
var speed = float32(0.01)
var lives = int(3)
var best = float64(0)

var pos_x = float32(-0.05)
var pos_z = [15][15][5]float32{}

var block_list = [15][15][5]*graphic.Mesh{}

var cylinder_mesh = graphic.NewMesh(nil, nil)
var player_mesh = graphic.NewMesh(nil, nil)

var standard = 0.1

var rotation = float32(-120)

var score = float64(0)

var move = int(0)

var shot = false

var blocks = int(0)

func main() {
	wmgr, err := window.Manager("glfw")
	if err != nil {
		panic(err)
	}
	win, err := wmgr.CreateWindow(800, 600, "GoBall by Mateusz Dera", false)
	if err != nil {
		panic(err)
	}

	runtime.LockOSThread()

	gs, err := gls.New()
	if err != nil {
		panic(err)
	}

	scene_gui := gui.NewRoot(gs, win)
	scene_gui.SetColor(math32.NewColor("darkgray"))

	width, height := win.Size()
	gs.Viewport(0, 0, int32(width), int32(height))
	scene_gui.SetSize(float32(width), float32(height))

	aspect := float32(width) / float32(height)
	camera := camera.NewPerspective(65, aspect, 0.01, 1000)
	camera.SetPosition(0, 1.5, -1.5)
	camera.LookAt(&math32.Vector3{0, 0, 0})

	win.Subscribe(window.OnWindowSize, func(evname string, ev interface{}) {
		width, height := win.FramebufferSize()
		gs.Viewport(0, 0, int32(width), int32(height))
		scene_gui.SetSize(float32(width), float32(height))
		aspect := float32(width) / float32(height)
		camera.SetAspect(aspect)
	})
	control.NewOrbitControl(camera, win)

	napis := gui.NewLabel("---")
	napis.SetPosition(10, 10)
	napis.SetPaddings(2, 2, 2, 2)
	scene_gui.Add(napis)

	b1 := gui.NewButton("<---")
	b1.SetPosition(50, 500)
	b1.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		move = 1
	})
	b1.Subscribe(gui.OnCursorLeave, func(name string, ev interface{}) {
		move = 0
	})
	scene_gui.Add(b1)

	b2 := gui.NewButton("--->")
	b2.SetPosition(100, 500)
	b2.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		move = -1
	})
	b2.Subscribe(gui.OnCursorLeave, func(name string, ev interface{}) {
		move = 0
	})
	scene_gui.Add(b2)

	b3 := gui.NewButton("Shot")
	b3.SetPosition(715, 500)
	b3.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		shot = true
	})
	scene_gui.Add(b3)

	scene.Add(camera)

	//Światło
	ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	scene.Add(ambientLight)
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(1, 0, 2)
	scene.Add(pointLight)

	//Shader
	rend := renderer.NewRenderer(gs)
	err = rend.AddDefaultShaders()
	if err != nil {
		panic(err)
	}

	//Renderowanie
	rend.SetScene(scene)
	rend.SetGui(scene_gui)

	// Czyszczenie ekranu
	gs.ClearColor(0.21, 0.21, 0.21, 1.0)

	create()

	//-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-\\

	// Główna pętla
	for !win.ShouldClose() {
		rend.Render(camera)
		win.SwapBuffers()
		wmgr.PollEvents()

		//Pozycja cylindra
		if lives > 0 {

			if move == -1 {
				if pos_x > -0.645 {
					pos_x = pos_x - speed
				}
			}
			if move == 1 {
				if pos_x < 0.545 {
					pos_x = pos_x + speed
				}
			}

			scale_x = math32.Cos(math32.DegToRad(rotation))
			scale_y = math32.Sin(math32.DegToRad(rotation))

			cylinder_x = cylinder_x + speed*scale_x
			cylinder_y = cylinder_y + speed*scale_y

			cylinder_mesh.SetPositionX(cylinder_x)
			cylinder_mesh.SetPositionZ(cylinder_y)

			if cylinder_y < -11*0.1 {
				rotation = 90
				cylinder_x = -0.05
				cylinder_y = -9 * 0.1
				lives = lives - 1
				shot = false
			}

			player_mesh.SetPositionX(pos_x)

			//Kolizje
			end := false
			for x := 0; x < 15; x++ {
				for y := 0; y < 15; y++ {
					for z := 0; z < 5; z++ {

						if block_list[x][y][z] != nil {
							if z > 0 {
								if pos_z[x][y][z-1] == 0 {
									if block_list[x][y][z].Position().Y > float32(0) {
										block_list[x][y][z].SetPositionY(float32(block_list[x][y][z].Position().Y) - speed)
									}
									if block_list[x][y][z].Position().Y < float32(0) {
										block_list[x][y][z].SetPositionY(0)
									}
								}
								if pos_z[x][y][z-1] == 1 {
									if block_list[x][y][z].Position().Y > block_list[x][y][z-1].Position().Y+0.1 {
										block_list[x][y][z].SetPositionY(float32(block_list[x][y][z].Position().Y) - speed)
									}
									if block_list[x][y][z].Position().Y < block_list[x][y][z-1].Position().Y+0.1 {
										block_list[x][y][z].SetPositionY(block_list[x][y][z-1].Position().Y + 0.1)
									}
								}
							}
						}
						if pos_z[x][y][z] == 1 {
							if float32(block_list[x][y][z].Position().Y) < float32(standard)*0.75 {
								if block_list[x][y][z].Position().Z < cylinder_mesh.Position().Z+0.08 {
									if block_list[x][y][z].Position().Z > cylinder_mesh.Position().Z-0.08 {
										if block_list[x][y][z].Position().X < cylinder_mesh.Position().X+0.08 {
											if block_list[x][y][z].Position().X > cylinder_mesh.Position().X-0.08 {
												if end == false {
													rotation = -(90 - rotation) + ((rand.Float32() * 20) - 10)
													end = true
												}
												pos_z[x][y][z] = 0
												scene.Remove(block_list[x][y][z])
												blocks = blocks - 1

												score = score + mnożnik
											}
										}
									}
								}
							}
						}
					}
				}
			}

			if player_mesh.Position().Z-0.05 < cylinder_mesh.Position().Z+0.08 {
				if player_mesh.Position().Z+0.05 > cylinder_mesh.Position().Z-0.08 {
					if player_mesh.Position().X-0.15 < cylinder_mesh.Position().X+0.08 {
						if player_mesh.Position().X+0.15 > cylinder_mesh.Position().X-0.08 {
							if end == false {
								rotation = -(90 - rotation) + ((rand.Float32() * 20) - 10)
								end = true
							}
						}
					}
				}
			}

			if cylinder_x > 0.65 {
				if scale_y > float32(0) {
					rotation = 120
				}
				if scale_y < float32(0) {
					rotation = -120
				}
			}

			if cylinder_x < -0.75 {
				if scale_y > float32(0) {
					rotation = 30
				}
				if scale_y < float32(0) {
					rotation = -30
				}
			}

			if cylinder_y > 1.4 {
				if scale_x > float32(0) {
					rotation = -30
				}
				if scale_x < float32(0) {
					rotation = -120
				}
			}

			if int(rotation)%30 == 0 {
				rotation = rotation + ((rand.Float32() * 10) - 5)
			}

			if shot == false {
				cylinder_mesh.SetPositionX(player_mesh.Position().X)
				cylinder_mesh.SetPositionZ(player_mesh.Position().Z + 0.15)
				cylinder_x = cylinder_mesh.Position().X
				cylinder_y = cylinder_mesh.Position().Z
			}
			mnożnik = mnożnik - 0.01
		}

		//Napisy
		napis_końcowy := fmt.Sprint("Mnożnik: ", math.Round(mnożnik), "    Wynik: ", math.Round(score), "   Najlepszy wynik: ", math.Round(best), "   Życia: ", lives)

		if lives <= 0 || blocks <= 0 {
			napis_końcowy = fmt.Sprint("KONIEC GRY   Twój wynik: ", math.Round(score), "   Najlepszy wynik: ", math.Round(best))
		}

		napis.SetText(napis_końcowy)

		if lives <= 0 {
			if score > best {
				f, _ := os.Create("./game.save")

				save := fmt.Sprint(score)

				f.WriteString(save)

				defer f.Close()
			}
		}
	}
}

func create() {
	for x := 0; x < 15; x++ {
		for y := 0; y < 15; y++ {
			for z := 0; z < 5; z++ {
				rand := (rand.Float32() * 10)

				pos_z[x][y][z] = 1

				if rand < 5 {
					pos_z[x][y][z] = 0
				}

				if z > 1 {
					if pos_z[x][y][z-1] == 0 {
						pos_z[x][y][z] = 0
					}
				}

				if pos_z[x][y][z] == 1 {
					blocks = blocks + 1
				}

			}
		}
	}

	for x := 0; x < 15; x++ {
		for y := 0; y < 15; y++ {
			for z := 0; z < 5; z++ {
				if pos_z[x][y][z] != 0 {
					create_block(x, y, z, rand.Intn(12))
				}
			}
		}
	}

	create_cylinder(-1, -1)
	create_player()
	bottom()
}

func create_block(positionX int, positionY int, positionZ int, color int) {
	geom := geometry.NewBox(float32(standard), float32(standard), float32(standard))
	mat := material.NewPhong(math32.NewColor("Red"))
	switch color {
	case 0:
		mat = material.NewPhong(math32.NewColor("White"))
	case 1:
		mat = material.NewPhong(math32.NewColor("Red"))
	case 2:
		mat = material.NewPhong(math32.NewColor("Green"))
	case 3:
		mat = material.NewPhong(math32.NewColor("Blue"))
	case 4:
		mat = material.NewPhong(math32.NewColor("Yellow"))
	case 5:
		mat = material.NewPhong(math32.NewColor("Orange"))
	case 6:
		mat = material.NewPhong(math32.NewColor("Pink"))
	case 7:
		mat = material.NewPhong(math32.NewColor("Violet"))
	case 8:
		mat = material.NewPhong(math32.NewColor("Brown"))
	case 9:
		mat = material.NewPhong(math32.NewColor("Lime"))
	case 10:
		mat = material.NewPhong(math32.NewColor("Cyan"))
	case 11:
		mat = material.NewPhong(math32.NewColor("Magenta"))
	default:
		mat = material.NewPhong(math32.NewColor("Black"))
	}

	block := graphic.NewMesh(geom, mat)
	block.SetPosition(float32(standard)*float32(positionX)-float32(standard)*15/2, float32(standard)*float32(positionZ), float32(standard)*float32(positionY))
	scene.Add(block)
	block_list[positionX][positionY][positionZ] = block
}

func create_cylinder(positionX int, positionY int) {
	geom := geometry.NewCylinder(standard/2, standard/2, standard*0.75, 360, 360, 360, 360, true, true)
	mat := material.NewPhong(math32.NewColor("Red"))
	cylinder := graphic.NewMesh(geom, mat)
	cylinder.SetPosition(float32(positionX), 0, float32(positionY))
	scene.Add(cylinder)
	cylinder_mesh = cylinder
}

func bottom() {
	geom := geometry.NewBox(1.5, 0.1, 2.5)
	mat := material.NewPhong(math32.NewColor("Black"))
	block := graphic.NewMesh(geom, mat)
	block.SetPosition(-0.05, -0.1, 0.2)
	scene.Add(block)
}

func create_player() {
	geom := geometry.NewBox(0.3, 0.1, 0.1)
	mat := material.NewPhong(math32.NewColor("Lime"))
	player := graphic.NewMesh(geom, mat)
	player.SetPosition(-0.05, 0, -float32(10*standard))
	scene.Add(player)
	player_mesh = player
}
