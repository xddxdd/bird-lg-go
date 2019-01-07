package main

import (
    "net/http"
    "net/url"
    "io/ioutil"
    "strconv"
)

type channelData struct {
    id int
    data string
}

func batchRequest(servers []string, endpoint string, command string) []string {
    var ch chan channelData = make(chan channelData)
    var response_array []string = make([]string, len(servers))

    for i, server := range servers {
        var isValidServer bool = false
        for _, validServer := range settingServers {
            if validServer == server {
                isValidServer = true
                break
            }
        }
        if !isValidServer {
            go func (i int) {
                ch <- channelData{i, "request failed: invalid server\n"}
            } (i)
        } else {
            url := "http://" + server + "." + settingServersDomain + ":" + strconv.Itoa(settingServersPort) + "/" + url.PathEscape(endpoint) + "?q=" + url.QueryEscape(command)
            go func (url string, i int){
                response, err := http.Get(url)
                if err != nil {
                    ch <- channelData{i, "request failed: " + err.Error() + "\n"}
                    return
                }
                text, _ := ioutil.ReadAll(response.Body)
                ch <- channelData{i, string(text)}
            } (url, i)
        }
    }

    for range servers {
        var output channelData = <-ch
        response_array[output.id] = output.data
    }

    return response_array
}
