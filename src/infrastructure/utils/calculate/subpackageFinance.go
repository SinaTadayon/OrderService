package calculate

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	"time"
)

type SubpackageFinance struct {
	SId    uint64
	Status string
	Items  []*ItemFinance
}

type ItemFinance struct {
	SKU         string
	InventoryId string
	Quantity    int32
	Invoice     ItemInvoiceFinance
}

type ItemInvoiceFinance struct {
	Unit       *decimal.Decimal
	Total      *decimal.Decimal
	Original   *decimal.Decimal
	Special    *decimal.Decimal
	Discount   *decimal.Decimal
	Commission *ItemCommissionFinance
	Share      *ItemShareFinance
	Voucher    *ItemVoucherFinance
	SSO        *ItemSSOFinance
	VAT        *ItemVATFinance
}

type ItemShareFinance struct {
	RawItemGross              *decimal.Decimal
	RoundupItemGross          *decimal.Decimal
	RawTotalGross             *decimal.Decimal
	RoundupTotalGross         *decimal.Decimal
	RawItemNet                *decimal.Decimal
	RoundupItemNet            *decimal.Decimal
	RawTotalNet               *decimal.Decimal
	RoundupTotalNet           *decimal.Decimal
	RawUnitBusinessShare      *decimal.Decimal
	RoundupUnitBusinessShare  *decimal.Decimal
	RawTotalBusinessShare     *decimal.Decimal
	RoundupTotalBusinessShare *decimal.Decimal
	RawUnitSellerShare        *decimal.Decimal
	RoundupUnitSellerShare    *decimal.Decimal
	RawTotalSellerShare       *decimal.Decimal
	RoundupTotalSellerShare   *decimal.Decimal
	CreatedAt                 *time.Time
	UpdatedAt                 *time.Time
}

type ItemCommissionFinance struct {
	ItemCommission    float32
	RawUnitPrice      *decimal.Decimal
	RoundupUnitPrice  *decimal.Decimal
	RawTotalPrice     *decimal.Decimal
	RoundupTotalPrice *decimal.Decimal
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
}

type ItemSSOFinance struct {
	RawUnitPrice      *decimal.Decimal
	RoundupUnitPrice  *decimal.Decimal
	RawTotalPrice     *decimal.Decimal
	RoundupTotalPrice *decimal.Decimal
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
}

type ItemVoucherFinance struct {
	RawUnitPrice      *decimal.Decimal
	RoundupUnitPrice  *decimal.Decimal
	RawTotalPrice     *decimal.Decimal
	RoundupTotalPrice *decimal.Decimal
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
}

type ItemVATFinance struct {
	SellerVat   *SellerVATFinance
	BusinessVat *BusinessVATFinance
}

type SellerVATFinance struct {
	Rate              float32
	IsObliged         bool
	RawUnitPrice      *decimal.Decimal
	RoundupUnitPrice  *decimal.Decimal
	RawTotalPrice     *decimal.Decimal
	RoundupTotalPrice *decimal.Decimal
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
}

type BusinessVATFinance struct {
	Rate              float32
	RawUnitPrice      *decimal.Decimal
	RoundupUnitPrice  *decimal.Decimal
	RawTotalPrice     *decimal.Decimal
	RoundupTotalPrice *decimal.Decimal
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
}

