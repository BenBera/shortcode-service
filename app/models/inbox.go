package models

type Inbox struct {

	//Msisdn mobile number
	Msisdn string `json:"msisdn"`

	//Message sms sent by user
	Message string `json:"message"`

	//InboxID id of the message
	InboxID int64 `json:"inbox_id"`
}
type MnoInbox struct {
	Network string      `json:"network"`
	Payload interface{} `json:"payload"`
}
type StatusChange struct {
	//ProfileId
	ProfileID int32 `json:"profileID"`
	//Status
	Status int32 `json:"status"`
}

type USSD struct {

	//Msisdn mobile number
	Msisdn int64 `json:"msisdn"`

	//Message sms sent by user
	Text string `json:"text"`

	//user Session
	Session string `json:"session"`

	//network
	Network string `json:"network"`

	//ussd code
	Code string `json:"code"`

	//level
	Level int64 `json:"level"`
	//INPUT
	Input string `json:"input"`
}

type UssdResponse struct {
	Text         string `json:"text"`
	ResponseType string `json:"responseType"`
}

type UssdGame struct {
	GameID   int
	Name     string
	Date     string
	Priority int
	Home     float32
	Away     float32
	Draw     float32
}
type UssdGameSlip struct {
	GameID    int
	Action    string
	Odds      float32
	ProfileID int
}

type InboxDLR struct {
	MSISDN      int64  `json:"msisdn"`
	Status      int    `json:"status"`
	Description string `json:"description"`
	MessageType string `json:"message_type"`
	//InboxID id of the message
	InboxID int64 `json:"transaction_id"`
}


type SMSError struct {
	ID             int
	ErrorType      string
	NumberOfSends  int
	NoOfErrors     int
	PreviousErrors int
	Latency        float64
}

type EmailSend struct {
	Subject string            `json:"subject"`
	Message string            `json:"message"`
	Summary string            `json:"summary"`
	Template string 	      `json:"template"`
	Name string               `json:"name"`
	Email string			  `json:"email"`
	ClientID int64            `json:"client_id"`
}