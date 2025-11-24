package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"context"
	"strconv"
	"strings"

	"os"

	"github.com/BenBera/shortcode-service/app/constants"

	"github.com/BenBera/shortcode-service/app/library"
	"github.com/BenBera/shortcode-service/app/models"
	"github.com/labstack/echo/v4"
	goutils "github.com/mudphilo/go-utils"
	"github.com/sirupsen/logrus"

	"github.com/labstack/gommon/log"

	"time"
)

var ussdGame = "{number}.{match_name} {time} {game_id} {line} "

var ussdgamesTemplate = "Top games:{line}{line}{games}{line}00. Previous{line}88. Next"

//const  AMOUNT_LEVEL = "888"

const UssdGameCount = 3

var INVALID_INPUT = models.UssdResponse{
	Text:         "Invalid input. Please try again.",
	ResponseType: "END",
}

const MAINRESPONSE = "Karibu 9UBet:\n1. Top games\n\n2. Deposit/withdraw\n3. My account\n4. Customer Care \n5. Exit"

func (controller *Controller) IncomingUSSD(c echo.Context) error {

	ctx, span := controller.Tracer.Start(c.Request().Context(), "USSD")
	defer span.End()

	payload := goutils.GetJSONRawBody(c)

	session, _ := goutils.GetString(payload, "sessionId", "")
	msisdnstr, _ := goutils.GetString(payload, "msisdn", "")
	msisdn, _ := strconv.ParseInt(msisdnstr, 10, 64)
	//network,_:= goutils.GetString(payload,"network","")
	text, _ := goutils.GetString(payload, "text", "")
	level := 0

	parts := strings.Split(text, "*")
	userResponse := strings.TrimSpace(parts[len(parts)-1])

	log.Printf("the request has %v", payload)

	ipAddress := c.RealIP()

	//register user if not existing
	t0 := time.Now().UnixMilli()

	t1 := time.Now().UnixMilli()
	log.Printf("time to save %dms ", t1-t0)

	//incoming request
	//get latest  session from db
	firstlevelKey := fmt.Sprintf("FIRST:%s:%d", session, msisdn)
	key1, er1 := controller.getLevelFromRedis(firstlevelKey)
	log.Printf("first error is %v AND  key is %v", er1, key1)

	if key1 == -1 || key1 == -2 || key1 == -3 || key1 == 0 {
		response, err := controller.handleFirstLevelKey(ctx, session, msisdn, userResponse, firstlevelKey)
		if err != nil {
			return RespondRaw(c, http.StatusInternalServerError, models.ErrorResponse{
				ErrorCode:    http.StatusInternalServerError,
				ErrorMessage: fmt.Sprintf("internal server error %v", err),
			})
		}
		return RespondRaw(c, http.StatusOK, response)
	}

	dbUtils := goutils.Db{DB: controller.DB, Context: ctx}
	dbUtilslave := goutils.Db{DBSlave: controller.DBSlave, Context: ctx}
	dbUtilslave.SetQuery("SELECT ussd_level FROM ussd_request WHERE session_id = ? LIMIT 1 ")
	dbUtilslave.SetParams(session)

	er := dbUtilslave.FetchOneSlave().Scan(&level)

	if er != nil {
		level := 0

		inserts := map[string]interface{}{
			"ussd_level": level,
			"session_id": session,
			"msisdn":     msisdn,
			"text":       text,
		}

		_, err := dbUtils.UpsertWithContext("ussd_request", inserts, nil)
		if err != nil {

			logrus.WithContext(ctx).
				WithFields(logrus.Fields{
					constants.DESCRIPTION: "Failed to create ussd_request",
					constants.DATA:        inserts,
				}).
				Error(err.Error())

			return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
				ErrorCode:    http.StatusBadRequest,
				ErrorMessage: fmt.Sprintf("internal server error %v", err),
			})
		}

	}

	profilekey := fmt.Sprintf("profid:%v", session)
	profid, _ := library.GetRedisKey(controller.RedisConn, profilekey)
	//if err != nil {
	//	userprofile, err := controller.IdentityServiceClient.GetProfileByMsisdn(ctx, &usermsisdn)
	//	if err != nil {
	//		firstresponse := models.UssdResponse{"Session failed. Try again later", "END"}
	//		return RespondRaw(c, http.StatusOK, firstresponse)
	//
	//	}
	//	profid = fmt.Sprintf("%v", userprofile.Id)
	//}
	profileid, _ := strconv.ParseInt(profid, 10, 64)
	log.Printf("the profile id is %v", profileid)
	response := controller.Ussdflow(ctx, userResponse, level, profileid, session, msisdn, ipAddress)
	log.Printf("the profile id is %v | msisdn %v | user Response %v | response %v", profileid, msisdn, userResponse, response.Text)

	//store in db each session hopes request

	sessionhopeinserts := map[string]interface{}{
		"user_input": userResponse,
		"session_id": session,
		"msisdn":     msisdn,
		"response":   response.Text,
	}
	dbUtils.Insert("session_hopes", sessionhopeinserts)

	return RespondRaw(c, http.StatusOK, response)

}

