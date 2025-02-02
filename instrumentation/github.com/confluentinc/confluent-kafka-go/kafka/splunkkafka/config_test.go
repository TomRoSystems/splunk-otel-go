// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build cgo && (linux || darwin)
// +build cgo
// +build linux darwin

package splunkkafka

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	splunkotel "github.com/signalfx/splunk-otel-go"
)

type fnTracerProvider struct {
	tracer func(string, ...trace.TracerOption) trace.Tracer
}

func (fn *fnTracerProvider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return fn.tracer(name, opts...)
}

type fnTracer struct {
	start func(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
}

func (fn *fnTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return fn.start(ctx, name, opts...)
}

func TestConfigTracerFromGlobal(t *testing.T) {
	c := newConfig()
	expected := otel.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	assert.Equal(t, expected, c.Tracer)
}

func TestConfigTracerFromConfig(t *testing.T) {
	tp := &fnTracerProvider{
		tracer: func(string, ...trace.TracerOption) trace.Tracer {
			return &fnTracer{}
		},
	}
	c := newConfig(WithTracerProvider(tp))
	expected := tp.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(splunkotel.Version()),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	assert.Equal(t, expected, c.Tracer)
}

type testTextMapProp struct{}

var _ propagation.TextMapPropagator = (*testTextMapProp)(nil)

func (*testTextMapProp) Inject(context.Context, propagation.TextMapCarrier) {}

func (*testTextMapProp) Extract(ctx context.Context, _ propagation.TextMapCarrier) context.Context {
	return ctx
}

func (*testTextMapProp) Fields() []string { return nil }

func TestConfigDefaultPropagator(t *testing.T) {
	c := newConfig()
	expected := otel.GetTextMapPropagator()
	assert.Same(t, expected, c.Propagator)
}

func TestConfigUserPropagator(t *testing.T) {
	prop := new(testTextMapProp)
	c := newConfig(WithPropagator(prop))
	assert.Same(t, prop, c.Propagator)
}
