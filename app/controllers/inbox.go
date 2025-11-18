package controllers

import (
	"context"
	"fmt"

	"github.com/BenBera/shortcode-service/app/constants"

	"net/http"
	"strconv"
	"strings"

	"github.com/BenBera/shortcode-service/app/models"
	goutils "github.com/mudphilo/go-utils"
	"github.com/sirupsen/logrus"
)

func (controller *Controller) ProcessInbox(ctx context.Context, u *models.Inbox, sdpAutoResponse bool, ipAddress string) (int, interface{}) {
	//ctx, span := controller.Tracer.Start(ctx, "ProcessInbox")
	//defer span.End()
	//
	//profile, err := controller.createOrGetUser(ctx, u.Msisdn)
	//if err != nil {
	//	return controller.handleError(ctx, err, "Failed to create user", u)
	//}
	//
	//if err := controller.saveInbox(ctx, u); err != nil {
	//	return controller.handleError(ctx, err, "Failed to create inbox", u)
	//}
	//
	//message := u.Message
	//if message == "" {
	//	message = "join"
	//}
	//
	//return controller.processMessage(ctx, message,  u.InboxID, ipAddress, sdpAutoResponse)
	return 0, nil
}

func (controller *Controller) processMessage(ctx context.Context, message string, inboxID int64, ipAddress string, sdpAutoResponse bool) (int, interface{}) {
	//logrus.WithContext(ctx).Infof("Processing message: '%s', ID %d ", message, inboxID)
	//profileID := profile.Id
	//// First, check if the message matches the SMS betting format
	//parts := strings.Split(message, "#")
	//if len(parts) > 2 {
	//	logrus.WithContext(ctx).Info("Message identified as SMS bet")
	//	betResponse := controller.smsBet(ctx, profileID, message, ipAddress)
	//	controller.AutoResponse(ctx, inboxID, betResponse, sdpAutoResponse)
	//	return http.StatusOK, models.SuccessResponse{
	//		Status:  http.StatusOK,
	//		Message: betResponse,
	//	}
	//}
	//
	//// If not an SMS bet, proceed with keyword matching
	//keywords, err := controller.getKeywordsFromDB(ctx)
	//if err != nil {
	//	logrus.WithContext(ctx).Error("Failed to get keywords from DB: ", err)
	//	return http.StatusInternalServerError, models.ErrorResponse{
	//		ErrorCode:    http.StatusInternalServerError,
	//		ErrorMessage: "Internal server error",
	//	}
	//}
	//
	////logrus.WithContext(ctx).Infof("Retrieved keywords from DB: %v", keywords)
	//
	//lowercaseMessage := strings.ToLower(strings.TrimSpace(message))
	//firstWord := strings.Split(lowercaseMessage, "#")[0]
	//
	//logrus.WithContext(ctx).Infof("Lowercase message: '%s', First word: '%s'", lowercaseMessage, firstWord)
	//
	//// Implement strict matching logic
	//for category, keywordList := range keywords {
	//	for _, keyword := range keywordList {
	//		if keyword == firstWord {
	//			logrus.WithContext(ctx).Infof("Strict keyword match found: category '%s', keyword '%s'", category, keyword)
	//			response := controller.handleMessageByCategory(ctx, category, message, profile, inboxID, ipAddress)
	//			controller.AutoResponse(ctx, inboxID, response, sdpAutoResponse)
	//			return http.StatusOK, models.SuccessResponse{
	//				Status:  http.StatusOK,
	//				Message: response,
	//			}
	//		}
	//	}
	//}
	//
	//logrus.WithContext(ctx).Info("No keyword match found, checking for autobet message")
	//
	//// Check for autobet messages if no other category matches
	//autoBetResponse := controller.handleAutoBetMessage(ctx, profileID, message)
	//if autoBetResponse != "" {
	//	logrus.WithContext(ctx).Info("Autobet message identified")
	//	controller.AutoResponse(ctx, inboxID, autoBetResponse, sdpAutoResponse)
	//	return http.StatusOK, models.SuccessResponse{
	//		Status:  http.StatusOK,
	//		Message: autoBetResponse,
	//	}
	//}
	//
	//logrus.WithContext(ctx).Info("No matching category or autobet found, returning default response")
	//
	//// Default response if no keyword matches
	//defaultResponse := controller.GetSMSTemplate(ctx, "JOIN")
	//controller.AutoResponse(ctx, inboxID, defaultResponse, sdpAutoResponse)
	return http.StatusOK, models.SuccessResponse{
		Status:  http.StatusOK,
		Message: "",
	}
}

