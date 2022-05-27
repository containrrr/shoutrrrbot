// Package p contains an HTTP Cloud Function.
package p

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var groupChatNote string

// PostWebhook ...
func PostWebhook(w http.ResponseWriter, r *http.Request) {
	botToken := os.Getenv("BOT_API_TOKEN")
	gcpProject := os.Getenv("GCP_PROJECT")
	funcRegion := os.Getenv("FUNCTION_REGION")
	funcName := os.Getenv("K_SERVICE")
	if funcName == "" || gcpProject == "" || botToken == "" || funcRegion == "" {
		log.Print("Environment variables are missing!")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	webhookURL := fmt.Sprintf("https://%v-%v.cloudfunctions.net/%v", funcRegion, gcpProject, funcName)
	log.Println("Using webhook URL:", webhookURL)

	bot, err := tg.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	botName := "@" + bot.Self.UserName
	bot.Debug = true

	log.Printf("Authorized on account %s", botName)
	groupChatNote = fmt.Sprintf("To get the ID for a group chat, invite %v to the chat", botName)

	switch r.URL.RawQuery {
	case "register":
		log.Printf("Registering webhook...")
		wh, _ := tg.NewWebhook(webhookURL)
		if _, err := bot.Request(wh); err != nil {
			fmt.Fprint(w, "Failed to register webhook")
			log.Print(err)
		} else {
			fmt.Fprint(w, "Webhook registered")
		}
		return
	case "unregister":
		log.Printf("Unregistering webhook...")
		payload := tg.DeleteWebhookConfig{DropPendingUpdates: true}
		if _, err := bot.Request(payload); err != nil {
			fmt.Fprint(w, "Failed to unregister webhook")
			log.Print(err)
		} else {
			fmt.Fprint(w, "Webhook unregistered")
		}
		return
	default:
		update := tg.Update{}

		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			if err != io.EOF {
				log.Printf("JSON Error: %v", err)
			}
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if msg := update.Message; msg != nil {

			var reply *tg.MessageConfig

			isDirectedMessage := strings.HasPrefix(msg.Text, botName)
			isForwarded := msg.ForwardFrom != nil
			isPrivate := msg.Chat != nil && msg.Chat.Type == "private"

			if isPrivate {
				if isForwarded {
					if msg.ForwardFromChat != nil {
						reply = replyFromChat(msg, msg.ForwardFromChat)
					} else {
						reply = replyFromUser(msg, msg.ForwardFrom)
					}
				} else {
					reply_msg := tg.NewMessage(msg.From.ID, "To get the ID of a chat, forward a message to this chat or start your message with "+botName)
					reply = &reply_msg
				}
			}

			if isDirectedMessage {
				reply = replyFromChat(msg, msg.Chat)
			}

			if reply != nil {
				if _, err = bot.Send(reply); err != nil {
					log.Println("Error sending message:", err)
				}
			}

		}

		updateParts := getUpdateParts(&update)

		log.Println("Got update with:", updateParts)
	}

	fmt.Fprint(w, "OK")
}

func getUpdateParts(update *tg.Update) (updateParts []string) {
	if update.CallbackQuery != nil {
		updateParts = append(updateParts, "CallbackQuery")
	}
	if update.ChannelPost != nil {
		updateParts = append(updateParts, "ChannelPost")
	}
	if update.ChatJoinRequest != nil {
		updateParts = append(updateParts, "ChatJoinRequest")
	}
	if update.ChatMember != nil {
		updateParts = append(updateParts, "ChatMember")
	}
	if update.ChosenInlineResult != nil {
		updateParts = append(updateParts, "ChosenInlineResult")
	}
	if update.EditedChannelPost != nil {
		updateParts = append(updateParts, "EditedChannelPost")
	}
	if update.EditedMessage != nil {
		updateParts = append(updateParts, "EditedMessage")
	}
	if update.InlineQuery != nil {
		updateParts = append(updateParts, "InlineQuery")
	}
	if update.Message != nil {
		updateParts = append(updateParts, "Message")
	}
	if update.MyChatMember != nil {
		updateParts = append(updateParts, "MyChatMember")
	}
	if update.Poll != nil {
		updateParts = append(updateParts, "Poll")
	}
	if update.PollAnswer != nil {
		updateParts = append(updateParts, "PollAnswer")
	}
	if update.PreCheckoutQuery != nil {
		updateParts = append(updateParts, "PreCheckoutQuery")
	}
	if update.ShippingQuery != nil {
		updateParts = append(updateParts, "ShippingQuery")
	}
	return updateParts
}

func replyFromChat(msg *tg.Message, chat *tg.Chat) *tg.MessageConfig {
	return createReply(msg, chat.ID, chat.Type, chat.Title, chat.UserName, "")
}

func replyFromUser(msg *tg.Message, user *tg.User) *tg.MessageConfig {
	title := user.FirstName + " " + user.LastName

	return createReply(msg, user.ID, "user", title, user.UserName, groupChatNote)
}

func createReply(msg *tg.Message, chatID int64, chatType string, title string, username string, note string) *tg.MessageConfig {
	text := strings.Builder{}
	addKeyVal(&text, "Chat ID", chatID)
	addKeyVal(&text, "Type", chatType)
	addKeyVal(&text, "Title", title)
	addKeyVal(&text, "Username", username)
	if note != "" {
		text.WriteString("<i><b>Note:</b> ")
		text.WriteString(note)
		text.WriteString("</i>")
	}
	reply := tg.NewMessage(msg.From.ID, text.String())
	reply.ParseMode = "HTML"

	// For some reason, this causes "no such message" when referring to a message in a group chat
	// reply.ReplyToMessageID = msg.MessageID
	return &reply
}

func addKeyVal(sb *strings.Builder, key string, val interface{}) {
	sb.WriteString("<b>")
	sb.WriteString(key)
	sb.WriteString(": </b>")
	if val == "" {
		sb.WriteString("<i>none</i>\n")
	} else {
		sb.WriteString("<code>")
		sb.WriteString(fmt.Sprintf("%v", val))
		sb.WriteString("</code>\n")
	}
}
