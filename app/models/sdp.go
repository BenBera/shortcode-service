package models

type InboxPayloadInteractive struct {
	Operation    string `json:"operation"`
	RequestID    string `json:"requestId"`
	Channel      string `json:"channel"`
	RequestParam struct {
		AdditionalData []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"additional_data"`
		Data []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"data"`
	} `json:"requestParam"`
	RequestTimeStamp string `json:"requestTimeStamp"`
}


type DataParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type RequestParam struct {
	Data []DataParam `json:"data"`
}

type SendSMS struct {
	RequestID        int64        `json:"requestId"`
	RequestTimeStamp string       `json:"requestTimeStamp"`
	Channel          string       `json:"channel"`
	SourceAddress    string       `json:"sourceAddress"`
	Operation        string       `json:"operation"`
	RequestParam     RequestParam `json:"requestParam"`
}

type SendSMSResponse struct {
	RequestID        interface{}   `json:"requestId"`
	RequestTimeStamp string        `json:"requestTimeStamp"`
	Channel          string        `json:"channel"`
	SourceAddress    string        `json:"sourceAddress"`
	Operation        string        `json:"operation"`
	RequestParam     RequestParam  `json:"requestParam"`
	ResponseParam    ResponseParam `json:"responseParam"`
}

type ResponseParam struct {
	Status      string `json:"status"`
	StatusCode  string `json:"statusCode"`
	Description string `json:"description"`
}