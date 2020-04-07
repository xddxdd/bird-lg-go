package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type tgChat struct {
	ID int `json:"id"`
}

type tgMessage struct {
	MessageID int    `json:"message_id"`
	Chat      tgChat `json:"chat"`
	Text      string `json:"text"`
}

type tgWebhookRequest struct {
	Message tgMessage `json:"message"`
}

type tgWebhookResponse struct {
	Method string `json:"method"`
	ChatID int    `json:"chat_id"`
	Text   string `json:"text"`
}

func telegramIsCommand(message string, command string) bool {
	b := false
	b = b || strings.HasPrefix(message, "/"+command+"@")
	b = b || strings.HasPrefix(message, "/"+command+" ")
	b = b || message == "/"+command
	return b
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
	if request.Message.Text[0] != '/' {
		return
	}

	// Select only one server based on webhook URL
	server := r.URL.Path[len("/telegram/"):]
	if len(server) == 0 {
		server = setting.servers[0]
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
		commandResult = batchRequest([]string{server}, "traceroute", target)[0]
	} else if telegramIsCommand(request.Message.Text, "trace6") {
		commandResult = batchRequest([]string{server}, "traceroute6", target)[0]

	} else if telegramIsCommand(request.Message.Text, "route") || telegramIsCommand(request.Message.Text, "route4") {
		commandResult = batchRequest([]string{server}, "bird", "show route for "+target+" primary")[0]
	} else if telegramIsCommand(request.Message.Text, "route6") {
		commandResult = batchRequest([]string{server}, "bird6", "show route for "+target+" primary")[0]

	} else if telegramIsCommand(request.Message.Text, "path") || telegramIsCommand(request.Message.Text, "path4") {
		tempResult := batchRequest([]string{server}, "bird", "show route for "+target+" all primary")[0]
		for _, s := range strings.Split(tempResult, "\n") {
			if strings.Contains(s, "BGP.as_path: ") {
				commandResult = strings.Split(s, "BGP.as_path: ")[1]
			}
		}
	} else if telegramIsCommand(request.Message.Text, "path6") {
		tempResult := batchRequest([]string{server}, "bird6", "show route for "+target+" all primary")[0]
		for _, s := range strings.Split(tempResult, "\n") {
			if strings.Contains(s, "BGP.as_path: ") {
				commandResult = strings.Split(s, "BGP.as_path: ")[1]
			}
		}

	} else if telegramIsCommand(request.Message.Text, "whois") {
		commandResult = whois(target)

	} else if telegramIsCommand(request.Message.Text, "help") {
		commandResult = strings.TrimSpace(`
/[path|path6] <IP>
/[route|route6] <IP>
/[trace|trace6] <IP>
/whois <Target>
		`)
	}

	// Create a JSON response
	w.Header().Add("Content-Type", "application/json")
	response := &tgWebhookResponse{
		Method: "sendMessage",
		ChatID: request.Message.Chat.ID,
		Text:   commandResult,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		println(err.Error())
		return
	}
}
