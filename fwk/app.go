package fwk

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"
)

type App struct {
	*Base
	ctx     context.Context
	modules []Module
}

func New(name string, modules ...Module) (*App, error) {
	for _, module := range modules {
		dev := module.(Device)
		if System.Device(module.Name()) == dev {
			continue
		}
		System.Register(dev)
	}

	return &App{
		Base:    NewBase(name),
		ctx:     context.Background(),
		modules: modules,
	}, nil
}

func (app *App) AddModule(m Module) {
	app.modules = append(app.modules, m)
}

func (app *App) Run() error {
	var err error

	err = app.sysBoot()
	if err != nil {
		return err
	}

	err = app.sysStart()
	if err != nil {
		return err
	}

	err = app.sysRun()
	if err != nil {
		return err
	}

	err = app.sysStop()
	if err != nil {
		return err
	}

	err = app.sysShutdown()
	if err != nil {
		return err
	}

	return err
}

func (app *App) visit(node Node) error {
	type named struct {
		Name string
		Lvl  int
	}
	var nodes []named
	var visit func(node Node, lvl int)
	visit = func(node Node, lvl int) {
		nodes = append(nodes, named{node.Name, lvl})
		for _, node := range node.Nodes {
			visit(node, lvl+1)
		}
	}
	visit(node, 0)
	for _, n := range nodes {
		fmt.Printf("%s--> %s\n", strings.Repeat("  ", n.Lvl), n.Name)
	}

	return nil
}

func (app *App) sysBoot() error {
	var err error
	ctx, cancel := context.WithCancel(app.ctx)
	defer cancel()

	for _, m := range app.modules {
		err = m.Boot(ctx)
		if err != nil {
			return err
		}
	}

	return err
}

func (app *App) sysStart() error {
	ctx, cancel := context.WithCancel(app.ctx)
	defer cancel()
	for _, m := range app.modules {
		err := m.Start(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *App) sysRun() error {
	return nil
}

func (app *App) sysStop() error {
	ctx, cancel := context.WithCancel(app.ctx)
	defer cancel()
	for _, m := range app.modules {
		err := m.Stop(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *App) sysShutdown() error {
	ctx, cancel := context.WithCancel(app.ctx)
	defer cancel()
	for _, m := range app.modules {
		err := m.Shutdown(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