// first level before main menu
func (controller *Controller) handleFirstLevelKey(ctx context.Context, session string, msisdn int64, userResponse string, firstlevelKey string) (models.UssdResponse, error) {
	key1, er1 := controller.getLevelFromRedis(firstlevelKey)
	log.Printf("first error is %v AND key is %v", er1, key1)

	switch key1 {
	//case -3:
	//	return controller.handleKeyMinusThree(ctx, session, userResponse, firstlevelKey, codeRequest)
	case -2:
		return controller.handleKeyMinusTwo(ctx, session, msisdn, userResponse, firstlevelKey)
	case -1:
		return controller.handleKeyMinusOne(ctx, session, msisdn, userResponse, firstlevelKey)
	case 0:
		return controller.handleKeyZero(ctx, session, msisdn, firstlevelKey)
	case -4:
		return controller.handleKeyMinusFour(ctx, session, msisdn, userResponse, firstlevelKey)
	default:
		return models.UssdResponse{Text: "Invalid key value.", ResponseType: "END"}, nil
	}
}

//// request otp
//func (controller *Controller) handleKeyMinusThree(ctx context.Context, session string, userResponse string, firstlevelKey string) (models.UssdResponse, error) {
//
//	cacheDuration := 5 * time.Minute
//	response, err := controller.IdentityServiceClient.RequestCode(ctx, codeRequest)
//	ussdtext := "An error occurred while processing your request. Please try again later."
//	ussdresponse := "END"
//	if err == nil {
//		ussdtext = response.Description + "\nVisit www.maybets.com/Maybets APP to complete your Registration"
//		ussdresponse = "CON"
//	}
//	firstresponse := models.UssdResponse{ussdtext, ussdresponse}
//	library.SetRedisKeyWithExpiry(controller.RedisConn, firstlevelKey, "-3", int(cacheDuration.Seconds()))
//	return firstresponse, nil
//
//}

// referral sub menu
func (controller *Controller) handleKeyMinusTwo(ctx context.Context, session string, msisdn int64, userResponse string, firstlevelKey string) (models.UssdResponse, error) {
	if len(userResponse) < 3 {
		return models.UssdResponse{Text: "Referral code should be greater than three characters", ResponseType: "END"}, nil
	}

	firstresponse := models.UssdResponse{Text: MAINRESPONSE, ResponseType: "CON"}

	return firstresponse, nil
}

// registration
func (controller *Controller) handleKeyMinusOne(ctx context.Context, session string, msisdn int64, userResponse string, firstlevelKey string) (models.UssdResponse, error) {
	firstresponse := models.UssdResponse{MAINRESPONSE, "CON"}
	cacheDuration := 5 * time.Minute
	if userResponse == "1" {
		firstresponse = models.UssdResponse{Text: "Enter Password to match xxxx requirements.", ResponseType: "CON"}
		library.SetRedisKeyWithExpiry(controller.RedisConn, firstlevelKey, "-4", int(cacheDuration.Seconds()))
	} else if userResponse == "2" {
		firstresponse = models.UssdResponse{"Enter Referral CODE.", "CON"}
		library.SetRedisKeyWithExpiry(controller.RedisConn, firstlevelKey, "-2", int(cacheDuration.Seconds()))
	} else {
		firstresponse = models.UssdResponse{"You have canceled your registration. For help, call 0701001000.", "END"}
	}
	return firstresponse, nil
}

// passowrd sub menu
func (controller *Controller) handleKeyMinusFour(ctx context.Context, session string, msisdn int64, userResponse string, firstlevelKey string) (models.UssdResponse, error) {
	firstresponse := models.UssdResponse{Text: MAINRESPONSE, ResponseType: "CON"}
	//regex check to match password requirements
	if len(userResponse) < 6 {
		return models.UssdResponse{Text: "Password does not meet requirements", ResponseType: "END"}, nil
	}

	endpoint := "https://9ubet.co.ke/api/userRegByMobile"
	payload := models.UserRegByMobileRequest{
		Mobile:   strconv.FormatInt(msisdn, 10),
		VKey:     "",
		VCode:    "",
		Password: userResponse, //determine which fields are required
		TGCode:   "",
		TID:      "",
		DeviceID: "",
		AC:       "userRegByMobile",
	}
	status, _ := goutils.HTTPPost(endpoint, nil, payload)
	if status != 200 {
		firstresponse = models.UssdResponse{Text: "Failed to create account. Try again later", ResponseType: "END"}
		return firstresponse, nil
	}

	return firstresponse, nil
}

