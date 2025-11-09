package crontask

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/BenBera/shortcode-service/app/constants"
	"github.com/BenBera/shortcode-service/app/library"
	"github.com/BenBera/shortcode-service/app/models"
	goutils "github.com/mudphilo/go-utils"
	"github.com/sirupsen/logrus"
	"time"
)

func (cron *Crontask) SendDashboardReports(ctx context.Context) {

	ctx, span := cron.Tracer.Start(ctx, "SendDashboardReports")
	defer span.End()

	ticker := time.NewTicker(1 * time.Minute)

	for _ = range ticker.C {

		cron.sendInboxReports(ctx)

	}

	select {}
}

func (cron *Crontask) sendInboxReports(ctx context.Context) {

	ctx, span := cron.Tracer.Start(ctx, "sendInboxReports")
	defer span.End()

	redisKey := "inbox"
	lastTimeStamp := cron.getLastTimeStamp(ctx, redisKey)

	dbUtils := goutils.Db{DBSlave: cron.DBSlave, Context: ctx}

	if len(lastTimeStamp) == 0 {

		dbUtils.SetQuery("SELECT id, msisdn, inbox_id, message,response,created,updated " +
			"FROM inbox ")

	} else {

		dbUtils.SetQuery("SELECT id, msisdn, inbox_id, message,response,created,updated " +
			"FROM inbox WHERE updated > ? ")

		dbUtils.SetParams(lastTimeStamp)
	}

	rows, err := dbUtils.FetchSlaveWithContext()
	if err == sql.ErrNoRows {

		return
	}

	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error retrieving bet  ",
			}).
			Error(err.Error())

		return
	}

	defer rows.Close()

	table := "inbox"

	for rows.Next() {

		var id, msisdn, inbox_id sql.NullInt64
		var message, response sql.NullString
		var created, updated sql.NullTime

		err = rows.Scan(&id, &msisdn, &inbox_id, &message, &response, &created, &updated)
		if err != nil {

			logrus.WithContext(ctx).
				WithFields(logrus.Fields{
					constants.DESCRIPTION: "error scanning inbox  ",
				}).
				Warn(err.Error())

			continue
		}

		data := map[string]interface{}{
			"id":       id.Int64,
			"msisdn":   msisdn.Int64,
			"inbox_id": inbox_id.Int64,
			"message":  message.String,
			"response": response.String,
			"created":  goutils.ToMysql(created.Time),
		}

		payload := models.UpsertData{
			Table:  table,
			Fields: getFields(data, []string{"id"}),
			Data:   data,
		}

		_ = library.Publish(ctx, cron.RabbitMQConn, fmt.Sprintf("reports-service.%s.create", table), payload, 0)

		lastTimeStamp = goutils.ToMysql(updated.Time)

	}

	cron.setLastTimeStamp(ctx, redisKey, lastTimeStamp)

}

func getFields(data map[string]interface{}, excludes []string) []string {

	var fields []string

	for k := range data {

		if !goutils.Contains(excludes, k) {

			fields = append(fields, k)
		}
	}

	return fields
}

func (cron *Crontask) getLastTimeStamp(ctx context.Context, reportName string) string {

	ctx, span := cron.Tracer.Start(ctx, "getLastTimeStamp")
	defer span.End()

	dbUtils := goutils.Db{DBSlave: cron.DBSlave, Context: ctx}
	dbUtils.SetQuery("SELECT last_timestamp FROM reports_sync WHERE report_name = ? ")
	dbUtils.SetParams(reportName)

	var lastTimestamp sql.NullTime

	err := dbUtils.FetchOneSlaveWithContext().Scan(&lastTimestamp)
	if err == sql.ErrNoRows {

		return ""
	}

	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error retrieving last_timestamp  ",
				constants.DATA:        reportName,
			}).
			Error(err.Error())

		return goutils.ToMysql(time.Now())
	}

	return goutils.ToMysql(lastTimestamp.Time)

}

func (cron *Crontask) setLastTimeStamp(ctx context.Context, reportName, lastTimestamp string) error {

	ctx, span := cron.Tracer.Start(ctx, "setLastTimeStamp")
	defer span.End()

	dbUtils := goutils.Db{DB: cron.DB, Context: ctx}

	inserts := map[string]interface{}{
		"report_name":    reportName,
		"last_timestamp": lastTimestamp,
	}

	_, err := dbUtils.UpsertWithContext("reports_sync", inserts, []string{"last_timestamp"})
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error inserting reports_sync  ",
				constants.DATA:        inserts,
			}).
			Error(err.Error())

	}

	return nil

}
