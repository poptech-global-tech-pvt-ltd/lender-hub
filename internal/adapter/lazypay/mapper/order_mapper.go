package mapper

import (
	"fmt"

	orderReq "lending-hub-service/internal/domain/order/dto/request"
	orderResp "lending-hub-service/internal/domain/order/dto/response"
	lpCommon "lending-hub-service/internal/adapter/lazypay/dto/common"
	lpReq "lending-hub-service/internal/adapter/lazypay/dto/request"
	lpResp "lending-hub-service/internal/adapter/lazypay/dto/response"
)

// ToLPCreateOrderRequest converts canonical CreateOrderRequest → LP format
func ToLPCreateOrderRequest(
	req orderReq.CreateOrderRequest,
	accessKey, merchantID, signature string,
) *lpReq.LPCreateOrderRequest {
	lpRequest := &lpReq.LPCreateOrderRequest{
		AccessKey:     accessKey,
		MerchantID:    merchantID,
		MerchantTxnID: req.PaymentID,
		User: lpCommon.LPUserDetails{
			Mobile: req.Mobile,
			Email:  "", // Not in canonical request
		},
		Amount: lpCommon.LPAmount{
			Value:    fmt.Sprintf("%.2f", req.Amount),
			Currency: req.Currency,
		},
		ReturnURL:  req.ReturnURL,
		Signature:  signature,
		EMITenure:  req.EmiSelection.Tenure,
	}

	// Map address if provided
	if req.Address != nil && req.Address.Street1 != "" {
		lpRequest.Address = &lpCommon.LPAddress{
			Street1: req.Address.Street1,
			City:    req.Address.City,
			State:   req.Address.State,
			Zip:     req.Address.Zip,
		}
	}

	// Map product lines if provided
	if len(req.ProductLines) > 0 {
		productLines := make([]lpReq.LPProductLine, len(req.ProductLines))
		for i, pl := range req.ProductLines {
			productLines[i] = lpReq.LPProductLine{
				Name:     pl.Name,
				SKU:      pl.SKU,
				Quantity: pl.Quantity,
				Price:    fmt.Sprintf("%.2f", pl.Price),
			}
		}
		lpRequest.ProductLines = productLines
	}

	return lpRequest
}

// FromLPOrderResponse converts LP response → canonical OrderResponse
func FromLPOrderResponse(
	lp *lpResp.LPOrderResponse,
	paymentID string,
) *orderResp.OrderResponse {
	lenderOrderID := lp.OrderID
	redirectURL := lp.RedirectURL
	return &orderResp.OrderResponse{
		PaymentID:     paymentID,
		Status:        lp.Status,
		LenderOrderID: &lenderOrderID,
		RedirectURL:   &redirectURL,
	}
}
