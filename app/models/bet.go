package models

type Bet struct {

	//Stake stake amount
	Stake float64 `json:"stake" validate:"optional"`

	//BetType bet type 1 (mobile web), 2 (desktop web), 3 (sms), 4 (opera), 5 (android app)
	BetType int64 `json:"bet_type" validate:"optional" enums:"1,2,3,4,5"`

	//Source bet source 1 (mobile web), 2 (desktop web), 3 (sms), 4 (opera), 5 (android app) enums:"1,2,3,4,5"
	Source int64 `json:"source" validate:"optional" enums:"1,2,3,4,5"`

	//IPAddress customer device Ip address
	IPAddress string `json:"ip_address" validate:"optional"`

	//StakeType type of balance to use to place abet, 1 (cash bet), 2 (bonus bet)
	StakeType int64 `json:"stake_type" validate:"optional" enums:"1,2" default:"1"`

	//BookingCode booking code if any
	BookingCode string `param:"booking_code" query:"booking_code" form:"booking_code" json:"booking_code" xml:"booking_code" validate:"optional"`

	//Bets if booking code is supplied this field will be ignored
	Bets []BetSlip `param:"bets" query:"bets" form:"bets" json:"bets" xml:"bets" validate:"optional"`

	//UTMSource utm source
	UTMSource string `param:"utm_source" query:"utm_source" form:"utm_source" json:"utm_source" xml:"utm_source" validate:"optional"`

	//Campaign utm campaign
	Campaign string `param:"campaign" query:"campaign" form:"campaign" json:"campaign" xml:"campaign" validate:"optional"`

	//Medium utm medium
	Medium string `param:"medium" query:"medium" form:"medium" json:"medium" xml:"medium" validate:"optional"`
}

type BetSlip struct {
	//OddID odd id received from fixtures
	OddID int64 `json:"odd_id"`

	//ProducerID producer ID
	ProducerID int64 `json:"producer_id" enums:"1,3,4"`
}

type CloseID struct {

	ID int64 `json:"id"`
}

type Template struct {

	//Name name of the sms template
	Name string `json:"name" validate:"required" enums:"JOIN,HELP,BALANCE_QUERY,SMS_GAMES"`

	//Content content of the sms template
	Content string `json:"content" validate:"required"`

}