package library

import (
	"bytes"
	goutils "github.com/mudphilo/go-utils"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	goutilsmodels "github.com/mudphilo/go-utils/models"

	"github.com/labstack/echo/v4"
)
func GetUssdhopsAsTable(dbSlave *sql.DB,c echo.Context, search goutilsmodels.VueTable) (httpStatus int, resp interface{}, err error) {
	
	table := "session_hops"
	primaryKey := fmt.Sprintf("%s.id", table)

	// params

	var fields []string
	var joins []string
	var andWhere []string
	var params []interface{}

	fields = append(fields, "session_hops.id")
	fields = append(fields, "session_hops.user_input")
	fields = append(fields, "session_hops.session_id")
	fields = append(fields, "session_hops.msisdn")
	fields = append(fields, "session_hops.response")
	fields = append(fields, "session_hops.created")

	if search.ID > 0 {

		andWhere = append(andWhere, "session_hops.id = ? ")
		params = append(params, search.ID)

	}

	if len(search.StartDate) > 0 {

		andWhere = append(andWhere, "session_hops.created >= ? ")
		params = append(params, search.StartDate)
	}

	if len(search.EndDate) > 0 {

		andWhere = append(andWhere, "session_hops.created <= ? ")
		params = append(params, search.EndDate)
	}
	paginator := goutilsmodels.Paginator{
		VueTable:   search,
		TableName:  table,
		PrimaryKey: primaryKey,
		Fields:     fields,
		Joins:      joins,
		GroupBy:    []string{},
		OrWhere:    andWhere,
		Params:     params,
		Results:    ResultsFc,
	}

	if search.Download == 1 {
		rowData, headers := DownloadVueTableData(dbSlave, paginator)
		if rowData == nil || len(rowData) == 0 {
			return http.StatusNotFound,nil, fmt.Errorf("No data available for download")
		}
		return http.StatusAccepted,GenerateCSVResponse(c, rowData, headers, "profile_data"),nil
	}
	resp = goutils.PaginateDataSlave(dbSlave, paginator)

	return http.StatusOK, resp, nil

}
func ResultsFc(rows *sql.Rows) []interface{} {

	var data []interface{}

	for rows.Next() {
	
		var id, msisdn sql.NullInt64
		var userinput,sessionid,response sql.NullString
		var created sql.NullTime
		_ = rows.Scan(&id, &userinput, &sessionid, &msisdn, &response, &created)
		data = append(data, map[string]interface{}{
			"id":             id.Int64,
			"user_input":          userinput.String,
			"session_id":         sessionid.String,
			"msisdn": msisdn.Int64,
			"response":   response.String,
			"created":        goutils.ToMysqlDateTime(created.Time),
		})
	}

	return data

}


func GenerateCSVResponse(c echo.Context, data []interface{}, headers []string, baseFileName string) error {
	// Generate CSV content
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write headers
	if err := writer.Write(headers); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error writing headers: %v", err))
	}

	// Write data
	for _, row := range data {
		var csvRow []string
		r := reflect.ValueOf(row)
		for _, header := range headers {
			field := r.FieldByName(header)
			if field.IsValid() {
				csvRow = append(csvRow, fmt.Sprintf("%v", field.Interface()))
			} else {
				csvRow = append(csvRow, "")
			}
		}
		if err := writer.Write(csvRow); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error writing row: %v", err))
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error flushing writer: %v", err))
	}

	// Generate a unique filename
	fileName := fmt.Sprintf("%s_%s.csv", baseFileName, time.Now().Format("20060102_150405"))

	// Set headers for file download
	c.Response().Header().Set(echo.HeaderContentType, "text/csv")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", fileName))

	// Write CSV content to response
	_, err := c.Response().Write(buf.Bytes())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error writing CSV to response: %v", err))
	}

	return nil
}

func DownloadVueTableData(db *sql.DB, paginator goutilsmodels.Paginator) (rowData []interface{}, headrs []string) {

	search := paginator.VueTable
	joins := paginator.Joins
	fields := paginator.Fields
	orWhere := paginator.OrWhere
	groupBy := paginator.GroupBy
	params := paginator.Params
	tableName := paginator.TableName
	primaryKey := paginator.PrimaryKey
	isDebug, _ := strconv.ParseInt(os.Getenv("DEBUG"), 10, 64)

	joinQuery := strings.Join(joins[:], " ")
	field := strings.Join(fields[:], ",")

	var headers []string

	for _, h := range fields {

		parts := strings.Split(h, " ")
		headers = append(headers, parts[len(parts)-1])
	}

	whereQuery := func() string {

		if len(orWhere) > 0 {

			return strings.Join(orWhere[:], " AND ")
		}
		return "1"
	}

	group := func() string {

		if len(groupBy) > 0 {

			return fmt.Sprintf("GROUP BY %s", strings.Join(groupBy[:], " , "))

		}

		return ""
	}

	// build order by query

	orderBy := ""

	if len(search.Sort) > 0 {

		sortPrams := strings.Split(search.Sort, "|")

		column := sortPrams[0]
		direction := sortPrams[1]
		orderBy = fmt.Sprintf("ORDER BY %s %s ", column, direction)
	}

	hardLimit, _ := strconv.ParseInt(os.Getenv("HARD_SQL_FETCH_LIMIT"), 10, 64)
	if hardLimit == 0 {

		hardLimit = 200000
	}

	var countQuery string

	if hardLimit == -1 {

		countQuery = fmt.Sprintf("SELECT count(%s) as total FROM %s %s WHERE %s ", primaryKey, tableName, joinQuery, whereQuery())

	} else {

		countQuery = fmt.Sprintf("SELECT count(%s) as total FROM %s %s WHERE %s LIMIT %d", primaryKey, tableName, joinQuery, whereQuery(), hardLimit)

	}
	// count query

	total := 0
	dbUtil := goutils.Db{DBSlave: db}
	dbUtil.SetQuery(countQuery)
	dbUtil.SetParams(params...)
	if isDebug != 0 {

		log.Printf("Count Query | %s", countQuery)
		log.Printf("Params | %v", params...)

	}
	err := dbUtil.FetchOne().Scan(&total)
	if err != nil {

		log.Printf("got error retrieving total number of records %s ", err.Error())
		return nil, headers
	}

	var sqlQuery string

	if hardLimit == -1 {

		sqlQuery = fmt.Sprintf("SELECT %s FROM %s %s WHERE %s %s %s ", field, tableName, joinQuery, whereQuery(), group(), orderBy)

	} else {

		sqlQuery = fmt.Sprintf("SELECT %s FROM %s %s WHERE %s %s %s LIMIT %d", field, tableName, joinQuery, whereQuery(), group(), orderBy, hardLimit)

	}

	// pull records

	// retrieve user roles
	dbUtil.SetQuery(sqlQuery)
	if isDebug != 0 {

		log.Printf("Data Query | %s", sqlQuery)

	}
	rows, err := dbUtil.Fetch()
	if err != nil {

		log.Printf("error pulling vuetable data %s", err.Error())
		return nil, headers

	}

	defer rows.Close()

	rowData = paginator.Results(rows)
	return rowData, headers
}