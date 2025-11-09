package models

import "encoding/xml"

type MarketDescriptionsSpecifiers struct {
	Text      string `xml:",chardata"`
	Specifier []struct {
		Text        string `xml:",chardata"`
		Name        string `xml:"name,attr"`
		Type        string `xml:"type,attr"`
		Description string `xml:"description,attr"`
	} `xml:"specifier"`
}

type MarketDescriptionsOutcomes struct {
	Text    string `xml:",chardata"`
	Outcome []struct {
		Text string `xml:",chardata"`
		ID   string `xml:"id,attr"`
		Name string `xml:"name,attr"`
	} `xml:"outcome"`
}

type MarketDescriptions struct {
	XMLName      xml.Name `xml:"market_descriptions"`
	Text         string   `xml:",chardata"`
	ResponseCode string   `xml:"response_code,attr"`
	Market       []struct {
		Text                   string                       `xml:",chardata"`
		ID                     string                       `xml:"id,attr"`
		Name                   string                       `xml:"name,attr"`
		Groups                 string                       `xml:"groups,attr"`
		IncludesOutcomesOfType string                       `xml:"includes_outcomes_of_type,attr"`
		OutcomeType            string                       `xml:"outcome_type,attr"`
		Outcomes               MarketDescriptionsOutcomes   `xml:"outcomes"`
		Specifiers             MarketDescriptionsSpecifiers `xml:"specifiers"`
		Attributes             struct {
			Text      string `xml:",chardata"`
			Attribute struct {
				Text        string `xml:",chardata"`
				Name        string `xml:"name,attr"`
				Description string `xml:"description,attr"`
			} `xml:"attribute"`
		} `xml:"attributes"`
	} `xml:"market"`
}

type CreateMarketAlias struct {

	//MarketID market ID
	MarketID int64 `json:"market_id"`

	//Specifier market line specifier
	Specifier string `json:"specifier"`

	//OutcomeID market outcome ID
	OutcomeID string `json:"outcome_id"`

	//Alias custom alias
	Alias string `json:"alias"`

}

type MarketAlias struct {

	//ID
	ID int64 `json:"id"`

	//MarketID market ID
	MarketID int64 `json:"market_id"`

	//Specifier market line specifier
	Specifier string `json:"specifier"`

	//OutcomeID market outcome ID
	OutcomeID string `json:"outcome_id"`

	//Alias custom alias
	Alias string `json:"alias"`

	//MarketName market name
	MarketName string `json:"market_name"`

}