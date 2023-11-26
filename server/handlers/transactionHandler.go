package handlers

import (
	"Qpay/models"
	"Qpay/services"
	"Qpay/utils"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type CreateTransactionRequest struct {
	PhoneNumber   string  `json:"phone_number"`
	PaymentAmount float64 `json:"payment_amount"`
}
type CreateTransactionResponse struct {
	TransactionID uint `json:"id"`
}

type TransactionStartResponse struct {
	PaymentAmount float64 `json:"payment_amount"`
	OwnerName     string  `json:"owner_name"`
}
type TransactionResponse struct {
	TrackingCode  string  `json:"tracking_code"`
	Status        string  `json:"status"`
	PurchaserCard string  `json:"purchaser_card"`
	PaymentAmount float64 `json:"payment_amount"`
	PhoneNumber   string  `json:"phone_number"`
	PaymentDate   string  `json:"payment_date"`
}

func (h *Handler) ListAllTransaction(ctx echo.Context) error {
	h.SetUserID(ctx)
	return nil
}

func (h *Handler) FindTransaction(ctx echo.Context) error {
	h.SetUserID(ctx)
	return nil
}

func (h *Handler) FilterTransaction(ctx echo.Context) error {
	h.SetUserID(ctx)
	//	امکان فیلتر کردن تراکنش‌ها بر حسب تاریخ و یا قیمت (بازه زمانی و یا قیمتی)
	return nil
}
func (h *Handler) SearchTransaction(ctx echo.Context) error {
	h.SetUserID(ctx)
	//	امکان جستجو در تراکنش‌های ثبت شده بر حسب تاریخ و یا قیمت (بازه زمانی و یا قیمتی)
	return nil
}

func (h *Handler) CreateTransaction(ctx echo.Context) error {
	route := ctx.Param("route")
	var req CreateTransactionRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "Bind Error")
	}
	gateway, err := services.GetGateway(h.DB, "route", route)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, "No get way with such route")
	}

	commission := req.PaymentAmount*gateway.Commission.PercentagePerTrans + gateway.Commission.AmountPerTrans
	model, err := services.CreateTransaction(h.DB, gateway.ID, req.PaymentAmount, req.PhoneNumber, commission)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, CreateTransactionResponse{TransactionID: model.ID})
}

//	func (tr *Handler) RequestPersonalTransaction(ctx echo.Context) error {
//		// دریافت پست مقادیر زیر
//		//	آدرس درگاه - Route - باید به آی دی تبدیل بشه
//		//	مقدار پرداخت
//		//	شماره موبایل
//
//		//	ریسپانس مقادیر زیر
//		//	آی دی تراکنش
//		return nil
//	}
//
//	func (tr *Handler) RequestBusinessTransaction(ctx echo.Context) error {
//		// دریافت پست مقادیر زیر
//		//	آی دی درگاه merchantId
//		//	مقدار پرداخت
//		//	شماره موبایل
//
//		//	ریسپانس مقادیر زیر
//		//	آی دی تراکنش
//		return nil
//	}
func (h *Handler) GetTransactionForStart(ctx echo.Context) error {
	transactionid, _ := strconv.ParseUint(ctx.Param("id"), 10, 64)
	transaction, err := services.GetTransactionByID(h.DB, uint(transactionid))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	gateway, err := services.GetGateway(h.DB, "id", fmt.Sprintf("%v", transaction.GatewayID))
	return ctx.JSON(http.StatusOK, TransactionStartResponse{PaymentAmount: transaction.PaymentAmount, OwnerName: gateway.BankAccount.AccountOwner})
}

type PaymentTransactionRequest struct {
	PurchaserCard       string `json:"purchaser_card"`
	CardMonth           int    `json:"card_month"`
	CardYear            int    `json:"card_year"`
	TransactionID       uint   `json:"transaction_id"`
	PaymentConfirmation bool   `json:"payment_confirmation"` //	دستور پرداخت و کم کردن موجودی (کنسل تراکنش - پرداخت)
}
type PaymentTransactionResponse struct {
	TransactionID uint    `json:"transaction_id"`
	TrackingCode  string  `json:"tracking_code"`
	Status        uint8   `json:"status"`
	PaymentAmount float64 `json:"payment_amount"`
	PurchaserCard string  `json:"purchaser_card"`
}

