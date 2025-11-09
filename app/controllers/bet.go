package controllers

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BenBera/shortcode-service/app/constants"
	"github.com/BenBera/shortcode-service/app/grpc/betting"
	fx "github.com/BenBera/shortcode-service/app/grpc/fixture"
	"github.com/BenBera/shortcode-service/app/grpc/jackpot"
	"github.com/BenBera/shortcode-service/app/grpc/wallet"
	"github.com/BenBera/shortcode-service/app/library"
	"github.com/BenBera/shortcode-service/app/models"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	goutils "github.com/mudphilo/go-utils"
	"github.com/sirupsen/logrus"
)

var smsGame = "{game_id} {match_name}{line}{time}{line}1={home} | X={draw} | 2={away}"
var smsFreebetGame = "Today Freebet Game{line}{line}{name}{line}{line}1={home},X={draw},2={away}{line}{line}To bet on Freebet Game SMS FREE#PICK TO 29098"
var gamesTemplate = "TOP GAMES{line}{line}{games}{line}--{line}To bet use format GameID#pick#amount{line}PayBill 498098{line}Help 0701001000"
var jackpotGames = "MB8 Jackpot GAMES {date}{line}BETTING CLOSES AT {closing_time}{line}--{line}{games}{line}--{line}SEND JP#YOUR PICKS to 29098 to bet{line}T & Cs Apply."

const NOBETSTATUSGAMES = 2
const USSDCON = "CON"
const USSDEND = "END"

var (
	once      sync.Once
	netClient *http.Client
)

func (controller *Controller) Inbox(c echo.Context) error {
	t0 := time.Now().UnixMilli()
	ctx, span := controller.Tracer.Start(c.Request().Context(), "Inbox")
	defer span.End()

	payload := goutils.GetJSONRawBody(c)

	u := new(models.Inbox)

	jsB, _ := json.Marshal(payload)

	err := json.Unmarshal(jsB, u)
	if err != nil {
		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "Error binding data",
				constants.DATA:        err.Error(),
			}).
			Error(err.Error())

		return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid values",
		})
	}

	// Get the IP address from the Echo context
	ipAddress := c.RealIP()

	statusCode, response := controller.ProcessInbox(ctx, u, false, ipAddress)
	t1 := time.Now().UnixMilli()
	delay := t1 - t0
	if delay > 3000 {
		go library.InsertOrUpdateSMSError(controller.DB, constants.CODELATENCY, float64(delay))
	}
	return RespondRaw(c, statusCode, response)
}

func (controller *Controller) AutoResponse(ctx context.Context, inboxID int64, message string, sdpAutoResponse bool) {

	if sdpAutoResponse || inboxID < 1000 {

		_ = controller.SDPAutoresponse(inboxID, message)
		return

	}
	callbackurl := os.Getenv("shortcode_callback")

	ctx, span := controller.Tracer.Start(ctx, "AutoResponse")
	defer span.End()

	t4 := time.Now().UnixMilli()

	inserts := map[string]interface{}{
		"response": message,
	}

	conditions := map[string]interface{}{
		"inbox_id": inboxID,
	}

	dbUtils := goutils.Db{DB: controller.DB, Context: ctx}

	_, err := dbUtils.UpdateWithContext("inbox", conditions, inserts)
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "Failed to update inbox",
				constants.DATA:        fmt.Sprintf("inbox id %d | %s", inboxID, message),
			}).
			Error(err.Error())

		return
	}

	t5 := time.Now().UnixMilli()
	log.Printf("Update inbox ttl %dms ", t5-t4)

	payload := map[string]interface{}{
		"inbox_id":     inboxID,
		"message":      message,
		"callback_url": callbackurl,
	}

	endpoint := os.Getenv("sms_autoresponse_endpoint")

	apiKey, err := controller.getIntouchToken()
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "Failed to get api-key",
			}).
			Error(err.Error())

		return
	}

	headers := map[string]string{
		"api-key": apiKey,
	}

	st, response := goutils.HTTPPost(endpoint, headers, payload)
	if st < 200 || st > 210 {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: fmt.Sprintf("Invalid status %d sending auto response", st),
				constants.DATA:        response,
			}).
			Error(fmt.Sprintf("Invalid status %d sending auto response", st))
	}

	t6 := time.Now().UnixMilli()
	log.Printf("Send response  ttl %dms ", t6-t5)

	return

}

