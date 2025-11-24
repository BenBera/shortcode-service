package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/BenBera/shortcode-service/app/constants"
	"github.com/BenBera/shortcode-service/app/library"
	"github.com/BenBera/shortcode-service/app/models"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	goutils "github.com/mudphilo/go-utils"
	"github.com/sirupsen/logrus"
)

func (controller *Controller) SDPIncomingSMS(c echo.Context) error {

	ctx, span := controller.Tracer.Start(c.Request().Context(), "Inbox")
	defer span.End()

	data := goutils.GetJSONRawBody(c)

	payload := new(models.InboxPayloadInteractive)

	jsB, _ := json.Marshal(data)

	err := json.Unmarshal(jsB, payload)
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

	requestId := payload.RequestID
	RequestId := payload.RequestID
	operation := payload.Operation
	Operation := payload.Operation
	RequestTimestamp := payload.RequestTimeStamp
	requestTimeStamp := payload.RequestTimeStamp
	LinkId := ""
	OfferCode := ""
	RefernceId := ""
	Message := ""

	Msisdn := ""
	Command := ""

	//correlatorId := ""
	//deliveryStatus := ""
	//campaignId := ""
	if operation == "INTERACTIVE" {

		OfferCode = strconv.Itoa(78655453748)
		RefernceId = strconv.Itoa(rand.Intn(100000))
		LinkId = strconv.Itoa(rand.Intn(100000))

		for _, v := range payload.RequestParam.Data {

			switch v.Name {

			case "Msisdn":
				Msisdn = v.Value
				break
			case "Command":
				Command = v.Value
				break
			}
		}

		for _, w := range payload.RequestParam.AdditionalData {

			switch w.Name {

			case "SMS":
				Message = w.Value
				break
			case "DA":
				log.Printf("ShortCode = %s", w.Value)
				break
			}
		}

		db := goutils.Db{DB: controller.DB}

		inserts := map[string]interface{}{
			"offer_code":        OfferCode,
			"message":           Message,
			"request_id":        RequestId,
			"operation":         operation,
			"request_timestamp": RequestTimestamp,
			"reference_id":      RefernceId,
			"msisdn":            Msisdn,
			"link_id":           LinkId,
			"command":           Command,
			"created":           goutils.MysqlNow(),
		}

		inboxID, err := db.Upsert("safaricom_inbox", inserts, nil)
		if err != nil {

			logrus.WithFields(logrus.Fields{
				constants.DESCRIPTION: "got error preparing interactive database statement ",
			}).Error(err.Error())
			return err

		}

		u := models.Inbox{
			Msisdn:  Msisdn,
			Message: Message,
			InboxID: inboxID,
		}
		ipAddress := c.RealIP()

		statusCode, response := controller.ProcessInbox(ctx, &u, true, ipAddress, LinkId)

		return c.JSON(statusCode, response)

	} else {

		for _, v := range payload.RequestParam.Data {

			switch v.Name {

			case "LinkId":
				LinkId = v.Value
				break

			case "OfferCode":
				OfferCode = v.Value
				break

			case "RefernceId":
				RefernceId = v.Value
				break

			case "USER_DATA":

				Message = v.Value

			case "SMS":
				if v.Value != "" {
					Message = v.Value
				}

			case "Msisdn":
				Msisdn = v.Value
				break

			case "Type":
				Command = v.Value
				break

			}
		}

		_Msisdn, _ := strconv.ParseInt(Msisdn, 10, 64)

		db := goutils.Db{DB: controller.DB}
		inserts := map[string]interface{}{
			"offer_code":        OfferCode,
			"message":           Message,
			"request_id":        requestId,
			"operation":         Operation,
			"request_timestamp": requestTimeStamp,
			"reference_id":      RefernceId,
			"msisdn":            _Msisdn,
			"link_id":           LinkId,
			"command":           Command,
			"created":           goutils.MysqlNow(),
		}

		inboxID, err := db.Upsert("safaricom_inbox", inserts, nil)
		if err != nil {

			logrus.WithFields(logrus.Fields{
				constants.DESCRIPTION: "got error preparing database statement ",
			}).Error(err.Error())
			return err

		}

		u := models.Inbox{
			Msisdn:  Msisdn,
			Message: Message,
			InboxID: inboxID,
		}
		ipAddress := c.RealIP()

		statusCode, response := controller.ProcessInbox(ctx, &u, false, ipAddress, LinkId)

		return c.JSON(statusCode, response)

	}

}

