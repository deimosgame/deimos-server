package main

import (
	"bufio"
	"os"
	"strings"
)

// ParseCommands starts a separate routine to parse stdin (ATM) commands
func ParseCommands() {
	go Parser()
}

func Parser() {
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
	case "stop":
		HandleStopCommand(args)
	case "kick":
		HandleKickCommand(args)
	}
}

/**
 *  Command handlers
 */

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
			reason += reasonWord
		}
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
			reason += reasonWord
		}
	}

	Stop(reason)
}
