package main

import (
	"bitbucket.org/deimosgame/go-akadok/entity"
)

type World struct {
	players  []*entity.Player
	entities []*entity.Entity
}
