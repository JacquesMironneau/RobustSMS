package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/jacobsa/go-serial/serial"
)

const AllowCloud = true
const demo = true

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

type messagePublish struct {
	Message string   `json:"Message"`
	To      []string `json:"To"`
}

type messageBatch []message

type messageQueue []message

var messages = messageQueue{
	{
		ID: "1", To: "0651392313", Message: "Salut !!monde!!",
	},
}

type SMSSender interface {
	SendSMS(msgPtr *string, phoneNumberPtr *string) bool
}

type OnPremSMSSEnder struct {
	options serial.OpenOptions
}

type AWS_SNS_Sender struct {
	svc *sns.SNS
}

// Initialize a session that the SDK will use to load
// credentials from the shared credentials file. (~/.aws/credentials).
var sess = session.Must(session.NewSessionWithOptions(session.Options{
	SharedConfigState: session.SharedConfigEnable,
}))

var svc = sns.New(sess)

var LocalSMSSender = OnPremSMSSEnder{options}
var CloudSMSSender = AWS_SNS_Sender{svc}

func (sender OnPremSMSSEnder) sendSMS(msgPtr *string, phoneNumberPtr *string) bool {

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		fmt.Println("serial.Open: %v", err)
		return false
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

	if !AllowCloud {
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
	fmt.Printf("Sending %s to %s\n", *msgPtr, *phoneNumberPtr)

	var sent bool
	if demo {
		sent = CloudSMSSender.sendSMS(msgPtr, phoneNumberPtr)

	} else {
		sent = LocalSMSSender.sendSMS(msgPtr, phoneNumberPtr)
		if !sent {
			fmt.Println("Localrequest failed: Send with cloud")
			sent = CloudSMSSender.sendSMS(msgPtr, phoneNumberPtr)
		}
	}
	return sent
}
