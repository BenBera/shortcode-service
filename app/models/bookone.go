package models

type ShortCodeAutoResponse struct {
	Mobile     string `json:"mobile"`
	SenderName string `json:"sender_name"`
	ServiceID  int64  `json:"service_id"`
	LinkID     string `json:"link_id"`
	Message    string `json:"message"`
}
type SMSResponse struct {
	StatusCode    string `json:"status_code"`
	StatusDesc    string `json:"status_desc"`
	MessageID     int64  `json:"message_id"`
	MobileNumber  string `json:"mobile_number"`
	NetworkID     string `json:"network_id"`
	MessageCost   string `json:"message_cost"`
	CreditBalance string `json:"credit_balance"`
}