func FactoryFromSubPkg(ctx context.Context, subpackage *entities.Subpackage) (*SubpackageFinance, error) {

	financeItems := make([]*ItemFinance, 0, len(subpackage.Items))
	for i := 0; i < len(subpackage.Items); i++ {
		itemFinance := &ItemFinance{
			SKU:         subpackage.Items[i].SKU,
			InventoryId: subpackage.Items[i].InventoryId,
			Quantity:    subpackage.Items[i].Quantity,
			Invoice:     ItemInvoiceFinance{},
		}

		unit, err := decimal.NewFromString(subpackage.Items[i].Invoice.Unit.Amount)
		if err != nil {
			applog.GLog.Logger.FromContext(ctx).Error("Invoice.Unit.Amount invalid",
				"fn", "FactoryFromSubPkg",
				"unit", subpackage.Items[i].Invoice.Unit.Amount,
				"inventoryId", subpackage.Items[i].InventoryId,
				"sid", subpackage.SId,
				"pid", subpackage.PId,
				"oid", subpackage.OrderId,
				"error", err)
			return nil, err
		}
		itemFinance.Invoice.Unit = &unit

		total, err := decimal.NewFromString(subpackage.Items[i].Invoice.Total.Amount)
		if err != nil {
			applog.GLog.Logger.FromContext(ctx).Error("Invoice.Total.Amount invalid",
				"fn", "FactoryFromSubPkg",
				"total", subpackage.Items[i].Invoice.Total.Amount,
				"inventoryId", subpackage.Items[i].InventoryId,
				"sid", subpackage.SId,
				"pid", subpackage.PId,
				"oid", subpackage.OrderId,
				"error", err)
			return nil, err
		}
		itemFinance.Invoice.Total = &total

		original, err := decimal.NewFromString(subpackage.Items[i].Invoice.Original.Amount)
		if err != nil {
			applog.GLog.Logger.FromContext(ctx).Error("Invoice.Original.Amount invalid",
				"fn", "FactoryFromSubPkg",
				"original", subpackage.Items[i].Invoice.Original.Amount,
				"inventoryId", subpackage.Items[i].InventoryId,
				"sid", subpackage.SId,
				"pid", subpackage.PId,
				"oid", subpackage.OrderId,
				"error", err)
			return nil, err
		}
		itemFinance.Invoice.Original = &original

		special, err := decimal.NewFromString(subpackage.Items[i].Invoice.Special.Amount)
		if err != nil {
			applog.GLog.Logger.FromContext(ctx).Error("Invoice.Special.Amount invalid",
				"fn", "FactoryFromSubPkg",
				"special", subpackage.Items[i].Invoice.Special.Amount,
				"inventoryId", subpackage.Items[i].InventoryId,
				"sid", subpackage.SId,
				"pid", subpackage.PId,
				"oid", subpackage.OrderId,
				"error", err)
			return nil, err
		}
		itemFinance.Invoice.Special = &special

		discount, err := decimal.NewFromString(subpackage.Items[i].Invoice.Discount.Amount)
		if err != nil {
			applog.GLog.Logger.FromContext(ctx).Error("Invoice.Discount.Amount invalid",
				"fn", "FactoryFromSubPkg",
				"discount", subpackage.Items[i].Invoice.Discount.Amount,
				"inventoryId", subpackage.Items[i].InventoryId,
				"sid", subpackage.SId,
				"pid", subpackage.PId,
				"oid", subpackage.OrderId,
				"error", err)
			return nil, err
		}
		itemFinance.Invoice.Discount = &discount

		// Invoice ItemCommission
		if subpackage.Items[i].Invoice.Commission != nil {
			itemFinance.Invoice.Commission = &ItemCommissionFinance{}
			itemFinance.Invoice.Commission.ItemCommission = subpackage.Items[i].Invoice.Commission.ItemCommission
			if subpackage.Items[i].Invoice.Commission.RawUnitPrice != nil {
				rawUnitPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.Commission.RawUnitPrice.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.ItemCommission.RawUnitPrice.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawUnitPrice", subpackage.Items[i].Invoice.Commission.RawUnitPrice.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Commission.RawUnitPrice = &rawUnitPrice
			}

			if subpackage.Items[i].Invoice.Commission.RoundupUnitPrice != nil {
				roundupUnitPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.Commission.RoundupUnitPrice.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.ItemCommission.RoundupUnitPrice.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupUnitPrice", subpackage.Items[i].Invoice.Commission.RoundupUnitPrice.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Commission.RoundupUnitPrice = &roundupUnitPrice
			}

			if subpackage.Items[i].Invoice.Commission.RawTotalPrice != nil {
				rawTotalPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.Commission.RawTotalPrice.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.ItemCommission.RawTotalPrice.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawTotalPrice", subpackage.Items[i].Invoice.Commission.RawTotalPrice.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Commission.RawTotalPrice = &rawTotalPrice
			}

			if subpackage.Items[i].Invoice.Commission.RoundupTotalPrice != nil {
				roundupTotalPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.Commission.RoundupTotalPrice.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.ItemCommission.RoundupTotalPrice.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupTotalPrice", subpackage.Items[i].Invoice.Commission.RoundupTotalPrice.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Commission.RoundupTotalPrice = &roundupTotalPrice
			}

			itemFinance.Invoice.Commission.CreatedAt = subpackage.Items[i].Invoice.Commission.CreatedAt
			itemFinance.Invoice.Commission.UpdatedAt = subpackage.Items[i].Invoice.Commission.UpdatedAt
		}

		// Invoice Share
		if subpackage.Items[i].Invoice.Share != nil {
			itemFinance.Invoice.Share = &ItemShareFinance{}
			if subpackage.Items[i].Invoice.Share.RawItemGross != nil {
				rawItemGross, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RawItemGross.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RawItemGross.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawItemGross", subpackage.Items[i].Invoice.Share.RawItemGross.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RawItemGross = &rawItemGross
			}

			if subpackage.Items[i].Invoice.Share.RoundupItemGross != nil {
				roundupItemGross, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RoundupItemGross.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RoundupItemGross.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupItemGross", subpackage.Items[i].Invoice.Share.RoundupItemGross.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RoundupItemGross = &roundupItemGross
			}

			if subpackage.Items[i].Invoice.Share.RawTotalGross != nil {
				rawTotalGross, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RawTotalGross.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RawTotalGross.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawTotalGross", subpackage.Items[i].Invoice.Share.RawItemGross.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RawTotalGross = &rawTotalGross
			}

			if subpackage.Items[i].Invoice.Share.RoundupTotalGross != nil {
				roundupTotalGross, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RoundupTotalGross.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RoundupTotalGross.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupTotalGross", subpackage.Items[i].Invoice.Share.RoundupTotalGross.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RoundupTotalGross = &roundupTotalGross
			}

			if subpackage.Items[i].Invoice.Share.RawItemNet != nil {
				rawItemNet, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RawItemNet.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RawItemNet.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawItemNet", subpackage.Items[i].Invoice.Share.RawItemNet.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RawItemNet = &rawItemNet
			}

			if subpackage.Items[i].Invoice.Share.RoundupItemNet != nil {
				roundupItemNet, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RoundupItemNet.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RoundupItemNet.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupItemNet", subpackage.Items[i].Invoice.Share.RoundupItemNet.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RoundupItemNet = &roundupItemNet
			}

			if subpackage.Items[i].Invoice.Share.RawTotalNet != nil {
				rawTotalNet, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RawTotalNet.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RawTotalNet.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawTotalNet", subpackage.Items[i].Invoice.Share.RawTotalNet.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RawTotalNet = &rawTotalNet
			}

			if subpackage.Items[i].Invoice.Share.RoundupTotalNet != nil {
				roundupTotalNet, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RoundupTotalNet.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RoundupTotalNet.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupTotalNet", subpackage.Items[i].Invoice.Share.RoundupTotalNet.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RoundupTotalNet = &roundupTotalNet
			}

			if subpackage.Items[i].Invoice.Share.RawUnitBusinessShare != nil {
				rawUnitBusinessShare, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RawUnitBusinessShare.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RawUnitBusinessShare.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawUnitBusinessShare", subpackage.Items[i].Invoice.Share.RawUnitBusinessShare.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RawUnitBusinessShare = &rawUnitBusinessShare
			}

			if subpackage.Items[i].Invoice.Share.RoundupUnitBusinessShare != nil {
				roundupUnitBusinessShare, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RoundupUnitBusinessShare.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RoundupUnitBusinessShare.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupUnitBusinessShare", subpackage.Items[i].Invoice.Share.RoundupUnitBusinessShare.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RoundupUnitBusinessShare = &roundupUnitBusinessShare
			}

			if subpackage.Items[i].Invoice.Share.RawTotalBusinessShare != nil {
				rawTotalBusinessShare, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RawTotalBusinessShare.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RawTotalBusinessShare.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawTotalBusinessShare", subpackage.Items[i].Invoice.Share.RawTotalBusinessShare.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RawTotalBusinessShare = &rawTotalBusinessShare
			}

			if subpackage.Items[i].Invoice.Share.RoundupTotalBusinessShare != nil {
				roundupTotalBusinessShare, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RoundupTotalBusinessShare.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RoundupTotalBusinessShare.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupTotalBusinessShare", subpackage.Items[i].Invoice.Share.RoundupTotalBusinessShare.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RoundupTotalBusinessShare = &roundupTotalBusinessShare
			}

			if subpackage.Items[i].Invoice.Share.RawUnitSellerShare != nil {
				rawUnitSellerShare, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RawUnitSellerShare.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RawUnitSellerShare.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawUnitSellerShare", subpackage.Items[i].Invoice.Share.RawUnitSellerShare.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RawUnitSellerShare = &rawUnitSellerShare
			}

			if subpackage.Items[i].Invoice.Share.RoundupUnitSellerShare != nil {
				roundupUnitSellerShare, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RoundupUnitSellerShare.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RoundupUnitSellerShare.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupUnitSellerShare", subpackage.Items[i].Invoice.Share.RoundupUnitSellerShare.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RoundupUnitSellerShare = &roundupUnitSellerShare
			}

			if subpackage.Items[i].Invoice.Share.RawTotalSellerShare != nil {
				rawTotalSellerShare, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RawTotalSellerShare.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RawTotalSellerShare.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawTotalSellerShare", subpackage.Items[i].Invoice.Share.RawTotalSellerShare.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RawTotalSellerShare = &rawTotalSellerShare
			}

			if subpackage.Items[i].Invoice.Share.RoundupTotalSellerShare != nil {
				roundupTotalSellerShare, err := decimal.NewFromString(subpackage.Items[i].Invoice.Share.RoundupTotalSellerShare.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Share.RoundupTotalSellerShare.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupTotalSellerShare", subpackage.Items[i].Invoice.Share.RoundupTotalSellerShare.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Share.RoundupTotalSellerShare = &roundupTotalSellerShare
			}

			itemFinance.Invoice.Share.CreatedAt = subpackage.Items[i].Invoice.Share.CreatedAt
			itemFinance.Invoice.Share.UpdatedAt = subpackage.Items[i].Invoice.Share.UpdatedAt
		}

		// Invoice Voucher
		if subpackage.Items[i].Invoice.Voucher != nil {
			itemFinance.Invoice.Voucher = &ItemVoucherFinance{}
			if subpackage.Items[i].Invoice.Voucher.RawUnitPrice != nil {
				rawUnitPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.Voucher.RawUnitPrice.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Voucher.RawUnitPrice.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawUnitPrice", subpackage.Items[i].Invoice.Voucher.RawUnitPrice.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Voucher.RawUnitPrice = &rawUnitPrice
			}

			if subpackage.Items[i].Invoice.Voucher.RoundupUnitPrice != nil {
				roundupUnitPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.Voucher.RoundupUnitPrice.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Voucher.RoundupUnitPrice.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupUnitPrice", subpackage.Items[i].Invoice.Voucher.RoundupUnitPrice.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Voucher.RoundupUnitPrice = &roundupUnitPrice
			}

			if subpackage.Items[i].Invoice.Voucher.RawTotalPrice != nil {
				rawTotalPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.Voucher.RawTotalPrice.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Voucher.RawTotalPrice.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawTotalPrice", subpackage.Items[i].Invoice.Voucher.RawTotalPrice.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Voucher.RawTotalPrice = &rawTotalPrice
			}

			if subpackage.Items[i].Invoice.Voucher.RoundupTotalPrice != nil {
				roundupTotalPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.Voucher.RoundupTotalPrice.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.Voucher.RoundupTotalPrice.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupTotalPrice", subpackage.Items[i].Invoice.Voucher.RoundupTotalPrice.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.Voucher.RoundupTotalPrice = &roundupTotalPrice
			}

			itemFinance.Invoice.Voucher.CreatedAt = subpackage.Items[i].Invoice.Voucher.CreatedAt
			itemFinance.Invoice.Voucher.UpdatedAt = subpackage.Items[i].Invoice.Voucher.UpdatedAt
		}

		// Invoice SSO
		if subpackage.Items[i].Invoice.SSO != nil {
			itemFinance.Invoice.SSO = &ItemSSOFinance{}
			if subpackage.Items[i].Invoice.SSO.RawUnitPrice != nil {
				rawUnitPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.SSO.RawUnitPrice.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.SSO.RawUnitPrice.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawUnitPrice", subpackage.Items[i].Invoice.SSO.RawUnitPrice.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.SSO.RawUnitPrice = &rawUnitPrice
			}

			if subpackage.Items[i].Invoice.SSO.RoundupUnitPrice != nil {
				roundupUnitPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.SSO.RoundupUnitPrice.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.SSO.RoundupUnitPrice.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupUnitPrice", subpackage.Items[i].Invoice.SSO.RoundupUnitPrice.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.SSO.RoundupUnitPrice = &roundupUnitPrice
			}

			if subpackage.Items[i].Invoice.SSO.RawTotalPrice != nil {
				rawTotalPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.SSO.RawTotalPrice.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.SSO.RawTotalPrice.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"rawTotalPrice", subpackage.Items[i].Invoice.SSO.RawTotalPrice.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.SSO.RawTotalPrice = &rawTotalPrice
			}

			if subpackage.Items[i].Invoice.SSO.RoundupTotalPrice != nil {
				roundupTotalPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.SSO.RoundupTotalPrice.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("Invoice.SSO.RoundupTotalPrice.Amount invalid",
						"fn", "FactoryFromSubPkg",
						"roundupTotalPrice", subpackage.Items[i].Invoice.SSO.RoundupTotalPrice.Amount,
						"inventoryId", subpackage.Items[i].InventoryId,
						"sid", subpackage.SId,
						"pid", subpackage.PId,
						"oid", subpackage.OrderId,
						"error", err)
					return nil, err
				}
				itemFinance.Invoice.SSO.RoundupTotalPrice = &roundupTotalPrice
			}

			itemFinance.Invoice.SSO.CreatedAt = subpackage.Items[i].Invoice.SSO.CreatedAt
			itemFinance.Invoice.SSO.UpdatedAt = subpackage.Items[i].Invoice.SSO.UpdatedAt
		}

		// Invoice VAT
		if subpackage.Items[i].Invoice.VAT != nil {
			itemFinance.Invoice.VAT = &ItemVATFinance{}
			if subpackage.Items[i].Invoice.VAT.SellerVat != nil {
				itemFinance.Invoice.VAT.SellerVat = &SellerVATFinance{}

				itemFinance.Invoice.VAT.SellerVat.Rate = subpackage.Items[i].Invoice.VAT.SellerVat.Rate
				itemFinance.Invoice.VAT.SellerVat.IsObliged = subpackage.Items[i].Invoice.VAT.SellerVat.IsObliged

				if subpackage.Items[i].Invoice.VAT.SellerVat.RawUnitPrice != nil {
					rawUnitPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.VAT.SellerVat.RawUnitPrice.Amount)
					if err != nil {
						applog.GLog.Logger.FromContext(ctx).Error("Invoice.VAT.SellerVat.RawUnitPrice.Amount invalid",
							"fn", "FactoryFromSubPkg",
							"rawUnitPrice", subpackage.Items[i].Invoice.VAT.SellerVat.RawUnitPrice.Amount,
							"inventoryId", subpackage.Items[i].InventoryId,
							"sid", subpackage.SId,
							"pid", subpackage.PId,
							"oid", subpackage.OrderId,
							"error", err)
						return nil, err
					}
					itemFinance.Invoice.VAT.SellerVat.RawUnitPrice = &rawUnitPrice
				}

				if subpackage.Items[i].Invoice.VAT.SellerVat.RoundupUnitPrice != nil {
					roundupUnitPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.VAT.SellerVat.RoundupUnitPrice.Amount)
					if err != nil {
						applog.GLog.Logger.FromContext(ctx).Error("Invoice.VAT.SellerVat.RoundupUnitPrice.Amount invalid",
							"fn", "FactoryFromSubPkg",
							"roundupUnitPrice", subpackage.Items[i].Invoice.VAT.SellerVat.RoundupUnitPrice.Amount,
							"inventoryId", subpackage.Items[i].InventoryId,
							"sid", subpackage.SId,
							"pid", subpackage.PId,
							"oid", subpackage.OrderId,
							"error", err)
						return nil, err
					}
					itemFinance.Invoice.VAT.SellerVat.RoundupUnitPrice = &roundupUnitPrice
				}

				if subpackage.Items[i].Invoice.VAT.SellerVat.RawTotalPrice != nil {
					rawTotalPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.VAT.SellerVat.RawTotalPrice.Amount)
					if err != nil {
						applog.GLog.Logger.FromContext(ctx).Error("Invoice.VAT.SellerVat.RawTotalPrice.Amount invalid",
							"fn", "FactoryFromSubPkg",
							"rawTotalPrice", subpackage.Items[i].Invoice.VAT.SellerVat.RawTotalPrice.Amount,
							"inventoryId", subpackage.Items[i].InventoryId,
							"sid", subpackage.SId,
							"pid", subpackage.PId,
							"oid", subpackage.OrderId,
							"error", err)
						return nil, err
					}
					itemFinance.Invoice.VAT.SellerVat.RawTotalPrice = &rawTotalPrice
				}

				if subpackage.Items[i].Invoice.VAT.SellerVat.RoundupTotalPrice != nil {
					roundupTotalPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.VAT.SellerVat.RoundupTotalPrice.Amount)
					if err != nil {
						applog.GLog.Logger.FromContext(ctx).Error("Invoice.VAT.SellerVat.RoundupTotalPrice.Amount invalid",
							"fn", "FactoryFromSubPkg",
							"roundupTotalPrice", subpackage.Items[i].Invoice.VAT.SellerVat.RoundupTotalPrice.Amount,
							"inventoryId", subpackage.Items[i].InventoryId,
							"sid", subpackage.SId,
							"pid", subpackage.PId,
							"oid", subpackage.OrderId,
							"error", err)
						return nil, err
					}
					itemFinance.Invoice.VAT.SellerVat.RoundupTotalPrice = &roundupTotalPrice
				}

				itemFinance.Invoice.VAT.SellerVat.CreatedAt = subpackage.Items[i].Invoice.VAT.SellerVat.CreatedAt
				itemFinance.Invoice.VAT.SellerVat.UpdatedAt = subpackage.Items[i].Invoice.VAT.SellerVat.UpdatedAt
			}

			if subpackage.Items[i].Invoice.VAT.BusinessVat != nil {
				itemFinance.Invoice.VAT.BusinessVat = &BusinessVATFinance{}
				itemFinance.Invoice.VAT.BusinessVat.Rate = subpackage.Items[i].Invoice.VAT.BusinessVat.Rate

				if subpackage.Items[i].Invoice.VAT.BusinessVat.RawUnitPrice != nil {
					rawUnitPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.VAT.BusinessVat.RawUnitPrice.Amount)
					if err != nil {
						applog.GLog.Logger.FromContext(ctx).Error("Invoice.VAT.BusinessVat.RawUnitPrice.Amount invalid",
							"fn", "FactoryFromSubPkg",
							"rawUnitPrice", subpackage.Items[i].Invoice.VAT.BusinessVat.RawUnitPrice.Amount,
							"inventoryId", subpackage.Items[i].InventoryId,
							"sid", subpackage.SId,
							"pid", subpackage.PId,
							"oid", subpackage.OrderId,
							"error", err)
						return nil, err
					}
					itemFinance.Invoice.VAT.BusinessVat.RawUnitPrice = &rawUnitPrice
				}

				if subpackage.Items[i].Invoice.VAT.BusinessVat.RawTotalPrice != nil {
					rawTotalPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.VAT.BusinessVat.RawTotalPrice.Amount)
					if err != nil {
						applog.GLog.Logger.FromContext(ctx).Error("Invoice.VAT.BusinessVat.RawTotalPrice.Amount invalid",
							"fn", "FactoryFromSubPkg",
							"rawTotalPrice", subpackage.Items[i].Invoice.VAT.BusinessVat.RawTotalPrice.Amount,
							"inventoryId", subpackage.Items[i].InventoryId,
							"sid", subpackage.SId,
							"pid", subpackage.PId,
							"oid", subpackage.OrderId,
							"error", err)
						return nil, err
					}
					itemFinance.Invoice.VAT.BusinessVat.RawTotalPrice = &rawTotalPrice
				}

				if subpackage.Items[i].Invoice.VAT.BusinessVat.RoundupTotalPrice != nil {
					roundupTotalPrice, err := decimal.NewFromString(subpackage.Items[i].Invoice.VAT.BusinessVat.RoundupTotalPrice.Amount)
					if err != nil {
						applog.GLog.Logger.FromContext(ctx).Error("Invoice.VAT.BusinessVat.RoundupTotalPrice.Amount invalid",
							"fn", "FactoryFromSubPkg",
							"roundupTotalPrice", subpackage.Items[i].Invoice.VAT.BusinessVat.RoundupTotalPrice.Amount,
							"inventoryId", subpackage.Items[i].InventoryId,
							"sid", subpackage.SId,
							"pid", subpackage.PId,
							"oid", subpackage.OrderId,
							"error", err)
						return nil, err
					}
					itemFinance.Invoice.VAT.BusinessVat.RoundupTotalPrice = &roundupTotalPrice
				}

				itemFinance.Invoice.VAT.BusinessVat.CreatedAt = subpackage.Items[i].Invoice.VAT.BusinessVat.CreatedAt
				itemFinance.Invoice.VAT.BusinessVat.UpdatedAt = subpackage.Items[i].Invoice.VAT.BusinessVat.UpdatedAt
			}
		}

		financeItems = append(financeItems, itemFinance)
	}

	return &SubpackageFinance{
		SId:    subpackage.SId,
		Status: subpackage.Status,
		Items:  financeItems,
	}, nil
}

