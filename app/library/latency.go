package library

import (
	"database/sql"
	"log"
	"time"

	goutils "github.com/mudphilo/go-utils"
)

func InsertOrUpdateSMSError(db *sql.DB, errorType string, latency float64) {
	// Load the Nairobi timezone
	loc, _ := time.LoadLocation("Africa/Nairobi")

	// Get today's date in the required format (YYYY-MM-DD) in Nairobi timezone
	today := time.Now().In(loc).Format("2006-01-02")
	

	// Initialize DB utilities
	
	dbUtilMaster := goutils.Db{DB: db} // Use master DB for updates and inserts

	// Check if there's an existing record for today and the given error type
	var id, noOfErrors sql.NullInt64
	var avgLatency sql.NullFloat64

	dbUtilMaster.SetQuery("SELECT id, no_of_errors, average_latency FROM sms_error WHERE DATE(created) = ? AND error_type = ?")
	dbUtilMaster.SetParams(today, errorType)

	err := dbUtilMaster.FetchOne().Scan(&id, &noOfErrors, &avgLatency)
	if err == sql.ErrNoRows {
		// No record exists for today, insert a new one
		insertData := map[string]interface{}{
			"error_type":       errorType,
			"number_of_sends":  0,
			"no_of_errors":     1,
			"previous_errors":  0,
			"average_latency":  latency,
			
		}

		_, err = dbUtilMaster.Upsert("sms_error", insertData,nil)
		if err != nil {
			log.Printf("Error inserting sms_error: %v", err)
		} else {
			log.Printf("Inserted new sms_error record for %s", errorType)
		}
	} else if err == nil {
		// Update the existing record
		newAvgLatency := avgLatency.Float64
		if latency > 0 {
			newAvgLatency = (avgLatency.Float64*float64(noOfErrors.Int64) + latency) / float64(noOfErrors.Int64+1)
		}

		updateData := map[string]interface{}{
			"no_of_errors":     noOfErrors.Int64 + 1,
			"average_latency":  newAvgLatency,
			
		}

		conditions := map[string]interface{}{
			"id": id.Int64,
		}

		_, err = dbUtilMaster.Update("sms_error", conditions, updateData)
		if err != nil {
			log.Printf("Error updating sms_error for %s: %v", errorType, err)
		} else {
			log.Printf("Updated sms_error for %s", errorType)
		}
	} else {
		log.Printf("Error fetching sms_error record: %v", err)
	}
}