func (controller *Controller) SDPAutoresponse(LinkID, message, msisdn string) error {
	endpoint := os.Getenv("SHORTCODE_URL")
	payload := models.ShortCodeAutoResponse{
		Mobile:     msisdn,
		SenderName: os.Getenv("SHORTCODE_NAME"),
		ServiceID:  1,
		LinkID:     fmt.Sprintf("%d", LinkID), //assumption link id is inbox is
		Message:    message,
	}
	jsP, _ := json.MarshalIndent(payload, " ", "\t")
	headers := map[string]string{
		"h_api_key": os.Getenv("H_API_KEY"),
	}
	st, body := goutils.HTTPPost(endpoint, headers, payload)

	var resp []models.SMSResponse
	err := json.Unmarshal([]byte(body), &resp)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			constants.DESCRIPTION: "got error unmarshalling shortcode response",
		}).Error(err.Error())
	}
	for _, v := range resp {
		if v.StatusCode != "1000" {
			x, _ := json.MarshalIndent(v, "", "\t")
			logrus.WithFields(logrus.Fields{
				constants.DESCRIPTION: "got error submitting shortcode auto  response",
				"response_body":       string(x),
			}).Debug()
			return fmt.Errorf("error submitting shortcode auto response")

		}
	}
	log.Printf("%s | %s | response %d | %v ", endpoint, string(jsP), st, resp)

	return nil

}

func ToTimestamp(t time.Time) string {

	//yyyyMMddHHmmss 2019 09 25 14 22 25

	tt := t.Format("20060102150405")

	return tt
}

func SDPRefreshAccessToken(redisConn *redis.Client) {

	// check if we have a token first
	redisKey := "SAFARICOM:SDP:TOKEN"
	log.Printf("here is the sdp redis key %s", redisKey)

	data, err := library.GetRedisKey(redisConn, redisKey)

	if err != nil || len(data) == 0 {

		logrus.WithFields(logrus.Fields{
			constants.DESCRIPTION: fmt.Sprintf("got error retrieving Redis data for key %s", redisKey),
		}).Info()

		SDPgetAccessToken(redisConn)
	}

	go func() {

		ticker := time.NewTicker(20 * time.Minute)

		for t := range ticker.C {

			logrus.WithFields(logrus.Fields{
				constants.DESCRIPTION: fmt.Sprintf(" getAccessToken cronjob running here at %s ", t.Format("2006-01-02 15:04:05")),
			}).Info()
			SDPgetAccessToken(redisConn)
		}

		select {}

	}()

}

func SDPgetAccessToken(redisConn *redis.Client) string {

	// get sdp credentials
	api_username := os.Getenv("SDP_USERNAME")
	api_password := os.Getenv("SDP_PASSWORD")

	// check if we have a token first

	token_api_url := os.Getenv("SDP_TOKEN_URL")
	if len(token_api_url) == 0 {

		token_api_url = "https://dsvc.safaricom.com:9480/api/auth/login"

	}

	payload := map[string]string{
		"username": api_username,
		"password": api_password,
	}

	headers := map[string]string{
		"X-Requested-With": "XMLHttpRequest",
		"Content-Type":     "application/json",
	}

	status, response := goutils.HTTPPost(token_api_url, headers, payload)

	if status >= 200 && status <= 202 {

		var resp map[string]interface{}

		if err := json.Unmarshal([]byte(response), &resp); err != nil {

			logrus.WithFields(logrus.Fields{
				constants.DESCRIPTION: "Got error retrieving SDP Token ",
			}).Error(err.Error())

		}

		token, err := goutils.GetString(resp, "token", "")
		if err != nil {

			logrus.WithFields(logrus.Fields{
				constants.DESCRIPTION: "Got error retrieving SDP Token ",
			}).Error(err.Error())

		}

		refreshToken, err := goutils.GetString(resp, "refreshToken", "")
		if err != nil {

			logrus.WithFields(logrus.Fields{
				constants.DESCRIPTION: "Got error retrieving SDP Token ",
			}).Error(err.Error())

		}

		redisKeyRefreshToken := "SAFARICOM:SDP:REFRESHTOKEN"

		err = library.SetRedisKey(redisConn, redisKeyRefreshToken, refreshToken)
		if err != nil {

			logrus.WithFields(logrus.Fields{
				constants.DESCRIPTION: "Got error retrieving SDP Token ",
			}).Error(err.Error())

		}

		redisKey := "SAFARICOM:SDP:TOKEN"
		err = library.SetRedisKeyWithExpiry(redisConn, redisKey, token, 60*29)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				constants.DESCRIPTION: "Got error retrieving SDP Token ",
			}).Error(err.Error())

		}
		log.Printf("Here is the access token %s", token)

		return token

	}

	return ""

}

