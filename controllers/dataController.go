package controllers

import (
	"bytes"
	"core-service/crud"
	"core-service/models"
	"core-service/services"
	"core-service/utils"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// blockfint
const apiurlfacecomparison string = "https://e-kyc.blockfint.com/api/aai/face-identity/v3/face-comparison"
const apiurllicense string = "https://e-kyc.blockfint.com/api/aai/face-identity/v1/auth-license"
const apiurlocr string = "https://e-kyc.blockfint.com/api/aai/face-identity/v3/id-card-ocr"
const apiurltoken string = "https://e-kyc.blockfint.com/api/aai/auth/ticket/v1/generate-token"
const LivenessDetectionURL string = "https://e-kyc.blockfint.com/api/aai/face-identity/v1/liveness-detection"
const accessKey string = "f338abffd7284936"
const secretKey string = "90157bf8468e10bc"

// DOPA
const client_id string = "6614cb84"
const client_secret string = "5346668e3d71bd0165b02a9c6402e937"
const grant_type string = "client_credentials"
const apiurltoken_dopa string = "https://sso.api-sb.gsb.or.th/auth/realms/apim/protocol/openid-connect/token"
const apidopa string = "https://oidcs.api-sb.gsb.or.th/v1/gsbidcard/verify-laser"

// @Summary Pre Calculator V1
// @Description
// @Tags  General
// @Param  request body models.PreCalculatorReq true "Input data"
// @Produce json
// @Router /api/core/pre-calculator-v1 [post]
func Precalculatorv1(c echo.Context) error {
	body := new(models.PreCalculatorReq)
	if err := c.Bind(body); err != nil {
		return err
	}

	if err := c.Validate(body); err != nil {
		res := utils.RespValidateError(err)
		return c.JSON(http.StatusBadRequest, res)
	}

	res, err := precal(*body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["VALIDATE_ERROR"].Code,
			RespMsg:  err.Error(),
		})

	}

	return c.JSON(http.StatusOK, utils.StdResponse{
		RespCode: utils.CommonRespCode["OK"].Code,
		RespMsg:  utils.CommonRespCode["OK"].Message,
		Data:     res,
	})
}

func precal(body models.PreCalculatorReq) (res models.PreCalculatorRes, err error) {
	MaxDSR := 0.95
	InterestRate := 20.5
	MaxProductCredit := 0.0

	body.Amount = utils.RoundUp(body.Amount, -3)

	switch body.LoanType {
	case "NANO":
		MaxProductCredit = 100000.0
		if body.Amount > MaxProductCredit {
			body.Amount = MaxProductCredit
		}
		if body.Installment > 24 {
			body.Installment = 24
		}
	case "P_LOAN":
		MaxProductCredit = 1000000.0
		if body.Amount > MaxProductCredit {
			body.Amount = MaxProductCredit
		}
		if body.Installment > 48 {
			body.Installment = 48
		}
	default:
		return models.PreCalculatorRes{}, utils.ErrPreCalInvaildType
	}

	if !((MaxDSR * body.IncomePerMonth) >= body.ExpensesPerMonth) {
		return models.PreCalculatorRes{
			LoanAmount:        0,
			Installment:       body.Installment,
			InterestRate:      InterestRate,
			InstallmentAmount: 0,
		}, nil
	}
	CreditLine1 := 0.0
	paymentAmount := ((MaxDSR - (body.ExpensesPerMonth / body.IncomePerMonth)) * body.IncomePerMonth)
	if body.Installment == 0 {
		CreditLine1 = paymentAmount / (5.0 / 100)
	} else {
		CreditLine1 = utils.Pv((InterestRate/100.0)/(12.0), float64(body.Installment), -1*paymentAmount, 0, 0)
	}

	MaxApprove := 0.0
	if body.IncomePerMonth >= 30000 {
		MaxApprove = math.Min(body.IncomePerMonth*5, CreditLine1)
	} else {
		MaxApprove = math.Min(body.IncomePerMonth*1.5, CreditLine1)
	}
	MaxApprove = math.Min(MaxApprove, math.Trunc(body.Amount/100)*100)
	MaxApprove = math.Min(MaxApprove, MaxProductCredit)
	MaxApprove = utils.RoundUp(MaxApprove, -2)

	if MaxApprove <= 100000 {
		if body.Installment > 24 {
			body.Installment = 24
		}
	} else {
		if body.Installment > 48 {
			body.Installment = 48
		}
	}

	if body.Installment != 0 {
		res.InstallmentAmount = utils.RoundUp(utils.PMTYear(utils.RoundDown(MaxApprove, -2), InterestRate, float64(body.Installment)), -2)
	} else {
		res.InstallmentAmount = 0
	}
	res.InterestRate = InterestRate
	res.LoanAmount = MaxApprove
	res.Installment = body.Installment

	switch body.LoanType {
	case "NANO":
		if MaxApprove < 8000 {
			res.LoanAmount = 0
			res.InstallmentAmount = 0
		}

	case "P_LOAN":
		if MaxApprove < 10000 {
			res.LoanAmount = 0
			res.InstallmentAmount = 0
		}

	}
	return
}

