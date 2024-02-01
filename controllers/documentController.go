package controllers

import (
	"core-service/config"
	"core-service/services"
	"core-service/utils"
	"crypto/tls"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/go-mail/mail"
	"github.com/labstack/echo/v4"
	"github.com/tealeg/xlsx"
)

// @Summary Generate a document
// @Description Generate an Excel document and send it via email
// @Tags Document
// @Produce json
// @Router /api/core/document [post]
func Gendocument(c echo.Context) error {
	currentTime := time.Now()
	startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day()-1, 7, 0, 0, 0, time.UTC)
	endTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 7, 0, 0, 0, time.UTC)
	formattedTime := currentTime.Format("02.01.2006 Time 15:04 PM")
	startTimeStr := startTime.Format("02.01.2006")
	endTimeStr := endTime.Format("02.01.2006")
	// formattedTime2 := currentTime.Format("02012006")
	f := excelize.NewFile()

	sheetName := "Sheet1"

	// สร้างสไตล์ที่รวมทั้งสองสไตล์เข้าด้วยกัน
	combinedStyle, _ := f.NewStyle(`{
		"alignment":{"wrap_text":true,"vertical":"center","horizontal":"center"},
		"font": {"bold": false, "italic": false},
		"border": [
			{"type": "top", "color": "000000", "style": 1},
			{"type": "bottom", "color": "000000", "style": 1},
			{"type": "left", "color": "000000", "style": 1},
			{"type": "right", "color": "000000", "style": 1}
		],
		"alignment": {"horizontal": "center"},
		"fill": {"type": "pattern", "color": ["C6C6D1"], "pattern": 1}
	}`)

	dataCellStyle, _ := f.NewStyle(`{"border":[{"type":"top","color":"000000","style":1},{"type":"bottom","color":"000000","style":1},{"type":"left","color":"000000","style":1},{"type":"right","color":"000000","style":1}],"alignment":{"horizontal":"left"}}`)

	stylecenter, _ := f.NewStyle(`{"alignment":{"horizontal":"center","vertical":"center"},
	    "font":{"bold":false,"italic":false}}`)
	styleleft, _ := f.NewStyle(`{"alignment":{"horizontal":"left"}, 
        "font":{"bold":false,"italic":false}}`)

	f.SetCellValue(sheetName, "A1", "บริษัท เงินดีดี จำกัด")
	f.SetColWidth(sheetName, "A", "R", 20)
	f.MergeCell(sheetName, "A1", "R1")

	f.SetCellValue(sheetName, "A2", "Onboarding Report")
	f.SetColWidth(sheetName, "A", "R", 20)
	f.MergeCell(sheetName, "A2", "R2")

	f.SetCellValue(sheetName, "A3", "ข้อมูลตั้งแต่วันที่ "+startTimeStr+" to "+endTimeStr)
	f.SetColWidth(sheetName, "A", "R", 20)
	f.MergeCell(sheetName, "A3", "R3")

	f.SetCellValue(sheetName, "A4", "วันที่รายงาน "+formattedTime)
	f.SetColWidth(sheetName, "A", "R", 20)
	f.MergeCell(sheetName, "A4", "R4")

	f.SetCellValue(sheetName, "A5", "ลำดับ")
	f.SetColWidth(sheetName, "A", "A", 20)
	f.MergeCell(sheetName, "A5", "A7")

	f.SetCellValue(sheetName, "B5", "วันที่ลงทะเบียน")
	f.SetColWidth(sheetName, "B", "B", 20)
	f.MergeCell(sheetName, "B5", "B7")

	f.SetCellValue(sheetName, "C5", "User ID")
	f.SetColWidth(sheetName, "C", "C", 20)
	f.MergeCell(sheetName, "C5", "C7")

	f.SetCellValue(sheetName, "D5", "Device ID")
	f.SetColWidth(sheetName, "D", "D", 20)
	f.MergeCell(sheetName, "D5", "D7")

	f.SetCellValue(sheetName, "E5", "Liveness ID")
	f.SetColWidth(sheetName, "E", "E", 20)
	f.MergeCell(sheetName, "E5", "E7")

	f.SetCellValue(sheetName, "F5", "เลขที่บัตรประชาชน")
	f.SetColWidth(sheetName, "F", "F", 20)
	f.MergeCell(sheetName, "F5", "F7")

	f.SetCellValue(sheetName, "G5", "ชื่อ - นามสกุล")
	f.SetColWidth(sheetName, "G", "G", 20)
	f.MergeCell(sheetName, "G5", "G7")

	f.SetCellValue(sheetName, "H5", "เบอร์โทร")
	f.SetColWidth(sheetName, "H", "H", 20)
	f.MergeCell(sheetName, "H5", "H7")

	f.SetCellValue(sheetName, "I5", "รูปแบบการยืนยันตัวตน (MyMo/E-KYC/NDID)")
	f.SetColWidth(sheetName, "I", "I", 20)
	f.MergeCell(sheetName, "I5", "I7")

	f.SetCellValue(sheetName, "J5", "Liveness")
	f.SetColWidth(sheetName, "J", "L", 20)
	f.MergeCell(sheetName, "J5", "L5")

	f.SetCellValue(sheetName, "J6", "เวลาเริ่ม Liveness")
	f.SetColWidth(sheetName, "J", "J", 20)
	f.MergeCell(sheetName, "J6", "J7")

	f.SetCellValue(sheetName, "K6", "สถานะ การทำ Liveness\n(Success/Not Success)")
	f.SetColWidth(sheetName, "K", "K", 20)
	f.MergeCell(sheetName, "K6", "K7")

	f.SetCellValue(sheetName, "L6", "เวลาที่ทำ Liveness สำเร็จ")
	f.SetColWidth(sheetName, "L", "L", 20)
	f.MergeCell(sheetName, "L6", "L7")

	f.SetCellValue(sheetName, "M5", "OCR")
	f.SetColWidth(sheetName, "M", "P", 20)
	f.MergeCell(sheetName, "M5", "P5")

	f.SetCellValue(sheetName, "M6", "เวลาเริ่ม OCR")
	f.SetColWidth(sheetName, "M", "M", 20)
	f.MergeCell(sheetName, "M6", "M7")

	f.SetCellValue(sheetName, "N6", "OCR ID")
	f.SetColWidth(sheetName, "N", "N", 20)
	f.MergeCell(sheetName, "N6", "N7")

	f.SetCellValue(sheetName, "O6", "สถานะ การทำ OCR\n(Success/Not Success)")
	f.SetColWidth(sheetName, "O", "O", 20)
	f.MergeCell(sheetName, "O6", "O7")

	f.SetCellValue(sheetName, "P6", "เวลาที่ทำ OCR สำเร็จ")
	f.SetColWidth(sheetName, "P", "P", 20)
	f.MergeCell(sheetName, "P6", "P7")

	f.SetCellValue(sheetName, "Q5", "ผลเปรียบเทียบใบหน้า")
	f.SetColWidth(sheetName, "Q", "Q", 20)
	f.MergeCell(sheetName, "Q5", "Q7")

	f.SetCellValue(sheetName, "R5", "Result of registration")
	f.SetColWidth(sheetName, "R", "R", 20)
	f.MergeCell(sheetName, "R5", "R7")

	f.SetCellStyle(sheetName, "A1", "Q1", stylecenter)
	f.SetCellStyle(sheetName, "A2", "Q2", stylecenter)
	f.SetCellStyle(sheetName, "A3", "Q3", styleleft)
	f.SetCellStyle(sheetName, "A4", "Q4", styleleft)
	f.SetCellStyle(sheetName, "A5", "A7", combinedStyle)
	f.SetCellStyle(sheetName, "B5", "B7", combinedStyle)
	f.SetCellStyle(sheetName, "C5", "C7", combinedStyle)
	f.SetCellStyle(sheetName, "D5", "D7", combinedStyle)
	f.SetCellStyle(sheetName, "E5", "E7", combinedStyle)
	f.SetCellStyle(sheetName, "F5", "F7", combinedStyle)
	f.SetCellStyle(sheetName, "G5", "G7", combinedStyle)
	f.SetCellStyle(sheetName, "H5", "H7", combinedStyle)
	f.SetCellStyle(sheetName, "I5", "I7", combinedStyle)
	f.SetCellStyle(sheetName, "J5", "L5", combinedStyle)
	f.SetCellStyle(sheetName, "J6", "L7", combinedStyle)
	f.SetCellStyle(sheetName, "K6", "K7", combinedStyle)
	f.SetCellStyle(sheetName, "L6", "L7", combinedStyle)
	f.SetCellStyle(sheetName, "M5", "P5", combinedStyle)
	f.SetCellStyle(sheetName, "M6", "M7", combinedStyle)
	f.SetCellStyle(sheetName, "N6", "N7", combinedStyle)
	f.SetCellStyle(sheetName, "O6", "O7", combinedStyle)
	f.SetCellStyle(sheetName, "P6", "P7", combinedStyle)
	f.SetCellStyle(sheetName, "Q5", "Q7", combinedStyle)
	f.SetCellStyle(sheetName, "R5", "R7", combinedStyle)

	rows, err := config.DbPostgres.Query(`
	SELECT

	COALESCE(CAST(to_char(u.created_at,'DD-MM-YYYY HH24:MI:SS') AS TEXT), '-') AS  created_at,

	COALESCE(CAST(to_char(u.updated_at,'DD-MM-YYYY HH24:MI:SS')  AS TEXT), '-') AS  user_updated_at,
	COALESCE(CAST(u.user_id AS TEXT), '-') AS  user_id,
	COALESCE(CAST( u.device_id AS TEXT), '-') AS  device_id,
	
	COALESCE(CAST( u.idcardno AS TEXT), '-') AS  idcardno,
	COALESCE(CAST(u.th_firstname AS TEXT), '-') AS th_firstname,
	COALESCE(CAST(u.th_lastname AS TEXT), '-') AS th_lastname,
	COALESCE(CAST(u.phone AS TEXT), '-') AS phone,
	
	
	COALESCE(CAST(tl.liveness_id AS TEXT), '-') AS liveness_id,
	COALESCE(CAST(to_char(tl.liveness_start_date,'DD-MM-YYYY HH24:MI:SS') AS TEXT), '-') AS liveness_start_date,
	COALESCE(CAST(to_char(tl.liveness_end_date,'DD-MM-YYYY HH24:MI:SS')  AS TEXT), '-') AS liveness_end_date,
	COALESCE(CAST(tl.liveness_status AS TEXT), '-') AS liveness_status,
	COALESCE(CAST(to_char(tl.created_at,'DD-MM-YYYY HH24:MI:SS') AS TEXT), '-') AS liveness_created_at,
	COALESCE(CAST(to_char(tl.updated_at,'DD-MM-YYYY HH24:MI:SS') AS TEXT), '-') AS liveness_updated_at,


	COALESCE(CAST(to_char(o.ocr_start_date,'DD-MM-YYYY HH24:MI:SS') AS TEXT), '-') AS ocr_start_date,
	COALESCE(CAST(o.ocr_id AS TEXT), '-') AS ocr_id,
	COALESCE(CAST(to_char(o.ocr_end_date,'DD-MM-YYYY HH24:MI:SS') AS TEXT), '-') AS ocr_end_date,
	COALESCE(CAST(o.ocr_status AS TEXT), '-') AS ocr_status,
	COALESCE(CAST(to_char(o.created_at,'DD-MM-YYYY HH24:MI:SS') AS TEXT), '-') AS ocr_created_at,
	COALESCE(CAST(to_char(o.updated_at,'DD-MM-YYYY HH24:MI:SS') AS TEXT), '-') AS ocr_updated_at,
	CASE 
		when tl.liveness_status = 'Success' and o.ocr_status ='Success' then  'Success'
		when tl.liveness_status = 'Success' and o.ocr_status ='False' then  '-'
		when tl.liveness_status = 'Success' and o.ocr_status isnull then  '-'
		when tl.liveness_status = 'Fail' and o.ocr_status = 'Fail' then  'Fail'
		when tl.liveness_status = 'Fail' and o.ocr_status isnull  then  '-'
		else  '-'
	END AS compair_result,
	CASE
	when (
	SELECT 1 FROM moneydd."user" us 
	JOIN moneydd."trans_ocr" os ON os.user_id =  us.user_id 
		WHERE us.user_id=u.user_id 
		AND (SELECT MAX(oss.created_at)  FROM moneydd."trans_ocr" oss  WHERE   os.ocr_status ='Success' AND  oss.user_id=u.user_id AND u.th_firstname IS NOT NULL LIMIT 1 ) = o.created_at LIMIT 1)  = 1 then 'Success'
		else  'Fail'
		END AS   register_result

	FROM 
	moneydd."trans_liveness" AS tl
	INNER JOIN moneydd."user" AS u ON u.user_id = tl.user_id
	LEFT JOIN moneydd."trans_ocr" AS o ON tl.liveness_id = o.liveness_id    
	WHERE DATE(tl.created_at) = CURRENT_DATE -1
	ORDER BY u.created_at
`)
	if err != nil {
		fmt.Println("ddddd")
		fmt.Println(err)
		return err
	}
	defer rows.Close()

	i := 1   // Initialize i outside of the loop
	row := 8 // Start from row index 6 to skip headers
	for rows.Next() {
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("Q%d", row), dataCellStyle)
		var created_at string
		var updated_at string
		var user_id string
		var device_id string
		var liveness_id string
		var idcardno string
		var firstname string
		var lastname string
		var phone string
		// var identity_type sql.NullString
		var liveness_start_date string
		var liveness_end_date string
		var liveness_status string
		var liveness_created_at string
		var liveness_updated_at string
		var ocr_id string
		var ocr_start_date string
		var ocr_end_date string
		var ocr_status string
		var ocr_created_at string
		var ocr_updated_at string
		var compair_result string
		var register_result string

		err := rows.Scan(
			&created_at, &updated_at, &user_id, &device_id,
			&idcardno, &firstname, &lastname, &phone,
			&liveness_id, &liveness_start_date, &liveness_end_date, &liveness_status, &liveness_created_at, &liveness_updated_at,
			&ocr_start_date, &ocr_id, &ocr_end_date, &ocr_status, &ocr_created_at, &ocr_updated_at, &compair_result, &register_result)
		if err != nil {
			// fmt.Println("dddd")
			fmt.Println(err)
			return err
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), i)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), created_at)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), user_id)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), device_id)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), liveness_id)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), idcardno)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), firstname+" "+lastname)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), phone)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), "E-KYC") /*identity_type*/
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), liveness_start_date)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), liveness_status)
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), liveness_end_date)
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), ocr_start_date)
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), ocr_id)
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", row), ocr_status)
		f.SetCellValue(sheetName, fmt.Sprintf("P%d", row), ocr_end_date)
		f.SetCellValue(sheetName, fmt.Sprintf("Q%d", row), compair_result)
		f.SetCellValue(sheetName, fmt.Sprintf("R%d", row), register_result)

		i++
		row++
	}

	Report := "Onboading_report.xlsx"
	// ReportDirectory := "report"
	// ReportPath := filepath.Join(ReportDirectory, Report)
	err = f.SaveAs(Report)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, utils.StdResponse{
			RespCode: utils.CommonRespCode["Excel_ERROR"].Code,
			RespMsg:  utils.CommonRespCode["Excel_ERROR"].Message,
		})
	}

	emailAddresses := []string{
		"kitsada.s@extend-it-resource.com",
		"chaipipat.p@extend-it-resource.com",
		"waterzaza00@gmail.com",
		"arun@moneydd.co.th",
		"ratsuda@moneydd.co.th",
		"warangkana@moneydd.co.th",
		"thanongsak@moneydd.co.th",
		"narongrit@moneydd.co.th",
		"kornranut@moneydd.co.th",
	}

	isOK, errMsg := EmailSMTP(emailAddresses, Report)
	if isOK {
		successString := strconv.FormatBool(true)
		statusCodeString := strconv.Itoa(http.StatusOK)
		go services.Log("", "", "REPORT_SUCCESS", "CORE-SERVICE", statusCodeString, successString, "OK")
		return c.JSON(http.StatusOK, utils.StdResponse{
			RespCode: utils.CommonRespCode["OK"].Code,
			RespMsg:  utils.CommonRespCode["OK"].Message,
		})
	} else {
		notsuccessString := strconv.FormatBool(false)
		statusCodeString := strconv.Itoa(http.StatusOK)
		go services.Log("", "", "REPORT_ERROR", "CORE-SERVICE", statusCodeString, notsuccessString, errMsg)
		fmt.Println(errMsg)
		return c.JSON(http.StatusInternalServerError, utils.StdResponse{
			RespCode: utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Code,
			RespMsg:  utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Message,
		})
	}
}

