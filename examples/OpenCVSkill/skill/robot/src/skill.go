package examples

import (
	"bytes"
	"encoding/base64"
	"image/jpeg"
	"math"
	"mind/core/framework"
	"mind/core/framework/drivers/distance"
	"mind/core/framework/drivers/hexabody"
	"mind/core/framework/drivers/media"
	"mind/core/framework/log"
	"mind/core/framework/skill"
	"os"
	"time"

	"github.com/lazywei/go-opencv/opencv"
)

const (
	FAST_DURATION         = 80
	SLOW_DURATION         = 500
	TIME_TO_NEXT_REACTION = 2000 // milliseconds
	DISTANCE_TO_REACTION  = 250  // millimeters
	MOVE_HEAD_DURATION    = 500  // milliseconds
	ROTATE_DEGREES        = 60   // degrees out of 360
	WALK_SPEED            = 1.0  // cm per second
	SENSE_INTERVAL        = 250  // four times per second
)

type OpenCVSkill struct {
	skill.Base
	stop      chan bool
	cascade   *opencv.HaarCascade
	direction float64
}

func NewSkill() skill.Interface {
	return &OpenCVSkill{
		stop:    make(chan bool),
		cascade: opencv.LoadHaarClassifierCascade("assets/haarcascade_frontalface_alt.xml"),
	}
}

func newDirection(direction float64) float64 {
	return math.Mod(direction+ROTATE_DEGREES, 360) * -1
}

func (d *OpenCVSkill) distance() float64 {
	distance, err := distance.Value()
	if err != nil {
		log.Error.Println(err)
	}
	return distance
}

func (d *OpenCVSkill) changeDirection() {
	d.direction = newDirection(d.direction)
	hexabody.MoveHead(d.direction, MOVE_HEAD_DURATION)
	hexabody.WalkContinuously(0, WALK_SPEED)
	time.Sleep(TIME_TO_NEXT_REACTION * time.Millisecond)
}

func (d *OpenCVSkill) shouldChangeDirection() bool {
	return d.distance() < DISTANCE_TO_REACTION
}

func (d *OpenCVSkill) walk() {
	hexabody.WalkContinuously(0, WALK_SPEED)
	for {
		select {
		case <-d.stop:
			return
		default:
			if d.shouldChangeDirection() {
				d.changeDirection()
			}
			time.Sleep(SENSE_INTERVAL * time.Millisecond)
		}
	}
}

func (d *OpenCVSkill) sight() {
	for {
		select {
		case <-d.stop:
			return
		default:
			image := media.SnapshotRGBA()
			buf := new(bytes.Buffer)
			jpeg.Encode(buf, image, nil)
			str := base64.StdEncoding.EncodeToString(buf.Bytes())
			framework.SendString(str)
			cvimg := opencv.FromImageUnsafe(image)
			faces := d.cascade.DetectObjects(cvimg)
			log.Info.Println("Number of faces: ", len(faces))
			if len(faces) >= 1 {
				hexabody.StopWalkingContinuously()
				// hexabody.RelaxLegs()
				// hexabody.Relax()
				// d.play()
			}
			// hexabody.StandWithHeight(float64(len(faces)) * 50)
		}
	}
}

func ready() {
	hexabody.Stand()
	hexabody.MoveHead(0.0, FAST_DURATION)
	// Using goroutines to make some commands be executed at the same time
	legPosition2 := hexabody.NewLegPosition().SetCoordinates(-100, 50.0, 70.0)
	// legPositionInfo() is used to check, adjust or return the legposition's infomation and it's not necessary.
	legPositionInfo(legPosition2)
	// Moveleg
	go hexabody.MoveLeg(2, legPosition2, SLOW_DURATION)
	hexabody.MoveLeg(5, hexabody.NewLegPosition().SetCoordinates(100, 50.0, 70.0), SLOW_DURATION)
	go hexabody.MoveJoint(0, 1, 90, SLOW_DURATION)
	hexabody.MoveJoint(0, 2, 45, SLOW_DURATION)
	go hexabody.MoveJoint(1, 1, 90, FAST_DURATION)
	hexabody.MoveJoint(1, 2, 45, FAST_DURATION)
}

func legPositionInfo(legPosition *hexabody.LegPosition) {
	if !legPosition.IsValid() {
		log.Info.Println("The position is not valid, means it's unreachale, fit it.")
		legPosition.Fit()
	}
	x, y, z, err := legPosition.Coordinates()
	if err != nil {
		log.Info.Println("Get coordinates of legposition error:", err)
	} else {
		log.Info.Println("The coordinates of legposition are:", x, y, z)
	}
}

func moveLegs(v float64) {
	go hexabody.MoveJoint(0, 1, 45*math.Sin(v*math.Pi/180)+70, FAST_DURATION)
	go hexabody.MoveJoint(0, 0, 35*math.Cos(v*math.Pi/180)+60, FAST_DURATION)
	hexabody.MoveJoint(1, 1, 45*math.Cos(v*math.Pi/180)+70, FAST_DURATION)
}

func (d *OpenCVSkill) play() {
	ready()
	v := 0.0
	for {
		select {
		case <-d.stop:
			return
		default:
			moveLegs(v)
			v += 10
		}
	}
}

func (d *OpenCVSkill) OnStart() {
}

func (d *OpenCVSkill) OnConnect() {
	err := hexabody.Start()
	if err != nil {
		log.Error.Println("Hexabody start err:", err)
		return
	}
	if !media.Available() {
		log.Error.Println("Media driver not available")
		return
	}
	if err := media.Start(); err != nil {
		log.Error.Println("Media driver could not start")
	}
	if err := distance.Start(); err != nil {
		log.Error.Println("Distance start err:", err)
	}
	if !distance.Available() {
		log.Error.Println("Distance sensor is not available")
	}
}

func (d *OpenCVSkill) OnClose() {
	hexabody.Close()
	distance.Close()
}

func (d *OpenCVSkill) OnDisconnect() {
	os.Exit(0) // Closes the process when remote disconnects
}

func (d *OpenCVSkill) OnRecvString(data string) {
	log.Info.Println(data)
	switch data {
	case "start":
		go d.walk()
		go d.sight()
	case "stop":
		d.stop <- true
		hexabody.StopWalkingContinuously()
		hexabody.RelaxLegs()
		hexabody.Relax()
	}
}
