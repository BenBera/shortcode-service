package library

import (
	"bitbucket.org/maybets/shortcode-service/app/constants"
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