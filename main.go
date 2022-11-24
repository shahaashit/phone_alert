package main

import (
	"encoding/json"
	"errors"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"net/http"
	"net/url"
	"phone_alert/config"
	"phone_alert/log"
	"strconv"
	"strings"
	"time"
)

func sendMessage(message string, errorType bool) {
	if errorType {
		log.Error("Sending error message: ", message)
		sendMessageToFlock("ERROR: "+message, false)
	} else {
		log.Info("Sending success message: ", message)
		sendMessageToFlock("SUCCESS: "+message, false)
	}
}

func initializeCrons() {
	c := cron.New()
	c.AddFunc("@every 15m", func() {
		sendMessageToFlock("healthy", false)
	})
	c.AddFunc("@every 15s", func() {
		log.Info(time.Now())
	})
	c.Start()
}

func main() {
	initializeCrons()
	for {
		for _, pincode := range config.Pincodes {
			response, err := callUrlAndGetResponse(makeUrl(pincode))
			/*//start
			if err == nil {
				sendMessage("working now", false)
			}
			continue
			//end*/
			if err != nil {
				sendMessage(err.Error(), true)
				continue
			}
			checkForAvailability(response, pincode)
		}
		time.Sleep(time.Second * 30)
	}
}

func checkForAvailability(response *apiResponse, pincode string) {
	if response == nil {
		sendMessage("Nil response for pincode: "+pincode, true)
		return
	}
	if len(response.Body.Content.PickupMessage.Stores) < config.MinNumberOfStores {
		sendMessage("Stores number is less than threshold for pincode: "+pincode, true)
		return
	}

	for _, storeObj := range response.Body.Content.PickupMessage.Stores {
		for modelId, resp := range storeObj.PartsAvailability {
			if resp.isAvailable() && !inStringArray(storeObj.StoreName, config.IgnoreStoreNames) {
				sendMessage("Available for pincode: "+pincode+". At store: "+storeObj.StoreName+". For model: "+config.PhoneModels[modelId]+".", false)
			}
		}
	}
}

func makeUrl(pincode string) string {
	q := url.Values{}
	q.Set("location", pincode)
	q.Set("pl", "true")
	q.Set("mts.0", "regular")
	q.Set("cppart", "UNLOCKED/US")
	q.Set("cppart", "UNLOCKED/US")

	i := 0
	for modelId := range config.PhoneModels {
		q.Set("parts."+strconv.Itoa(i), modelId)
		i++
	}

	u := url.URL{
		Scheme:   "https",
		Host:     "www.apple.com",
		Path:     "shop/fulfillment-messages",
		RawQuery: q.Encode(),
	}
	urlToCall := u.String()
	log.Info("Calling Url: ", urlToCall)
	return urlToCall
}

func callUrlAndGetResponse(urlToCall string) (*apiResponse, error) {
	finalResp := &apiResponse{}

	var req *http.Request
	req, err := http.NewRequest("GET", urlToCall, nil)
	if err != nil {
		log.Error("error while making request to apple: ", err)
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Accept-Language", "hi_IN")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36")

	var defaultClient = &http.Client{}
	response, err := defaultClient.Do(req)
	if err != nil {
		return nil, errors.New("error while calling url:" + err.Error())
	}
	apResponseInBytes, err := ioutil.ReadAll(response.Body)
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Error("error while closing body : ", err)
		}
	}()

	err = json.Unmarshal(apResponseInBytes, &finalResp)
	if err != nil {
		//fmt.Println("error while unmarshalling response: " + err.Error() + ". Raw response: " + string(apResponseInBytes) + ".")
		return nil, errors.New("error while unmarshalling response: " + err.Error())
	}

	if response.StatusCode != 200 {
		return nil, errors.New("invalid status code: " + strconv.Itoa(response.StatusCode))
	}
	return finalResp, err
}

func inStringArray(needle string, haystack []string) bool {
	for _, v := range haystack {
		if strings.ToLower(v) == strings.ToLower(needle) {
			return true
		}
	}
	return false
}
