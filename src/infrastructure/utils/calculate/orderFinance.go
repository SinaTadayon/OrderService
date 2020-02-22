package calculate

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	"time"
)

type OrderFinance struct {
	OrderId  uint64
	Status   string
	Invoice  InvoiceFinance
	Packages []*PackageItemFinance
}

type InvoiceFinance struct {
	GrandTotal    *decimal.Decimal
	Subtotal      *decimal.Decimal
	Discount      *decimal.Decimal
	ShipmentTotal *decimal.Decimal
	Share         *ShareFinance
	Commission    *CommissionFinance
	Voucher       *VoucherFinance
	SSO           *SSOFinance
	VAT           *VATFinance
}

type CommissionFinance struct {
	RawTotalPrice     *decimal.Decimal
	RoundupTotalPrice *decimal.Decimal
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
}

type ShareFinance struct {
	RawTotalShare     *decimal.Decimal
	RoundupTotalShare *decimal.Decimal
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
}

type VoucherFinance struct {
	Percent                     float64
	AppliedPrice                *decimal.Decimal
	RoundupAppliedPrice         *decimal.Decimal
	RawShipmentAppliedPrice     *decimal.Decimal
	RoundupShipmentAppliedPrice *decimal.Decimal
	Price                       *decimal.Decimal
	//Code                        string
	//Details                     *VoucherDetailsFinance
}

