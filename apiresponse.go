package main

type apiResponse struct {
	Body body `json:"body"`
}

type body struct {
	Content content `json:"content"`
}

type content struct {
	PickupMessage pickupMessage `json:"pickupMessage"`
}

type pickupMessage struct {
	Stores []store `json:"stores"`
}

type store struct {
	PartsAvailability map[string]partsAvailability `json:"partsAvailability"`
	StoreName         string                       `json:"storeName"`
}

type partsAvailability struct {
	PickupDisplay string `json:"pickupDisplay"`
}

func (p *partsAvailability) isAvailable() bool {
	return p.PickupDisplay == "available"
}