func (controller *Controller) getKeywordsFromDB(ctx context.Context) (map[string][]string, error) {
	dbUtils := goutils.Db{DBSlave: controller.DBSlave, Context: ctx}

	sqlQuery := `
        SELECT c.category_name, k.keyword 
        FROM sms_keywords k 
        JOIN sms_categories c ON k.category_id = c.id 
        WHERE k.is_active = 1
        ORDER BY c.category_name , k.keyword `

	dbUtils.SetQuery(sqlQuery)

	rows, err := dbUtils.FetchSlaveWithContext()
	if err != nil {
		return nil, fmt.Errorf("error fetching keywords: %w", err)
	}
	defer rows.Close()

	keywords := make(map[string][]string)

	for rows.Next() {
		var categoryName, keyword string

		err = rows.Scan(&categoryName, &keyword)
		if err != nil {
			logrus.WithContext(ctx).
				WithFields(logrus.Fields{
					constants.DESCRIPTION: "Error scanning keyword row",
				}).
				Warn(err.Error())
			continue
		}

		keywords[categoryName] = append(keywords[categoryName], strings.ToLower(keyword))
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over keyword rows: %w", err)
	}

	return keywords, nil
}

func (controller *Controller) handleMessageByCategory(ctx context.Context, category, message string, inboxID int64, ipAddress string) string {
	//profileID := profile.Id
	//MSISDN := profile.Msisdn
	//
	//switch category {
	//case "sports":
	//	return controller.handleSportsMessage(ctx, profileID, message)
	//case "balance":
	//	return controller.handleBalanceMessage(ctx, profileID)
	//case "withdraw":
	//	return controller.handleWithdrawMessage(ctx, profileID, message)
	//case "deposit":
	//	return controller.handleDepositMessage(ctx, profileID, message)
	//case "bet_status":
	//	return controller.handleBetStatusMessage(ctx, profileID, message)
	//case "bet_cancel":
	//	return controller.handleBetCancelMessage(ctx, profileID, MSISDN, message, ipAddress)
	//case "jackpot":
	//	return controller.handleWeeklyJackpotMessage(ctx, profileID, message)
	//case "daily_jackpot":
	//	return controller.handleDailyJackpotMessage(ctx, profileID, message)
	//case "referral":
	//	return controller.handleDailyJackpotMessage(ctx, profileID, message)
	//case "help":
	//	return controller.GetSMSTemplate(ctx, "HELP")
	//default:
	//	return controller.GetSMSTemplate(ctx, "JOIN")
	//}
	return ""
}

// Helper functions

func (controller *Controller) handleSportsMessage(ctx context.Context, profileID int64, message string) string {
	res := controller.GetSMSGames(ctx, profileID, 10)
	return res
}

func (controller *Controller) handleBalanceMessage(ctx context.Context, profileID int64) string {
	return controller.GetBalance(ctx, profileID)
}

func (controller *Controller) handleWithdrawMessage(ctx context.Context, profileID int64, message string) string {
	parts := strings.Split(message, "#")
	if len(parts) < 2 {
		return "Sorry you have entered invalid amount to withdraw\nTo withdraw send the word\nwithdraw#AMOUNT"
	}
	amount, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return "Sorry you have entered invalid amount to withdraw\nTo withdraw send the word\nwithdraw#AMOUNT"
	}
	return controller.DoWithdraw(ctx, profileID, amount)
}

