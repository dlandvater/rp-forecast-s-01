package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//Exists in common-pub.go
//type pushRequest struct {
//	Message struct {
//		Attributes map[string]string
//		Data       []byte
//		ID         string `json:"message_id"`
//	}
//	Subscription string
//}

//Exists in common-pub.go
//type NetChangeTrigger struct {
//	OrgId      string
//	ItemId     string
//	LocationId string
//}

func retrieveTriggers(w http.ResponseWriter, r *http.Request) []NetChangeTrigger {

	var triggers []NetChangeTrigger

	// Verify the token.
	if r.URL.Query().Get("token") != token {
		http.Error(w, "Bad token", http.StatusBadRequest)
		return triggers
	}

	msg := &pushRequest{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		//TODO remove testing
		fmt.Println("after json.NewDecoder")
		fmt.Println("err:", err)
		http.Error(w, fmt.Sprintf("Could not decode body: %v", err), http.StatusBadRequest)
		return triggers
	}

	messagesMu.Lock()
	defer messagesMu.Unlock()

	err := json.Unmarshal(msg.Message.Data, &triggers)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return triggers
	}
	//TODO remove testing
	fmt.Println("received forecast message ID: ", msg.Message.ID)
	fmt.Println("triggers:", triggers)

	return triggers
}