// @Summary Pre Calculator
// @Description
// @Tags  General
// @Param  request body utils.Pre_Calculator true "Input data"
// @Produce json
// @Router /api/core/pre-calculator [post]
func Precalculator(c echo.Context) error {
	body := new(utils.Pre_Calculator)
	if err := c.Bind(body); err != nil {
		return err
	}
	if err := c.Validate(body); err != nil {
		res := utils.RespValidateError(err)
		return c.JSON(http.StatusBadRequest, res)
	}
	if body.LoanType != "Trem_Loan" && body.LoanType != "Revolving_Loan" {
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["INPUT_FAIL"].Code,
			RespMsg:  utils.CommonRespCode["INPUT_FAIL"].Message,
		})
	}

	var req models.PreCalculatorReq
	req.Installment = body.Installment
	req.ExpensesPerMonth = float64(body.ExpensesPerMonth)
	req.IncomePerMonth = body.IncomePerMonth
	req.Amount = body.Creditlimit

	switch body.EmploymentType {
	case "อาชีพอิสระ":
		req.LoanType = "NANO"
	default:
		req.LoanType = "P_LOAN"
	}

	res, err := precal(req)

	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["VALIDATE_ERROR"].Code,
			RespMsg:  err.Error(),
		})

	}

	if res.Installment > 0 {
		response := utils.TremLoan{
			RecommendCreditLine: fmt.Sprintf("%.2f", res.LoanAmount),
			Installment:         fmt.Sprintf("%.2f", float64(res.Installment)),
			ResultPayment:       res.InstallmentAmount,
		}
		return c.JSON(http.StatusOK, utils.StdResponse{
			RespCode: utils.CommonRespCode["OK"].Code,
			RespMsg:  utils.CommonRespCode["OK"].Message,
			Data:     response,
		})
	} else {
		response := utils.RevolvingLoan{
			RecommendCreditLine: fmt.Sprintf("%.2f", res.LoanAmount),
			InterestRate:        fmt.Sprintf("%.1f", res.InterestRate),
		}
		return c.JSON(http.StatusOK, utils.StdResponse{
			RespCode: utils.CommonRespCode["OK"].Code,
			RespMsg:  utils.CommonRespCode["OK"].Message,
			Data:     response,
		})
	}

}

func calculateLoanParameters(body *utils.Pre_Calculator, resultinstallment float64) (float64, float64) {
	var MaxApproveCreditLine float64
	var interestRate float64

	switch body.LoanType {
	case "Trem_Loan":
		if body.EmploymentType == "อาชีพอิสระ" || body.EmploymentType == "เจ้าของกิจการที่ไม่ได้จดทะเบียนบริษัท" {
			MaxApproveCreditLine, _ = Nano_Loan_Trem(resultinstallment, body.Creditlimit, body.Installment, body.IncomePerMonth)
		} else {
			MaxApproveCreditLine, _ = P_Loan_Trem(resultinstallment, body.Creditlimit, body.Installment, body.IncomePerMonth)
		}
		interestRate = 0.205 / 12
	case "Revolving_Loan":
		if body.EmploymentType == "อาชีพอิสระ" || body.EmploymentType == "เจ้าของกิจการที่ไม่ได้จดทะเบียนบริษัท" {
			MaxApproveCreditLine, _ = Nano_Loan_Revo(resultinstallment, body.IncomePerMonth, body.Creditlimit)
		} else {
			MaxApproveCreditLine, _ = P_Loan_Revo(resultinstallment, body.IncomePerMonth, body.Creditlimit)
		}
		interestRate = 0.205 / 12
	}

	return MaxApproveCreditLine, interestRate
}

func calculateLoanPayment(body *utils.Pre_Calculator, MaxApproveCreditLine, interestRate float64) float64 {
	RecommendCreditLine := IF((MOD(MaxApproveCreditLine, 100) > 99), MaxApproveCreditLine, ROUNDDOWN(MaxApproveCreditLine, -2))
	result := PMT(interestRate, body.Installment, RecommendCreditLine)
	resultPayment := roundToNearest(result, 100.0)
	return resultPayment
}

func P_Loan_Trem(resultinstallment float64, Creditlimit float64, installment int, IncomePerMonth float64) (float64, error) {
	MaxTermLoan := 48
	InterestRate := 0.205 / 12                                              // แปลง 20% อัตราดอกเบี้ยรายปีเป็นรายเดือน
	byGapDSR, err := PV(InterestRate, MaxTermLoan, resultinstallment, 0, 0) // คำนวณ Present Value (PV)
	if err != nil {
		return 0, err
	}

	var maxMultiple float64
	if IncomePerMonth >= 30000 {
		maxMultiple = 5.0
	} else {
		maxMultiple = 1.5
	}
	ApproveCreditLine := maxMultiple * IncomePerMonth
	MaxProductCreditLine := 1000000.0
	minValue := math.Min(byGapDSR, math.Min(ApproveCreditLine, MaxProductCreditLine))
	if minValue > Creditlimit {
		minValue = Creditlimit
	}
	return minValue, nil
}

func P_Loan_Revo(resultinstallment, IncomePerMonth, Creditlimit float64) (float64, error) {
	InterestRate := 0.05
	byGapDSR := resultinstallment / InterestRate
	var maxMultiple float64
	if IncomePerMonth >= 30000 {
		maxMultiple = 5.0
	} else {
		maxMultiple = 1.5
	}
	ApproveCreditLine := maxMultiple * IncomePerMonth
	MaxProductCreditLine := 1000000.0
	minCreditLine := math.Min(byGapDSR, ApproveCreditLine)
	minCreditLine = math.Min(minCreditLine, MaxProductCreditLine)
	if minCreditLine > Creditlimit {
		minCreditLine = Creditlimit
	}
	return minCreditLine, nil
}

