package game

import (
	"github.com/alidadar7676/gimulator/types"
	"log"
)

func Update(action types.Action, world types.World) types.World {
	const (
		changeTurn = "otherPlayer"
		fixedTurn  = "fixedTurn"
		noTurn     = "noTurn"
	)

	updateWorld := func(turnState string) {
		world.BallPos = action.To
		world.Moves = append(world.Moves,
			types.Move{
				Name: action.PlayerName,
				A:    action.From,
				B:    action.To,
			})
		switch turnState {
		case changeTurn:
			world.Turn = world.OtherPlayer(action.PlayerName)
		case noTurn:
			world.Turn = ""
		}
	}

	if action.PlayerName != world.Turn {
		return world
	}

	world.UpdateTimer(action.PlayerName)
	if world.Player1.Duration <= 0 {
		world.Turn = ""
		world.Winner = world.OtherPlayer(world.Player1.Name)
		return world
	}
	if world.Player2.Duration <= 0 {
		world.Turn = ""
		world.Winner = world.OtherPlayer(world.Player2.Name)
		return world
	}

	actionRes := Judge(action, world)
	log.Printf("Action with result: %s", actionRes)
	switch actionRes {
	case InvalidAction:
		return world
	case ValidAction:
		updateWorld(changeTurn)
	case ValidActionWithPrice:
		updateWorld(fixedTurn)
	case WinningAction:
		updateWorld(noTurn)
		world.Winner = action.PlayerName
	case LosingAction:
		updateWorld(noTurn)
		world.Winner = world.OtherPlayer(action.PlayerName)
	}

	world.SetLastAction()

	return world
}
