//go:build !linux || !cgo

package gst

type dummyPipeline struct {
	sample chan Sample
}

func (p *dummyPipeline) Play()  {}
func (p *dummyPipeline) Pause() {}
func (p *dummyPipeline) Destroy() {
	close(p.sample)
}
func (p *dummyPipeline) Sample() chan Sample {
	return p.sample
}
func (p *dummyPipeline) EmitVideoKeyframe() {}

func CreatePipeline(pipelineStr string) (Pipeline, error) {
	return &dummyPipeline{
		sample: make(chan Sample, 100),
	}, nil
}
