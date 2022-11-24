package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"phone_alert/config"
	"phone_alert/log"
)

type flockResponse struct {
	Uid         string `json:"uid"`
	Error       string `json:"error"`
	Description string `json:"description"`
}

func sendMessageToFlock(message string, flockmlType bool) error {
	u := url.URL{
		Scheme: "https",
		Host:   "api.flock.com",
		Path:   "hooks/sendMessage/" + config.FlockToken,
	}
	urlToCall := u.String()

	values := map[string]string{}
	if flockmlType {
		values["flockml"] = message
	} else {
		values["text"] = message
	}
	jsonValue, _ := json.Marshal(values)
	bytes.NewBuffer(jsonValue)
	resp, err := http.Post(urlToCall, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Error("error while sending data via flock: ", err)
		return err
	}

	res := flockResponse{}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		log.Error("error while decoding flock message response: ", err)
		return err
	}

	if res.Error != "" {
		log.Error("error response from flock while sending message: ", res)
		return errors.New(res.Description)
	}
	return nil
}
