package crontask

import (
	"bitbucket.org/maybets/shortcode-service/app/grpc/identity"
	"context"
	"database/sql"
	"github.com/go-co-op/gocron/v2"
	goutils "github.com/mudphilo/go-utils"
	"log"
)

func (cron *Crontask) ScheduleCleanUp(db *sql.DB) {
	ctx, span := cron.Tracer.Start(context.Background(), "campaign")
	defer span.End()
	log.Printf("ScheduleCleanUp here ")

	// create a scheduler
	s, err := gocron.NewScheduler()
	if err != nil {

		log.Panicf("failed to setup cron job %s ", err.Error())
	}

	// add a job to the scheduler this runs daily at 08:00:00 hours
	j, err := s.NewJob(gocron.CronJob("0 0 8 * * * ", true), gocron.NewTask(cron.SMSCampaign, ctx, db))
	if err != nil {

		log.Panicf("failed to setup cron job %s ", err.Error())

	}

	// each job has a unique id
	log.Printf("Cron job created with task ID %s ", j.ID().String())

	// start the scheduler
	s.Start()

	select {}
}

func (cron *Crontask) SMSCampaign(ctx context.Context, db *sql.DB) {
	ctx, span := cron.Tracer.Start(ctx, "smsCampaign")
	defer span.End()

	dbUtils := goutils.Db{DB: db}
	dbUtils.SetQuery("select distinct inbox.msisdn from inbox where inbox.created > curdate() - interval 2 week ")
	dbUtils.SetParams()

	rows, err := dbUtils.Fetch()
	if err != nil {
		log.Printf("error retrieving last 2 weeks numbers: %s", err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		var msisdn sql.NullInt64
		err = rows.Scan(&msisdn)
		if err != nil {
			log.Printf("error scanning campaign msisdn: %s", err.Error())
			continue
		}

		identityResponse, err := cron.IdentityServiceClient.GetProfileByMsisdn(context.Background(), &identity.Msisdn{
			Msisdn: msisdn.Int64,
		})
		if err != nil {
			log.Printf("Error retrieving profileDetails: %s", err.Error())
			continue
		}

		smsGames := cron.Controller.GetSMSGames(ctx, identityResponse.Id, 10)

		cron.Controller.AutoResponse(ctx, 0, smsGames, false)

	}
}
