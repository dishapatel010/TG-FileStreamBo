package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"
	"fmt"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/celestix/gotgproto/tg"
)

func (m *command) LoadStart(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewCommand("start", start))
}

func start(ctx *ext.Context, u *ext.Update) error {
	chatId := u.EffectiveChat().GetID()
	peerChatId := ctx.PeerStorage.GetPeerById(chatId)

	// Check if the peer is a user
	if peerChatId.Type != int(storage.TypeUser) {
		return dispatcher.EndGroups
	}

	// Check if the user is in the allowed list
	if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, "You are not allowed to use this bot.", nil)
		return dispatcher.EndGroups
	}

	// Channel ID where the bot checks for membership
	const channelId int64 = -1002108741045

	// Check if the user is a member of the required channel
	member, err := GetChannelMember(ctx, channelId, chatId)
	if err != nil {
		ctx.Logger().Error("Failed to get channel member", "error", err)
		ctx.Reply(u, "Could not verify your channel membership. Please try again later.", nil)
		return dispatcher.EndGroups
	}
	if member == nil {
		ctx.Reply(u, "You must be a member of the required channel to use this bot.", nil)
		return dispatcher.EndGroups
	}

	// Successful response
	ctx.Reply(u, "Hi, send me any file to get a direct streamable link to that file.", nil)
	return dispatcher.EndGroups
}

// GetChannelMember checks if a user is a member of a specific channel or supergroup
func GetChannelMember(ctx *ext.Context, chatId, userId int64) (*tg.ChannelsChannelParticipant, error) {
	// Retrieve channel peer
	channel := ctx.PeerStorage.GetInputPeerById(chatId)
	peerChannel, ok := channel.(*tg.InputPeerChannel)
	if !ok {
		return nil, fmt.Errorf("unsupported peer type %T", channel)
	}

	// Retrieve user peer
	user := ctx.PeerStorage.GetInputPeerById(userId)

	// Perform the request
	return ctx.Raw.ChannelsGetParticipant(
		ctx,
		&tg.ChannelsGetParticipantRequest{
			Channel: &tg.InputChannel{
				ChannelID:  peerChannel.ChannelID,
				AccessHash: peerChannel.AccessHash,
			},
			Participant: user,
		},
	)
}