type SSOFinance struct {
	RawTotal     *decimal.Decimal
	RoundupTotal *decimal.Decimal
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

type VATFinance struct {
	Rate         float32
	RawTotal     *decimal.Decimal
	RoundupTotal *decimal.Decimal
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

//type VoucherDetailsFinance struct {
//	Title            string
//	Prefix           string
//	UseLimit         int32
//	Count            int32
//	Length           int32
//	Categories       []string
//	Products         []string
//	Users            []string
//	Sellers          []string
//	IsFirstPurchase  bool
//	StartDate        time.Time
//	EndDate          time.Time
//	Type             string
//	MaxDiscountValue uint64
//	MinBasketValue   uint64
//}

func FactoryFromOrder(ctx context.Context, order *entities.Order) (*OrderFinance, error) {
	financeInvoice := InvoiceFinance{}
	grandTotal, err := decimal.NewFromString(order.Invoice.GrandTotal.Amount)
	if err != nil {
		applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.GrandTotal.Amount invalid",
			"fn", "FactoryFromOrder",
			"grandTotal", order.Invoice.GrandTotal.Amount,
			"oid", order.OrderId,
			"error", err)
		return nil, err
	}
	financeInvoice.GrandTotal = &grandTotal

	subtotal, err := decimal.NewFromString(order.Invoice.Subtotal.Amount)
	if err != nil {
		applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.Subtotal.Amount invalid",
			"fn", "FactoryFromOrder",
			"subtotal", order.Invoice.Subtotal.Amount,
			"oid", order.OrderId,
			"error", err)
		return nil, err
	}
	financeInvoice.Subtotal = &subtotal

	discount, err := decimal.NewFromString(order.Invoice.Discount.Amount)
	if err != nil {
		applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.Discount.Amount invalid",
			"fn", "FactoryFromOrder",
			"discount", order.Invoice.Discount.Amount,
			"oid", order.OrderId,
			"error", err)
		return nil, err
	}
	financeInvoice.Discount = &discount

	shipmentTotal, err := decimal.NewFromString(order.Invoice.ShipmentTotal.Amount)
	if err != nil {
		applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.ShipmentTotal.Amount invalid",
			"fn", "FactoryFromOrder",
			"shipmentTotal", order.Invoice.ShipmentTotal.Amount,
			"oid", order.OrderId,
			"error", err)
		return nil, err
	}
	financeInvoice.ShipmentTotal = &shipmentTotal

	// invoice share
	if order.Invoice.Share != nil {
		financeInvoice.Share = &ShareFinance{}

		if order.Invoice.Share.RawTotalShare != nil {
			rawTotalShare, err := decimal.NewFromString(order.Invoice.Share.RawTotalShare.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.Share.RawTotalShare.Amount invalid",
					"fn", "FactoryFromOrder",
					"rawTotalShare", order.Invoice.Share.RawTotalShare.Amount,
					"oid", order.OrderId,
					"error", err)
				return nil, err
			}
			financeInvoice.Share.RawTotalShare = &rawTotalShare
		}

		if order.Invoice.Share.RoundupTotalShare != nil {
			roundupTotalShare, err := decimal.NewFromString(order.Invoice.Share.RoundupTotalShare.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.Share.RoundupTotalShare.Amount invalid",
					"fn", "FactoryFromOrder",
					"roundupTotalShare", order.Invoice.Share.RoundupTotalShare.Amount,
					"oid", order.OrderId,
					"error", err)
				return nil, err
			}
			financeInvoice.Share.RoundupTotalShare = &roundupTotalShare
		}

		financeInvoice.Share.CreatedAt = order.Invoice.Share.CreatedAt
		financeInvoice.Share.UpdatedAt = order.Invoice.Share.UpdatedAt
	}

	// invoice share
	if order.Invoice.Voucher != nil {
		financeInvoice.Voucher = &VoucherFinance{}
		financeInvoice.Voucher.Percent = order.Invoice.Voucher.Percent

		if order.Invoice.Voucher.AppliedPrice != nil {
			appliedPrice, err := decimal.NewFromString(order.Invoice.Voucher.AppliedPrice.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.Voucher.AppliedPrice.Amount invalid",
					"fn", "FactoryFromOrder",
					"appliedPrice", order.Invoice.Voucher.AppliedPrice.Amount,
					"oid", order.OrderId,
					"error", err)
				return nil, err
			}
			financeInvoice.Voucher.AppliedPrice = &appliedPrice
		}

		if order.Invoice.Voucher.RoundupAppliedPrice != nil {
			roundupAppliedPrice, err := decimal.NewFromString(order.Invoice.Voucher.RoundupAppliedPrice.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.Voucher.RoundupAppliedPrice.Amount invalid",
					"fn", "FactoryFromOrder",
					"roundupAppliedPrice", order.Invoice.Voucher.RoundupAppliedPrice.Amount,
					"oid", order.OrderId,
					"error", err)
				return nil, err
			}
			financeInvoice.Voucher.RoundupAppliedPrice = &roundupAppliedPrice
		}

		if order.Invoice.Voucher.RawShipmentAppliedPrice != nil {
			RawShipmentAppliedPrice, err := decimal.NewFromString(order.Invoice.Voucher.RawShipmentAppliedPrice.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.Voucher.RawShipmentAppliedPrice.Amount invalid",
					"fn", "FactoryFromOrder",
					"rawShipmentAppliedPrice", order.Invoice.Voucher.RawShipmentAppliedPrice.Amount,
					"oid", order.OrderId,
					"error", err)
				return nil, err
			}
			financeInvoice.Voucher.RawShipmentAppliedPrice = &RawShipmentAppliedPrice
		}

		if order.Invoice.Voucher.RoundupShipmentAppliedPrice != nil {
			roundupShipmentAppliedPrice, err := decimal.NewFromString(order.Invoice.Voucher.RoundupShipmentAppliedPrice.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.Voucher.RoundupShipmentAppliedPrice.Amount invalid",
					"fn", "FactoryFromOrder",
					"roundupShipmentAppliedPrice", order.Invoice.Voucher.RoundupShipmentAppliedPrice.Amount,
					"oid", order.OrderId,
					"error", err)
				return nil, err
			}
			financeInvoice.Voucher.RoundupShipmentAppliedPrice = &roundupShipmentAppliedPrice
		}

		if order.Invoice.Voucher.Price != nil {
			price, err := decimal.NewFromString(order.Invoice.Voucher.Price.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.Voucher.Price.Amount invalid",
					"fn", "FactoryFromOrder",
					"price", order.Invoice.Voucher.Price.Amount,
					"oid", order.OrderId,
					"error", err)
				return nil, err
			}
			financeInvoice.Voucher.Price = &price
		}
	}

	// invoice sso
	if order.Invoice.SSO != nil {
		financeInvoice.SSO = &SSOFinance{}

		if order.Invoice.SSO.RawTotal != nil {
			rawTotal, err := decimal.NewFromString(order.Invoice.SSO.RawTotal.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.SSO.RawTotal.Amount invalid",
					"fn", "FactoryFromOrder",
					"rawTotal", order.Invoice.SSO.RawTotal.Amount,
					"oid", order.OrderId,
					"error", err)
				return nil, err
			}
			financeInvoice.SSO.RawTotal = &rawTotal
		}

		if order.Invoice.SSO.RoundupTotal != nil {
			roundupTotal, err := decimal.NewFromString(order.Invoice.SSO.RoundupTotal.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.SSO.RoundupTotal.Amount invalid",
					"fn", "FactoryFromOrder",
					"roundupTotal", order.Invoice.SSO.RoundupTotal.Amount,
					"oid", order.OrderId,
					"error", err)
				return nil, err
			}
			financeInvoice.SSO.RoundupTotal = &roundupTotal
		}

		financeInvoice.SSO.CreatedAt = order.Invoice.SSO.CreatedAt
		financeInvoice.SSO.UpdatedAt = order.Invoice.SSO.UpdatedAt
	}

	// invoice vat
	if order.Invoice.VAT != nil {
		financeInvoice.VAT = &VATFinance{}
		financeInvoice.VAT.Rate = order.Invoice.VAT.Rate

		if order.Invoice.VAT.RawTotal != nil {
			rawTotal, err := decimal.NewFromString(order.Invoice.VAT.RawTotal.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.VAT.RawTotal.Amount invalid",
					"fn", "FactoryFromOrder",
					"rawTotal", order.Invoice.VAT.RawTotal.Amount,
					"oid", order.OrderId,
					"error", err)
				return nil, err
			}
			financeInvoice.VAT.RawTotal = &rawTotal
		}

		if order.Invoice.VAT.RoundupTotal != nil {
			roundupTotal, err := decimal.NewFromString(order.Invoice.VAT.RoundupTotal.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.VAT.RoundupTotal.Amount invalid",
					"fn", "FactoryFromOrder",
					"roundupTotal", order.Invoice.VAT.RoundupTotal.Amount,
					"oid", order.OrderId,
					"error", err)
				return nil, err
			}
			financeInvoice.VAT.RoundupTotal = &roundupTotal
		}

		financeInvoice.VAT.CreatedAt = order.Invoice.VAT.CreatedAt
		financeInvoice.VAT.UpdatedAt = order.Invoice.VAT.UpdatedAt
	}

	// invoice commission
	if order.Invoice.Commission != nil {
		financeInvoice.Commission = &CommissionFinance{}

		if order.Invoice.Commission.RawTotalPrice != nil {
			rawTotalPrice, err := decimal.NewFromString(order.Invoice.Commission.RawTotalPrice.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.Commission.RawTotalPrice.Amount invalid",
					"fn", "FactoryFromPkg",
					"rawTotalPrice", order.Invoice.Commission.RawTotalPrice.Amount,
					"oid", order.OrderId,
					"error", err)
				return nil, err
			}
			financeInvoice.Commission.RawTotalPrice = &rawTotalPrice
		}

		if order.Invoice.Commission.RoundupTotalPrice != nil {
			roundupTotalPrice, err := decimal.NewFromString(order.Invoice.Commission.RoundupTotalPrice.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("order.Invoice.Commission.RoundupTotalPrice.Amount invalid",
					"fn", "FactoryFromPkg",
					"roundupTotalPrice", order.Invoice.Commission.RoundupTotalPrice.Amount,
					"oid", order.OrderId,
					"error", err)
				return nil, err
			}
			financeInvoice.Commission.RoundupTotalPrice = &roundupTotalPrice
		}

		financeInvoice.Commission.CreatedAt = order.Invoice.Commission.CreatedAt
		financeInvoice.Commission.UpdatedAt = order.Invoice.Commission.UpdatedAt
	}

	financePackages := make([]*PackageItemFinance, 0, len(order.Packages))
	for i := 0; i < len(order.Packages); i++ {
		financePkg, err := FactoryFromPkg(ctx, order.Packages[i])
		if err != nil {
			return nil, err
		}

		financePackages = append(financePackages, financePkg)
	}

	return &OrderFinance{
		OrderId:  order.OrderId,
		Status:   order.Status,
		Invoice:  financeInvoice,
		Packages: financePackages,
	}, nil
}