func (controller *Controller) handleDepositMessage(ctx context.Context, profileID int64, message string) string {
	parts := strings.Split(message, "#")
	if len(parts) < 2 {
		return "Sorry you have entered invalid amount to deposit\nTo deposit send the word\ndeposit#AMOUNT"
	}
	amount, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return "Sorry you have entered invalid amount to deposit\nTo deposit send the word\ndeposit#AMOUNT"
	}
	return controller.DoStk(ctx, profileID, amount)
}

func (controller *Controller) handleBetStatusMessage(ctx context.Context, profileID int64, message string) string {
	parts := strings.Split(message, "#")
	if len(parts) != 2 {
		return "You have entered invalid key word, to check your bet, reply with status#betID"
	}
	return controller.betStatus(ctx, profileID, parts[1])
}

func (controller *Controller) handleBetCancelMessage(ctx context.Context, profileID, MSISDN int64, message, IpAddress string) string {
	parts := strings.Split(message, "#")
	if len(parts) != 2 {
		return "You have entered invalid key word, to cancel your bet, reply with cancel#Code"
	}
	//controller.betCancel(ctx, profileID, MSISDN, parts[1], IpAddress)
	return "Not Available"
}

func (controller *Controller) handleReferral(ctx context.Context, profileID int64, message string) string {
	parts := strings.Split(message, "#")
	if len(parts) != 2 {
		return "You have entered invalid key word, to check your bet, reply with refer#referralCode"
	}
	return controller.betStatus(ctx, profileID, parts[1])
}

// Utility function to check for exact matches
func (controller *Controller) isExactMatch(message string, keywords []string) bool {
	for _, keyword := range keywords {
		if message == keyword {
			return true
		}
	}
	return false
}

//func (controller *Controller) createOrGetUser(ctx context.Context, msisdn string) (*identity.Profile, error) {
//	ms := strings.ReplaceAll(msisdn, "+", "")
//	parsedMsisdn, err := strconv.ParseInt(ms, 10, 64)
//	if err != nil {
//		return nil, fmt.Errorf("invalid phone number %s: %w", ms, err)
//	}
//
//	identityResponse, err := controller.IdentityServiceClient.CreateUser(ctx, &identity.NewUser{Msisdn: parsedMsisdn})
//	if err != nil {
//		return nil, err
//	}
//
//	if identityResponse.Profile.Status == -1 {
//		return controller.updateUserStatus(ctx, identityResponse.Profile.Id, 1)
//	}
//
//	return identityResponse.Profile, nil
//}
//
//func (controller *Controller) updateUserStatus(ctx context.Context, profileID int64, newStatus int) (*identity.Profile, error) {
//	changeStatus, err := controller.IdentityServiceClient.ChangeStatus(ctx, &identity.StatusRequest{
//		Status:    int32(newStatus),
//		ProfileID: int32(profileID),
//	})
//	if err != nil {
//		return nil, err
//	}
//	return &identity.Profile{Id: int64(changeStatus.ProfileID), Status: int64(newStatus)}, nil
//}

func (controller *Controller) saveInbox(ctx context.Context, u *models.Inbox) error {
	dbUtils := goutils.Db{DB: controller.DB, Context: ctx}
	inserts := map[string]interface{}{
		"inbox_id": u.InboxID,
		"message":  u.Message,
		"msisdn":   u.Msisdn,
	}
	_, err := dbUtils.UpsertWithContext("inbox", inserts, nil)
	return err
}

func (controller *Controller) handleError(ctx context.Context, err error, description string, data interface{}) (int, interface{}) {
	logrus.WithContext(ctx).
		WithFields(logrus.Fields{
			constants.DESCRIPTION: description,
			constants.DATA:        data,
		}).
		Error(err.Error())

	return http.StatusInternalServerError, models.ErrorResponse{
		ErrorCode:    http.StatusInternalServerError,
		ErrorMessage: "Internal server error",
	}
}