func Nano_Loan_Trem(resultinstallment float64, Creditlimit float64, installment int, IncomePerMonth float64) (float64, error) {
	MaxTermLoan := 48
	InterestRate := 0.205 / 12                                              // แปลง 20% อัตราดอกเบี้ยรายปีเป็นรายเดือน
	byGapDSR, err := PV(InterestRate, MaxTermLoan, resultinstallment, 0, 0) // คำนวณ Present Value (PV)
	if err != nil {
		return 0, err
	}
	var maxMultiple float64
	if IncomePerMonth >= 30000 {
		maxMultiple = 5.0
	} else {
		maxMultiple = 1.5
	}

	ApproveCreditLine := maxMultiple * IncomePerMonth
	MaxProductCreditLine := 100000.0
	minValue := math.Min(byGapDSR, math.Min(ApproveCreditLine, MaxProductCreditLine))

	if minValue > Creditlimit {
		minValue = Creditlimit
	}

	return minValue, nil
}

func Nano_Loan_Revo(resultinstallment, IncomePerMonth, Creditlimit float64) (float64, error) {
	InterestRate := 0.05
	byGapDSR := resultinstallment / InterestRate
	var maxMultiple float64
	if IncomePerMonth >= 30000 {
		maxMultiple = 5.0
	} else {
		maxMultiple = 1.5
	}
	ApproveCreditLine := maxMultiple * IncomePerMonth
	MaxProductCreditLine := 100000.0
	minCreditLine := math.Min(byGapDSR, ApproveCreditLine)
	minCreditLine = math.Min(minCreditLine, MaxProductCreditLine)
	if minCreditLine > Creditlimit {
		minCreditLine = Creditlimit
	}
	return minCreditLine, nil
}
func PV(rate float64, nper int, pmt float64, fv float64, pmtType int) (float64, error) {
	pv := 0.0
	denominator := math.Pow(1+rate, float64(nper))
	if pmtType == 0 {
		pv = (pmt * (1 - (1 / denominator)) / rate) - (fv / denominator)
	} else if pmtType == 1 {
		pv = (pmt * (1 - (1 / denominator)) / rate)
	}

	return pv, nil
}
func PMT(interestRate float64, NumberOfPeriods int, Loanamount float64) float64 {
	monthlyInterestRate := interestRate
	numerator := Loanamount * monthlyInterestRate * math.Pow(1+monthlyInterestRate, float64(NumberOfPeriods))
	denominator := math.Pow(1+monthlyInterestRate, float64(NumberOfPeriods)) - 1
	monthlyPayment := numerator / denominator
	monthlyPayment = math.Round(monthlyPayment * 100 / 100)

	return monthlyPayment
}
func Responsedsr(loanType string, Installment int) interface{} {
	switch loanType {
	case "Trem_Loan":
		return utils.TremLoan{
			RecommendCreditLine: fmt.Sprintf("%.2f", 0.00),
			Installment:         fmt.Sprintf("%.2f", float64(Installment)),
			ResultPayment:       0,
		}
	case "Revolving_Loan":
		return utils.RevolvingLoan{
			RecommendCreditLine: fmt.Sprintf("%.2f", 0.00),
			InterestRate:        "20.5%",
		}
	default:
		return nil
	}
}
func roundToNearest(x, nearest float64) float64 {
	return math.Round(x/nearest+0.5) * nearest
}
func IF(condition bool, trueValue float64, falseValue float64) float64 {
	if condition {
		return trueValue
	}
	return falseValue
}
func MOD(x float64, y float64) float64 {
	return math.Mod(x, y)
}

func ROUNDDOWN(number float64, places int) float64 {
	rounding := math.Pow(10, float64(places))
	return math.Floor(number*rounding) / rounding
}

func NewGetTokenBF() utils.GetTokenBF {

	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	combined := accessKey + secretKey + timestamp
	signature := sha256Hex(combined)

	return utils.GetTokenBF{
		AccessKey:    accessKey,
		Timestamp:    timestamp,
		Signature:    signature,
		PeriodSecond: "3600",
	}
}
func sha256Hex(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
func GetTokenHandler(c echo.Context) error {
	tokenResponse, err := GetToken(c)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, tokenResponse)
}