func (controller *Controller) handleKeyZero(ctx context.Context, session string, msisdn int64, firstlevelKey string) (models.UssdResponse, error) {
	cacheDuration := 5 * time.Minute
	profilekey := fmt.Sprintf("profid:%v", session)
	profid, err := library.GetRedisKey(controller.RedisConn, profilekey)
	firstresponse := models.UssdResponse{Text: MAINRESPONSE, ResponseType: "CON"}
	if err != nil {
		//we need a way to check if user exists else we direct  them to account creation
		//    usermsisdn := identity.Msisdn{
		//	Msisdn: msisdn,
		//}
		//userprofile, err := controller.IdentityServiceClient.GetProfileByMsisdn(ctx, &usermsisdn)
		//if err != nil {
		//	firstresponsenew := models.UssdResponse{"Karibu Maybets to complete Registration Press 1 or 2 to accept Terms and conditions at maybets.com/terms-and-condition\n\n1.Join\n2.Join using Referral code ", "CON"}
		//	library.SetRedisKeyWithExpiry(controller.RedisConn, firstlevelKey, "-1", int(cacheDuration.Seconds()))
		//	return firstresponsenew, nil
		//}

		log.Printf("no first key")

		//if userprofile.Status == -1 {
		//	response, err := controller.IdentityServiceClient.RequestCode(ctx, codeRequest)
		//	ussdtext := "An error occurred while processing your request. Please try again later."
		//	ussdresponse := "END"
		//	if err == nil {
		//		ussdtext = response.Description + "\nVisit www.maybets.com/Maybets APP to complete your Registration"
		//		ussdresponse = "CON"
		//	}
		//	firstresponse = models.UssdResponse{ussdtext, ussdresponse}
		//	library.SetRedisKeyWithExpiry(controller.RedisConn, firstlevelKey, "-3", int(cacheDuration.Seconds()))
		//} else {
		ussdtext := "\nEnter your desired password to complete registration"
		ussdresponse := "CON"
		firstresponse = models.UssdResponse{Text: ussdtext, ResponseType: ussdresponse}
		library.SetRedisKeyWithExpiry(controller.RedisConn, firstlevelKey, "-4", int(cacheDuration.Seconds()))
		//}
		//profid = fmt.Sprintf("%v", userprofile.Id)
		library.SetRedisKeyWithExpiry(controller.RedisConn, profilekey, profid, int(cacheDuration.Seconds()))
	}

	return firstresponse, nil
}

func (controller *Controller) Ussdflow(ctx context.Context, userResponse string, level int, profileID int64, session string, msisdn int64, ipAddress string) models.UssdResponse {

	var response models.UssdResponse

	ussdtext := ""
	ussdresponse := ""
	dbUtils := goutils.Db{DB: controller.DB, Context: ctx}
	// Convert userResponse to an integer
	responseInt, err := strconv.Atoi(userResponse)

	if level == 0 {

		// Convert userResponse to an integer
		if err != nil || responseInt < 1 || responseInt > 6 {
			ussdtext = MAINRESPONSE
			ussdresponse = "CON"

		} else {

			inserts := map[string]interface{}{
				"ussd_level": responseInt,
			}

			conditions := map[string]interface{}{
				"session_id": session,
				"msisdn":     msisdn,
			}

			dbUtils.UpdateWithContext("ussd_request", conditions, inserts)

			//call the function again
			return controller.Ussdflow(ctx, userResponse, responseInt, profileID, session, msisdn, ipAddress)

		}

	} else if level == 1 {

		// get three games at time

		return controller.GetUSSDGames(ctx, profileID, session, 50, userResponse, msisdn, ipAddress)

	} else if level == 2 {

		//deposit and withdraw

		return controller.depositwithdraw(ctx, profileID, msisdn, session, userResponse)

	} else if level == 3 {

		return controller.account(ctx, profileID, msisdn, session, userResponse)

	} else if level == 4 {
		ussdtext = "For assistance, contact Customer Care:Call: +254701001000 WhatsApp: +254704498098. Thank you!"
		ussdresponse = "END"

	} else {
		ussdtext = "Thank you for using 9UBet USSD. Good luck and see you next time!"
		ussdresponse = "END"

	}
	response = models.UssdResponse{ussdtext, ussdresponse}
	return response
}

