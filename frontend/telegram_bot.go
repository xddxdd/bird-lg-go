package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type tgChat struct {
	ID int64 `json:"id"`
}

type tgMessage struct {
	MessageID int64  `json:"message_id"`
	Chat      tgChat `json:"chat"`
	Text      string `json:"text"`
}

type tgWebhookRequest struct {
	Message tgMessage `json:"message"`
}

type tgWebhookResponse struct {
	Method           string `json:"method"`
	ChatID           int64  `json:"chat_id"`
	Text             string `json:"text"`
	ReplyToMessageID int64  `json:"reply_to_message_id"`
	ParseMode        string `json:"parse_mode"`
}

func telegramIsCommand(message string, command string) bool {
	b := false
	b = b || strings.HasPrefix(message, "/"+command+"@")
	b = b || strings.HasPrefix(message, "/"+command+" ")
	b = b || message == "/"+command
	return b
}

func telegramDefaultPostProcess(s string) string {
	return strings.TrimSpace(s)
}

func telegramBatchRequestFormat(servers []string, endpoint string, command string, postProcess func(string) string) string {
	results := batchRequest(servers, endpoint, command)
	result := ""
	for i, r := range results {
		result += servers[i] + "\n"
		result += postProcess(r) + "\n\n"
	}
	return result
}

func webHandlerTelegramBot(w http.ResponseWriter, r *http.Request) {
	// Parse only needed fields of incoming JSON body
	var err error
	var request tgWebhookRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		println(err.Error())
		return
	}

	// Do not respond if not a tg Bot command (starting with /)
	if len(request.Message.Text) == 0 || request.Message.Text[0] != '/' {
		return
	}

	// Select only one server based on webhook URL
	var servers []string
	if len(r.URL.Path[len("/telegram/"):]) == 0 {
		servers = setting.servers
	} else {
		servers = strings.Split(r.URL.Path[len("/telegram/"):], "+")
	}

	// Parse target
	target := ""
	if strings.Contains(request.Message.Text, " ") {
		target = strings.Join(strings.Split(request.Message.Text, " ")[1:], " ")
	}

	// Execute command
	commandResult := ""

	// - traceroute
	if telegramIsCommand(request.Message.Text, "trace") || telegramIsCommand(request.Message.Text, "trace4") {
		commandResult = telegramBatchRequestFormat(servers, "traceroute", target, telegramDefaultPostProcess)
	} else if telegramIsCommand(request.Message.Text, "trace6") {
		commandResult = telegramBatchRequestFormat(servers, "traceroute6", target, telegramDefaultPostProcess)

	} else if telegramIsCommand(request.Message.Text, "route") || telegramIsCommand(request.Message.Text, "route4") {
		commandResult = telegramBatchRequestFormat(servers, "bird", "show route for "+target+" primary", telegramDefaultPostProcess)
	} else if telegramIsCommand(request.Message.Text, "route6") {
		commandResult = telegramBatchRequestFormat(servers, "bird6", "show route for "+target+" primary", telegramDefaultPostProcess)

	} else if telegramIsCommand(request.Message.Text, "path") || telegramIsCommand(request.Message.Text, "path4") {
		commandResult = telegramBatchRequestFormat(servers, "bird", "show route for "+target+" all primary", func(result string) string {
			for _, s := range strings.Split(result, "\n") {
				if strings.Contains(s, "BGP.as_path: ") {
					return strings.TrimSpace(strings.Split(s, ":")[1])
				}
			}
			return ""
		})
	} else if telegramIsCommand(request.Message.Text, "path6") {
		commandResult = telegramBatchRequestFormat(servers, "bird6", "show route for "+target+" all primary", func(result string) string {
			for _, s := range strings.Split(result, "\n") {
				if strings.Contains(s, "BGP.as_path: ") {
					return strings.TrimSpace(strings.Split(s, ":")[1])
				}
			}
			return ""
		})

	} else if telegramIsCommand(request.Message.Text, "whois") {
		if setting.netSpecificMode == "dn42" {
			targetNumber, err := strconv.ParseUint(target, 10, 64)
			if err == nil {
				if targetNumber < 10000 {
					targetNumber += 4242420000
					target = "AS" + strconv.FormatUint(targetNumber, 10)
				} else {
					target = "AS" + target
				}
			}
		}
		tempResult := whois(target)
		if setting.netSpecificMode == "dn42" {
			commandResult = dn42WhoisFilter(tempResult)
		} else {
			commandResult = tempResult
		}

	} else if telegramIsCommand(request.Message.Text, "help") {
		commandResult = `
/[path|path6] <IP>
/[route|route6] <IP>
/[trace|trace6] <IP>
/whois <Target>
		`
	}

	commandResult = strings.TrimSpace(commandResult)
	if len(commandResult) > 0 {
		// Create a JSON response
		w.Header().Add("Content-Type", "application/json")
		response := &tgWebhookResponse{
			Method:           "sendMessage",
			ChatID:           request.Message.Chat.ID,
			Text:             "```\n" + strings.TrimSpace(commandResult) + "\n```",
			ReplyToMessageID: request.Message.MessageID,
			ParseMode:        "Markdown",
		}
		data, err := json.Marshal(response)
		if err != nil {
			println(err.Error())
			return
		}
		// println(string(data))
		w.Write(data)
	}
}
