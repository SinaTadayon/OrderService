package calculate

import (
	"context"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
)

type FinanceCalcType uint64
type FinanceMode string

const (
	ORDER_FINANCE  FinanceMode = "ORDER_FINANCE"
	SELLER_FINANCE FinanceMode = "SELLER_FINANCE"
	BUYER_FINANCE  FinanceMode = "BUYER_FINANCE"
)

const (
	VOUCHER_CALC FinanceCalcType = 1 << iota
	SELLER_VAT_CALC
	NET_COMMISSION_CALC
	BUSINESS_VAT_CALC
	SELLER_SSO_CALC
	SHARE_CALC
)

func Set(b, flag FinanceCalcType) FinanceCalcType    { return b | flag }
func Clear(b, flag FinanceCalcType) FinanceCalcType  { return b &^ flag }
func Toggle(b, flag FinanceCalcType) FinanceCalcType { return b ^ flag }
func Has(b, flag FinanceCalcType) bool               { return b&flag != 0 }

type FinanceCalculator interface {
	FinanceCalc(ctx context.Context, order *entities.Order, fct FinanceCalcType, mode FinanceMode) error
}