func (controller *Controller) GetSMSGames(ctx context.Context, profileID int64, count int) string {

	// retrieve last timestamp

	key := fmt.Sprintf("LAST_SMS_QUERY_PRIORITY:%s:%d", goutils.Today(), profileID)

	lastTimestamp := "0"

	lt, err := library.GetRedisKey(controller.RedisConn, key)
	if err != nil {

		log.Printf("error retrieving redis key %s ", err.Error())

	}

	lastTimestamp = lt

	lastPriority, err := strconv.Atoi(lastTimestamp)
	if err != nil {

		lastPriority = 0
	}

	dt := fx.RequestSMSGame{
		Count:        int64(count),
		FromPriority: int64(lastPriority),
	}

	games, err := controller.FixtureServiceClient.GetSMSGames(ctx, &dt)
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "Failed to get games from fixture service",
				constants.DATA:        dt.String(),
			}).
			Error(err.Error())

		return "Welcome to Maybets where winners are made"

	}

	if len(games.Games) == 0 {

		return "No more SMS games for the day"

	}

	var parts []string
	for _, r := range games.Games {

		sms := smsGame
		sms = strings.Replace(sms, "{game_id}", fmt.Sprintf("%d", r.GameID), -1)
		sms = strings.Replace(sms, "{match_name}", r.Name, -1)
		sms = strings.Replace(sms, "{home}", fmt.Sprintf("%0.2f", r.Home), -1)
		sms = strings.Replace(sms, "{draw}", fmt.Sprintf("%0.2f", r.Draw), -1)
		sms = strings.Replace(sms, "{away}", fmt.Sprintf("%0.2f", r.Away), -1)
		sms = strings.Replace(sms, "{time}", goutils.StringToTime(r.Date).Format("3:04 PM"), -1)
		lastPriority = int(r.Priority)
		parts = append(parts, sms)
	}

	message := strings.Join(parts, "{line}--{line}")

	msg := controller.GetSMSTemplate(ctx, "SMS_GAMES")

	if len(msg) == 0 {

		msg = gamesTemplate
	}

	msg = strings.Replace(msg, "{games}", message, -1)
	msg = strings.Replace(msg, "{line}", "\n", -1)

	if profileID > -1 {

		err = library.SetRedisKeyWithExpiry(controller.RedisConn, key, fmt.Sprintf("%d", lastPriority), 60*10)
		if err != nil {

			logrus.WithContext(ctx).
				WithFields(logrus.Fields{
					constants.DESCRIPTION: "error saving redis key",
					constants.DATA:        key,
				}).
				Error(err.Error())
		}

	}

	return msg
}

func (controller *Controller) GetJackpotSMSGames(ctx context.Context, CategoryID int64) string {
	dt := jackpot.JackpotGamesRequest{
		CategoryID: CategoryID,
	}

	games, err := controller.JackpotServiceClient.JpGames(ctx, &dt)
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "Failed to get games from Jackpot service",
				constants.DATA:        dt.String(),
			}).
			Error(err.Error())

		return "Welcome to Maybets where winners are made"

	}

	if len(games.Games) == 0 {

		return "Sorry we do not have active jackpot at the moment. Please keep checking for a surprise "

	}

	return games.Games
}

func (controller *Controller) GetBalance(ctx context.Context, profileID int64) string {

	dt := wallet.BalanceRequest{ProfileID: profileID}

	walletRes, err := controller.WalletServiceClient.GetBalance(ctx, &dt)
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "Failed to get balance from wallet service",
				constants.DATA:        dt.String(),
			}).
			Error(err.Error())

		return "Welcome to Maybets where winners are made"

	}

	templateName := "BALANCE_QUERY"
	sms := controller.GetSMSTemplate(ctx, templateName)

	replacements := map[string]string{
		"balance": fmt.Sprintf("%d", int64(walletRes.CurrentBalance)),
	}

	sms = BuildSMS(sms, replacements)

	return sms

}

func (controller *Controller) DoWithdraw(ctx context.Context, profileID int64, amount float64) string {

	dt := wallet.WithdrawRequest{ProfileID: profileID, Amount: float32(amount)}

	walletRes, err := controller.WalletServiceClient.Withdraw(ctx, &dt)
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "Failed to perfom withdraw from wallet service",
				constants.DATA:        dt.String(),
			}).
			Error(err.Error())

		return "we cannot perfom your request at the moment, try again later"

	}

	return walletRes.Description

}