// @Summary Get TokenBF
// @Description
// @Tags General
// @Accept json
// @Produce json
// @Router /api/core/gettoken [post]
func GetToken(c echo.Context) (*utils.ResponseToken, error) {
	_gettoken := NewGetTokenBF()
	url := apiurltoken
	method := "POST"

	payload := map[string]string{
		"accessKey":    _gettoken.AccessKey,
		"signature":    _gettoken.Signature,
		"timestamp":    _gettoken.Timestamp,
		"periodSecond": _gettoken.PeriodSecond,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(payloadBytes))
	if err != nil {
		fmt.Println(err)
		return nil, err

	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	response := utils.ResponseToken{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &response, nil
}

// @Summary Liveness License
// @Description
// @Tags General
// @Accept json
// @Produce json
// @Router /api/core/liveness-license [post]
func LivenessLicense(c echo.Context) error {

	// fmt.Println("LivenessLicense: ", 1)
	// userID := uuid.New().String()
	userID := c.Get("userID").(string)
	tranId := uuid.New().String()
	payload := map[string]string{
		"LicenseEffectiveSeconds": "86400",
		"ApplicationID":           "com.moneydd.goodmoney",
	}

	// fmt.Println("LivenessLicense: ", 2)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return err
	}
	client := &http.Client{}

	// fmt.Println("LivenessLicense: ", 3)
	req, err := http.NewRequest("POST", apiurllicense, bytes.NewReader(payloadBytes))
	if err != nil {
		fmt.Println(err)
		return err
	}

	// fmt.Println("LivenessLicense: ", 4)
	tokenResponse, err := GetToken(c)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// fmt.Println("LivenessLicense: ", 5)
	token := tokenResponse.Data.Token
	req.Header.Add("X-ACCESS-TOKEN", token)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	// fmt.Println("LivenessLicense: ", 6)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	payloadStr := string(payloadBytes)

	// fmt.Println("LivenessLicense: ", 7)
	_license := utils.Liveness_License{}
	err = json.Unmarshal(body, &_license)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// fmt.Println("LivenessLicense: ", 8)
	_licenseData := _license.Data
	_licenseDataJSON, err := json.Marshal(_licenseData)
	_licenseDataStr := string(_licenseDataJSON)
	if err != nil {
		fmt.Println(err)
		go services.Log(userID, tranId, "LivenessLicense_ERROR", apiurllicense, fmt.Sprintf("%d", res.StatusCode), payloadStr, _licenseDataStr)
		return err
	}

	// fmt.Println("LivenessLicense: ", 9)
	go services.Log(userID, tranId, "LivenessLicense_ERROR", apiurllicense, fmt.Sprintf("%d", res.StatusCode), payloadStr, _licenseDataStr)
	if _license.Code == "LIVENESS_ID_NOT_EXISTED" {
		go services.Log(userID, tranId, "LivenessLicense_ERROR", apiurllicense, fmt.Sprintf("%d", res.StatusCode), payloadStr, _licenseDataStr)
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["LIVENESS_ID_NOT_ERROR"].Code,
			RespMsg:  utils.CommonRespCode["LIVENESS_ID_NOT_ERROR"].Message,
		})
	}

	go services.Log(userID, tranId, "LivenessLicense_SUCCESS", apiurllicense, fmt.Sprintf("%d", res.StatusCode), payloadStr, _licenseDataStr)
	return c.JSON(http.StatusOK, utils.StdResponse{
		RespCode: utils.CommonRespCode["OK"].Code,
		RespMsg:  utils.CommonRespCode["OK"].Message,
		Data:     _license,
	})
}

// @Summary Insert Trans Liveness
// @Description
// @Tags General
// @Accept json
// @Param request body utils.TransLiveness true "Body"
// @Produce json
// @Router /api/core/insert-trans-liveness [post]
// @Security BearerAuth
func InsertTransLiveness(c echo.Context) error {

	var body utils.TransLiveness
	// fmt.Println("InsertTransLiveness: ", 1)
	if err := c.Bind(&body); err != nil {
		fmt.Println("1")
		return c.String(http.StatusBadRequest, err.Error())
	}

	// fmt.Println("InsertTransLiveness: ", 2)
	if err := c.Validate(body); err != nil {
		fmt.Println("2")
		res := utils.RespValidateError(err)
		return c.JSON(http.StatusBadRequest, res)
	}

	// fmt.Println("InsertTransLiveness: ", 3)
	body.UserId = c.Get("userID").(string)
	//body.UserId = uuid.NewString()
	body.CreateAt = time.Now()
	body.UpdateAt = time.Now()
	fmt.Println(body.UserId)

	query := `INSERT INTO moneydd.trans_liveness ( user_id, liveness_start_date, liveness_end_date, liveness_status, created_at, updated_at) VALUES ( :user_id, :liveness_start_date, :liveness_end_date, :liveness_status, :created_at, :updated_at)`

	// fmt.Println("InsertTransLiveness: ", 4)
	err := crud.Create(c, query, body)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Code,
			RespMsg:  utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Message,
			Error:    err,
		})
	}

	return c.JSON(http.StatusOK, utils.StdResponse{
		RespCode: utils.CommonRespCode["OK"].Code,
		RespMsg:  utils.CommonRespCode["OK"].Message,
	})
}

// @Summary Insert Trans Ocr
// @Description
// @Tags General
// @Accept json
// @Param request body utils.TransOcr true "Body"
// @Produce json
// @Router /api/core/insert-trans-ocr [post]
// @Security BearerAuth
func InsertTransOcr(c echo.Context) error {
	var body utils.TransOcr
	if err := c.Bind(&body); err != nil {
		fmt.Println("1")
		return c.String(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(body); err != nil {
		fmt.Println("2")
		res := utils.RespValidateError(err)
		return c.JSON(http.StatusBadRequest, res)
	}

	body.UserId = c.Get("userID").(string)
	//body.UserId = uuid.NewString()
	body.CreateAt = time.Now()
	body.UpdateAt = time.Now()
	fmt.Println(body.UserId)
	query := `INSERT  INTO moneydd.trans_ocr ( user_id, ocr_start_date, ocr_end_date, ocr_status, created_at, updated_at,liveness_id ) VALUES ( :user_id, :ocr_start_date, :ocr_end_date, :ocr_status, :created_at, :updated_at ,(SELECT liveness_id FROM moneydd.trans_liveness WHERE user_id = :user_id ORDER BY liveness_id DESC LIMIT 1) )`

	err := crud.Create(c, query, body)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Code,
			RespMsg:  utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Message,
			Error:    err,
		})
	}

	return c.JSON(http.StatusOK, utils.StdResponse{
		RespCode: utils.CommonRespCode["OK"].Code,
		RespMsg:  utils.CommonRespCode["OK"].Message,
	})
}

