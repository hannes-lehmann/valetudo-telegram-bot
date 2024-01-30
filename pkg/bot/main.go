package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/SkaceKamen/valetudo-telegram-bot/pkg/valetudo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	robotApi    *valetudo.ValetudoClient
	telegramApi *tgbotapi.BotAPI
	chatIds     []int64

	/** capabilities supported by the robot */
	capabilities []string
}

func NewBot(robotApi *valetudo.ValetudoClient, telegramApi *tgbotapi.BotAPI) Bot {
	return Bot{robotApi: robotApi, telegramApi: telegramApi}
}

func (bot *Bot) AddUserId(id int64) {
	bot.chatIds = append(bot.chatIds, id)
}

func (bot *Bot) Start() error {
	capabilities, err := bot.robotApi.GetRobotCapabilities()

	if err != nil {
		return fmt.Errorf("failed to get robot capabilities: %w", err)
	}

	bot.capabilities = *capabilities

	err = bot.publishMyCommands()

	if err != nil {
		return fmt.Errorf("failed to publish commands: %w", err)
	}

	go func() {
		err = bot.listenToStateChanges()
		if err != nil {
			log.Println(fmt.Errorf("failed to listen to state changes: %w", err))
		}
	}()

	err = bot.listenToMessages()
	if err != nil {
		return fmt.Errorf("listening for new messages failed: %w", err)
	}

	return nil
}

func (bot *Bot) listenToStateChanges() error {
	lastState, err := bot.getParsedState()

	if err != nil {
		return err
	}

	for {
		log.Println("Listening for state changes...")

		err = bot.robotApi.ListenToStateAttributesChanges(func(state *[]valetudo.RobotStateAttribute, err error) {
			if err != nil {
				log.Println(err)
				return
			}

			parsed := stateObjToData(state)

			log.Println("Received state, status: ", parsed.Status, " batteryStatus:", parsed.BatteryStatus, " batteryLevel:", parsed.BatteryLevel)

			if lastState.BatteryStatus != parsed.BatteryStatus {
				bot.handleBatteryStatusChange(lastState, parsed)
			}

			if lastState.Status != parsed.Status {
				bot.handleStatusChange(lastState, parsed)
			}

			lastState = parsed
		})

		if err != nil {
			log.Println(err)
		}
	}
}

func (bot *Bot) handleStatusChange(previous *CurrentState, new *CurrentState) {
	for _, user := range bot.chatIds {
		newStatusLabel := localizeRobotStatus(new.Status)
		newStatusIcon := robotStatusEmoji(new.Status)
		statusMessage := newStatusIcon + " " + newStatusLabel

		// Special status transitions that aren't actually a separate statuses
		switch new.Status {
		case "returning":
			if previous.Status == "cleaning" {
				statusMessage = "✅ Cleaning complete, returning home"
			}
		}

		bot.Send(user, statusMessage)
	}
}

func (bot *Bot) handleBatteryStatusChange(previous *CurrentState, new *CurrentState) {
	for _, user := range bot.chatIds {
		statusMessage := ""

		switch new.BatteryStatus {
		case "charging":
			statusMessage = fmt.Sprintf("🪫 Charging battery from %d %%", new.BatteryLevel)
		case "charged":
			statusMessage = "🔋 Battery fully charged"
		}

		if statusMessage != "" {
			bot.Send(user, statusMessage)
		}
	}
}

func (bot *Bot) isAllowedUserId(id int64) bool {
	for _, allowedId := range bot.chatIds {
		if allowedId == id {
			return true
		}
	}

	return false
}