func (controller *Controller) GetUSSDGames(ctx context.Context, profileID int64, session string, count int, action string, msisdn int64, ipAddress string) models.UssdResponse {

	key := fmt.Sprintf("USSD_GAMES:%s:%d", session, profileID)
	pageKey := fmt.Sprintf("USSD_GAMES_PAGE:%s:%d", session, profileID)
	levelKey := fmt.Sprintf("USSD_LEVEL:%s:%d", session, profileID) // move to next levels to help go back and
	betslipKey := fmt.Sprintf("BETSLIP:%d", profileID)
	cacheDuration := 5 * time.Minute
	dbUtils := goutils.Db{DB: controller.DB, Context: ctx}
	var games []models.UssdGame
	var level int
	currentPage := 0
	firstCall := 0

	// Retrieve current page from Redis
	currentPage, er := controller.getLevelFromRedis(pageKey)
	if er != nil {
		firstCall = 1
	}

	// Retrieve games from Redis
	redisData, err := library.GetRedisKey(controller.RedisConn, key)
	if err == nil {
		err = json.Unmarshal([]byte(redisData), &games)
		if err != nil {
			log.Printf("error unmarshalling redis data: %s", err.Error())
		}
	}

	// Retrieve level from Redis
	level, _ = controller.getLevelFromRedis(levelKey)

	// If Redis data is empty or expired, fetch from the database
	if len(games) == 0 {
		//	lastPriority := 0
		//data := fx.RequestSMSGame{
		//	Count:        int64(count),
		//	FromPriority: int64(lastPriority),
		//}
		//
		//response, err := controller.FixtureServiceClient.GetSMSGames(ctx, &data)
		//
		//if err != nil {
		//	logrus.WithContext(ctx).
		//		WithFields(logrus.Fields{
		//			"DESCRIPTION": "Failed to get games from fixture service",
		//			"DATA":        &data,
		//		}).
		//		Error(err.Error())
		//	return models.UssdResponse{
		//		Text:         "Welcome to Maybets where winners are made",
		//		ResponseType: "END",
		//	}
		//}
		//
		//// Convert response.Games (slice of pointers to fixture.SMSGame) to slice of UssdGame
		//for _, smsGame := range response.Games {
		//	games = append(games, ConvertSMSGameToUssdGame(*smsGame))
		//}
		//
		//// Store games in Redis
		//gamesData, err := json.Marshal(games)
		//if err == nil {
		//	controller.setLevelInRedis(key, string(gamesData), cacheDuration)
		//
		//}
	}

	//  place bet
	if level == 888 {
		amount, err := strconv.Atoi(action)
		if err != nil {
			return models.UssdResponse{
				Text:         "Enter Valid Bet Amount",
				ResponseType: "CON",
			}
		}

		// Retrieve the betslip from Redis
		betslipData, err := library.GetRedisKey(controller.RedisConn, betslipKey)
		if err != nil {
			return models.UssdResponse{
				Text:         "Error retrieving betslip. Please try again later.",
				ResponseType: "END",
			}
		}

		var betslip []models.UssdGameSlip
		err = json.Unmarshal([]byte(betslipData), &betslip)
		if err != nil {
			return models.UssdResponse{
				Text:         "Error processing betslip. Please try again later.",
				ResponseType: "END",
			}
		}

		// Build the betslip string
		var betslipParts []string
		for _, game := range betslip {
			action := game.Action

			// Replace 2 with X and 3 with 2
			if action == "2" {
				action = "X"
			} else if action == "3" {
				action = "2"
			}

			betslipParts = append(betslipParts, fmt.Sprintf("%d#%s", game.GameID, action))
		}

		// Join the betslip parts with "#" and append the amount
		betslipString := strings.Join(betslipParts, "#")
		message := fmt.Sprintf("%s#%d", betslipString, amount)

		//message, err = controller.ussdBet(ctx, profileID, message, ipAddress)
		//if err == nil {
		//	// Clear the betslip by deleting the Redis key
		//	err := library.DeleteRedisKey(controller.RedisConn, betslipKey)
		//	if err != nil {
		//		logrus.WithContext(ctx).
		//			WithFields(logrus.Fields{
		//				"DESCRIPTION": "Error clearing betslip from redis",
		//				"DATA":        betslipKey,
		//			}).
		//			Error(err.Error())
		//	}
		//}

		// Return the USSD response with the message
		return models.UssdResponse{
			Text:         message,
			ResponseType: "END",
		}
	}

	// enter amount for bet
	if level == 999 {
		switch action {
		case "1":
			// Store levelKey in Redis with the value of the action
			err = library.SetRedisKeyWithExpiry(controller.RedisConn, levelKey, "888", int(cacheDuration.Seconds()))
			if err != nil {
				logrus.WithContext(ctx).
					WithFields(logrus.Fields{
						"DESCRIPTION": "Error saving level key",
						"DATA":        levelKey,
					}).
					Error(err.Error())
			}

			return models.UssdResponse{
				Text:         "Enter Bet Amount",
				ResponseType: "CON",
			}
		case "2", "3":

			ussdLevel := 1

			if action == "3" {
				ussdLevel = 0
				// Clear the betslip by deleting the Redis key
				err := library.DeleteRedisKey(controller.RedisConn, betslipKey)
				if err != nil {
					logrus.WithContext(ctx).
						WithFields(logrus.Fields{
							"DESCRIPTION": "Error clearing betslip from redis",
							"DATA":        betslipKey,
						}).
						Error(err.Error())
				}
			}
			log.Printf("main level %v", ussdLevel)
			inserts := map[string]interface{}{
				"ussd_level": ussdLevel,
			}
			conditions := map[string]interface{}{
				"session_id": session,
				"msisdn":     msisdn,
			}
			//clear level to start a fresh

			dbUtils.UpdateWithContext("ussd_request", conditions, inserts)

			//delete page key to prevent from selecting 1 game
			controller.clearRedisKeys(levelKey, pageKey)
			log.Printf("ussd last step of placing bet %v", action)

			return controller.Ussdflow(ctx, "1", ussdLevel, profileID, session, msisdn, ipAddress)
		default:

			return INVALID_INPUT

		}
	}

	// Handle pagination and selection based on the action
	var selectedGame models.UssdGame

	switch action {
	case "88":
		if currentPage < (len(games)+UssdGameCount-1)/UssdGameCount-1 {
			currentPage++
		}

	case "00":
		if currentPage > 0 {
			currentPage--
		} else {

			//go to main menu
			inserts := map[string]interface{}{
				"ussd_level": 0,
			}
			conditions := map[string]interface{}{
				"session_id": session,
				"msisdn":     msisdn,
			}
			//clear level to start a fresh
			dbUtils.UpdateWithContext("ussd_request", conditions, inserts)

			//clear redis key
			controller.clearRedisKeys(levelKey, pageKey)

			return controller.Ussdflow(ctx, "", 0, profileID, session, msisdn, ipAddress)

		}
	default:
		if action == "1" && firstCall == 1 {

			currentPage = 0
		} else {
			// Convert action to int
			actionint, _ := strconv.Atoi(action)

			if actionint < 1 || actionint > len(games) {
				return INVALID_INPUT
			}

			// Determine page and selection within page
			currentPage = (actionint - 1) / UssdGameCount
			selectedGameIndex := (actionint - 1) % UssdGameCount

			// Ensure valid selection
			startIndex := currentPage * UssdGameCount
			index := startIndex + selectedGameIndex

			if index >= len(games) {
				return INVALID_INPUT
			}

			selectedGame = games[index]

			// Store selected level in Redis
			cachedLevel := action
			if level > 0 {
				cachedLevel = "999"
			}

			err := library.SetRedisKeyWithExpiry(controller.RedisConn, levelKey, cachedLevel, int(cacheDuration.Seconds()))
			if err != nil {
				logrus.WithContext(ctx).
					WithFields(logrus.Fields{
						"DESCRIPTION": "Error saving level key",
						"DATA":        levelKey,
					}).
					Error(err.Error())
			}

			// Handle betslip addition
			if level > 0 {
				log.Printf("Selected game is %v", selectedGame)
				return models.UssdResponse{
					Text:         controller.Addtobetslip(ctx, profileID, selectedGame, action, level),
					ResponseType: "CON",
				}
			}

		}

	}
	// Handle fetching and displaying the selected game with the desired format
	if selectedGame.GameID != 0 {
		msg := fmt.Sprintf(
			"%s %d %s\n1. Home %.2f\n2. Draw %.2f\n3. Away %.2f",
			selectedGame.Name,
			selectedGame.GameID,
			goutils.StringToTime(selectedGame.Date).Format("3:04 PM"),
			selectedGame.Home,
			selectedGame.Draw,
			selectedGame.Away,
		)

		return models.UssdResponse{
			Text:         msg,
			ResponseType: "CON",
		}
	}

	// Paginate and display games
	startIndex := currentPage * UssdGameCount
	endIndex := startIndex + UssdGameCount

	if endIndex > len(games) {
		endIndex = len(games)
	}

	var currentGames []models.UssdGame = games[startIndex:endIndex]
	var parts []string
	for i, r := range currentGames {
		sms := ussdGame
		sms = strings.Replace(sms, "{number}", fmt.Sprintf("%d", startIndex+i+1), -1)
		sms = strings.Replace(sms, "{game_id}", fmt.Sprintf("%d", r.GameID), -1)
		sms = strings.Replace(sms, "{match_name}", r.Name, -1)
		sms = strings.Replace(sms, "{time}", goutils.StringToTime(r.Date).Format("3:04 PM"), -1)
		parts = append(parts, sms)
	}

	message := strings.Join(parts, "{space}")
	msg := ussdgamesTemplate
	msg = strings.Replace(msg, "{games}", message, -1)
	msg = strings.Replace(msg, "{line}", "\n", -1)
	msg = strings.Replace(msg, "{space}", " ", -1)
	if len(currentGames) == 0 {

		msg = "No more USSD games for the day"

	}

	// Update the page key in Redis
	err = library.SetRedisKeyWithExpiry(controller.RedisConn, pageKey, strconv.Itoa(currentPage), int(cacheDuration.Seconds()))
	if err != nil {
		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				"DESCRIPTION": "Error saving page key",
				"DATA":        pageKey,
			}).
			Error(err.Error())
	}

	return models.UssdResponse{
		Text:         msg,
		ResponseType: "CON",
	}
}

