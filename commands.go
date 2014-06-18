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
}

// RegisterCommandHandler saves command handlers into a dedicated map
func RegisterCommandHandler(command string, handler interface{}) {
	CommandHandlers[command] = handler
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

	for _, currentPlayer := range players {
		if currentPlayer.Match(args[0]) {
			currentPlayer.Kick(reason)
			return "Kicked " + currentPlayer.Name
		}
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