func (bot *Bot) listenToMessages() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.telegramApi.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			if !bot.isAllowedUserId(update.CallbackQuery.Message.Chat.ID) {
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "You are not allowed to do that")
				if _, err := bot.telegramApi.Request(callback); err != nil {
					log.Println(err)
				}

				continue
			}

			data := strings.Split(update.CallbackQuery.Data, " ")
			switch data[0] {
			case "pause":
				err := bot.robotApi.Pause()
				if err != nil {
					log.Println(err)
					bot.Send(update.CallbackQuery.Message.Chat.ID, "❌ Error pausing robot: "+err.Error())
				} else {
					callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "⏸ Paused")
					if _, err := bot.telegramApi.Request(callback); err != nil {
						log.Println(err)
					}
				}

			case "stop":
				err := bot.robotApi.Stop()
				if err != nil {
					log.Println(err)
					bot.Send(update.CallbackQuery.Message.Chat.ID, "❌ Error stopping robot: "+err.Error())
				} else {
					callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "⏹ Stopped")
					if _, err := bot.telegramApi.Request(callback); err != nil {
						log.Println(err)
					}
				}

			case "home":
				err := bot.robotApi.Home()
				if err != nil {
					log.Println(err)
					bot.Send(update.CallbackQuery.Message.Chat.ID, "❌ Error sending robot home: "+err.Error())
				} else {
					callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "🏠 Going home")
					if _, err := bot.telegramApi.Request(callback); err != nil {
						log.Println(err)
					}
				}

			case "mode":
				bot.handleOneTimeCallback(update.CallbackQuery, data[1:], func(query *tgbotapi.CallbackQuery, args []string) (string, error) {
					target := args[0]
					err := bot.robotApi.SetOperationModeControlCapabilityPreset(target)
					if err != nil {
						return "", err
					}

					return "✔️ Mode set to " + localizeOperationMode(target), nil
				})

			case "fan":
				bot.handleOneTimeCallback(update.CallbackQuery, data[1:], func(query *tgbotapi.CallbackQuery, args []string) (string, error) {
					targetSpeed := args[0]
					err := bot.robotApi.SetFanSpeedControlCapabilityPreset(targetSpeed)
					if err != nil {
						return "", err
					}

					return "✔️ Fan speed set to " + localizeFanSpeed(targetSpeed), nil
				})

			case "water":
				bot.handleOneTimeCallback(update.CallbackQuery, data[1:], func(query *tgbotapi.CallbackQuery, args []string) (string, error) {
					target := args[0]
					err := bot.robotApi.SetWaterUsageControlCapabilityPreset(target)
					if err != nil {
						return "", err
					}

					return "✔️ Water usage set to " + localizeWaterGrade(target), nil
				})

			case "clean":
				if len(data) < 2 {
					bot.handleCleanCommand(update.CallbackQuery.Message.Chat.ID, "")
					callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "You need to pick what")
					if _, err := bot.telegramApi.Request(callback); err != nil {
						log.Println(err)
					}

					continue
				}

				err := bot.robotApi.CleanMapSegments([]string{data[1]}, 1)

				if err != nil {
					bot.Send(update.CallbackQuery.Message.Chat.ID, "❌ Error cleaning room: "+err.Error())
					log.Println(err)
				} else {
					roomName := data[1]
					rooms, err := bot.getRooms()

					if err == nil {
						for _, room := range *rooms {
							if *room.Metadata.SegmentId == data[1] {
								roomName = *room.Metadata.Name
								break
							}
						}
					} else {
						log.Println(err)
					}

					bot.telegramApi.Request(
						tgbotapi.EditMessageTextConfig{
							BaseEdit: tgbotapi.BaseEdit{
								ChatID:      update.CallbackQuery.Message.Chat.ID,
								MessageID:   update.CallbackQuery.Message.MessageID,
								ReplyMarkup: nil,
							},
							Text: "🧹 Cleaning " + roomName,
						},
					)
				}
			}

			continue
		}

		if update.Message == nil {
			continue
		}

		if !bot.isAllowedUserId(update.Message.From.ID) {
			bot.Send(update.Message.From.ID, "⚠️ You're not allowed to access this bot. Your ID: "+fmt.Sprintf("%d", update.Message.From.ID))

			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		switch update.Message.Command() {
		case "start":
			bot.Send(update.CallbackQuery.Message.Chat.ID, "👋 I'm ready, /status or /clean")
		case "status":
			err := bot.handleStatusCommand(update.Message.Chat.ID, update.Message.CommandArguments())
			if err != nil {
				log.Println(err)
				bot.Send(update.Message.Chat.ID, "❌ Error fetching status: "+err.Error())
			}
		case "stop":
			err := bot.robotApi.Stop()
			if err != nil {
				log.Println(err)
				bot.Send(update.Message.Chat.ID, "❌ Error stopping robot: "+err.Error())
			}
		case "home":
			err := bot.robotApi.Home()
			if err != nil {
				log.Println(err)
				bot.Send(update.Message.Chat.ID, "❌ Error sending robot home: "+err.Error())
			}
		case "pause":
			err := bot.robotApi.Pause()
			if err != nil {
				log.Println(err)
				bot.Send(update.Message.Chat.ID, "❌ Error pausing robot: "+err.Error())
			}
		case "clean":
			err := bot.handleCleanCommand(update.Message.Chat.ID, update.Message.CommandArguments())
			if err != nil {
				log.Println(err)
				bot.Send(update.Message.Chat.ID, "❌ Error cleaning: "+err.Error())
			}
		case "mode":
			err := bot.handleModeCommand(update.Message.Chat.ID, update.Message.CommandArguments())
			if err != nil {
				log.Println(err)
				bot.Send(update.Message.Chat.ID, "❌ Error setting mode: "+err.Error())
			}
		case "fan":
			err := bot.handleFanCommand(update.Message.Chat.ID, update.Message.CommandArguments())
			if err != nil {
				log.Println(err)
				bot.Send(update.Message.Chat.ID, "❌ Error setting fan speed: "+err.Error())
			}
		case "water":
			err := bot.handleWaterCommand(update.Message.Chat.ID, update.Message.CommandArguments())
			if err != nil {
				log.Println(err)
				bot.Send(update.Message.Chat.ID, "❌ Error setting water grade: "+err.Error())
			}
		}
	}

	return nil
}

func (bot *Bot) HasCapability(capability string) bool {
	for _, cap := range bot.capabilities {
		if cap == capability {
			return true
		}
	}

	return false
}