//func ConvertSMSGameToUssdGame(smsGame fx.SMSGame) models.UssdGame {
//	return models.UssdGame{
//		GameID:   int(smsGame.GameID),
//		Name:     smsGame.Name,
//		Date:     smsGame.Date,
//		Priority: int(smsGame.Priority),
//		Home:     smsGame.Home,
//		Away:     smsGame.Away,
//		Draw:     smsGame.Draw,
//	}
//}

func (controller *Controller) Addtobetslip(ctx context.Context, profileID int64, game models.UssdGame, action string, level int) string {
	// Generate Redis key for betslip
	betslipKey := fmt.Sprintf("BETSLIP:%d", profileID)

	// Retrieve the current betslip from Redis
	var betslip []models.UssdGameSlip
	redisData, err := library.GetRedisKey(controller.RedisConn, betslipKey)
	if err == nil {
		err = json.Unmarshal([]byte(redisData), &betslip)
		if err != nil {
			log.Printf("error unmarshalling betslip redis data: %s", err.Error())
		}
	}

	// Convert game to UssdGameSlip
	newSlip := models.UssdGameSlip{
		GameID:    int(game.GameID),
		Action:    action,
		ProfileID: int(profileID),
	}

	// Set the odds based on the action
	switch action {
	case "1":
		newSlip.Odds = game.Home
	case "2":
		newSlip.Odds = game.Draw
	case "3":
		newSlip.Odds = game.Away
	default:
		return "Invalid action. Please try again."
	}
	log.Printf("selected odd is %v and home is  %v away is %v and draw is %v", newSlip.Odds, game.Home, game.Away, game.Draw)

	// Check if the game already exists in the betslip
	gameExists := false
	for i, bet := range betslip {
		if bet.GameID == newSlip.GameID {
			// Update action and odds for the existing game
			betslip[i].Action = newSlip.Action
			betslip[i].Odds = newSlip.Odds
			gameExists = true
			break
		}
	}

	if !gameExists {
		// Add the new game to the betslip
		betslip = append(betslip, newSlip)
	}

	// Store the updated betslip in Redis
	betslipData, err := json.Marshal(betslip)
	if err == nil {
		err = library.SetRedisKeyWithExpiry(controller.RedisConn, betslipKey, string(betslipData), 10*60) // 10 minutes
		if err != nil {
			logrus.WithContext(ctx).
				WithFields(logrus.Fields{
					"DESCRIPTION": "Error saving betslip to redis",
					"DATA":        betslipKey,
				}).
				Error(err.Error())
		}
	}

	// Count the number of games in the betslip
	gameCount := len(betslip)

	// Calculate the total odds
	totalOdds := 1.0
	for _, bet := range betslip {
		totalOdds *= float64(bet.Odds)
	}

	// Construct the message to return to the user
	msg := fmt.Sprintf(
		"Your betslip has %d game(s)\nTotal odds %.2f\n1. Place bet\n2. Add more games\n3. Clear betslip\n88. Main menu",
		gameCount,
		totalOdds,
	)

	return msg
}

