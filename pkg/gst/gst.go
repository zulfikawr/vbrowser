package gst

import "time"

type Sample struct {
	Data       []byte
	Duration   time.Duration
	IsKeyframe bool
}

type Pipeline interface {
	Play()
	Pause()
	Destroy()
	Sample() chan Sample
	EmitVideoKeyframe()
}