func BeginTransactionResponse(transaction models.Transaction) PaymentTransactionResponse {
	return PaymentTransactionResponse{
		TransactionID: transaction.ID,
		TrackingCode:  transaction.TrackingCode,
		Status:        transaction.Status,
		PaymentAmount: transaction.PaymentAmount,
		PurchaserCard: transaction.PurchaserCard,
	}
}
func (h *Handler) BeginTransaction(ctx echo.Context) error {
	var req PaymentTransactionRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, "Bind Error")
	}
	// اینجا چک میکنه اگه طرف فیلد پیمنت کانفیرم رو فالس داده بود
	//یعنی میخواد پرداخت رو کنسل کنه و پرداخت انجام نده
	if !req.PaymentConfirmation {
		if err := services.CancelledTransaction(h.DB, req.TransactionID); err != nil {
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}
		return ctx.JSON(http.StatusNotAcceptable, "your Payment Transaction is Canceled")
	}

	// حالا که همه چیز آماده انجام تراکنش هست باید اول بررسی شود که
	// فلیدهای لازم درون درخواست وجود داند یا خیر؟
	if err := ValidateTransaction(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	// اینجا باید به ماک متصل بشم و یه خروجی ازش بگیرم که مثلا از کارت مشتری پول کم شده
	transaction, err := services.PaymentTransaction(h.DB, req.TransactionID, req.CardYear, req.CardMonth, req.PurchaserCard)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}
	return ctx.JSON(http.StatusOK, BeginTransactionResponse(transaction))

}
func ValidateTransaction(transaction *PaymentTransactionRequest) error {

	requiredFields := map[string]string{
		"purchaser_card": transaction.PurchaserCard,
	}
	requiredFieldsInt := map[string]int{
		"card_month": transaction.CardMonth,
		"card_year":  transaction.CardYear,
	}
	requiredFieldsUint := map[string]uint{
		"transaction_id": transaction.TransactionID,
	}
	if err := utils.IsRequired(requiredFields); err != nil {
		return err
	}
	if err := utils.IsRequiredInt(requiredFieldsInt); err != nil {
		return err
	}
	if err := utils.IsRequiredUint(requiredFieldsUint); err != nil {
		return err
	}

	return nil
}

func (h *Handler) VerifyTransaction(ctx echo.Context) error {
	var transaction models.Transaction
	trackingCode := ctx.Param("tracking_code")
	transaction, err := services.GetSpecificTransaction(h.DB, trackingCode)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, "Transaction does not exist!")
	}
	return ctx.JSON(http.StatusOK, SetVerifyTransactionResponse(transaction))
}
func SetVerifyTransactionResponse(transaction models.Transaction) TransactionResponse {
	var status string
	if transaction.Status == models.NotPaid {
		status = "NotPaid"
	} else if transaction.Status == models.NotSuccessfully {
		status = "NotSuccessfully"
	} else if transaction.Status == models.IssueOccurred {
		status = "IssueOccurred"
	} else if transaction.Status == models.Blocked {
		status = "Blocked"
	} else if transaction.Status == models.Refund {
		status = "Refund"
	} else if transaction.Status == models.Cancelled {
		status = "Cancelled"
	} else if transaction.Status == models.ReturnToGateway {
		status = "ReturnToGateway"
	} else if transaction.Status == models.AwaitingConfirmation {
		status = "AwaitingConfirmation"
	} else if transaction.Status == models.Confirmed {
		status = "Confirmed"
	} else if transaction.Status == models.Paid {
		status = "Paid"
	}

	return TransactionResponse{
		TrackingCode:  transaction.TrackingCode,
		Status:        status,
		PurchaserCard: transaction.PurchaserCard,
		PaymentAmount: transaction.PaymentAmount,
		PhoneNumber:   transaction.PhoneNumber,
		PaymentDate:   transaction.CreatedAt.Format("2006-01-02 15:04"),
	}
}