// Get and parse level from Redis
func (controller *Controller) getLevelFromRedis(jplevelKey string) (int, error) {
	levelData, err := library.GetRedisKey(controller.RedisConn, jplevelKey)
	if err != nil {

		return 0, err
	}

	return strconv.Atoi(levelData)
}

// Set level in Redis
func (controller *Controller) setLevelInRedis(jplevelKey, updatelevel string, cacheDuration time.Duration) error {
	return library.SetRedisKeyWithExpiry(controller.RedisConn, jplevelKey, updatelevel, int(cacheDuration.Seconds()))
}

// Reset game data and return to the main menu
func (controller *Controller) resetGame(ctx context.Context, jplevelKey string, level int, session string, msisdn int64) {
	library.DeleteRedisKey(controller.RedisConn, jplevelKey)

	// Change level to 0 in the database
	dbUtils := goutils.Db{DB: controller.DB, Context: ctx}
	inserts := map[string]interface{}{"ussd_level": 0}
	conditions := map[string]interface{}{"session_id": session, "msisdn": msisdn}
	dbUtils.UpdateWithContext("ussd_request", conditions, inserts)
}

// Get the category ID based on the level and action
func (controller *Controller) getJackpotCategoryID(level int) int64 {
	switch level {
	case 11, 111:

		// return controller.getCategoryID("jackpot_category_id")
		return 6

	case 21, 121:

		// return controller.getCategoryID("mb8_jackpot_category_id")
		return 5

	}
	return 0
}

// Retrieve the category ID from environment variables
func (controller *Controller) getCategoryID(envVar string) int64 {
	categoryID, _ := strconv.ParseInt(os.Getenv(envVar), 10, 64)
	return categoryID
}

// Check if the action is valid
func (controller *Controller) isValidAction(action string) bool {
	return action == "1" || action == "2" || action == "3"
}

// Retrieve game data from Redis
func (controller *Controller) retrieveGameData(key, headerKey, pageKey string) (int, []string, string) {
	// Fetch games and header from Redis
	games := controller.getGamesFromRedis(key)
	jackpotHeader := controller.getHeaderFromRedis(headerKey)
	currentPage := controller.getCurrentPageFromRedis(pageKey)

	return currentPage, games, jackpotHeader
}

// Get games from Redis
func (controller *Controller) getGamesFromRedis(key string) []string {
	var games []string
	redisData, err := library.GetRedisKey(controller.RedisConn, key)
	if err == nil {
		_ = json.Unmarshal([]byte(redisData), &games)
	}

	return games
}

// Get header from Redis
func (controller *Controller) getHeaderFromRedis(headerKey string) string {
	header, err := library.GetRedisKey(controller.RedisConn, headerKey)
	if err != nil {
		return ""
	}
	return header
}

// Get current page from Redis
func (controller *Controller) getCurrentPageFromRedis(pageKey string) int {
	pageData, err := library.GetRedisKey(controller.RedisConn, pageKey)
	if err != nil {
		return 0
	}

	currentPage, err := strconv.Atoi(pageData)
	if err != nil {
		return 0
	}

	return currentPage
}

// Update betslip in Redis
func (controller *Controller) updateBetslip(betslipKey, action string) {
	betslipData, err := library.GetRedisKey(controller.RedisConn, betslipKey)
	if err != nil || betslipData == "" {
		betslipData = action
	} else {
		betslipData += "," + action
	}
	library.SetRedisKeyWithExpiry(controller.RedisConn, betslipKey, betslipData, 900)
}

// Generate game display string

