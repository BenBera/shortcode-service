package crontask

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"bitbucket.org/maybets/shortcode-service/app/constants"
	"bitbucket.org/maybets/shortcode-service/app/models"

	goutils "github.com/mudphilo/go-utils"
)

func Notificationerror(db *sql.DB) {

	ticker := time.NewTicker(5 * time.Minute)

	for _ = range ticker.C {

		errorNotification(db)
		
	}

	select {}
}




func errorNotification(db *sql.DB) {
	
	dbUtilMaster := goutils.Db{DB: db} // Use master DB for updates

	// Load the Nairobi timezone
	loc, _ := time.LoadLocation("Africa/Nairobi")

	// Get today's date in the required format (YYYY-MM-DD) in Nairobi timezone
	today := time.Now().In(loc).Format("2006-01-02")

	// Query to fetch errors recorded today where the number of errors is greater than 10
	query := `
		SELECT id, error_type, number_of_sends, no_of_errors, previous_errors, average_latency
		FROM sms_error 
		WHERE DATE(created) = ? AND no_of_errors > 10
	`

	dbUtilMaster.SetQuery(query)
	dbUtilMaster.SetParams(today)

	rows, err := dbUtilMaster.Fetch()
	if err != nil {
		log.Printf("Error fetching SMS errors: %v", err)
		return
	}
	defer rows.Close()

	// Iterate over query results
	for rows.Next() {
		var errorData models.SMSError

		err := rows.Scan(&errorData.ID, &errorData.ErrorType, &errorData.NumberOfSends, &errorData.NoOfErrors, &errorData.PreviousErrors, &errorData.Latency)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Check conditions for alerting
		if errorData.NumberOfSends < 3 &&
			errorData.NoOfErrors > 10 &&
			errorData.NoOfErrors > errorData.PreviousErrors {
			
			sendMailNotification(errorData, dbUtilMaster)
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
	}
}
func sendMailNotification(errorData models.SMSError, dbUtilMaster goutils.Db) {
	// Generate alert message
	message := fmt.Sprintf(
		"Dear Team,\n\n"+
			"SMS delivery delays detected from Maybets to SDP.\n\n"+
			"**Details:**\n"+
			"- Error Type: %s\n"+
			"- Latency: %.2f sec\n"+
			"- Errors Today: %d\n"+
			"- Previous Errors: %d\n\n"+
			"Immediate attention is required to resolve this issue.\n\n"+
			"Best Regards,\nIntouchVAS Tech Team",
		errorData.ErrorType, errorData.Latency/1000, errorData.NoOfErrors, errorData.PreviousErrors,
	)

	// Prepare update parameters
	updates := map[string]interface{}{
		"number_of_sends": errorData.NumberOfSends + 1,
		"previous_errors": errorData.PreviousErrors + errorData.NoOfErrors,
	}

	conditions := map[string]interface{}{
		"id": errorData.ID,
	}

	// Update database
	_, updateErr := dbUtilMaster.Update("sms_error", conditions, updates)
	if updateErr != nil {
		log.Printf("Error updating sms_error for ID %d: %v", errorData.ID, updateErr)
		return
	}

	log.Printf("Updated sms_error for ID %d: number_of_sends+1, previous_errors=previousErrors + noOfErrors", errorData.ID)

	// Send notification via email
	data := models.EmailSend{
		Subject:  "SMS ERROR",
		Name:     "INTOUCHVAS",
		Email:    "tech@intouchvas.io",
		Template: "notification",
		Summary:  "SMS TO SDP SMS ERROR",
		Message:  message,
		ClientID: 1,
	}

	SendMail(1, data)

	// Second email to Maybets team
	emailData2 := models.EmailSend{
		Subject:  "SMS ERROR",
		Name:     "MAYBETS",
		Email:    "tech-support@maybets.com",
		Template: "notification",
		Summary:  "SMS TO SDP SMS ERROR",
		Message:  message,
		ClientID: 1,
	}
	SendMail(1, emailData2)

	
}



func SendMail(clientID int64, data models.EmailSend) int {
	endpoint := constants.INTOUCHMAILENDPOINT
	headers := map[string]string{
		"x-token":    os.Getenv("INTOUCH_MAIL_TOKEN"),
		"x-client-id": fmt.Sprintf("%d", clientID),
	}

	dt := data

	

	status, _ := goutils.HTTPPost(endpoint, headers, dt)

	

	return status
}

