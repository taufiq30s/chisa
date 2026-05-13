package handlers

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/utils"
)

// Register creates all Discord application commands and registers all event
// and interaction handlers. It calls wg.Done when complete.
func (r *Registry) Register(wg *sync.WaitGroup, guildId string) {
	defer wg.Done()
	utils.InfoLog.Println("Registering Handler")
	defer utils.InfoLog.Println("Registering Handlers Successfully")

	r.registerCommand(guildId)
	r.registerCommandHandlers()
	r.registerEvents()
}

// Unregister removes all Discord application commands previously registered.
func (r *Registry) Unregister(guildId string) {
	utils.InfoLog.Println("Unregistering Handler")
	defer utils.InfoLog.Println("Unregistering Handler Successfully")

	r.commands = r.getCommands(guildId)
	r.unregisterCommands(guildId, r.commands)
}

func (r *Registry) registerEvents() {
	utils.InfoLog.Println("Registering Events")
	for _, handler := range r.eventHandlers {
		fn := handler(r.bot)
		r.bot.Session.AddHandler(fn)
	}
}

func (r *Registry) lookupButtonHandler(id string) (componentFunction, bool) {
	for key := range r.buttonHandlers {
		if len(id) >= len(key) && id[:len(key)] == key {
			return r.buttonHandlers[key], true
		}
	}
	return nil, false
}

func (r *Registry) lookupSelectHandler(id string) (componentFunction, bool) {
	for key := range r.selectHandlers {
		if len(id) >= len(key) && id[:len(key)] == key {
			return r.selectHandlers[key], true
		}
	}
	return nil, false
}

func (r *Registry) registerCommandHandlers() {
	utils.InfoLog.Println("Registering Command Handlers")
	r.bot.Session.AddHandler(func(c *discordgo.Session, interaction *discordgo.InteractionCreate) {
		switch interaction.Type {
		case discordgo.InteractionApplicationCommand:
			if handle, ok := r.commandHandlers[interaction.ApplicationCommandData().Name]; ok {
				handle(r.bot, interaction)
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
			if handle, ok := r.commandAutofillHandlers[interaction.ApplicationCommandData().Name]; ok {
				handle(r.bot, interaction)
			}
		case discordgo.InteractionMessageComponent:
			switch interaction.MessageComponentData().ComponentType {
			case discordgo.ButtonComponent:
				if handle, ok := r.lookupButtonHandler(interaction.MessageComponentData().CustomID); ok {
					handle(r.bot.Session, interaction)
				}
			case discordgo.SelectMenuComponent:
				if handle, ok := r.lookupSelectHandler(interaction.MessageComponentData().CustomID); ok {
					handle(r.bot.Session, interaction)
				}
			}
		}
	})
}

func (r *Registry) registerCommand(guildId string) {
	utils.InfoLog.Println("Registering Commands")
	registerCommands := make([]*discordgo.ApplicationCommand, len(r.commands))
	isFailed := false
	for i, command := range r.commands {
		cmd, err := r.bot.Session.ApplicationCommandCreate(r.bot.Session.State.User.ID, guildId, command)
		if err != nil {
			utils.ErrorLog.Printf("Failed to create '%v' command: %v\n", command.Name, err)
			isFailed = true
			break
		}
		registerCommands[i] = cmd
	}

	if isFailed {
		utils.InfoLog.Println("Executing Rollback")
		r.unregisterCommands(guildId, registerCommands)
		return
	}
	r.commands = registerCommands
}

