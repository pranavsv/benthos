// Copyright (c) 2019 Ashley Jeffs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, sub to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package tracer

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

//------------------------------------------------------------------------------

func init() {
	Constructors[TypeJaeger] = TypeSpec{
		constructor: NewJaeger,
		description: `
Send spans to a [Jaeger](https://www.jaegertracing.io/) agent.

Available sampler types are: const, probabilistic, ratelimiting and remote.`,
	}
}

//------------------------------------------------------------------------------

// JaegerConfig is config for the Jaeger metrics type.
type JaegerConfig struct {
	AgentAddress          string            `json:"agent_address" yaml:"agent_address"`
	ServiceName           string            `json:"service_name" yaml:"service_name"`
	SamplerType           string            `json:"sampler_type" yaml:"sampler_type"`
	SamplerManagerAddress string            `json:"sampler_manager_address" yaml:"sampler_manager_address"`
	SamplerParam          float64           `json:"sampler_param" yaml:"sampler_param"`
	Tags                  map[string]string `json:"tags" yaml:"tags"`
	FlushInterval         string            `json:"flush_interval" yaml:"flush_interval"`
}

// NewJaegerConfig creates an JaegerConfig struct with default values.
func NewJaegerConfig() JaegerConfig {
	return JaegerConfig{
		AgentAddress:          "localhost:6831",
		ServiceName:           "benthos",
		SamplerType:           "const",
		SamplerManagerAddress: "",
		SamplerParam:          1.0,
		Tags:                  map[string]string{},
		FlushInterval:         "",
	}
}

//------------------------------------------------------------------------------

// Jaeger is a tracer with the capability to push spans to a Jaeger instance.
type Jaeger struct {
	closer io.Closer
}

// NewJaeger creates and returns a new Jaeger object.
func NewJaeger(config Config, opts ...func(Type)) (Type, error) {
	j := &Jaeger{}

	for _, opt := range opts {
		opt(j)
	}

	var sampler *jaegercfg.SamplerConfig
	if sType := config.Jaeger.SamplerType; len(sType) > 0 {
		sampler = &jaegercfg.SamplerConfig{
			Param:             config.Jaeger.SamplerParam,
			SamplingServerURL: config.Jaeger.SamplerManagerAddress,
		}
		switch strings.ToLower(sType) {
		case "const":
			sampler.Type = jaeger.SamplerTypeConst
		case "probabilistic":
			sampler.Type = jaeger.SamplerTypeProbabilistic
		case "ratelimiting":
			sampler.Type = jaeger.SamplerTypeRateLimiting
		case "remote":
			sampler.Type = jaeger.SamplerTypeRemote
		default:
			return nil, fmt.Errorf("unrecognised sampler type: %v", sType)
		}
	}

	cfg := jaegercfg.Configuration{
		ServiceName: config.Jaeger.ServiceName,
		Sampler:     sampler,
	}

	if tags := config.Jaeger.Tags; len(tags) > 0 {
		var jTags []opentracing.Tag
		for k, v := range config.Jaeger.Tags {
			jTags = append(jTags, opentracing.Tag{
				Key:   k,
				Value: v,
			})
		}
		cfg.Tags = jTags
	}

	reporterConf := &jaegercfg.ReporterConfig{}

	if i := config.Jaeger.FlushInterval; len(i) > 0 {
		flushInterval, err := time.ParseDuration(i)
		if err != nil {
			return nil, fmt.Errorf("failed to parse flush interval '%s': %v", i, err)
		}
		reporterConf.BufferFlushInterval = flushInterval
		cfg.Reporter = reporterConf
	}

	if i := config.Jaeger.AgentAddress; len(i) > 0 {
		reporterConf.LocalAgentHostPort = i
		cfg.Reporter = reporterConf
	}

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, err
	}
	opentracing.SetGlobalTracer(tracer)
	j.closer = closer

	return j, nil
}

//------------------------------------------------------------------------------

// Close stops the tracer.
func (j *Jaeger) Close() error {
	if j.closer != nil {
		j.closer.Close()
		j.closer = nil
	}
	return nil
}

//------------------------------------------------------------------------------
