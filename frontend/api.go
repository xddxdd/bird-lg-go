package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

type apiRequest struct {
	Servers []string `json:"servers"`
	Type    string   `json:"type"`
	Args    string   `json:"args"`
}

type apiGenericResultPair struct {
	Server string `json:"server"`
	Data   string `json:"data"`
}

type apiSummaryResultPair struct {
	Server string           `json:"server"`
	Data   []SummaryRowData `json:"data"`
}

type apiResponse struct {
	Error  string        `json:"error"`
	Result []interface{} `json:"result"`
}

var apiHandlerMap = map[string](func(request apiRequest) apiResponse){
	"summary":     apiSummaryHandler,
	"bird":        apiGenericHandlerFactory("bird"),
	"traceroute":  apiGenericHandlerFactory("traceroute"),
	"whois":       apiWhoisHandler,
	"server_list": apiServerListHandler,
}

func apiGenericHandlerFactory(endpoint string) func(request apiRequest) apiResponse {
	return func(request apiRequest) apiResponse {
		results := batchRequest(request.Servers, endpoint, request.Args)
		var response apiResponse

		for i, result := range results {
			response.Result = append(response.Result, &apiGenericResultPair{
				Server: request.Servers[i],
				Data:   result,
			})
		}

		return response
	}
}

func apiServerListHandler(request apiRequest) apiResponse {
	var response apiResponse

	for _, server := range setting.servers {
		response.Result = append(response.Result, apiGenericResultPair{
			Server: server,
		})
	}

	return response
}

func apiSummaryHandler(request apiRequest) apiResponse {
	results := batchRequest(request.Servers, "bird", "show protocols")
	var response apiResponse

	for i, result := range results {
		parsedSummary, err := summaryParse(result, request.Servers[i])
		if err != nil {
			return apiResponse{
				Error: err.Error(),
			}
		}

		response.Result = append(response.Result, &apiSummaryResultPair{
			Server: request.Servers[i],
			Data:   parsedSummary.Rows,
		})
	}

	return response
}

func apiWhoisHandler(request apiRequest) apiResponse {
	return apiResponse{
		Error: "",
		Result: []interface{}{
			apiGenericResultPair{
				Server: "",
				Data:   whois(request.Args),
			},
		},
	}
}

func apiErrorHandler(err error) apiResponse {
	return apiResponse{
		Error: err.Error(),
	}
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	var request apiRequest
	var response apiResponse
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response = apiResponse{
			Error: err.Error(),
		}
	} else {
		handler := apiHandlerMap[request.Type]
		if handler == nil {
			response = apiErrorHandler(errors.New("invalid request type"))
		} else {
			response = handler(request)
		}
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	bytes, err := json.Marshal(response)
	if err != nil {
		println(err.Error())
		return
	}
	w.Write(bytes)
}