// @Summary Ocr ID Card
// @Description
// @Tags General
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "id card image file"
// @Router /api/core/ocr-id-card [post]
// @Security BearerAuth
func OcrIDCard(c echo.Context) error {
	// userID := uuid.New().String()
	userID := c.Get("userID").(string)
	tranId := uuid.New().String()
	file, errFile := c.FormFile("image")
	if errFile != nil {
		fmt.Println(errFile)
		return errFile
	}
	src, errOpen := file.Open()
	if errOpen != nil {
		fmt.Println(errOpen)
		return errOpen
	}
	defer src.Close()

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	part, errCreate := writer.CreateFormFile("image", file.Filename)
	if errCreate != nil {
		fmt.Println(errCreate)
		return errCreate
	}
	_, errCopy := io.Copy(part, src)
	if errCopy != nil {
		fmt.Println(errCopy)
		return errCopy
	}
	side := "front"
	_ = writer.WriteField("side", side)

	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	if os.Getenv("ENV") == "loadtest" {
		fakeOCR := utils.OCRIDCard{}
		gofakeit.Struct(&fakeOCR)
		return c.JSON(http.StatusOK, utils.StdResponse{
			RespCode: utils.CommonRespCode["OK"].Code,
			RespMsg:  utils.CommonRespCode["OK"].Message,
			Data:     fakeOCR,
		})
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", apiurlocr, payload)
	if err != nil {
		fmt.Println(err)
		return err
	}

	tokenResponse, err := GetToken(c)
	if err != nil {
		fmt.Println(err)
		return err
	}
	token := tokenResponse.Data.Token
	req.Header.Add("X-ACCESS-TOKEN", token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	ocr_idcard := utils.OCRIDCard{}
	err = json.Unmarshal(body, &ocr_idcard)
	if err != nil {
		fmt.Println(err)
		return err

	}
	payloadStr := payload.String()
	ocr_idcardStr, err := json.Marshal(ocr_idcard)
	if ocr_idcard.Code == "PARAMETER_ERROR" || ocr_idcard.Code == "TOO_MANY_CARDS" || ocr_idcard.Data.SerialNumber == "" || ocr_idcard.Code == "OCR_NO_RESULT" {
		go services.Log(userID, tranId, "OCR_ERROR", apiurlocr, fmt.Sprintf("%d", res.StatusCode), payloadStr, string(ocr_idcardStr))
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["OCR_NO_RESULT"].Code,
			RespMsg:  utils.CommonRespCode["OCR_NO_RESULT"].Message,
		})
	}
	//ที่เพิ่มล่าสุด
	homeaddress := strings.TrimSpace(strings.Replace(ocr_idcard.Data.HomeAddressTH, "ที่อยู่", "", -1))
	if len(homeaddress) >= 4 {
		laneTH := ""
		roadTH := ""
		addressParts := strings.Fields(homeaddress)
		for i := 1; i < len(addressParts); i++ {
			part := addressParts[i]
			if strings.HasPrefix(part, "ซ.") {
				laneTH = strings.TrimSpace(strings.Replace(part, "ซ.", "", -1))
				homeaddress = strings.TrimSpace(strings.Replace(homeaddress, "ซ."+laneTH, "", -1))
			} else if strings.HasPrefix(part, "ถ.") {
				roadTH = strings.TrimSpace(strings.Replace(part, "ถ.", "", -1))
				homeaddress = strings.TrimSpace(strings.Replace(homeaddress, "ถ."+roadTH, "", -1))

			}
		}

		subdistrictTH := strings.TrimSpace(strings.Replace(ocr_idcard.Data.SubDistrictTH, "ต.", "", -1))
		subdistrictTH = strings.TrimSpace(strings.Replace(subdistrictTH, "แขวง", "", -1))
		districtTH := strings.TrimSpace(strings.Replace(ocr_idcard.Data.DistrictTH, "อ.", "", -1))
		districtTH = strings.TrimSpace(strings.Replace(districtTH, "เขต", "", -1))
		provinceTH := strings.TrimSpace(strings.Replace(ocr_idcard.Data.ProvinceTH, "จ.", "", -1))

		namePartsTH := strings.SplitN(ocr_idcard.Data.NameTH, " ", 3)
		if len(namePartsTH) >= 3 {
			prefixnameTH := strings.TrimSpace(namePartsTH[0])
			firstNameTH := strings.TrimSpace(namePartsTH[1])
			lastNameTH := strings.TrimSpace(namePartsTH[2])
			namePartsEN := strings.SplitN(ocr_idcard.Data.NameEN, " ", 2)
			prefixNameEN := ""
			firstNameEn := ""
			LastNameEN := ""
			if len(namePartsEN) >= 2 {
				prefixNameEN = strings.TrimSpace(strings.Replace(namePartsEN[0], ".", "", -1))
				firstNameEn = namePartsEN[1]
				LastNameEN = strings.TrimSpace(ocr_idcard.Data.LastNameEN)
			}

			updatedOCRData := utils.OCRIDCardData{
				IDNumber:         ocr_idcard.Data.IDNumber,
				SerialNumber:     ocr_idcard.Data.SerialNumber,
				TypeEN:           ocr_idcard.Data.TypeEN,
				TypeTH:           ocr_idcard.Data.TypeTH,
				NameEN:           ocr_idcard.Data.NameEN,
				NameTH:           ocr_idcard.Data.NameTH,
				BirthdayEN:       ocr_idcard.Data.BirthdayEN,
				BirthdayTH:       ocr_idcard.Data.BirthdayTH,
				IssueDateEN:      ocr_idcard.Data.IssueDateEN,
				IssueDateTH:      ocr_idcard.Data.IssueDateTH,
				ExpiryDateEN:     ocr_idcard.Data.ExpiryDateEN,
				ExpiryDateTH:     ocr_idcard.Data.ExpiryDateTH,
				ReligionTH:       ocr_idcard.Data.ReligionTH,
				IssuingOfficerTH: ocr_idcard.Data.IssuingOfficerTH,
				AddressAllTH:     ocr_idcard.Data.AddressAllTH,
				HomeAddressTH:    homeaddress,
				SubDistrictTH:    subdistrictTH,
				DistrictTH:       districtTH,
				ProvinceTH:       provinceTH,
				AddressOthersTH:  ocr_idcard.Data.AddressOthersTH,
				PrefixNameTH:     prefixnameTH,
				FirstNameTH:      firstNameTH,
				LastNameTH:       lastNameTH,
				PrefixNameEN:     prefixNameEN,
				FirstNameEN:      firstNameEn,
				LastNameEN:       LastNameEN,
				LaneTH:           laneTH,
				RoadTH:           roadTH,
			}

			ocr_idcard.Data = updatedOCRData

		}
		if err != nil {
			fmt.Println(err)
			go services.Log(userID, tranId, "OCR_Result", apiurlocr, fmt.Sprintf("%d", res.StatusCode), payloadStr, string(ocr_idcardStr))

			return err
		}
		go services.Log(userID, tranId, "OCR_Result", apiurlocr, fmt.Sprintf("%d", res.StatusCode), payloadStr, string(ocr_idcardStr))
		return c.JSON(http.StatusOK, utils.StdResponse{
			RespCode: utils.CommonRespCode["OK"].Code,
			RespMsg:  utils.CommonRespCode["OK"].Message,
			Data:     ocr_idcard,
		})
	}
	return c.JSON(http.StatusOK, utils.StdResponse{
		RespCode: utils.CommonRespCode["OK"].Code,
		RespMsg:  utils.CommonRespCode["OK"].Message,
		Data:     ocr_idcard,
	})
}

