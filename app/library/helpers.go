package library

import (
	"github.com/BenBera/shortcode-service/app/constants"
	"github.com/BenBera/shortcode-service/app/models"
)

// Helper function to determine bet status
func DetermineBetStatus(wonStatus int64) string {
	switch wonStatus {
	case constants.STATUS_NOT_LOST_OR_WON:
		return "pending"
	case constants.STATUS_LOST:
		return "lost"
	case constants.STATUS_WON:
		return "Won"
	case constants.STATUS_CASHOUT:
		return "Cashed Out"
	case constants.STATUS_PARTIAL_CASHOUT:
		return "lost"
	case constants.STATUS_CANCELLED:
		return "Cancelled"
	default:
		return "Unknown"
	}
}
func GetItemValue(name string, data models.ShortCodeIncomingRequest) string {
	for _, item := range data.RequestParam.Data {
		if item.Name == name {
			return item.Value
		}
	}
	return ""
}
