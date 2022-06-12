package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

// type SafeCounter struct {
// 	mu sync.Mutex
// 	v  map[string]int
// }

func prepareMessage(newMessage message) bool {

	num := newMessage.To
	inat, _ := regexp.Match("^\\+3", []byte(num))

	frToFormat, _ := regexp.Match("^(06|07)", []byte(num))

	sent := false
	// France

	if frToFormat {
		newMessage.To = "+33" + num[1:]
		fmt.Println("Formatting", num, newMessage.To)
	}

	num = newMessage.To
	fr, _ := regexp.Match("^\\+33", []byte(num))

	if fr {
		fmt.Println("Sending locally")
		sent = sendLocalWithCloudFallback(&newMessage.Message, &newMessage.To)
	} else if inat {
		sent = CloudSMSSender.sendSMS(&newMessage.Message, &newMessage.To)
	} else {
		fmt.Println("Wrong format")
	}
	return sent
}

func createMessage(w http.ResponseWriter, r *http.Request) {

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}

	var newMessage message
	err = json.Unmarshal(reqBody, &newMessage)
	if err != nil {
		fmt.Println("failed to parse message", err)
	}

	messages = append(messages, newMessage)

	// 	if *msgPtr == "" || *phoneNumberPtr == "" {
	// 		fmt.Println("You must supply a message and a phoneNumber (+33XXXXXXXX)")
	// 		fmt.Println("Usage: go run SnsPublish.go -m MESSAGE -n PhoneNumber")
	// 		os.Exit(1)
	// 	}
	log.Print(newMessage)

	sent := prepareMessage(newMessage)
	fmt.Println("Message sent", sent)
	if sent {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(newMessage)
}

func createMessageBulk(w http.ResponseWriter, r *http.Request) {

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}

	var newMessage message
	var batch messageBatch
	err = json.Unmarshal(reqBody, &batch)
	if err != nil {
		fmt.Println("failed to parse array")
	}

	messages = append(messages, batch[:]...)

	var sent bool
	for _, v := range batch {
		fmt.Printf("aaa %#v", v)
		sent = prepareMessage(v) && sent
	}

	if sent {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(newMessage)
}

func createMessagePublish(w http.ResponseWriter, r *http.Request) {

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}

	var newMessagePublish messagePublish
	err = json.Unmarshal(reqBody, &newMessagePublish)
	if err != nil {
		fmt.Println("failed to parse message publish")
	}
	var sent bool

	for _, v := range newMessagePublish.To {
		tmpMessage := message{
			To:      v,
			Message: newMessagePublish.Message,
		}
		messages = append(messages, tmpMessage)
		sent = prepareMessage(tmpMessage) && sent
	}
	if sent {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(newMessagePublish)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(messages)
}
func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome on the RobustSMS API !")
}

func main() {

	port := ":8010"

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/messages", createMessage).Methods("POST")
	router.HandleFunc("/messagesBulk", createMessageBulk).Methods("POST")
	router.HandleFunc("/messagesPublish", createMessagePublish).Methods("POST")
	router.HandleFunc("/messages", getMessages).Methods("GET")
	router.HandleFunc("/", homeLink).Methods("GET")

	fmt.Printf("Starting the server on localhost%s (cloud enabled: %t)", port, AllowCloud)
	log.Fatal(http.ListenAndServe(port, router))

}
