package main

import (
	"bufio"
	"os"
	"strings"
)

// CommandParser is the routine that parses stdin commands
func CommandParser() {
	input := bufio.NewReader(os.Stdin)
	for {
		line, _, err := input.ReadLine()
		if err != nil {
			log.Error("Command parsing error")
			continue
		}
		commandLine := string(line)
		splittedLine := strings.Split(commandLine, " ")
		HandleCommand(splittedLine[0], splittedLine[1:])
	}
}

// HandleCommand handles the commands and their arguments
func HandleCommand(command string, args []string) {
	switch command {
	case "config":
		HandleConfigCommand(args)
	case "stop":
		HandleStopCommand(args)
	case "kick":
		HandleKickCommand(args)
	}
}

/**
 *  Command handlers (convention: HandleXCommand)
 */

func HandleConfigCommand(args []string) {
	if len(args) == 0 {
		log.Info("config: Gets/changes the config of the server")
		log.Info("	Usage: config <item> [value]")
		return
	} else if len(args) == 1 {
		val, err := GetConfigItem(args[0])
		if err != nil {
			log.Error("Unknown config item!")
			return
		}
		log.Info(args[0] + ": " + val)
	}
}

// HandleKickCommand handles a player kick command
// Usage: kick <*|player> [reason]
func HandleKickCommand(args []string) {
	if len(args) == 0 {
		log.Info("kick: Kicks a player from the server")
		log.Info("	Usage: kick <*|player> [reason]")
		return
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
			log.Info("Kicked " + currentPlayer.Name)
		}
	}
}

// HandleStopCommand handles the stop commands and its arguments
// Usage: stop [reason]
func HandleStopCommand(args []string) {
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
}