func (controller *Controller) DoStk(ctx context.Context, profileID int64, amount float64) string {

	dt := wallet.STKRequest{ProfileID: profileID, Amount: float32(amount), Account: "sms"}

	walletRes, err := controller.WalletServiceClient.STK(ctx, &dt)
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "Failed to perfom STK from wallet service",
				constants.DATA:        dt.String(),
			}).
			Error(err.Error())

		return "we cannot perfom your request at the moment, try again later"

	}

	return walletRes.Description

}

func BuildSMS(message string, placeholder map[string]string) string {

	if placeholder != nil {

		for key, value := range placeholder {

			message = strings.Replace(message, fmt.Sprintf("[%s]", key), value, -1)
			message = strings.Replace(message, fmt.Sprintf("{%s}", key), value, -1)

		}
	}

	message = strings.Replace(message, "{line}", "\n", -1)
	message = strings.Replace(message, "[line]", "\n", -1)

	return message
}

func (controller *Controller) GetSMSTemplate(ctx context.Context, templateName string) string {

	// retrieve SMS template
	dbUtils := goutils.Db{DBSlave: controller.DBSlave, Context: ctx}
	dbUtils.SetQuery("SELECT message FROM sms_template WHERE name = ? LIMIT 1 ")
	dbUtils.SetParams(templateName)

	var content sql.NullString
	err := dbUtils.FetchOneSlave().Scan(&content)

	if err == sql.ErrNoRows {

		return ""
	}

	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error retrieving sms template",
				constants.DATA:        templateName,
			}).
			Error(err.Error())
	}

	return content.String
}

func (controller *Controller) betStatus(ctx context.Context, profileID int64, shareCode string) string {

	dt := betting.BetStatusRequest{
		ProfileID: profileID,
		ShareCode: shareCode,
	}

	betStatusRes, err := controller.BettingServiceClient.BetStatus(ctx, &dt)

	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "Failed to get bet status from betting service",
				constants.DATA:        dt.String(),
			}).
			Info(err.Error())

		return "we cannot perform your request at the moment, try again later"
	}

	var betSlips []string
	x := 0

	for _, v := range betStatusRes.SelectedSlips {

		x++
		sp := fmt.Sprintf("%d. %s - %s", x, v.MatchName, v.Description)
		betSlips = append(betSlips, sp)
	}

	selections := strings.Join(betSlips, "{line}")

	return BuildSMS(fmt.Sprintf("Bet Status betID#%s{line}--{line}%s", shareCode, selections), nil)

}

/*
func (controller *Controller) betCancel(ctx context.Context, profileID, MSISDN int64, shareCode, IpAddress string) string {

		dt := betting.CancelBetRequest{
			ProfileID: profileID,
			ShareCode: shareCode,
			MSISDN:    MSISDN,
			IpAddress: IpAddress,
		}

		if alias == "2" {

			return 1, "", "3", nil
		}

		if alias == "x" {

			return 1, "", "2", nil
		}

		if alias == "gg" {

			return 29, "", "74", nil
		}

		if alias == "ng" {

			return 29, "", "76", nil
		}

		var market_id sql.NullInt64
		var outcome_id, spec sql.NullString

		dbUtils := goutils.Db{DBSlave: controller.DBSlave}

		alias = strings.ReplaceAll(alias, ".", "")

		if strings.HasPrefix(alias, "ov") ||
			strings.HasPrefix(alias, "o") {

			specifier = strings.Replace(alias, "ov", "", -1)
			if len(specifier) == 0 {

				log.Printf("invalid alias %s ", alias)
				return 0, "", "", err

			}

			specifier = fmt.Sprintf("total=%s.%s", specifier[0:len(specifier)-1], specifier[len(specifier)-1:])
			marketID = 18
			outcomeID = "12"
			return marketID, outcomeID, specifier, nil

		} else if strings.HasPrefix(alias, "un") ||
			strings.HasPrefix(alias, "u") {

			specifier = strings.Replace(alias, "un", "", -1)
			if len(specifier) == 0 {

				log.Printf("invalid alias %s ", alias)
				return 0, "", "", err

			}
			specifier = fmt.Sprintf("total=%s.%s", specifier[0:len(specifier)-1], specifier[len(specifier)-1:])
			marketID = 18
			outcomeID = "13"
			return marketID, outcomeID, specifier, nil

		} else if strings.HasPrefix(alias, "cs") {

			specifier = strings.Replace(alias, "cs", "", -1)
			if len(specifier) == 0 {

				log.Printf("invalid alias %s ", alias)
				return 0, "", "", err

			}

			specifier = fmt.Sprintf("%s:%s", specifier[0:len(specifier)-1], specifier[len(specifier)-1:])
			marketID = 45
			outcomeName := strings.TrimSpace(specifier)

			query := fmt.Sprintf("SELECT outcome_id FROM outcomes WHERE market_id = 45 AND outcome_name = '%s' LIMIT 1 ", outcomeName)
			dbUtils.SetQuery(query)
			dbUtils.SetParams()
			var outcomeID sql.NullString

			er := dbUtils.FetchOneSlave().Scan(&outcomeID)
			if er == nil {

				return marketID, outcome_id.String, "", nil
			}

			return marketID, outcome_id.String, "", er
		}

		dbUtils.SetQuery(`SELECT sm.market_id,sm.outcome_id,sm.specifier
				FROM sms_market sm  WHERE LOWER(sm.alias) = ?`)

		dbUtils.SetParams(alias)

		err = dbUtils.FetchOneSlave().Scan(&market_id, &outcome_id, &spec)

		if err != nil {

			logrus.WithContext(ctx).
				WithFields(logrus.Fields{
					constants.DESCRIPTION: "Failed to cancel bet",
					constants.DATA:        dt.String(),
				}).
				Error(err.Error())

			return "we cannot perform your request at the moment, try again later"
		}

		return betCancelRes.Description
	}
*/