func SDPgetNewToken(redisConn *redis.Client) error {

	redisKeyRefreshToken := "SAFARICOM:SDP:REFRESHTOKEN"

	data, err := library.GetRedisKey(redisConn, redisKeyRefreshToken)

	if err != nil || len(data) == 0 {

		logrus.WithFields(logrus.Fields{
			constants.DESCRIPTION: "Got error retrieving Redis data ",
		}).Error(err.Error())
		return err
	}

	refresh_token_api_url := os.Getenv("SDP_REFRESH_TOKEN_URL")
	if len(refresh_token_api_url) == 0 {

		refresh_token_api_url = "https://dsvc.safaricom.com:9480/api/auth/RefreshToken"

	}

	headers := map[string]string{
		"X-Requested-With": "XMLHttpRequest",
		"Content-Type":     "application/json",
		"X-Authorization":  fmt.Sprintf("Bearer %s", data),
	}

	status, response := goutils.HTTPPost(refresh_token_api_url, headers, nil)

	if status >= 200 && status <= 202 {

		var resp map[string]interface{}

		if err := json.Unmarshal([]byte(response), &resp); err != nil {

			logrus.WithFields(logrus.Fields{
				constants.DESCRIPTION: "Got error retrieving SDP Token ",
			}).Error(err.Error())

			return err

		}

		token, err := goutils.GetString(resp, "token", "")
		if err != nil {

			logrus.WithFields(logrus.Fields{
				constants.DESCRIPTION: "Got error retrieving SDP Token ",
			}).Error(err.Error())
			return err

		}

		redisKey := "SAFARICOM:SDP:TOKEN"

		err = library.SetRedisKeyWithExpiry(redisConn, redisKey, token, 60*29)
		if err != nil {

			logrus.WithFields(logrus.Fields{
				constants.DESCRIPTION: "Got error retrieving SDP Token ",
			}).Error(err.Error())
			return err

		}

		return nil
	}

	return nil

}

func SDPGetToken(redisConn *redis.Client) string {

	// get token from redis
	redisKey := "SAFARICOM:SDP:TOKEN"
	data, err := library.GetRedisKey(redisConn, redisKey)

	if err != nil || len(data) == 0 {

		return SDPgetAccessToken(redisConn)
	}

	return data
}

// process dlr
func (controller *Controller) Inboxdlr(c echo.Context, db *sql.DB) error {

	ctx, span := controller.Tracer.Start(c.Request().Context(), "Inbox")
	defer span.End()

	payload := goutils.GetJSONRawBody(c)

	u := new(models.InboxDLR)

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
	if u.InboxID == 0 {
		return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "No inbox id",
		})
	}
	// store to db
	dlrTable := "inbox_dlr"

	inserts := map[string]interface{}{
		"inbox_id":     u.InboxID,
		"status":       u.Status,
		"description":  u.Description,
		"message_type": u.MessageType,
	}

	dbUtil := goutils.Db{DB: db}

	_, err = dbUtil.Upsert(dlrTable, inserts, []string{"inbox_id", "status", "description", "message_type"})
	if err != nil {
		return RespondRaw(c, http.StatusBadRequest, models.ErrorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
	}
	return RespondRaw(c, http.StatusOK, models.SuccessResponse{
		Message: "Inbox DLR successfully created/ Updated.",
		Status:  http.StatusCreated,
	})
}
