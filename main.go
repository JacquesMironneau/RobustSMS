package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/gorilla/mux"
	"github.com/jacobsa/go-serial/serial"
)

// type SafeCounter struct {
// 	mu sync.Mutex
// 	v  map[string]int
// }

var options = serial.OpenOptions{
	PortName:        "/dev/tty.usbserial-A8008HlV",
	BaudRate:        19200,
	DataBits:        8,
	StopBits:        1,
	MinimumReadSize: 4,
}

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

var allowCloud = false

type SMSSender interface {
	SendSMS(msgPtr *string, phoneNumberPtr *string) bool
}

type OnPremSMSSEnder struct {
	options serial.OpenOptions
}

type AWS_SNS_Sender struct {
	svc *sns.SNS
}

var LocalSMSSender = OnPremSMSSEnder{options}
var CloudSMSSender = AWS_SNS_Sender{svc}

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

	num := newMessage.To
	inat, _ := regexp.Match("^\\+3", []byte(num))

	frToFormat, _ := regexp.Match("^(06|07)", []byte(num))

	sent := false
	// France

	if frToFormat {
		newMessage.To = "+33" + num[1:]
		fmt.Println("Formatting", num, newMessage.To)
	}

	fr, _ := regexp.Match("^\\+33", []byte(num))

	if fr {
		fmt.Println("Sending locally")
		sent = sendLocalWithCloudFallback(&newMessage.Message, &newMessage.To)
	} else if inat {
		sent = CloudSMSSender.sendSMS(&newMessage.Message, &newMessage.To)
	} else {
		fmt.Println("Wrong format")
	}

	fmt.Println("Message sent", sent)
	if sent {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(newMessage)
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(messages)
}
func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome on the RobustSMS API !")
}

func (sender OnPremSMSSEnder) sendSMS(msgPtr *string, phoneNumberPtr *string) bool {

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	// Make sure to close it later.
	defer port.Close()

	// Write 4 bytes to the port.
	number := []byte("AT+CMGS=" + *phoneNumberPtr + "\r")
	n, err := port.Write(number)

	if err != nil {
		log.Fatalf("Failed to send on prem sms: %v", err)
		return false
	}
	fmt.Println("Wrote", n, "bytes.")

	message := []byte(*msgPtr + "\x1A")
	n, err = port.Write(message)

	if err != nil {
		log.Fatalf("Failed to send on prem sms: %v", err)
		return false
	}
	fmt.Println("Wrote", n, "bytes.")
	return true
}

func (sender AWS_SNS_Sender) sendSMS(msgPtr *string, phoneNumberPtr *string) bool {

	if !allowCloud {
		fmt.Println("Fake sending cloudly")
		return false
	}
	result, err := sender.svc.Publish(&sns.PublishInput{
		Message:     msgPtr,
		PhoneNumber: phoneNumberPtr,
	})
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	fmt.Println(*result.MessageId)
	return true
}

func sendLocalWithCloudFallback(msgPtr *string, phoneNumberPtr *string) bool {
	sent := LocalSMSSender.sendSMS(msgPtr, phoneNumberPtr)
	if !sent {
		sent = CloudSMSSender.sendSMS(msgPtr, phoneNumberPtr)
	}
	return sent
}

// Initialize a session that the SDK will use to load
// credentials from the shared credentials file. (~/.aws/credentials).
var sess = session.Must(session.NewSessionWithOptions(session.Options{
	SharedConfigState: session.SharedConfigEnable,
}))

var svc = sns.New(sess)

func main() {

	port := ":8010"

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/messages", createMessage).Methods("POST")
	router.HandleFunc("/messages", getMessages).Methods("GET")
	router.HandleFunc("/", homeLink).Methods("GET")

	fmt.Printf("Starting the server on localhost%s (cloud enabled: %t)", port, allowCloud)
	log.Fatal(http.ListenAndServe(port, router))

}