func (controller *Controller) ussdbetStatus(ctx context.Context, profileID int64, action string, session string, bettype string) (string, string) {

	key := fmt.Sprintf("BS_Games:%s:%d:%s", session, profileID, bettype)
	headerkey := fmt.Sprintf("BS_Games:H:%s:%d:%s", session, profileID, bettype)
	pageKey := fmt.Sprintf("BS_PAGE:%s:%d:%s", session, profileID, bettype)

	// Check if key exists in Redis

	cachedData, err := library.GetRedisKey(controller.RedisConn, key)

	if err != nil || cachedData == "" {
		betSlips := []string{}
		var status string
		var header string
		if bettype == "normal" {
			dt := betting.BetStatusRequest{
				ProfileID: profileID,
				ShareCode: action,
			}
			// Fetch from betting service if not in Redis
			betStatusRes, err := controller.BettingServiceClient.BetStatus(ctx, &dt)
			if err != nil {
				logrus.WithContext(ctx).
					WithFields(logrus.Fields{
						constants.DESCRIPTION: "Failed to get bet status from betting service",
						constants.DATA:        dt.String(),
					}).
					Info(err.Error())

				return "sorry we could not check your bet status at the moment, your have entered invalid bet ID", USSDEND
			}

			// Store fetched data in Redis with expiry (5 minutes)

			for i, v := range betStatusRes.SelectedSlips {
				betSlips = append(betSlips, fmt.Sprintf("%d. %s - %s", i+1, v.MatchName, v.Description))
			}
			// Determine bet status
			status = library.DetermineBetStatus(betStatusRes.Status)
			header = fmt.Sprintf("BetID: %s Status: %s", action, status)
		} else {
			jp := jackpot.JPBetStatusRequest{
				ProfileID: profileID,
				ShareCode: action,
			}
			betStatusRes, err := controller.JackpotServiceClient.JPBetStatus(ctx, &jp)
			if err != nil {
				logrus.WithContext(ctx).
					WithFields(logrus.Fields{
						constants.DESCRIPTION: "Failed to get bet status from jackpot service",
						constants.DATA:        jp.String(),
					}).
					Info(err.Error())

				return "sorry we could not check your bet status at the moment, your have entered invalid bet ID", USSDEND
			}
			for i, v := range betStatusRes.JackpotSelectedSlips {
				betSlips = append(betSlips, fmt.Sprintf("%d. %s - %s", i+1, v.MatchName, v.Description))
			}
			// Determine bet status
			status = library.DetermineBetStatus(betStatusRes.Status)
			jackpotname := "Weekly Jackpot"
			if betStatusRes.JackpotID == 5 {
				jackpotname = "Daily Jackpot"
			}

			header = fmt.Sprintf("%s BetID: %s Status: %s", jackpotname, action, status)
		}

		// Store bet slips in Redis
		library.SetRedisKeyWithExpiry(controller.RedisConn, key, strings.Join(betSlips, "|"), 300)
		library.SetRedisKeyWithExpiry(controller.RedisConn, headerkey, header, 300)
		library.SetRedisKeyWithExpiry(controller.RedisConn, pageKey, "0", 300)

		return fmt.Sprintf("%s\nPress 1 to view the betslip.\n1.View betslip\n00. Main Menu", header), USSDCON
	}
	pageStr, err := library.GetRedisKey(controller.RedisConn, pageKey)
	if err != nil {

		return "we cannot perform your request at the moment, try again later", USSDEND
	}
	page, _ := strconv.Atoi(pageStr)

	// Get stored bet slips
	betSlips := strings.Split(cachedData, "|")
	totalPages := (len(betSlips) + NOBETSTATUSGAMES - 1) / NOBETSTATUSGAMES // Two games per page

	switch action {
	case "0":
		if page > 0 {
			page--
		}
	case "88":
		if page < totalPages-1 {
			page++
		}
	}

	// Store updated page in Redis
	library.SetRedisKeyWithExpiry(controller.RedisConn, pageKey, strconv.Itoa(page), 300)

	// Display paginated games (games per page)
	start := page * NOBETSTATUSGAMES
	end := start + NOBETSTATUSGAMES
	if end > len(betSlips) {
		end = len(betSlips)
	}
	displayGames := strings.Join(betSlips[start:end], "\n")
	headerStr, _ := library.GetRedisKey(controller.RedisConn, headerkey)

	return fmt.Sprintf("%s\n%s\n0. Back\n88. Next\n00. Main Menu", headerStr, displayGames), USSDCON
}

