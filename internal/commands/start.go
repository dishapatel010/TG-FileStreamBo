package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/tg"
)

func (m *command) LoadStart(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewCommand("start", start))
}

func start(ctx *ext.Context, u *ext.Update) error {
	chatId := u.EffectiveChat().GetID()
	peerChatId := ctx.PeerStorage.GetPeerById(chatId)

	// Ensure the user is valid
	if peerChatId.Type != int(storage.TypeUser) {
		return dispatcher.EndGroups
	}

	// Check if user is in the allowed list (Optional, part of your existing logic)
	if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, "You are not allowed to use this bot.", nil)
		return dispatcher.EndGroups
	}

	// Force subscription check using channel ID
	requiredChannelID := int64(-1002108741045) // Replace with your actual channel ID
	inputChannel := &tg.InputChannel{
		ChannelID:  requiredChannelID,
		AccessHash: 0, // Typically 0 for public channels
	}

	// Create an InputUser object for the user
	inputUser := &tg.InputUser{
		UserID:     peerChatId.ID,
		AccessHash: peerChatId.AccessHash,
	}

	// Call the channels.getParticipant method
	response, err := ctx.Raw.ChannelsGetParticipant(ctx, &tg.ChannelsGetParticipantRequest{
		Channel: inputChannel,
		Participant: &tg.InputPeerUser{
			UserID:     inputUser.UserID,
			AccessHash: inputUser.AccessHash,
		},
	})
	if err != nil {
		ctx.Reply(u, "Failed to verify subscription. Please make sure you are subscribed to our channel.", nil)
		return dispatcher.EndGroups
	}

	// Check subscription status
	if _, ok := response.Participant.(*tg.ChannelParticipant); ok {
		// User is subscribed
		ctx.Reply(u, "Hi, send me any file to get a direct streamable link to that file.", nil)
		return dispatcher.EndGroups
	}

	// User is not subscribed
	ctx.Reply(u, "You must join our channel to use this bot: https://t.me/your_channel_username", nil)
	return dispatcher.EndGroups
}
