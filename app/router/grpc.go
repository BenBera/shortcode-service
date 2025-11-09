package router

import (
	"context"
	"fmt"
	"github.com/BenBera/shortcode-service/app/database"
	"github.com/BenBera/shortcode-service/app/grpc/shortcode"
	"github.com/BenBera/shortcode-service/app/models"
	"log"
	"strconv"
	"strings"
)

// IncomingSMS
func (a *App) IncomingSMS(ctx context.Context, in *shortcode.IncomingRequest) (*shortcode.IncomingResponse, error) {

	inbox := models.Inbox{
		Msisdn:  strconv.FormatInt(in.Msisdn, 10),
		Message: in.Message,
		InboxID: 0,
	}
	log.Printf("here is the inbox data for processing %v", inbox)

	// send to inbox processor
	a.Controller.ProcessInbox(ctx, &inbox, true, "0.0.0.0")

	res := shortcode.IncomingResponse{
		Status:      1,
		Description: "Success",
	}

	return &res, nil

}

func (a *App) GetSMSGames(ctx context.Context, in *shortcode.SmsRequest) (*shortcode.SmsResponse, error) {
	//log.Printf("here GetSMSGames count: %v", in.Count)

	// send to inbox processor
	sms := a.Controller.GetSMSGames(ctx, int64(in.ProfileId), int(in.Count))
	//log.Printf("here GetSMSGames text: %s", sms)
	res := shortcode.SmsResponse{
		Message: sms,
	}

	return &res, nil

}

func (a *App) Ping(ctx context.Context, in *shortcode.ShortCodePing) (*shortcode.ShortCodePong, error) {

	status, re := database.CheckConnectionStatus(ctx, a.DB)

	var st []string

	for k, v := range re {

		st = append(st, fmt.Sprintf("%s - %s ", k, v))

	}

	return &shortcode.ShortCodePong{
		Status: int64(status),
		Data:   strings.Join(st, ","),
	}, nil
}