func EmailSMTP(emailAddresses []string, excelFilePath string) (isOK bool, errMsg string) {
	currentTime := time.Now()
	formattedTime := currentTime.Format("02/01/2006")
	thaiYear := currentTime.Year() + 543
	formattedTimeThai := currentTime.Format("02/01/") + fmt.Sprint(thaiYear)
	formattedTime2 := currentTime.Format("02012006")

	for _, i := range emailAddresses {
		setemail := mail.NewMessage()
		setemail.SetHeader("From", "goodmoney-noreply@moneydd.co.th")
		setemail.SetHeader("To", i)
		setemail.SetHeader("Subject", "Onboding daliy report on "+formattedTime+" (Environment "+os.Getenv("ENV")+")")
		htmlTemplate, err := ioutil.ReadFile("email-report.html")
		if err != nil {
			return false, err.Error()
		}

		htmlContent := string(htmlTemplate)
		htmlContent = strings.Replace(htmlContent, "<!--CURRENT_DATE-->", formattedTimeThai, -1)
		setemail.SetBody("text/html", htmlContent)
		setemail.Attach(excelFilePath, mail.Rename("Onboading_report_"+formattedTime2+".xlsx"))

		d := mail.NewDialer("smtp.office365.com", 587, "goodmoney-noreply@moneydd.co.th", "Moneydd.0766")
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		if err := d.DialAndSend(setemail); err != nil {
			return false, err.Error()
		}
	}
	return true, ""
}

