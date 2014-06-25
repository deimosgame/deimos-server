package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var (
	CommandHandlers = make(map[string]interface{})
)

// CommandParser is the routine that parses stdin commands
func CommandParser() {
	input := bufio.NewReader(os.Stdin)
	for {
		line, _, err := input.ReadLine()
		if err != nil {
			HandleStopCommand([]string{}, nil)
		}
		if string(line) == "" {
			continue
		}
		HandleCommand(string(line), nil)
	}
}

// SetupCommandHandlers registers all necessary command handlers
func SetupCommandHandlers() {
	RegisterCommandHandler("stop", HandleStopCommand)
	RegisterCommandHandler("kick", HandleKickCommand)
	RegisterCommandHandler("config", HandleConfigCommand)
	RegisterCommandHandler("send", HandleSendCommand)
	RegisterCommandHandler("op", HandleOpCommand)
	RegisterCommandHandler("deop", HandleDeopCommand)
	RegisterCommandHandler("players", HandlePlayersCommand)
	RegisterCommandHandler("godmode", HandleGodmodeCommand)

	AllowClientCommand("debug")
	AllowClientCommand("noclip")
	AllowClientCommand("kill")
}

// RegisterCommandHandler saves command handlers into a dedicated map
func RegisterCommandHandler(command string, handler interface{}) {
	CommandHandlers[command] = handler
}

// AllowClientCommand allows some commands to be interpreted by the client
// instead of being executed by the server
func AllowClientCommand(command string) {
	CommandHandlers[command] = nil
}

func HandleCommand(rawArgs string, sender *Player) {
	// Special splitting to ignore quotes
	splitArgs := strings.Split(rawArgs, " ")
	args, toReunite := make([]string, 0), false
	for _, currentArg := range splitArgs {
		if toReunite {
			args[len(args)-1] = args[len(args)-1] + currentArg
		} else {
			args = append(args, currentArg)
		}
		if strings.Contains(currentArg, `"`) {
			toReunite = !toReunite
		}
	}

	handler, ok := CommandHandlers[args[0]]
	if !ok {
		if sender == nil {
			log.Notice("This command does not exists!")
		} else {
			sender.SendMessage("This command does not exists!")
		}
		return
	}

	// Commands allowed with AllowClientCommand()
	if handler == nil {
		return
	}

	if !h.Player.IsOperator() {
		h.Player.SendMessage("You are not allowed to run commands on the " +
			"server.")
		return
	}

	result := (handler.(func([]string, *Player) string))(args[1:], sender)

	// Send command result back to the sender
	if sender == nil {
		log.Notice(result)
	} else {
		sender.SendMessage(result)
	}
}

/**
 *  Command handlers (convention: HandleXCommand)
 */

func HandleConfigCommand(args []string, p *Player) string {
	if len(args) == 0 {
		return `config: Gets/changes the config of the server\n
Usage: config <item> [value]`
	} else if len(args) == 1 {
		val, err := GetConfigItem(args[0])
		if err != nil {
			return "Unknown config item!"
		}
		return fmt.Sprintf("%s=%s", args[0], val)
	}
	return "Not yet implemented."
}

// HandleKickCommand handles a player kick command
// Usage: kick <*|player> [reason]
func HandleKickCommand(args []string, p *Player) string {
	if len(args) == 0 {
		return `kick: Kicks a player from the server
Usage: kick <*|player> [reason]`
	}

	reason := ""
	if len(args) > 1 {
		for i, reasonWord := range args {
			if i == 0 {
				continue
			}
			reason += " " + reasonWord
		}
		reason = reason[1:]
	}

	for _, currentPlayer := range MatchPlayers(args[0]) {
		currentPlayer.Kick(reason)
		return "Kicked " + currentPlayer.Name
	}
	return "Couldn't find " + args[0] + "."
}

// HandleStopCommand handles the stop commands and its arguments
// Usage: stop [reason]
func HandleStopCommand(args []string, p *Player) string {
	reason := ""
	if len(args) > 0 {
		for i, reasonWord := range args {
			if i == 0 {
				continue
			}
			reason += " " + reasonWord
		}
		reason = reason[1:]
	}

	Stop(reason)
	return ""
}

// HanldeSendCommand handles /send, which displays a message as the following:
//  > [message]
//  It is used mainly by the console to comunicate with players
func HandleSendCommand(args []string, p *Player) string {
	SendMessage("> " + strings.Join(args, " "))
	return ""
}

func HandleOpCommand(args []string, p *Player) string {
	if len(args) != 1 {
		return `Invalid command arguments.
Usage: /op <player>`
	}
	players := MatchPlayers(args[0])
	if len(players) == 0 {
		return "No corresponding player has been found"
	}
	if len(players) > 1 {
		return `Your search has matched more than one player.
Please specify a single player for safety reasons.`
	}
	config.Operators = append(config.Operators, players[0].Account)
	players[0].SendMessage("You are now a server operator.")
	return players[0].Name + " has been granted operator powers."
}

func HandleDeopCommand(args []string, p *Player) string {
	if len(args) != 1 {
		return `Invalid command arguments.
Usage: /deop <player>`
	}
	players := MatchPlayers(args[0])
	if len(players) == 0 {
		return "No corresponding player has been found"
	}
	if len(players) > 1 {
		return `Your search has matched more than one player.
Please specify a single player for safety reasons.`
	}
	if !players[0].IsOperator() {
		return players[0].Name + " is not currently an operator."
	}
	newOperators := make([]string, 0)
	for _, currentOperator := range config.Operators {
		if strings.ToLower(currentOperator) ==
			strings.ToLower(players[0].Account) {
			continue
		}
		newOperators = append(newOperators, currentOperator)
	}
	config.Operators = newOperators
	players[0].SendMessage("You are not a server operator anymore.")
	return players[0].Name + " has lost his operator powers."
}

func HandlePlayersCommand(args []string, p *Player) string {
	playerList := ""
	for _, currentPlayer := range players {
		playerList += " " + currentPlayer.Name
	}
	if len(playerList) == 0 {
		return "No player is online."
	} else {
		return "Online players: " + playerList[1:]
	}
}

func HandleGodmodeCommand(args []string, p *Player) string {
	if len(args) == 0 && p != nil {
		if p.Godmode {
			p.Godmode = false
			return "God mode has been disabled"
		}
		p.Godmode = true
		return "God mode has been enabled"
	} else if len(args) == 0 {
		return "You cannot give god mode to the console."
	} else if len(args) > 1 {
		return "You can only give to this command zero or one arguments"
	}
	pl := MatchPlayers(args[0])
	if len(pl) == 0 {
		return "No player was found!"
	}
	for _, currentPlayer := range pl {
		if currentPlayer.Godmode {
			currentPlayer.Godmode = false
			currentPlayer.SendMessage("Your god mode has been disabled")
		} else {
			currentPlayer.Godmode = true
			currentPlayer.SendMessage("God mode was enabled for you")
		}
		p.SendMessage("Toggled god mode for " + currentPlayer.Name)
	}
	return ""
}
