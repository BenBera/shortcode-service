package models

type ErrorResponse struct {

	//ErrorCode error status code
	ErrorCode int `json:"error_code"  validate:"required"`

	//ErrorMessage status description
	ErrorMessage string `json:"error_message"  validate:"required"`
}

type SuccessResponse struct {
	//Status http status code
	Status int `json:"status"  validate:"required"`

	//Message status description
	Message string `json:"data"  validate:"required"`
}

type SuccessResponseWithID struct {
	//Status http status code
	Status int `json:"status"  validate:"required"`

	//ID created resource ID
	ID int64 `json:"id"  validate:"required"`

	//Message status description
	Message string `json:"data"  validate:"required"`
}

type BetSuccessResponse struct {
	//Status http status code
	Status int `json:"status"  validate:"required"`

	//BetID bet id that was created
	BetID string `json:"bet_id"  validate:"required"`

	//Message status description
	Message string `json:"data"  validate:"required"`
}

type ResponseMessage struct {
	Status  int         `json:"status"  validate:"required"`
	Message interface{} `json:"message"  validate:"required"`
}

