package crontask

import (
	"context"
	"encoding/xml"
	"github.com/BenBera/shortcode-service/app/models"
	goutils "github.com/mudphilo/go-utils"
	"log"
	"os"
)

func (cron *Crontask) GetMarketsData(ctx context.Context) {

	ctx, span := cron.Tracer.Start(ctx, "GetMarketsData")
	defer span.End()

	log.Printf("GetMarketsData running")

	headers := map[string]string{
		"x-access-token": os.Getenv("betrader_token"),
		"accept":         "*/*",
	}

	endpoint := "https://api.betradar.com/v1/descriptions/en/markets.xml?include_mappings=false"

	status, payload := goutils.HTTPGet(endpoint, headers, nil)
	if status < 200 || status > 299 {

		log.Printf("invalid status %d | %s ", status, payload)
		return
	}

	sports := new(models.MarketDescriptions)

	err := xml.Unmarshal([]byte(payload), sports)
	if err != nil {

		log.Printf("error decoding xml to BetradarSports %s ", err.Error())
		return
	}

	dbUtils := goutils.Db{DB: cron.DB, Context: ctx}

	for _, s := range sports.Market {

		marketID := s.ID
		//marketName := s.Name

		inserts := map[string]interface{}{
			"market_id":   marketID,
			"market_name": s.Name,
		}

		_, err = dbUtils.UpsertWithContext("markets", inserts, []string{"market_name"})
		if err != nil {

			log.Printf("error inserting markets %s ", err.Error())
		}

		for _, o := range s.Outcomes.Outcome {

			insertsOutcomes := map[string]interface{}{
				"market_id":    marketID,
				"outcome_id":   o.ID,
				"outcome_name": o.Name,
			}

			_, err = dbUtils.UpsertWithContext("outcomes", insertsOutcomes, []string{"outcome_name"})
			if err != nil {

				log.Printf("error inserting outcomes %s ", err.Error())
			}

		}

		for _, o := range s.Specifiers.Specifier {

			insertsSpecifiers := map[string]interface{}{
				"market_id": marketID,
				"specifier": o.Name,
				"data_type": o.Type,
			}

			_, err = dbUtils.UpsertWithContext("specifiers", insertsSpecifiers, []string{"data_type"})
			if err != nil {

				log.Printf("error inserting specifiers %s ", err.Error())
			}

		}

	}

}