func ConvertToSubPkg(ctx context.Context, finance *SubpackageFinance, subpackage *entities.Subpackage) error {

	if finance.SId != subpackage.SId {
		applog.GLog.Logger.FromContext(ctx).Error("subpackage finance sid invalid",
			"fn", "ConvertToSubPkg",
			"finance sid", finance.SId,
			"sid", subpackage.SId,
			"pid", subpackage.PId,
			"oid", subpackage.OrderId)
		return errors.New("subpackage finance sid not equal with subpackage sid")
	}

	var findFlag = false
	for i := 0; i < len(finance.Items); i++ {
		findFlag = false
		for j := 0; j < len(subpackage.Items); j++ {
			if finance.Items[i].InventoryId == subpackage.Items[j].InventoryId {
				findFlag = true

				// Invoice.ItemCommission
				if finance.Items[i].Invoice.Commission != nil {
					if finance.Items[i].Invoice.Commission.RawUnitPrice != nil {
						subpackage.Items[j].Invoice.Commission.RawUnitPrice = &entities.Money{
							Amount:   finance.Items[i].Invoice.Commission.RawUnitPrice.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Commission.RoundupTotalPrice != nil {
						subpackage.Items[j].Invoice.Commission.RoundupUnitPrice = &entities.Money{
							Amount:   finance.Items[i].Invoice.Commission.RoundupTotalPrice.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Commission.RawTotalPrice != nil {
						subpackage.Items[j].Invoice.Commission.RawTotalPrice = &entities.Money{
							Amount:   finance.Items[i].Invoice.Commission.RawTotalPrice.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Commission.RoundupTotalPrice != nil {
						subpackage.Items[j].Invoice.Commission.RoundupTotalPrice = &entities.Money{
							Amount:   finance.Items[i].Invoice.Commission.RoundupTotalPrice.String(),
							Currency: "IRR",
						}
					}
					subpackage.Items[j].Invoice.Commission.CreatedAt = finance.Items[i].Invoice.Commission.CreatedAt
					subpackage.Items[j].Invoice.Commission.UpdatedAt = finance.Items[i].Invoice.Commission.UpdatedAt
				}

				// Invoice.Share
				if finance.Items[i].Invoice.Share != nil {
					if subpackage.Items[j].Invoice.Share == nil {
						subpackage.Items[j].Invoice.Share = &entities.ItemShare{}
					}

					if finance.Items[i].Invoice.Share.RawItemGross != nil {
						subpackage.Items[j].Invoice.Share.RawItemGross = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RawItemGross.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RoundupItemGross != nil {
						subpackage.Items[j].Invoice.Share.RoundupItemGross = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RoundupItemGross.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RawTotalGross != nil {
						subpackage.Items[j].Invoice.Share.RawTotalGross = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RawTotalGross.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RoundupTotalGross != nil {
						subpackage.Items[j].Invoice.Share.RoundupTotalGross = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RoundupTotalGross.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RawItemNet != nil {
						subpackage.Items[j].Invoice.Share.RawItemNet = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RawItemNet.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RoundupItemNet != nil {
						subpackage.Items[j].Invoice.Share.RoundupItemNet = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RoundupItemNet.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RawTotalNet != nil {
						subpackage.Items[j].Invoice.Share.RawTotalNet = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RawTotalNet.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RoundupTotalNet != nil {
						subpackage.Items[j].Invoice.Share.RoundupTotalNet = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RoundupTotalNet.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RawUnitBusinessShare != nil {
						subpackage.Items[j].Invoice.Share.RawUnitBusinessShare = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RawUnitBusinessShare.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RoundupUnitBusinessShare != nil {
						subpackage.Items[j].Invoice.Share.RoundupUnitBusinessShare = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RoundupUnitBusinessShare.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RawTotalBusinessShare != nil {
						subpackage.Items[j].Invoice.Share.RawTotalBusinessShare = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RawTotalBusinessShare.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RoundupTotalBusinessShare != nil {
						subpackage.Items[j].Invoice.Share.RoundupTotalBusinessShare = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RoundupTotalBusinessShare.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RawUnitSellerShare != nil {
						subpackage.Items[j].Invoice.Share.RawUnitSellerShare = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RawUnitSellerShare.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RoundupUnitSellerShare != nil {
						subpackage.Items[j].Invoice.Share.RoundupUnitSellerShare = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RoundupUnitSellerShare.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RawTotalSellerShare != nil {
						subpackage.Items[j].Invoice.Share.RawTotalSellerShare = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RawTotalSellerShare.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Share.RoundupTotalSellerShare != nil {
						subpackage.Items[j].Invoice.Share.RoundupTotalSellerShare = &entities.Money{
							Amount:   finance.Items[i].Invoice.Share.RoundupTotalSellerShare.String(),
							Currency: "IRR",
						}
					}

					subpackage.Items[j].Invoice.Share.CreatedAt = finance.Items[i].Invoice.Share.CreatedAt
					subpackage.Items[j].Invoice.Share.UpdatedAt = finance.Items[i].Invoice.Share.UpdatedAt
				}

				// Invoice.Voucher
				if finance.Items[i].Invoice.Voucher != nil {
					if subpackage.Items[j].Invoice.Voucher == nil {
						subpackage.Items[j].Invoice.Voucher = &entities.ItemVoucher{}
					}

					if finance.Items[i].Invoice.Voucher.RawUnitPrice != nil {
						subpackage.Items[j].Invoice.Voucher.RawUnitPrice = &entities.Money{
							Amount:   finance.Items[i].Invoice.Voucher.RawUnitPrice.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Voucher.RoundupUnitPrice != nil {
						subpackage.Items[j].Invoice.Voucher.RoundupUnitPrice = &entities.Money{
							Amount:   finance.Items[i].Invoice.Voucher.RoundupUnitPrice.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Voucher.RawTotalPrice != nil {
						subpackage.Items[j].Invoice.Voucher.RawTotalPrice = &entities.Money{
							Amount:   finance.Items[i].Invoice.Voucher.RawTotalPrice.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.Voucher.RoundupTotalPrice != nil {
						subpackage.Items[j].Invoice.Voucher.RoundupTotalPrice = &entities.Money{
							Amount:   finance.Items[i].Invoice.Voucher.RoundupTotalPrice.String(),
							Currency: "IRR",
						}
					}

					subpackage.Items[j].Invoice.Voucher.CreatedAt = finance.Items[i].Invoice.Voucher.CreatedAt
					subpackage.Items[j].Invoice.Voucher.UpdatedAt = finance.Items[i].Invoice.Voucher.UpdatedAt
				}

				// Invoice.SSO
				if finance.Items[i].Invoice.SSO != nil {
					if subpackage.Items[j].Invoice.SSO == nil {
						subpackage.Items[j].Invoice.SSO = &entities.ItemSSO{}
					}

					if finance.Items[i].Invoice.SSO.RawUnitPrice != nil {
						subpackage.Items[j].Invoice.SSO.RawUnitPrice = &entities.Money{
							Amount:   finance.Items[i].Invoice.SSO.RawUnitPrice.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.SSO.RoundupUnitPrice != nil {
						subpackage.Items[j].Invoice.SSO.RoundupUnitPrice = &entities.Money{
							Amount:   finance.Items[i].Invoice.SSO.RoundupUnitPrice.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.SSO.RawTotalPrice != nil {
						subpackage.Items[j].Invoice.SSO.RawTotalPrice = &entities.Money{
							Amount:   finance.Items[i].Invoice.SSO.RawTotalPrice.String(),
							Currency: "IRR",
						}
					}

					if finance.Items[i].Invoice.SSO.RoundupTotalPrice != nil {
						subpackage.Items[j].Invoice.SSO.RoundupTotalPrice = &entities.Money{
							Amount:   finance.Items[i].Invoice.SSO.RoundupTotalPrice.String(),
							Currency: "IRR",
						}
					}

					subpackage.Items[j].Invoice.SSO.CreatedAt = finance.Items[i].Invoice.SSO.CreatedAt
					subpackage.Items[j].Invoice.SSO.UpdatedAt = finance.Items[i].Invoice.SSO.UpdatedAt
				}

				// Invoice.VAT
				if finance.Items[i].Invoice.VAT != nil {
					if subpackage.Items[j].Invoice.VAT == nil {
						subpackage.Items[j].Invoice.VAT = &entities.ItemVAT{}
					}

					if finance.Items[i].Invoice.VAT.SellerVat != nil {
						if subpackage.Items[j].Invoice.VAT.SellerVat == nil {
							subpackage.Items[j].Invoice.VAT.SellerVat = &entities.SellerVAT{}
						}

						if finance.Items[i].Invoice.VAT.SellerVat.RawUnitPrice != nil {
							subpackage.Items[j].Invoice.VAT.SellerVat.RawUnitPrice = &entities.Money{
								Amount:   finance.Items[i].Invoice.VAT.SellerVat.RawUnitPrice.String(),
								Currency: "IRR",
							}
						}

						if finance.Items[i].Invoice.VAT.SellerVat.RoundupUnitPrice != nil {
							subpackage.Items[j].Invoice.VAT.SellerVat.RoundupUnitPrice = &entities.Money{
								Amount:   finance.Items[i].Invoice.VAT.SellerVat.RoundupUnitPrice.String(),
								Currency: "IRR",
							}
						}

						if finance.Items[i].Invoice.VAT.SellerVat.RawTotalPrice != nil {
							subpackage.Items[j].Invoice.VAT.SellerVat.RawTotalPrice = &entities.Money{
								Amount:   finance.Items[i].Invoice.VAT.SellerVat.RawTotalPrice.String(),
								Currency: "IRR",
							}
						}

						if finance.Items[i].Invoice.VAT.SellerVat.RoundupTotalPrice != nil {
							subpackage.Items[j].Invoice.VAT.SellerVat.RoundupTotalPrice = &entities.Money{
								Amount:   finance.Items[i].Invoice.VAT.SellerVat.RoundupTotalPrice.String(),
								Currency: "IRR",
							}
						}

						subpackage.Items[j].Invoice.VAT.SellerVat.CreatedAt = finance.Items[i].Invoice.VAT.SellerVat.CreatedAt
						subpackage.Items[j].Invoice.VAT.SellerVat.UpdatedAt = finance.Items[i].Invoice.VAT.SellerVat.UpdatedAt
					}

					if finance.Items[i].Invoice.VAT.BusinessVat != nil {
						if subpackage.Items[j].Invoice.VAT.BusinessVat == nil {
							subpackage.Items[j].Invoice.VAT.BusinessVat = &entities.BusinessVAT{}
						}

						if finance.Items[i].Invoice.VAT.BusinessVat.RawUnitPrice != nil {
							subpackage.Items[j].Invoice.VAT.BusinessVat.RawUnitPrice = &entities.Money{
								Amount:   finance.Items[i].Invoice.VAT.BusinessVat.RawUnitPrice.String(),
								Currency: "IRR",
							}
						}

						if finance.Items[i].Invoice.VAT.BusinessVat.RoundupUnitPrice != nil {
							subpackage.Items[j].Invoice.VAT.BusinessVat.RoundupUnitPrice = &entities.Money{
								Amount:   finance.Items[i].Invoice.VAT.BusinessVat.RoundupUnitPrice.String(),
								Currency: "IRR",
							}
						}

						if finance.Items[i].Invoice.VAT.BusinessVat.RawTotalPrice != nil {
							subpackage.Items[j].Invoice.VAT.BusinessVat.RawTotalPrice = &entities.Money{
								Amount:   finance.Items[i].Invoice.VAT.BusinessVat.RawTotalPrice.String(),
								Currency: "IRR",
							}
						}

						if finance.Items[i].Invoice.VAT.BusinessVat.RoundupTotalPrice != nil {
							subpackage.Items[j].Invoice.VAT.BusinessVat.RoundupTotalPrice = &entities.Money{
								Amount:   finance.Items[i].Invoice.VAT.BusinessVat.RoundupTotalPrice.String(),
								Currency: "IRR",
							}
						}

						subpackage.Items[j].Invoice.VAT.BusinessVat.CreatedAt = finance.Items[i].Invoice.VAT.BusinessVat.CreatedAt
						subpackage.Items[j].Invoice.VAT.BusinessVat.UpdatedAt = finance.Items[i].Invoice.VAT.BusinessVat.UpdatedAt
					}
				}
			}
		}

		if !findFlag {
			applog.GLog.Logger.FromContext(ctx).Error("subpackage finance item invalid",
				"fn", "ConvertToSubPkg",
				"finance inventoryId", finance.Items[i].InventoryId,
				"sid", subpackage.SId,
				"pid", subpackage.PId,
				"oid", subpackage.OrderId)
			return errors.New("subpackage finance sid not equal with subpackage sid")
		}
	}

	return nil
}