func (controller *Controller) smsBet(ctx context.Context, profileID int64, message, ipAddress string) string {
	log.Printf("Input Message: %s", message)
	log.Printf("Input profile: %d", profileID)

	if ipAddress == "" {
		return "Error: IP address is required"
	}

	parts := strings.Split(message, "#")
	if len(parts) < 3 {
		log.Printf("Invalid Bet Format")
		return "Error: Invalid bet format. Please use GameID#pick#amount"
	}

	betType := 3
	source := 3

	amount, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil || amount <= 0 {
		log.Printf("Invalid Bet Stake %d", amount)

		return "Error: Invalid bet amount. Please provide a positive number."
	}

	bet, err := controller.parseBetSlips(ctx, parts[:len(parts)-1])
	if err != nil {
		log.Printf("Invalid Bet parsing %s", err.Error())

		return fmt.Sprintf("Error: %s", err.Error())
	}

	betResponse, err := controller.placeBet(ctx, profileID, float32(amount), source, constants.SMSCHANNELID, ipAddress, betType, bet)
	if err != nil {
		errMsg := err.Error()

		// Find and extract the "desc" part of the error message
		descIndex := strings.Index(errMsg, "desc =")
		if descIndex != -1 {
			// Extract the description part and trim any extra spaces
			description := strings.TrimSpace(errMsg[descIndex+len("desc ="):])
			log.Printf("Bet Placement Error: %s", description)
			return description
		} else {
			log.Printf("Bet Placement Error2: %s", errMsg)
			return errMsg
		}

	}

	if betResponse.Status == 201 || betResponse.Status == 200 {
		log.Printf("Success!! %d for %s", betResponse.Status, betResponse.ShareCode)

		return fmt.Sprintf("Bet #%s placed successfully. Please wait for a confirmation message.", betResponse.ShareCode)
	}

	return fmt.Sprintf("Bet placement failed: %s", betResponse.Description)
}

func (controller *Controller) parseBetSlips(ctx context.Context, parts []string) ([]*betting.Slips, error) {
	var bet []*betting.Slips

	for i := 0; i < len(parts)-1; i += 2 {
		gameID, err := strconv.Atoi(strings.TrimSpace(parts[i]))
		if err != nil {
			return nil, fmt.Errorf("invalid game ID: %s", parts[i])
		}

		pick := strings.TrimSpace(parts[i+1])
		marketID, specifier, outcomeID, err := controller.GetAlias(pick)
		if err != nil {
			return nil, fmt.Errorf("invalid market: %s", pick)
		}

		slip, err := controller.getOddsByGameID(ctx, int64(gameID), int32(marketID), specifier, outcomeID)
		if err != nil {
			return nil, fmt.Errorf("error fetching odds: %s", err)
		}

		bet = append(bet, slip)
	}

	return bet, nil
}

