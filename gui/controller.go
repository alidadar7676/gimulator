package main

import (
	"log"

	"github.com/alidadar7676/gimulator/simulator"
	"github.com/alidadar7676/gimulator/types"
)

type Controller struct {
	Name      string
	Namespace string

	gimulator simulator.Gimulator
	watcher   chan simulator.Reconcile
}

func NewController(name, namespace string, gimulator simulator.Gimulator) *Controller {
	return &Controller{
		Name:      name,
		Namespace: namespace,
		gimulator: gimulator,
		watcher:   make(chan simulator.Reconcile, 1024),
	}
}

func (c *Controller) Run() {
	worldFilter := simulator.Object{
		Key: simulator.Key{
			Type:      types.WorldType,
			Namespace: c.Namespace,
		}}

	worlds, err := c.gimulator.Find(worldFilter)
	switch {
	case err != nil:
		log.Printf("An error in finding: %s", err)
		return
	case len(worlds) > 1:
		log.Println("Number of World is more than one")
		return
	case len(worlds) == 1:
		var world types.World
		if err := worlds[0].Struct(&world); err != nil {
			log.Printf("Cannot Structing the world object, Error: %s", err)
			return
		}
		c.watchWorld(world)
	}

	if err = c.gimulator.Watch(worldFilter, c.watcher); err != nil {
		log.Printf("Cannot watch on the world, Error: %s", err)
	}

	go func() {
		for r := range c.watcher {
			var world types.World
			if err := r.Object.Struct(&world); err != nil {
				log.Printf("object %v is not world: %v\n", r.Object, err)
				continue
			}
			c.watchWorld(world)
		}
	}()
}

func (c *Controller) watchWorld(world types.World) {
	d := worldDrawer{
		World:  world,
		width:  width(),
		height: height(),
	}
	render(d)
	disableEvent = world.Turn != playerName
}

func (c *Controller) InitPlayer(playerName string) error {
	playerIntroObject := simulator.Object{
		Key: simulator.Key{
			Type:      types.PlayerIntroType,
			Name:      playerName,
			Namespace: c.Namespace,
		},
		Value: types.PlayerIntro{},
	}
	return c.gimulator.Set(playerIntroObject)
}

func (c *Controller) Act(action types.Action) error {
	actionKey := simulator.Key{
		Type:      types.ActionType,
		Name:      action.PlayerName,
		Namespace: c.Namespace,
	}
	actionObject := simulator.Object{
		Key:   actionKey,
		Value: action,
	}
	log.Println("controller: set", action)
	return c.gimulator.Set(actionObject)
}
