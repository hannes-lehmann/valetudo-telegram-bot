package bot

import (
	"fmt"
	"sort"
	"strings"

	"github.com/SkaceKamen/valetudo-telegram-bot/pkg/valetudo_map_renderer"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (bot *Bot) publishMyCommands() error {
	baseCommands := []tgbotapi.BotCommand{
		{
			Command:     "clean",
			Description: "Clean everything or a specific room",
		},
		{
			Command:     "pause",
			Description: "Pause the robot",
		},
		{
			Command:     "stop",
			Description: "Stop the robot",
		},
		{
			Command:     "home",
			Description: "Make robot go home",
		},
		{
			Command:     "status",
			Description: "Get current status",
		},
	}

	if bot.HasCapability("OperationModeControlCapability") {
		baseCommands = append(
			baseCommands,
			tgbotapi.BotCommand{
				Command:     "mode",
				Description: "Set operation mode",
			},
		)
	}

	if bot.HasCapability("FanSpeedControlCapability") {
		baseCommands = append(
			baseCommands,
			tgbotapi.BotCommand{
				Command:     "fan",
				Description: "Set fan speed",
			},
		)
	}

	if bot.HasCapability("WaterUsageControlCapability") {
		baseCommands = append(
			baseCommands,
			tgbotapi.BotCommand{
				Command:     "water",
				Description: "Set water grade",
			},
		)
	}

	_, err := bot.telegramApi.Request(
		tgbotapi.NewSetMyCommands(
			baseCommands...,
		),
	)

	if err != nil {
		return fmt.Errorf("failed to set my commands: %w", err)
	}

	return err
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

	mapImage := valetudo_map_renderer.RenderMap(&fullState.Map)
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

func (bot *Bot) handleModeCommand(requesterId int64, args string) error {
	if args == "" {
		return bot.sendModeKeyboard(requesterId)
	}

	err := bot.robotApi.SetOperationModeControlCapabilityPreset(args)
	if err != nil {
		return err
	}

	bot.telegramApi.Send(tgbotapi.NewMessage(requesterId, "✅ Mode set to "+localizeOperationMode(args)))

	return nil
}

func (bot *Bot) sendModeKeyboard(requesterId int64) error {
	modes, err := bot.robotApi.GetOperationModeControlCapabilityPresets()

	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(requesterId, "⚙️ Set Mode to:")

	keyboard := [][]tgbotapi.InlineKeyboardButton{}
	for _, mode := range *modes {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(operationModeEmoji(mode)+" "+localizeOperationMode(mode), "mode "+mode),
		))
	}

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	_, err = bot.telegramApi.Send(msg)

	return err
}

func (bot *Bot) handleFanCommand(requesterId int64, args string) error {
	if args == "" {
		return bot.sendFanKeyboard(requesterId)
	}

	err := bot.robotApi.SetFanSpeedControlCapabilityPreset(args)
	if err != nil {
		return err
	}

	bot.telegramApi.Send(tgbotapi.NewMessage(requesterId, "✅ Fan speed set to "+localizeFanSpeed(args)))

	return nil
}

func (bot *Bot) sendFanKeyboard(requesterId int64) error {
	fans, err := bot.robotApi.GetFanSpeedControlCapabilityPresets()

	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(requesterId, "🌀 Set Fan speed to:")

	keyboard := [][]tgbotapi.InlineKeyboardButton{}
	for _, fan := range *fans {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(localizeFanSpeed(fan), "fan "+fan),
		))
	}

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	_, err = bot.telegramApi.Send(msg)

	return err
}

func (bot *Bot) handleWaterCommand(requesterId int64, args string) error {
	if args == "" {
		return bot.sendWaterKeyboard(requesterId)
	}

	err := bot.robotApi.SetWaterUsageControlCapabilityPreset(args)
	if err != nil {
		return err
	}

	bot.telegramApi.Send(tgbotapi.NewMessage(requesterId, "✅ Water grade set to "+localizeWaterGrade(args)))

	return nil
}

func (bot *Bot) sendWaterKeyboard(requesterId int64) error {
	waters, err := bot.robotApi.GetWaterUsageControlCapabilityPresets()

	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(requesterId, "💧 Set Water usage level to:")

	keyboard := [][]tgbotapi.InlineKeyboardButton{}
	for _, water := range *waters {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(localizeWaterGrade(water), "water "+water),
		))
	}

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	_, err = bot.telegramApi.Send(msg)

	return err
}