func (controller *Controller) generateGameDisplay(games []string, jackpotHeader string, currentPage int, categoryID int64) string {
	jackpot := "Jackpot "
	if categoryID == 5 {
		jackpot = "Daily Jackpot "
	}

	// Current game e.g. "Team A vs Team B\n1= 2.46 | X= 3.00 | 2= 2.75"
	currentGame := games[currentPage]
	lines := strings.Split(currentGame, "\n")

	if len(lines) < 2 {
		return "Game data unavailable"
	}

	// First line is match name
	match := lines[0]
	// Second line is odds
	oddsLine := lines[1]

	// Split odds by '|'
	oddsParts := strings.Split(oddsLine, "|")
	labels := []string{"(Home)", "(Draw)", "(Away)"}
	var formattedOdds []string

	for i, part := range oddsParts {
		odd := strings.TrimSpace(strings.TrimPrefix(part, "1="))
		odd = strings.TrimPrefix(odd, "X=")
		odd = strings.TrimPrefix(odd, "2=")
		formattedOdds = append(formattedOdds, fmt.Sprintf("%d. %s %s", i+1, labels[i], strings.TrimSpace(odd)))
	}

	return fmt.Sprintf(
		"%s%s\nGame no. %d\n%s\n%s",
		jackpot,
		jackpotHeader,
		currentPage+1,
		match,
		strings.Join(formattedOdds, "\n"),
	)
}

// Clear Redis keys
func (controller *Controller) clearRedisKeys(keys ...string) {
	for _, key := range keys {
		library.DeleteRedisKey(controller.RedisConn, key)
	}
}

// Update current page in Redis
func (controller *Controller) updatePageInRedis(pageKey string, currentPage int) {
	currentPage++
	pageData := strconv.Itoa(currentPage)
	library.SetRedisKeyWithExpiry(controller.RedisConn, pageKey, pageData, 600)
}

// withdraw
func (controller *Controller) depositwithdraw(ctx context.Context, profileID int64, msisdn int64, session string, action string) models.UssdResponse {

	dwlevel := fmt.Sprintf("DW_LEVEL:%s:%d", session, profileID)
	cacheDuration := 2 * time.Minute
	updatelevel := "1"
	ussdresponse := "CON"
	ussdtext := ""

	// Retrieve and parse level from Redis
	level, err := controller.getLevelFromRedis(dwlevel)
	if err != nil {
		level = 0
	}
	if action == "00" {
		dbUtils := goutils.Db{DB: controller.DB, Context: ctx}

		updatelevel = "0"
		dblevel, _ := strconv.Atoi(updatelevel)

		inserts := map[string]interface{}{
			"ussd_level": dblevel,
		}

		conditions := map[string]interface{}{
			"session_id": session,
			"msisdn":     msisdn,
		}
		//clear level to start a fresh
		dbUtils.UpdateWithContext("ussd_request", conditions, inserts)

		controller.setLevelInRedis(dwlevel, updatelevel, cacheDuration)
		return controller.Ussdflow(ctx, "0", 0, profileID, session, msisdn, "")
	}
	switch level {
	case 1:
		if action == "1" {
			ussdtext = "Enter Amount to Deposit\n"
			updatelevel = "2"
		} else if action == "2" {
			ussdtext = "Enter Amount to Withdraw\n"
			updatelevel = "3"
		}
	case 2:
		// Action handling for level 2
		updatelevel = "2"
		amount, err := strconv.ParseFloat(action, 64)
		if err != nil || amount <= 0 {
			ussdtext = "Invalid input. Please enter a valid Deposit amount.\n"
		} else {
			// Process deposit amount

			ussdtext = controller.DoStk(ctx, profileID, amount)
			ussdresponse = "END"
			updatelevel = "0"

		}
	case 3:
		// Action handling for level 3
		updatelevel = "3"
		amount, err := strconv.ParseFloat(action, 64)
		if err != nil || amount <= 0 {
			ussdtext = "Invalid input. Please enter a valid Withdraw amount.\n"
		} else {
			// Process withdrawal amount
			ussdtext = controller.DoWithdraw(ctx, profileID, amount)
			ussdresponse = "END"
			updatelevel = "0"

		}
	default:
		ussdtext = "Withdraw/Deposit\n1. Deposit\n2. Withdraw.\n00. Main menu"
		updatelevel = "1"
	}

	controller.setLevelInRedis(dwlevel, updatelevel, cacheDuration)

	return models.UssdResponse{Text: ussdtext, ResponseType: ussdresponse}

}