// @Summary Face Comparison
// @Description Face Comparison
// @Tags General
// @Accept multipart/form-data
// @Produce json
// @Param firstImage formData file true "The first face image"
// @Param secondImage formData file true "The second face image"
// @Router /api/core/face-comparison [post]
// @Security BearerAuth
func FaceComparison(c echo.Context) error {
	firstFile, errFirstFile := c.FormFile("firstImage")
	if errFirstFile != nil {
		fmt.Println(errFirstFile)
		return errFirstFile
	}

	secondFile, errSecondFile := c.FormFile("secondImage")
	if errSecondFile != nil {
		fmt.Println(errSecondFile)
		return errSecondFile
	}

	firstSrc, errFirstOpen := firstFile.Open()
	if errFirstOpen != nil {
		fmt.Println(errFirstOpen)
		return errFirstOpen
	}
	defer firstSrc.Close()

	secondSrc, errSecondOpen := secondFile.Open()
	if errSecondOpen != nil {
		fmt.Println(errSecondOpen)
		return errSecondOpen
	}
	defer secondSrc.Close()

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	firstPart, errFirstCreate := writer.CreateFormFile("firstImage", firstFile.Filename)
	if errFirstCreate != nil {
		fmt.Println(errFirstCreate)
		return errFirstCreate
	}
	_, errFirstCopy := io.Copy(firstPart, firstSrc)
	if errFirstCopy != nil {
		fmt.Println(errFirstCopy)
		return errFirstCopy
	}

	secondPart, errSecondCreate := writer.CreateFormFile("secondImage", secondFile.Filename)
	if errSecondCreate != nil {
		fmt.Println(errSecondCreate)
		return errSecondCreate
	}
	_, errSecondCopy := io.Copy(secondPart, secondSrc)
	if errSecondCopy != nil {
		fmt.Println(errSecondCopy)
		return errSecondCopy
	}

	errWriterClose := writer.Close()
	if errWriterClose != nil {
		fmt.Println(errWriterClose)
		return errWriterClose
	}

	if os.Getenv("ENV") == "loadtest" {
		faceComparisonResult := utils.FaceComparison{}
		gofakeit.Struct(&faceComparisonResult)
		return c.JSON(http.StatusOK, utils.StdResponse{
			RespCode: utils.CommonRespCode["OK"].Code,
			RespMsg:  utils.CommonRespCode["OK"].Message,
			Data:     faceComparisonResult,
		})
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", apiurlfacecomparison, payload)
	if err != nil {
		fmt.Println(err)
		return err
	}

	tokenResponse, err := GetToken(c)
	if err != nil {
		fmt.Println(err)
		return err
	}
	token := tokenResponse.Data.Token
	req.Header.Add("X-ACCESS-TOKEN", token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	faceComparisonResult := utils.FaceComparison{}
	err = json.Unmarshal(body, &faceComparisonResult)
	if err != nil {
		fmt.Println(err)
		return err
	}

	faceComparisonResultData := faceComparisonResult.Data
	faceComparisonResultDataJSON, err := json.Marshal(faceComparisonResultData)
	faceComparisonResultstr := string(faceComparisonResultDataJSON)
	payloadStr := payload.String()

	if faceComparisonResult.Code == "PARAMETER_ERROR" || faceComparisonResult.Code == "FIRST_IMAGE_LOW_QUALITY_FACE" || faceComparisonResult.Code == "SECOND_IMAGE_LOW_QUALITY_FACE" ||
		faceComparisonResult.Code == "NO_FACE_DETECTED_FROM_FIRST_IMAGE" || faceComparisonResult.Code == "NO_FACE_DETECTED_FROM_SECOND_IMAGE" {
		go services.Log(c.Get("customerId").(string), c.Get("transactionId").(string), "FaceComparison_ERROR", apiurllicense, fmt.Sprintf("%d", res.StatusCode), payloadStr, faceComparisonResultstr)
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["FACE_DETECTED_ERROR"].Code,
			RespMsg:  utils.CommonRespCode["FACE_DETECTED_ERROR"].Message,
		})
	} else if faceComparisonResult.Data.Similarity <= 70 {
		go services.Log(c.Get("customerId").(string), c.Get("transactionId").(string), "FaceComparison_ERROR", apiurllicense, fmt.Sprintf("%d", res.StatusCode), payloadStr, faceComparisonResultstr)
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["SIMILARITIES_ERROR"].Code,
			RespMsg:  utils.CommonRespCode["SIMILARITIES_ERROR"].Message,
		})
	}
	ff, _ := utils.FileMultipartToByte(firstFile)
	go uploadFile(c, "users/"+c.Get("customerId").(string)+"/liveness.jpg", firstFile.Header["Content-Type"][0], ff)

	sf, _ := utils.FileMultipartToByte(secondFile)
	go uploadFile(c, "users/"+c.Get("customerId").(string)+"/id-card.jpg", secondFile.Header["Content-Type"][0], sf)
	go services.Log(c.Get("customerId").(string), c.Get("transactionId").(string), "FaceComparison_SUCCESS", apiurllicense, fmt.Sprintf("%d", res.StatusCode), payloadStr, faceComparisonResultstr)
	return c.JSON(http.StatusOK, utils.StdResponse{
		RespCode: utils.CommonRespCode["OK"].Code,
		RespMsg:  utils.CommonRespCode["OK"].Message,
		Data:     faceComparisonResult,
	})
}