func (controller *Controller) getOddsByGameID(ctx context.Context, gameID int64, marketID int32, specifier string, outcomeID string) (*betting.Slips, error) {
	dt := fx.OddsByGameIdRequest{
		GameID:     gameID,
		MarketID:   marketID,
		Specifier:  specifier,
		OutcomeID:  outcomeID,
		ProducerID: 3,
	}

	fxRes, err := controller.FixtureServiceClient.GetOddsByGameID(ctx, &dt)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			constants.DESCRIPTION: "error getting odd id from fixture service",
			constants.DATA:        dt.String(),
		}).Error(err.Error())
		return nil, fmt.Errorf("error fetching odds for game ID %d", gameID)
	}

	return &betting.Slips{
		MarketID:   fxRes.MarketId,
		Specifier:  fxRes.Specifier,
		OutcomeID:  fxRes.OutcomeId,
		ProducerID: fxRes.ProducerId,
		MatchID:    fxRes.MatchId,
	}, nil
}

func (controller *Controller) placeBet(ctx context.Context, profileID int64, stake float32, source int, ChannelID int, ipAddress string, betType int, slips []*betting.Slips) (*betting.PlaceBetResponseMessage, error) {
	betRequest := betting.PlaceBetRequest{
		ProfileID:   profileID,
		Stake:       stake,
		Source:      strconv.Itoa(source),
		IpAddress:   ipAddress,
		BetType:     int64(betType),
		Slips:       slips,
		ChannelID:   int64(ChannelID),
		BookingCode: "",
	}

	betResponse, err := controller.BettingServiceClient.PlaceBet(ctx, &betRequest)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			constants.DESCRIPTION: "error placing bet with betting service",
			constants.DATA:        betRequest.String(),
		}).Warn(err.Error())
		return nil, fmt.Errorf("failed to place bet: %s", err)
	}

	return betResponse, nil
}

func (controller *Controller) jpAuto(ctx context.Context, profileID int64, message, ipAddress string, jackpotCategoryID int64) string {

	// jpauto

	jpRes, err := controller.JackpotServiceClient.AutoPick(ctx, &jackpot.AutoPickRequest{
		ProfileID: profileID,
		Stake:     float32(0),
		IpAddress: ipAddress,
		JackpotID: jackpotCategoryID,
	})

	if err != nil {

		log.Printf("error placing auto pick from jackpot service %s ", err.Error())

		return "We cannot process your request at the moment"
	}

	return jpRes.Description

}

func (controller *Controller) jpBet(ctx context.Context, profileID int64, message, ipAddress string, jackpotCategoryID int64) string {

	// jpauto#stake

	parts := strings.Split(message, "#")
	message = parts[1]
	log.Printf("Here are the messsage %s", message)

	jpRes, err := controller.JackpotServiceClient.AliasPick(ctx, &jackpot.AliasPickRequest{
		ProfileID: profileID,
		Stake:     0,
		IpAddress: ipAddress,
		JackpotID: jackpotCategoryID,
		Alias:     message,
	})

	if err != nil {

		log.Printf("error placing user pick from jackpot service %s ", err.Error())
		return "We cannot process your request at the moment"
	}

	return jpRes.Description

}

func (controller *Controller) getIntouchToken() (string, error) {

	redisKeyName := "SMS:ACCESS:KEY:INTOUCH"

	token, err := library.GetRedisKey(controller.RedisConn, redisKeyName)
	if err == nil && len(token) > 0 {

		return token, nil
	}

	route := os.Getenv("intouchvas_token_url")
	userName := os.Getenv("intouchvas_username")
	password := os.Getenv("intouchvas_password")

	auth := fmt.Sprintf("%s:%s", userName, password)
	basicAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	headers := map[string]string{
		"Authorization": "Basic " + basicAuth,
	}

	status, response := goutils.HTTPGet(route, headers, nil)

	if status > 299 || status < 200 {

		return "", errors.New(fmt.Sprintf("Got invalid status code %d : response body %s", status, response))
	}

	var responseMap map[string]interface{}
	err = json.Unmarshal([]byte(response), &responseMap)
	if err != nil {

		return "", errors.New(fmt.Sprintf("Got invalid status code %d : response body %s", status, response))
	}

	lifetime, _ := goutils.GetInt64(responseMap, "lifetime", 10800)
	smsToken, _ := goutils.GetString(responseMap, "token", "")

	expires := float64(lifetime) * 0.75

	err = library.SetRedisKeyWithExpiry(controller.RedisConn, redisKeyName, smsToken, int(expires))
	if err != nil {

		log.Printf("Got error saving redis key %s", err.Error())
	}

	return smsToken, nil

}

