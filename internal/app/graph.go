package app

import (
	"context"

	clog "github.com/charmbracelet/log"
	"github.com/sourcegraph/conc/pool"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2oracle"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/util-go/go2"
)

type Action int8

const (
	Create Action = 0
	Set    Action = 1
	Finish Action = 2
)

type graphUpdate struct {
	action Action
	key    string
	value  string
}

type outerGraph struct {
	graph *d2graph.Graph
}

type GraphProvider interface {
	Create(context context.Context, update chan<- graphUpdate) error
}

func (app *App) CreateGraph(ctx context.Context, provider GraphProvider) (string, error) {
	graph, err := app.initGraph(ctx)
	if err != nil {
		return "", err
	}
	outerGraph := &outerGraph{graph: graph}

	graphUpdateChannel := make(chan graphUpdate)
	graphPool := pool.New().WithContext(ctx).WithFirstError()

	graphPool.Go(func(context context.Context) error {
		err := graphUpdateReceiver(outerGraph, context, graphUpdateChannel)
		if err != nil {
			return err
		}
		return nil
	})

	provider.Create(ctx, graphUpdateChannel)

	graphUpdateChannel <- graphUpdate{action: Finish}

	out, err := app.renderGraph(ctx, outerGraph.graph)
	if err != nil {
		return "", err
	}

	return out, nil
}

func (app *App) initGraph(context context.Context) (*d2graph.Graph, error) {
	context = log.Stderr(context)
	_, graph, _ := d2lib.Compile(context, "", app.D2CompileOptions, nil)

	// d2oracle.Set(graph, nil, "grid-columns", nil, go2.Pointer("2"))
	// d2oracle.Set(graph, nil, "grid-rows", nil, go2.Pointer("2"))

	return graph, nil
}

func (app *App) renderGraph(context context.Context, graph *d2graph.Graph) (string, error) {
	context = log.Stderr(context)
	script := d2format.Format(graph.AST)
	clog.Debug(script)

	diagram, _, _ := d2lib.Compile(context, script, app.D2CompileOptions, app.D2RenderOpts)

	out, _ := d2svg.Render(diagram, app.D2RenderOpts)

	return string(out), nil

}

func graphUpdateReceiver(outerGraph *outerGraph, context context.Context, update <-chan graphUpdate) error {
	newGraph := &d2graph.Graph{}
	for elem := range update {
		clog.Debug(elem)
		switch elem.action {
		case Create:
			newGraph, _, _ = d2oracle.Create(outerGraph.graph, nil, elem.key)
		case Set:
			newGraph, _ = d2oracle.Set(outerGraph.graph, nil, elem.key, nil, go2.Pointer(elem.value))
		case Finish:
			clog.Debug("graph compiled")
		}
	}
	outerGraph.graph = newGraph

	return nil
}