// @Summary Generate a document
// @Description Generate an Excel document
// @Tags document
// @Produce json
// @Router /api/core/documenttest [post]
func GendocumentTEST(c echo.Context) error {
	// Create a new Excel file
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Add headers to the Excel sheet
	headers := []string{
		"ลำดับ",
		"วันที่ลงทะเบียน",
		"User ID",
		"Device ID",
		"Liveness ID",
		"เลขที่บัตรประชาชน",
		"ชื่อ - นามสกุล",
		"เบอร์โทร",
		"รูปแบบการยืนยันตัวตน (MyMo/E-KYC/NDID)",
		"เวลาเริ่ม Liveness",
		"สถานะ การทำ Liveness (Success/Not Success)",
		"เวลาที่ทำ Liveness สำเร็จ",
		"เวลาเริ่ม OCR",
		"จำนวนครั้งในการถ่ายภาพ OCR",
		"สถานะ การทำ OCR (Success/Not Success)",
		"เวลาที่ทำ OCR สำเร็จ",
		"Result of registration",
	}
	i := 1
	row := sheet.AddRow()
	for _, header := range headers {
		cell := row.AddCell()
		cell.Value = header
	}
	rows, err := config.DbPostgres.Query(`SELECT updated_at, user_id, device_id, idcardno, th_firstname, th_lastname, phone
	 FROM moneydd."user"`)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		row = sheet.AddRow()
		var updated_at string
		var user_id string
		var device_id sql.NullString
		var liveness_id string //ยังไม่มี
		var idcardno sql.NullString
		var firstname sql.NullString
		var lastname sql.NullString
		var phone sql.NullString
		var identity_type string         //ยังไม่มี
		var liveness_start_time string   //ยังไม่มี
		var liveness_status string       //ยังไม่มี
		var liveness_success_time string //ยังไม่มี
		var ocr_start_time string        //ยังไม่มี
		var ocr_count int                //ยังไม่มี
		var ocr_status string            //ยังไม่มี
		var ocr_success_time string      //ยังไม่มี
		// var registration_result string

		err := rows.Scan(&updated_at, &user_id, &device_id, &idcardno, &firstname, &lastname, &phone)
		if err != nil {
			fmt.Println(err)
			return err
		}

		cell := row.AddCell()
		cell.SetInt(i)
		cell = row.AddCell()
		cell.Value = updated_at
		cell = row.AddCell()
		cell.Value = user_id
		cell = row.AddCell()
		if device_id.Valid {
			cell.Value = device_id.String
		} else {
			cell.Value = ""
		}
		cell = row.AddCell()
		cell.Value = liveness_id
		cell = row.AddCell()
		if idcardno.Valid {
			cell.Value = idcardno.String
		} else {
			cell.Value = ""
		}
		cell = row.AddCell()
		if idcardno.Valid {
			cell.Value = firstname.String + " " + lastname.String
		} else {
			cell.Value = ""
		}
		cell = row.AddCell()
		if idcardno.Valid {
			cell.Value = phone.String
		} else {
			cell.Value = ""
		}
		cell = row.AddCell()
		cell.Value = identity_type
		cell = row.AddCell()
		cell.Value = liveness_start_time
		cell = row.AddCell()
		cell.Value = liveness_status
		cell = row.AddCell()
		cell.Value = liveness_success_time
		cell = row.AddCell()
		cell.Value = ocr_start_time
		cell = row.AddCell()
		cell.SetInt(ocr_count)
		cell = row.AddCell()
		cell.Value = ocr_status
		cell = row.AddCell()
		cell.Value = ocr_success_time

		i++
	}

	// Save the Excel file
	err = file.Save("example.xlsx")
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Excel file created successfully.")

	return c.JSON(http.StatusOK, utils.StdResponse{
		RespCode: utils.CommonRespCode["OK"].Code,
		RespMsg:  utils.CommonRespCode["OK"].Message,
	})
}

