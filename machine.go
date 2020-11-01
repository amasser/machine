// Copyright © 2020 Jonathan Whitaker <github@whitaker.io>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package machine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/metric"
	"go.opentelemetry.io/otel/label"
)

type handler func([]*Packet)
type recorder func(string, string, string, []*Packet)

type root struct {
	id        string
	retrieve  Retriever
	next      *vertex
	option    *Option
	vertacies map[string]*vertex
	recorder
}

type vertex struct {
	id         string
	vertexType string
	input      *edge
	handler
	connector func(ctx context.Context, m *root) error
	metrics   *metrics
}

type metrics struct {
	labels             []label.KeyValue
	inCounter          metric.Int64ValueRecorder
	outCounter         metric.Int64ValueRecorder
	errorsCounter      metric.Int64ValueRecorder
	inTotalCounter     metric.Float64Counter
	outTotalCounter    metric.Float64Counter
	errorsTotalCounter metric.Float64Counter
	batchDuration      metric.Int64ValueRecorder
}

func (m *root) run(ctx context.Context) error {
	if m.next == nil {
		return fmt.Errorf("non-terminated machine")
	}

	input := m.retrieve(ctx)
	edge := newEdge()

	tracer := global.Tracer("retriever.begin")
	spanEnabled := *m.option.Span
	go func() {
	Loop:
		for {
			select {
			case <-ctx.Done():
				break Loop
			case data := <-input:
				if len(data) < 1 {
					continue
				}

				payload := make([]*Packet, len(data))
				for i, item := range data {
					packet := &Packet{
						ID:   uuid.New().String(),
						Data: item,
					}
					if spanEnabled {
						packet.newSpan(ctx, tracer, "retriever.begin", m.id, "retriever")
					}
					payload[i] = packet
				}

				edge.channel <- payload
			}
		}
	}()

	return m.next.cascade(ctx, m, edge)
}

func (m *root) inject(ctx context.Context, logs map[string][]*Packet) {
	if payload, ok := logs[m.id]; ok {
		if *m.option.Span {
			tracer := global.Tracer("retriever.inject")
			for _, packet := range payload {
				packet.newSpan(ctx, tracer, "retriever.inject", m.id, "retriever")
			}
		}
		m.next.input.channel <- payload
	}

	for node, payload := range logs {
		if v, ok := m.vertacies[node]; ok {
			if *m.option.Span {
				tracer := global.Tracer(v.vertexType + ".inject")
				for _, packet := range payload {
					packet.newSpan(ctx, tracer, v.vertexType+".inject", v.id, v.vertexType)
				}
			}
			v.input.channel <- payload
		}
	}
}

func (v *vertex) cascade(ctx context.Context, m *root, input *edge) error {
	if v.input != nil {
		input.sendTo(ctx, v.input)
		return nil
	}

	v.input = input

	h := v.handler

	if m.recorder != nil {
		h = m.recorder.wrap(v.id, v.vertexType, h)
	}

	if *m.option.Metrics {
		h = v.metrics.wrap(ctx, h)
	}

	if *m.option.Span {
		h = v.wrap(ctx, h)
	}

	do(ctx, *m.option.FIFO, h, input)

	m.vertacies[v.id] = v

	return v.connector(ctx, m)
}

func (v *vertex) wrap(ctx context.Context, h handler) handler {
	return func(payload []*Packet) {
		start := time.Now()

		for _, packet := range payload {
			packet.span.AddEvent(ctx, "vertex",
				label.String("vertex_id", v.id),
				label.String("vertex_type", v.vertexType),
				label.String("packet_id", packet.ID),
				label.Int64("when", start.UnixNano()),
			)
		}

		h(payload)

		for _, packet := range payload {
			if packet.Error != nil {
				packet.span.AddEvent(ctx, "error",
					label.String("vertex_id", v.id),
					label.String("vertex_type", v.vertexType),
					label.String("packet_id", packet.ID),
					label.Bool("error", packet.Error != nil),
				)
			}
			if v.vertexType == "sender" {
				packet.span.End()
			}
		}
	}
}

func (mtrx *metrics) wrap(ctx context.Context, h handler) handler {
	return func(payload []*Packet) {
		mtrx.inCounter.Record(ctx, int64(len(payload)), mtrx.labels...)
		mtrx.inTotalCounter.Add(ctx, float64(len(payload)), mtrx.labels...)
		start := time.Now()
		h(payload)
		duration := time.Since(start)
		failures := 0
		for _, packet := range payload {
			if packet.Error != nil {
				failures++
			}
		}
		mtrx.outCounter.Record(ctx, int64(len(payload)), mtrx.labels...)
		mtrx.outTotalCounter.Add(ctx, float64(len(payload)), mtrx.labels...)
		mtrx.errorsCounter.Record(ctx, int64(failures), mtrx.labels...)
		mtrx.errorsTotalCounter.Add(ctx, float64(failures), mtrx.labels...)
		mtrx.batchDuration.Record(ctx, int64(duration), mtrx.labels...)
	}
}

func createMetrics(id, vertexType string) *metrics {
	meter := global.Meter(id)
	return &metrics{
		labels: []label.KeyValue{
			label.String("vertex_id", id),
			label.String("vertex_type", vertexType),
		},
		inTotalCounter:     metric.Must(meter).NewFloat64Counter(vertexType + "." + id + ".total.incoming"),
		outTotalCounter:    metric.Must(meter).NewFloat64Counter(vertexType + "." + id + ".total.outgoing"),
		errorsTotalCounter: metric.Must(meter).NewFloat64Counter(vertexType + "." + id + ".total.errors"),
		inCounter:          metric.Must(meter).NewInt64ValueRecorder(vertexType + "." + id + ".incoming"),
		outCounter:         metric.Must(meter).NewInt64ValueRecorder(vertexType + "." + id + ".outgoing"),
		errorsCounter:      metric.Must(meter).NewInt64ValueRecorder(vertexType + "." + id + ".errors"),
		batchDuration:      metric.Must(meter).NewInt64ValueRecorder(vertexType + "." + id + ".duration"),
	}
}

func (r recorder) wrap(id, vertexType string, h handler) handler {
	return func(payload []*Packet) {
		r(id, vertexType, "start", payload)
		h(payload)
		r(id, vertexType, "done", payload)
	}
}

func do(ctx context.Context, fifo bool, h handler, input *edge) {
	go func() {
	Loop:
		for {
			select {
			case <-ctx.Done():
				break Loop
			case data := <-input.channel:
				if len(data) < 1 {
					continue
				}

				if fifo {
					h(data)
				} else {
					go h(data)
				}
			}
		}
	}()
}