func (controller *Controller) GetAlias(alias string) (marketID int64, specifier, outcomeID string, err error) {
	log.Printf("get alias for %s", alias)

	// Convert alias to lowercase for standardization
	alias = strings.ToLower(strings.TrimSpace(alias))

	// Special case handling for over/under bets
	if strings.HasPrefix(alias, "ov") ||
		strings.HasPrefix(alias, "un") {
		return handleOverUnderBet(alias)
	}

	// Special case handling for correct score bets
	if strings.HasPrefix(alias, "cs") {
		return handleCorrectScoreBet(controller, alias)
	}

	// General case: look up in the sms_market table
	dbUtils := goutils.Db{DB: controller.DB}
	query := `SELECT market_id, outcome_id, specifier 
              FROM sms_market 
              WHERE LOWER(alias) = ?`
	dbUtils.SetQuery(query)
	dbUtils.SetParams(alias)

	var marketId sql.NullInt64
	var outcomeId, spec sql.NullString

	err = dbUtils.FetchOne().Scan(&marketId, &outcomeId, &spec)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", "", fmt.Errorf("unknown alias: %s", alias)
		}
		log.Printf("error scanning rows: %s", err.Error())
		return 0, "", "", err
	}

	return marketId.Int64, spec.String, outcomeId.String, nil
}

func handleOverUnderBet(alias string) (marketID int64, specifier, outcomeID string, err error) {
	isOver := strings.HasPrefix(alias, "ov") || strings.HasPrefix(alias, "o")
	alias = strings.TrimPrefix(strings.TrimPrefix(alias, "ov"), "o")
	alias = strings.TrimPrefix(strings.TrimPrefix(alias, "un"), "u")

	if len(alias) == 0 {
		return 0, "", "", fmt.Errorf("invalid over/under alias: %s", alias)
	}

	specifier = fmt.Sprintf("total=%s.%s", alias[:len(alias)-1], alias[len(alias)-1:])
	marketID = 18
	if isOver {
		outcomeID = "12"
	} else {
		outcomeID = "13"
	}

	return marketID, specifier, outcomeID, nil
}

func handleCorrectScoreBet(controller *Controller, alias string) (marketID int64, specifier, outcomeID string, err error) {
	// Remove the "cs" prefix
	score := strings.TrimPrefix(strings.ToLower(alias), "cs")

	// Check if it's the "other" option
	if score == "other" {
		specifier = "Other"
	} else {
		// Split the score using the period
		parts := strings.Split(score, ".")
		if len(parts) != 2 {
			return 0, "", "", fmt.Errorf("invalid correct score format: %s. Use format cs<home>.<away> (e.g., cs1.0 or cs4.4) or csother", alias)
		}

		// Parse home and away scores
		homeScore, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, "", "", fmt.Errorf("invalid home team score in %s: %v", alias, err)
		}

		awayScore, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, "", "", fmt.Errorf("invalid away team score in %s: %v", alias, err)
		}

		// Format the specifier
		specifier = fmt.Sprintf("%d:%d", homeScore, awayScore)
	}

	// Set market ID for Correct Score
	marketID = 45

	// Query the database for the outcome ID
	dbUtils := goutils.Db{DB: controller.DB}
	query := `SELECT outcome_id 
              FROM outcomes 
              WHERE market_id = 45 AND outcome_name = ? 
              LIMIT 1`
	dbUtils.SetQuery(query)
	dbUtils.SetParams(specifier)

	var outcomeId sql.NullString
	err = dbUtils.FetchOne().Scan(&outcomeId)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", "", fmt.Errorf("unsupported correct score: %s. Please choose a valid score or 'other'", specifier)
		}
		return 0, "", "", fmt.Errorf("database error while fetching outcome ID: %v", err)
	}

	if !outcomeId.Valid {
		return 0, "", "", fmt.Errorf("no valid outcome ID found for score: %s", specifier)
	}

	return marketID, specifier, outcomeId.String, nil
}
