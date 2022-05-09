package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/gorilla/mux"
)

// type SafeCounter struct {
// 	mu sync.Mutex
// 	v  map[string]int
// }

type message struct {
	ID      string `json:"ID"`
	To      string `json:"To"`
	Message string `json:"Message"`
}

type messageQueue []message

var messages = messageQueue{
	{
		ID: "1", To: "0651392313", Message: "Salut !!monde!!",
	},
}

// func main() {
// 	msgPtr := flag.String("m", "", "The message to send to the subscribed users of the topic")
// 	phoneNumberPtr := flag.String("n", "", "The ARN of the topic to which the user subscribes")

// 	flag.Parse()

// 	if *msgPtr == "" || *phoneNumberPtr == "" {
// 		fmt.Println("You must supply a message and a phoneNumber (+33XXXXXXXX)")
// 		fmt.Println("Usage: go run SnsPublish.go -m MESSAGE -n PhoneNumber")
// 		os.Exit(1)
// 	}

// Initialize a session that the SDK will use to load
// credentials from the shared credentials file. (~/.aws/credentials).

//
// }

// func (c *SafeCounter) Inc(key string) {
// 	c.mu.Lock()
// 	// Lock so only one goroutine at a time can access the map c.v.
// 	c.v[key]++
// 	c.mu.Unlock()
// }

func createMessage(w http.ResponseWriter, r *http.Request) {
	var newMessage message
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}

	json.Unmarshal(reqBody, &newMessage)
	messages = append(messages, newMessage)

	// 	if *msgPtr == "" || *phoneNumberPtr == "" {
	// 		fmt.Println("You must supply a message and a phoneNumber (+33XXXXXXXX)")
	// 		fmt.Println("Usage: go run SnsPublish.go -m MESSAGE -n PhoneNumber")
	// 		os.Exit(1)
	// 	}
	log.Print(newMessage)

	message := newMessage.Message
	phoneNumber := newMessage.To
	sendCloudSMS(&message, &phoneNumber)

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newMessage)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(messages)
}
func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func sendSMS() {
	sendOnPremSMS()
	// sendCloudSMS()
}

func sendOnPremSMS() {
	// sm, err := NewStateMachine("")
	// sm.Connect()
	// sm.SendSMS(number, "Bonjour depuis l'application!")
}

func sendCloudSMS(msgPtr *string, phoneNumberPtr *string) {
	result, err := svc.Publish(&sns.PublishInput{
		Message:     msgPtr,
		PhoneNumber: phoneNumberPtr,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(*result.MessageId)
}

var sess = session.Must(session.NewSessionWithOptions(session.Options{
	SharedConfigState: session.SharedConfigEnable,
}))

var svc = sns.New(sess)

func main() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/messages", createMessage).Methods("POST")
	router.HandleFunc("/messages", getMessages).Methods("GET")
	router.HandleFunc("/", homeLink).Methods("GET")
	log.Fatal(http.ListenAndServe(":8010", router))

}
