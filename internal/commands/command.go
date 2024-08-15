package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

var commands []*discordgo.ApplicationCommand
var commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)

func init() {
	commands = append(commands)
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){}
}

func registerHandler(s *discordgo.Session) {
	log.Println("Registering Command Handles")
	s.AddHandler(func(c *discordgo.Session, interaction *discordgo.InteractionCreate) {
		handle, ok := commandHandlers[interaction.ApplicationCommandData().Name]
		if ok {
			handle(s, interaction)
		}
	})
}

func Unregister(s *discordgo.Session, guildId *string, registeredCommands []*discordgo.ApplicationCommand) {
	log.Println("Unregister Commands")
	for _, command := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, *guildId, command.ID)
		if err != nil {
			log.Fatalf("Failed to delete '%v' command: %v", command.Name, err)
			break
		}
	}
}

func Register(s *discordgo.Session, guildId *string) {
	log.Println("Registering Commands")
	registerCommands := make([]*discordgo.ApplicationCommand, len(commands))
	isFailed := false
	for i, command := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *guildId, command)
		if err != nil {
			log.Fatalf("Failed to create '%v' command: %v", command.Name, err)
			log.Fatal("Executing Rollback")
			isFailed = true
			break
		}
		registerCommands[i] = cmd
	}

	if isFailed {
		Unregister(s, guildId, registerCommands)
		return
	}
	registerHandler(s)
	defer log.Println("Registering Commands Successfully")
}
