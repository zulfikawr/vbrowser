package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0 gstreamer-video-1.0
#include "gst.h"
*/
import "C"
import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var (
	pSerial       int32
	pipelines     = make(map[int]*pipeline)
	pipelinesLock sync.Mutex
)

func init() {
	C.gst_init(nil, nil)
}

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

type pipeline struct {
	id     int
	ctx    *C.GstPipelineCtx
	sample chan Sample
}

func CreatePipeline(pipelineStr string) (Pipeline, error) {
	id := atomic.AddInt32(&pSerial, 1)

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	var gstError *C.GError
	ctx := C.gstreamer_pipeline_create(pipelineStrUnsafe, C.int(id), &gstError)

	if gstError != nil {
		return nil, fmt.Errorf("gst error: %s", C.GoString(gstError.message))
	}

	p := &pipeline{
		id:     int(id),
		ctx:    ctx,
		sample: make(chan Sample, 100),
	}

	C.gstreamer_pipeline_attach_appsink(ctx, C.CString("appsink"))

	pipelinesLock.Lock()
	pipelines[p.id] = p
	pipelinesLock.Unlock()

	return p, nil
}

func (p *pipeline) Play() {
	C.gstreamer_pipeline_play(p.ctx)
}

func (p *pipeline) Pause() {
	C.gstreamer_pipeline_pause(p.ctx)
}

func (p *pipeline) Destroy() {
	C.gstreamer_pipeline_destory(p.ctx)
	pipelinesLock.Lock()
	delete(pipelines, p.id)
	pipelinesLock.Unlock()
	close(p.sample)
}

func (p *pipeline) Sample() chan Sample {
	return p.sample
}

func (p *pipeline) EmitVideoKeyframe() {
	C.gstreamer_pipeline_emit_video_keyframe(p.ctx)
}

//export goHandlePipelineBuffer
func goHandlePipelineBuffer(id C.int, buffer unsafe.Pointer, bufferLen C.int, duration C.int, isKeyframe C.int) {
	pipelinesLock.Lock()
	p, ok := pipelines[int(id)]
	pipelinesLock.Unlock()

	if ok {
		p.sample <- Sample{
			Data:       C.GoBytes(buffer, bufferLen),
			Duration:   time.Duration(duration),
			IsKeyframe: isKeyframe == 1,
		}
	}
}