func ConvertToOrder(ctx context.Context, finance *OrderFinance, order *entities.Order) error {
	if finance.OrderId != order.OrderId {
		applog.GLog.Logger.FromContext(ctx).Error("order finance oid invalid",
			"fn", "ConvertToOrder",
			"order finance orderId", finance.OrderId,
			"oid", order.OrderId)
		return errors.New("orderId not equal with finance OrderId")
	}

	// Invoice.Share
	if finance.Invoice.Share != nil {
		if order.Invoice.Share == nil {
			order.Invoice.Share = &entities.OrderShare{}
		}

		if finance.Invoice.Share.RawTotalShare != nil {
			order.Invoice.Share.RawTotalShare = &entities.Money{
				Amount:   finance.Invoice.Share.RawTotalShare.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.Share.RoundupTotalShare != nil {
			order.Invoice.Share.RoundupTotalShare = &entities.Money{
				Amount:   finance.Invoice.Share.RoundupTotalShare.String(),
				Currency: "IRR",
			}
		}

		order.Invoice.Share.CreatedAt = finance.Invoice.Share.CreatedAt
		order.Invoice.Share.UpdatedAt = finance.Invoice.Share.UpdatedAt
	}

	// Invoice.Voucher
	if finance.Invoice.Voucher != nil {
		if finance.Invoice.Voucher.RawShipmentAppliedPrice != nil {
			order.Invoice.Voucher.RawShipmentAppliedPrice = &entities.Money{
				Amount:   finance.Invoice.Voucher.RawShipmentAppliedPrice.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.Voucher.RoundupShipmentAppliedPrice != nil {
			order.Invoice.Voucher.RoundupShipmentAppliedPrice = &entities.Money{
				Amount:   finance.Invoice.Voucher.RoundupShipmentAppliedPrice.String(),
				Currency: "IRR",
			}
		}
	}

	// Invoice.Voucher
	if finance.Invoice.SSO != nil {
		if order.Invoice.SSO == nil {
			order.Invoice.SSO = &entities.SSO{}
		}

		if finance.Invoice.SSO.RawTotal != nil {
			order.Invoice.SSO.RawTotal = &entities.Money{
				Amount:   finance.Invoice.SSO.RawTotal.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.SSO.RoundupTotal != nil {
			order.Invoice.SSO.RoundupTotal = &entities.Money{
				Amount:   finance.Invoice.SSO.RoundupTotal.String(),
				Currency: "IRR",
			}
		}

		order.Invoice.SSO.CreatedAt = finance.Invoice.SSO.CreatedAt
		order.Invoice.SSO.UpdatedAt = finance.Invoice.SSO.UpdatedAt
	}

	// invoice vat
	if finance.Invoice.VAT != nil {
		if finance.Invoice.VAT.RawTotal != nil {
			order.Invoice.VAT.RawTotal = &entities.Money{
				Amount:   finance.Invoice.VAT.RawTotal.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.VAT.RoundupTotal != nil {
			order.Invoice.VAT.RoundupTotal = &entities.Money{
				Amount:   finance.Invoice.VAT.RoundupTotal.String(),
				Currency: "IRR",
			}
		}

		order.Invoice.VAT.CreatedAt = finance.Invoice.VAT.CreatedAt
		order.Invoice.VAT.UpdatedAt = finance.Invoice.VAT.UpdatedAt
	}

	// invoice commission
	if finance.Invoice.Commission != nil {
		if order.Invoice.Commission == nil {
			order.Invoice.Commission = &entities.Commission{}
		}

		if finance.Invoice.Commission.RawTotalPrice != nil {
			order.Invoice.Commission.RawTotalPrice = &entities.Money{
				Amount:   finance.Invoice.Commission.RawTotalPrice.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.Commission.RoundupTotalPrice != nil {
			order.Invoice.Commission.RoundupTotalPrice = &entities.Money{
				Amount:   finance.Invoice.Commission.RoundupTotalPrice.String(),
				Currency: "IRR",
			}
		}

		order.Invoice.Commission.CreatedAt = finance.Invoice.Commission.CreatedAt
		order.Invoice.Commission.UpdatedAt = finance.Invoice.Commission.UpdatedAt
	}

	for i := 0; i < len(finance.Packages); i++ {
		err := ConvertToPkg(ctx, finance.Packages[i], order.Packages[i])
		if err != nil {
			return err
		}
	}

	return nil
}