// @Summary
// @Description
// @Tags document
// @Param  request body utils.Email true "Input data"
// @Produce json
// @Router /api/core/emailsmtp [post]
func EmailSMTPTEST(c echo.Context) error {
	currentTime := time.Now()
	formattedTime := currentTime.Format("02/01/2006")
	thaiYear := currentTime.Year() + 543
	formattedTimeThai := currentTime.Format("02/01/") + fmt.Sprint(thaiYear)
	// formattedTime2 := currentTime.Format("02012006")
	body := new(utils.Email)
	if err := c.Bind(body); err != nil {
		log.Printf("Error binding request body: %v", err)
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["INPUT_FAIL"].Code,
			RespMsg:  utils.CommonRespCode["INPUT_FAIL"].Message,
		})
	}

	// for _, i := range body.Email {
	// 	email := string(i)
	setemail := mail.NewMessage()
	setemail.SetHeader("From", "goodmoney-noreply@moneydd.co.th")
	setemail.SetHeader("To", body.Email)
	setemail.SetHeader("Subject", "Onboding daliy report on "+formattedTime)
	htmlTemplate, err := ioutil.ReadFile("email-report.html")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.StdResponse{
			RespCode: utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Code,
			RespMsg:  utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Message,
		})
	}

	htmlContent := string(htmlTemplate)
	htmlContent = strings.Replace(htmlContent, "<!--CURRENT_DATE-->", formattedTimeThai, -1)
	setemail.SetBody("text/html", htmlContent)
	// setemail.Attach(excelFilePath, mail.Rename("Onboading_report_"+formattedTime2+".xlsx"))

	d := mail.NewDialer("smtp.office365.com", 587, "goodmoney-noreply@moneydd.co.th", "Moneydd.0766")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(setemail); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.StdResponse{
			RespCode: utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Code,
			RespMsg:  utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Message,
		})
	}
	// }
	return c.JSON(http.StatusOK, utils.StdResponse{
		RespCode: utils.CommonRespCode["OK"].Code,
		RespMsg:  utils.CommonRespCode["OK"].Message,
	})
}