func (controller *Controller) account(ctx context.Context, profileID int64, msisdn int64, session string, action string) models.UssdResponse {
	accountlevel := fmt.Sprintf("AC_LEVEL:%s:%d", session, profileID)
	cacheDuration := 2 * time.Minute
	updatelevel := "1"
	ussdresponse := "CON"
	ussdtext := ""
	// Trim action before passing it to ussdbetStatus
	trimmedAction := strings.TrimSpace(action)
	bettype := "normal"
	// Retrieve and parse level from Redis
	level, err := controller.getLevelFromRedis(accountlevel)
	if err != nil {
		level = 0
	}
	if action == "00" {
		dbUtils := goutils.Db{DB: controller.DB, Context: ctx}

		updatelevel = "0"
		dblevel, _ := strconv.Atoi(updatelevel)

		inserts := map[string]interface{}{
			"ussd_level": dblevel,
		}
		conditions := map[string]interface{}{
			"session_id": session,
			"msisdn":     msisdn,
		}
		//clear level to start a fresh
		//clear bet status keys
		if level == 3 {
			bettype = "jackpot"
		}

		bskey := fmt.Sprintf("BS_Games:%s:%d:%s", session, profileID, bettype)
		bsheaderkey := fmt.Sprintf("BS_Games:H:%s:%d:%s", session, profileID, bettype)
		bspageKey := fmt.Sprintf("BS_PAGE:%s:%d:%s", session, profileID, bettype)

		dbUtils.UpdateWithContext("ussd_request", conditions, inserts)
		controller.clearRedisKeys(bskey, bsheaderkey, bspageKey)

		controller.setLevelInRedis(accountlevel, updatelevel, cacheDuration)
		return controller.Ussdflow(ctx, "0", 0, profileID, session, msisdn, "")
	}
	switch level {
	case 1:
		if action == "1" {
			//get status
			ussdtext = "Enter Bet ID \n"
			updatelevel = "2"
		} else if action == "2" {
			//get status
			ussdtext = "Enter Jackpot Bet ID \n"
			updatelevel = "3"
		} else {
			ussdtext = INVALID_INPUT.Text
			ussdresponse = INVALID_INPUT.ResponseType
		}
	case 2:

		ussdtext, ussdresponse = controller.ussdbetStatus(ctx, profileID, trimmedAction, session, bettype)
		updatelevel = "2"
	case 3:
		bettype = "jackpot"
		ussdtext, ussdresponse = controller.ussdbetStatus(ctx, profileID, trimmedAction, session, bettype)
		updatelevel = "2"

	default:
		//dt := wallet.BalanceRequest{ProfileID: profileID}
		//
		//walletRes, err := controller.WalletServiceClient.GetBalance(ctx, &dt)
		//log.Printf("wallet response %v", walletRes)
		//if err != nil {
		//	log.Printf("wallet error %v", err.Error())
		//}
		//bal := walletRes.CurrentBalance
		//
		//// Format the USSD text with the extracted balance
		//ussdtext = fmt.Sprintf("My Account\nBalance: %v\nBonus:\nMaybets Points:\n1.Check Bet status\n2.Check Jackpot Bet status\n00. main menu", bal)

	}
	controller.setLevelInRedis(accountlevel, updatelevel, cacheDuration)
	return models.UssdResponse{Text: ussdtext, ResponseType: ussdresponse}

}

//
//func (controller *Controller) ussdBet(ctx context.Context, profileID int64, message, ipAddress string) (string, error) {
//	log.Printf("Input Message: %s", message)
//	log.Printf("Input profile: %d", profileID)
//
//	if ipAddress == "" {
//		return "Error: IP address is required", fmt.Errorf("Error: IP address is required")
//	}
//
//	parts := strings.Split(message, "#")
//	if len(parts) < 3 {
//		log.Printf("Invalid Bet Format")
//		return "Error: Invalid bet format. Please use GameID#pick#amount", fmt.Errorf("Error: Invalid bet format. Please use GameID#pick#amount")
//
//	}
//
//	betType := constants.USSDBET
//	source := constants.SOURCE
//
//	amount, err := strconv.Atoi(parts[len(parts)-1])
//	if err != nil || amount <= 0 {
//		log.Printf("Invalid Bet Stake %d", amount)
//
//		return "Error: Invalid bet amount. Please provide a positive number.", fmt.Errorf("Error: Invalid bet amount. Please provide a positive number.")
//	}
//
//	bet, err := controller.parseBetSlips(ctx, parts[:len(parts)-1])
//	if err != nil {
//		log.Printf("Invalid Bet parsing %s", err.Error())
//
//		return fmt.Sprintf("Error: %s", err.Error()), fmt.Errorf("Error: %s", err.Error())
//	}
//
//	betResponse, err := controller.placeBet(ctx, profileID, float32(amount), source, constants.USSDCHANNELID, ipAddress, betType, bet)
//	if err != nil {
//		errMsg := err.Error()
//
//		// Find and extract the "desc" part of the error message
//		descIndex := strings.Index(errMsg, "desc =")
//		if descIndex != -1 {
//			// Extract the description part and trim any extra spaces
//			description := strings.TrimSpace(errMsg[descIndex+len("desc ="):])
//			log.Printf("Bet Placement Error: %s", description)
//			return description, fmt.Errorf("Bet Placement Error: %s", description)
//		} else {
//			log.Printf("Bet Placement Error2: %s", errMsg)
//			return errMsg, fmt.Errorf("Bet Placement Error2: %s", errMsg)
//		}
//
//	}
//
//	if betResponse.Status == 201 || betResponse.Status == 200 {
//		log.Printf("Success!! %d for %s", betResponse.Status, betResponse.ShareCode)
//
//		return fmt.Sprintf("Bet #%s placed successfully. Please wait for a confirmation message.", betResponse.ShareCode), nil
//	}
//
//	return fmt.Sprintf("Bet placement failed: %s", betResponse.Description), fmt.Errorf("Bet placement failed: %s", betResponse.Description)
//}