func GenDOPAToken() (response utils.DOPAToken, err error) {
	// payload := strings.NewReader("client_id=" + client_id + "&client_secret=" + client_secret + "&grant_type=" + grant_type)
	payload := url.Values{}
	payload.Set("client_id", client_id)
	payload.Set("client_secret", client_secret)
	payload.Set("grant_type", grant_type)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, apiurltoken_dopa, strings.NewReader(payload.Encode()))
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	// body, err := io.ReadAll(res.Body)
	respBodyBuff := new(bytes.Buffer)
	_, err = io.Copy(respBodyBuff, res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(respBodyBuff.Bytes(), &response)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

// @Summary DOPA
// @Description
// @Tags User
// @Param request body utils.Dopareq true "Body"
// @Produce json
// @Router /api/core/dopa [post]
func DOPA(c echo.Context) error {
	// userID := uuid.New().String()
	userID := c.Get("userID").(string)
	tranId := uuid.NewString()
	var Dopareq utils.Dopareq
	// fmt.Println("DOPA: ", 1)
	if err := c.Bind(&Dopareq); err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["INPUT_FAIL"].Code,
			RespMsg:  utils.CommonRespCode["INPUT_FAIL"].Message,
			Data: utils.DOPAResp{
				Code: "5",
				Desc: "ข้อมูลที่ใช้ในการตรวจสอบไม่ถูกต้อง",
			},
			Error: err.Error(),
		})
	}

	// fmt.Println("DOPA: ", Dopareq)
	// fmt.Println("DOPA: ", 2)
	tokenResponse, err := GenDOPAToken()
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusServiceUnavailable, utils.StdResponse{
			RespCode: utils.CommonRespCode["INTERNAL_SERVER_ERROR_DOPA"].Code,
			RespMsg:  utils.CommonRespCode["INTERNAL_SERVER_ERROR_DOPA"].Message,
			Data: utils.DOPAResp{
				Code: "5",
				Desc: "ข้อมูลที่ใช้ในการตรวจสอบไม่ถูกต้อง",
			},
			Error: err.Error(),
		})
	}
	token := tokenResponse.AccessToken
	payload := map[string]string{
		"AppId":     "MDD",
		"PID":       Dopareq.PID,
		"FirstName": Dopareq.FirstName,
		"LastName":  Dopareq.LastName,
		"BirthDay":  Dopareq.BirthDay,
		"Laser":     Dopareq.Laser,
	}

	if os.Getenv("ENV") == "loadtest" {
		c.JSON(http.StatusOK, utils.StdResponse{
			RespCode: utils.CommonRespCode["OK"].Code,
			RespMsg:  utils.CommonRespCode["OK"].Message,
			Data: utils.DOPAResp{
				Code: json.Number("0"),
				Desc: "สถานะปกติ",
			},
		})
	}

	// fmt.Println("DOPA: ", 3)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["INPUT_FAIL"].Code,
			RespMsg:  utils.CommonRespCode["INPUT_FAIL"].Message,
			Data: utils.DOPAResp{
				Code: "5",
				Desc: "ข้อมูลที่ใช้ในการตรวจสอบไม่ถูกต้อง",
			},
			Error: err.Error(),
		})
	}

	payloadReader := bytes.NewReader(payloadBytes)

	// fmt.Println("DOPA: ", 4)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest(http.MethodPost, apidopa, payloadReader)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, utils.StdResponse{
			RespCode: utils.CommonRespCode["INTERNAL_SERVER_ERROR_DOPA"].Code,
			RespMsg:  utils.CommonRespCode["INTERNAL_SERVER_ERROR_DOPA"].Message,
			Data: utils.DOPAResp{
				Code: "5",
				Desc: "ข้อมูลที่ใช้ในการตรวจสอบไม่ถูกต้อง",
			},
			Error: err.Error(),
		})
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// fmt.Println("DOPA: ", 5)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadGateway, utils.StdResponse{
			RespCode: utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Code,
			RespMsg:  utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Message,
			Data: utils.DOPAResp{
				Code: "5",
				Desc: "ข้อมูลที่ใช้ในการตรวจสอบไม่ถูกต้อง",
			},
			Error: err.Error(),
		})
	}
	defer res.Body.Close()

	// fmt.Println("DOPA: ", 6)
	// body, err := io.ReadAll(res.Body)
	respBodyBuff := new(bytes.Buffer)
	_, err = io.Copy(respBodyBuff, res.Body)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadGateway, utils.StdResponse{
			RespCode: utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Code,
			RespMsg:  utils.CommonRespCode["INTERNAL_SERVER_ERROR"].Message,
			Data: utils.DOPAResp{
				Code: "5",
				Desc: "ข้อมูลที่ใช้ในการตรวจสอบไม่ถูกต้อง",
			},
			Error: err.Error(),
		})
	}

	// fmt.Println("DOPA: ", 8)
	dopaResp := utils.DOPAResp{}
	err = json.Unmarshal(respBodyBuff.Bytes(), &dopaResp)
	if err != nil {
		fmt.Println(err)
		go services.Log(userID, tranId, "DOPA_ERROR", apidopa, fmt.Sprintf("%d", res.StatusCode), string(payloadBytes), respBodyBuff.String())
		return c.JSON(http.StatusInternalServerError, utils.StdResponse{
			RespCode: utils.CommonRespCode["INTERNAL_SERVER_ERROR_DOPA"].Code,
			RespMsg:  utils.CommonRespCode["INTERNAL_SERVER_ERROR_DOPA"].Message,
			Data: utils.DOPAResp{
				Code: "5",
				Desc: "ข้อมูลที่ใช้ในการตรวจสอบไม่ถูกต้อง",
			},
			Error: err.Error(),
		})
	}

	// fmt.Println("DOPA: ", 10)
	// fmt.Println("DOPA: ", _Dopares)
	switch dopaResp.Code.String() {
	case "0":
		go services.Log(userID, tranId, "DOPA_SUCCESS", apidopa, fmt.Sprintf("%d", res.StatusCode), string(payloadBytes), respBodyBuff.String())
		return c.JSON(http.StatusOK, utils.StdResponse{
			RespCode: utils.CommonRespCode["OK"].Code,
			RespMsg:  utils.CommonRespCode["OK"].Message,
			Data:     dopaResp,
		})
	case "2":
		go services.Log(userID, tranId, "DOPA_ERROR", apidopa, fmt.Sprintf("%d", res.StatusCode), string(payloadBytes), respBodyBuff.String())
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["END_OF_USE_PERIOD"].Code,
			RespMsg:  utils.CommonRespCode["END_OF_USE_PERIOD"].Message,
			Data:     dopaResp,
		})
	case "4":
		go services.Log(userID, tranId, "DOPA_ERROR", apidopa, fmt.Sprintf("%d", res.StatusCode), string(payloadBytes), respBodyBuff.String())
		if strings.Contains(dopaResp.Desc, "ข้อมูลไม่ตรง") {
			return c.JSON(http.StatusBadRequest, utils.StdResponse{
				RespCode: utils.CommonRespCode["INFORMATION_FAIL"].Code,
				RespMsg:  utils.CommonRespCode["INFORMATION_FAIL"].Message,
				Data:     dopaResp,
			})
		}
		if strings.Contains(dopaResp.Desc, "ไม่พบเลขรหัสกำกับบัตร") {
			return c.JSON(http.StatusBadRequest, utils.StdResponse{
				RespCode: utils.CommonRespCode["IDCARD_FAIL"].Code,
				RespMsg:  utils.CommonRespCode["IDCARD_FAIL"].Message,
				Data:     dopaResp,
			})
		}
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["INFORMATION_FAIL"].Code,
			RespMsg:  utils.CommonRespCode["INFORMATION_FAIL"].Message,
			Data:     dopaResp,
		})
	case "5":
		go services.Log(userID, tranId, "DOPA_ERROR", apidopa, fmt.Sprintf("%d", res.StatusCode), string(payloadBytes), respBodyBuff.String())
		return c.JSON(http.StatusBadRequest, utils.StdResponse{
			RespCode: utils.CommonRespCode["INPUT_DOPA"].Code,
			RespMsg:  utils.CommonRespCode["INPUT_DOPA"].Message,
			Data:     dopaResp,
		})
	default:
		go services.Log(userID, tranId, "DOPA_ERROR", apidopa, fmt.Sprintf("%d", res.StatusCode), string(payloadBytes), respBodyBuff.String())
		return c.JSON(http.StatusInternalServerError, utils.StdResponse{
			RespCode: utils.CommonRespCode["INTERNAL_SERVER_ERROR_DOPA"].Code,
			RespMsg:  utils.CommonRespCode["INTERNAL_SERVER_ERROR_DOPA"].Message,
			Data:     dopaResp,
		})
	}
}
