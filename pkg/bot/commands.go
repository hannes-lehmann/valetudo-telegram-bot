package bot

import (
	"fmt"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (bot *Bot) publishMyCommands() error {
	_, err := bot.telegramApi.Request(
		tgbotapi.NewSetMyCommands(
			tgbotapi.BotCommand{
				Command:     "clean",
				Description: "Clean everything or a specific room",
			},
			tgbotapi.BotCommand{
				Command:     "pause",
				Description: "Pause the robot",
			},
			tgbotapi.BotCommand{
				Command:     "stop",
				Description: "Stop the robot",
			},
			tgbotapi.BotCommand{
				Command:     "home",
				Description: "Make robot go home",
			},
			tgbotapi.BotCommand{
				Command:     "status",
				Description: "Get current status",
			},
		),
	)

	return fmt.Errorf("failed to set my commands: %w", err)
}

func (bot *Bot) handleCleanCommand(requesterId int64, args string) error {
	rooms, err := bot.getRooms()

	if err != nil {
		return err
	}

	if args != "" {
		if args == "all" {
			err := bot.robotApi.Start()
			if err != nil {
				return err
			}

			bot.telegramApi.Send(tgbotapi.NewMessage(requesterId, "✅ Cleaning all"))

			return nil
		}

		roomNames := strings.Split(args, ",")
		roomsToFind := map[string]bool{}

		for _, roomName := range roomNames {
			roomsToFind[roomName] = true
		}

		toClean := []string{}

		for _, layer := range *rooms {
			if roomsToFind[*layer.Metadata.Name] {
				toClean = append(toClean, *layer.Metadata.SegmentId)
				roomsToFind[*layer.Metadata.Name] = false
			}
		}

		anyRoomNotFound := false

		for roomName, notFound := range roomsToFind {
			if notFound {
				bot.telegramApi.Send(tgbotapi.NewMessage(requesterId, "❌ Room "+roomName+" not found"))
				anyRoomNotFound = true
			}
		}

		if anyRoomNotFound {
			return nil
		}

		err := bot.robotApi.CleanMapSegments(toClean, 1)
		if err != nil {
			return err
		}

		bot.telegramApi.Send(tgbotapi.NewMessage(requesterId, "🧹 Cleaning "+strings.Join(roomNames, ", ")))

		return nil
	}

	keyboardButtons := [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("💯 Everything", "clean_all")},
	}

	sort.Slice(*rooms, func(i, j int) bool {
		return strings.Compare(*((*rooms)[i].Metadata.Name), *((*rooms)[j].Metadata.Name)) < 0
	})

	for _, layer := range *rooms {
		keyboardButtons = append(
			keyboardButtons,
			[]tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(
					*layer.Metadata.Name,
					fmt.Sprintf("clean %s", *layer.Metadata.SegmentId),
				),
			},
		)
	}

	botMessage := tgbotapi.NewMessage(requesterId, "What do you want to clean?")
	botMessage.ParseMode = "MarkdownV2"

	botMessage.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: keyboardButtons,
	}

	bot.telegramApi.Send(botMessage)

	return nil
}

func (bot *Bot) handleStatusCommand(requesterId int64, args string) error {
	state, err := bot.getParsedState()

	if err != nil {
		return err
	}

	statusString := robotStatusEmoji(state.Status) + " " + localizeRobotStatus(state.Status)

	stateString := ""
	stateString += fmt.Sprintf(
		"%s\n🔋 *Battery:* %d%% \\(%s\\)",
		statusString,
		state.BatteryLevel,
		state.BatteryStatus,
	)

	if len(state.AttachedAttachments) > 0 {
		localizedAttachments := []string{}

		for _, attachment := range state.AttachedAttachments {
			localizedAttachments = append(localizedAttachments, localizeAttachmentType(attachment))
		}

		stateString += "\n⚙️ *Attachments:* " + strings.Join(localizedAttachments, ", ")
	}

	if state.OperationMode != "" {
		stateString += "\n🔧 *Mode:* " + localizeOperationMode(state.OperationMode)
	}

	if state.FanSpeed != "" {
		stateString += "\n🌀 *Fan speed:* " + localizeFanSpeed(state.FanSpeed)
	}

	if state.WaterGrade != "" {
		stateString += "\n💧 *Water grade:* " + localizeWaterGrade(state.WaterGrade)
	}

	keyboard := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🧹 Start cleaning", "clean"),
	)

	switch state.Status {
	case "idle":
		keyboard = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧹 Start cleaning", "clean"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Home", "home"),
		)
	case "cleaning":
		keyboard = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏸ Pause", "pause"),
			tgbotapi.NewInlineKeyboardButtonData("🛑 Stop", "stop"),
		)
	case "paused":
		keyboard = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧹 Resume", "start"),
			tgbotapi.NewInlineKeyboardButtonData("🛑 Stop", "stop"),
		)
	case "returning":
		keyboard = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛑 Stop", "stop"),
		)
	}

	fullState, err := bot.robotApi.GetRobotState()
	if err != nil {
		return err
	}

	mapImage := renderMap(&fullState.Map)
	mapMsg := tgbotapi.NewPhoto(requesterId, tgbotapi.FileBytes{
		Name:  "map.png",
		Bytes: mapImage,
	})

	mapMsg.Caption = stateString
	mapMsg.ParseMode = "MarkdownV2"
	mapMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard)

	_, err = bot.telegramApi.Send(mapMsg)

	if err != nil {
		return err
	}

	return nil
}
