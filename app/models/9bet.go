package models

// UserRegByMobileRequest represents the expected payload for "userRegByMobile".
type UserRegByMobileRequest struct {
	Mobile   string `json:"mobile" form:"mobile" binding:"required"` // e.g. "254791784445"
	VKey     string `json:"vkey" form:"vkey"`                        // verification key
	VCode    string `json:"vcode" form:"vcode"`                      // verification code
	Password string `json:"password" form:"password" binding:"required"`
	TGCode   string `json:"tg_code" form:"tg_code"`          // telegram code (optional)
	TID      string `json:"tid" form:"tid"`                  // transaction / tracking id (optional)
	DeviceID string `json:"device_id" form:"device_id"`      // device identifier (optional)
	AC       string `json:"ac" form:"ac" binding:"required"` // action e.g. "userRegByMobile"
}
