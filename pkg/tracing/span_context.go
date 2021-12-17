package tracing

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// FromObject takes a Kubernetes objects and returns a Span from the
// context found in its annotations, or nil if not found;
// also a logger connected to the span and a new Context set up with both.
func FromObject(ctx context.Context, operationName string, obj runtime.Object) (context.Context, trace.Span, logr.Logger) {
	log := ctrl.LoggerFrom(ctx)
	m, err := meta.Accessor(obj)
	if err != nil {
		return nil, nil, log
	}
	ctx, sp := SpanFromAnnotations(ctx, operationName, m.GetAnnotations())
	if sp == nil {
		return nil, nil, log
	}
	sp.SetAttributes(label.String("objectKey", m.GetNamespace()+"/"+m.GetName()))
	log = tracingLogger{Logger: log, Span: sp}
	ctx = ctrl.LoggerInto(ctx, log)
	return ctx, sp, log
}

func NewSpanFromObject(ctx context.Context, operationName string, obj runtime.Object) (context.Context, trace.Span, logr.Logger) {
	log := ctrl.LoggerFrom(ctx)
	m, err := meta.Accessor(obj)
	if err != nil {
		return nil, nil, log
	}
	// TODO Keys, values in span?
	// TODO Maybe this should live in the pkg
	ctx, sp := otel.Tracer("controller-runtime").Start(ctx, fmt.Sprintf("Start reconciling %s", m.GetName()))
	// TODO When does it end?
	/*
		ctx, sp := SpanFromAnnotations(ctx, operationName, m.GetAnnotations())
		if sp == nil {
			return nil, nil, log
		}
	*/
	sp.SetAttributes(label.String("objectKey", m.GetNamespace()+"/"+m.GetName()),
		label.String("objectType", obj.GetObjectKind().GroupVersionKind().String()))
	log = tracingLogger{Logger: log, Span: sp}

	// TODO Try out the tracing logger?
	ctx = ctrl.LoggerInto(ctx, log)
	ctx = trace.ContextWithSpan(ctx, sp)
	return ctx, sp, log
}
