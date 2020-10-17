package main

import (
	"fmt"
	"os"
	"strings"
)

//commands needs to be lowercase
const (
	Reload    = "reload" //TODO(link) implement
	Save      = "save"
	Alert     = "alert"     //TODO(link) implement
	Broadcast = "broadcast" //TODO(link) implement

	Nuke    = "nuke"    //TODO(link) implement
	Replace = "replace" //TODO(link) implement

	Ban       = "ban"       //TODO(link) implement
	Permaban  = "permaban"  //TODO(link) implement
	Shadowban = "shadowban" //TODO(link) implement
	Unban     = "unban"     //TODO(link) implement

	ChatBan      = "chatban"      //TODO(link) implement
	PermaChatBan = "permachatban" //TODO(link) implement
	UnChatBan    = "unchatban"    //TODO(link) implement
	ChatPurge    = "chatpurge"    //TODO(link) implement

	SetName    = "setname"    //TODO(link) implement
	UpdateName = "updatename" //TODO(link) implement
	AddRole    = "addrole"    //TODO(link) implement
	RemoveRole = "removerole" //TODO(link) implement
)

//TODO(link) find a better name
func StartCommands(canvas *Canvas) {
	var command string
	for true {
		command = ""
		fmt.Fscanln(os.Stdin, &command)
		HandleCommand(&command, canvas)
	}
}

func HandleCommand(line *string, canvas *Canvas) {
	args := strings.Split(*line, " ")

	switch command := strings.ToLower(args[0]); command {
	case Reload:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)
	case Save:
		err := saveCanvas(canvas)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Canvas could not be saved: %v\n", err)
			return
		}
		fmt.Fprintf(os.Stdout, "Canvas has been saved\n")
	case Alert:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case Broadcast:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case Nuke:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case Replace:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case Ban:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case Permaban:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case Shadowban:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case Unban:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case ChatBan:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case PermaChatBan:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case UnChatBan:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case ChatPurge:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case SetName, UpdateName:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case AddRole:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	case RemoveRole:
		fmt.Fprintf(os.Stderr, "%s has not been implemented yet\n", command)

	default:
		fmt.Fprintf(os.Stdout, "%s is not a command\n", command)
	}

}
