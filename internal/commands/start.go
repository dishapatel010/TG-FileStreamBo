package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"
	"fmt"

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

	// Check if the chat type is a user
	if peerChatId.Type != int(storage.TypeUser) {
		return dispatcher.EndGroups
	}

	// Restrict access to allowed users
	if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, "You are not allowed to use this bot.", &ext.ReplyOpts{})
		return dispatcher.EndGroups
	}

	// Check if the user is a participant of a required channel/group
	forceSubChannelID := int64(2108741045) // Replace with your actual channel ID
	isParticipant, err := isParticipant(ctx, forceSubChannelID, chatId)
	if err != nil {
		ctx.Reply(u, "An error occurred while checking your subscription status. Please try again later.", &ext.ReplyOpts{})
		return dispatcher.EndGroups
	}

	if !isParticipant {
		ctx.Reply(u, "You need to subscribe to our channel to use this bot. [Subscribe here](https://t.me/your_channel_link)", &ext.ReplyOpts{})
		return dispatcher.EndGroups
	}

	// Send welcome message if all checks pass
	ctx.Reply(u, "Hi, send me any file to get a direct streamable link to that file.", &ext.ReplyOpts{})
	return dispatcher.EndGroups
}

func isParticipant(ctx *ext.Context, chatID, userID int64) (bool, error) {
    channelPeer := ctx.PeerStorage.GetPeerById(chatID)
    if channelPeer == nil {
        return false, fmt.Errorf("channel not found")
    }

    cp, err := ctx.Raw.ChannelsGetParticipant(ctx, &tg.ChannelsGetParticipantRequest{
        Channel: &tg.InputChannel{
            ChannelID:  channelPeer.ID,
            AccessHash: channelPeer.AccessHash,
        },
        Participant: ctx.PeerStorage.GetInputPeerById(userID),
    })
    if err != nil {
        return false, fmt.Errorf("failed to get participant: %v", err)
    }

    // Check participant type
    switch cp.GetParticipant().(type) {
    case *tg.ChannelParticipantLeft, *tg.ChannelParticipantBanned:
        return false, nil
    default:
        return true, nil
    }
}
