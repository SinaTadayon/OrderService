package calculate

import (
	"context"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	applog "gitlab.faza.io/order-project/order-service/infrastructure/logger"
	"time"
)

type PackageItemFinance struct {
	PId         uint64
	Status      string
	Invoice     PackageInvoiceFinance
	Subpackages []*SubpackageFinance
}

type PackageInvoiceFinance struct {
	Subtotal       *decimal.Decimal
	Discount       *decimal.Decimal
	ShipmentAmount *decimal.Decimal
	Share          *PackageShareFinance
	Commission     *PackageCommissionFinance
	Voucher        *PackageVoucherFinance
	SSO            *PackageSSOFinance
	VAT            *PackageVATFinance
}

type PackageCommissionFinance struct {
	RawTotalPrice     *decimal.Decimal
	RoundupTotalPrice *decimal.Decimal
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
}

type PackageShareFinance struct {
	RawBusinessShare         *decimal.Decimal
	RoundupBusinessShare     *decimal.Decimal
	RawSellerShare           *decimal.Decimal
	RoundupSellerShare       *decimal.Decimal
	RawSellerShippingNet     *decimal.Decimal
	RoundupSellerShippingNet *decimal.Decimal
	CreatedAt                *time.Time
	UpdatedAt                *time.Time
}

type PackageVoucherFinance struct {
	RawTotal                 *decimal.Decimal
	RoundupTotal             *decimal.Decimal
	RawCalcShipmentPrice     *decimal.Decimal
	RoundupCalcShipmentPrice *decimal.Decimal
	CreatedAt                *time.Time
	UpdatedAt                *time.Time
}

type PackageSSOFinance struct {
	Rate         float32
	IsObliged    bool
	RawTotal     *decimal.Decimal
	RoundupTotal *decimal.Decimal
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

type PackageVATFinance struct {
	SellerVAT   *PackageSellerVATFinance
	BusinessVAT *PackageBusinessVATFinance
}

type PackageSellerVATFinance struct {
	RawTotal     *decimal.Decimal
	RoundupTotal *decimal.Decimal
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

type PackageBusinessVATFinance struct {
	RawTotal     *decimal.Decimal
	RoundupTotal *decimal.Decimal
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
}

func FactoryFromPkg(ctx context.Context, pkg *entities.PackageItem) (*PackageItemFinance, error) {

	pkgInvoiceFinance := PackageInvoiceFinance{}
	subtotal, err := decimal.NewFromString(pkg.Invoice.Subtotal.Amount)
	if err != nil {
		applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Subtotal.Amount invalid",
			"fn", "FactoryFromPkg",
			"subtotal", pkg.Invoice.Subtotal.Amount,
			"pid", pkg.PId,
			"oid", pkg.OrderId,
			"error", err)
		return nil, err
	}
	pkgInvoiceFinance.Subtotal = &subtotal

	discount, err := decimal.NewFromString(pkg.Invoice.Discount.Amount)
	if err != nil {
		applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Discount.Amount invalid",
			"fn", "FactoryFromPkg",
			"discount", pkg.Invoice.Discount.Amount,
			"pid", pkg.PId,
			"oid", pkg.OrderId,
			"error", err)
		return nil, err
	}
	pkgInvoiceFinance.Discount = &discount

	shipmentAmount, err := decimal.NewFromString(pkg.Invoice.ShipmentAmount.Amount)
	if err != nil {
		applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.ShipmentAmount.Amount invalid",
			"fn", "FactoryFromPkg",
			"shipmentAmount", pkg.Invoice.ShipmentAmount.Amount,
			"pid", pkg.PId,
			"oid", pkg.OrderId,
			"error", err)
		return nil, err
	}
	pkgInvoiceFinance.ShipmentAmount = &shipmentAmount

	// Invoice Share
	if pkg.Invoice.Share != nil {
		pkgInvoiceFinance.Share = &PackageShareFinance{}

		if pkg.Invoice.Share.RawBusinessShare != nil {
			rawBusinessShare, err := decimal.NewFromString(pkg.Invoice.Share.RawBusinessShare.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Share.RawBusinessShare.Amount invalid",
					"fn", "FactoryFromPkg",
					"rawBusinessShare", pkg.Invoice.Share.RawBusinessShare.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.Share.RawBusinessShare = &rawBusinessShare
		}

		if pkg.Invoice.Share.RoundupBusinessShare != nil {
			roundupBusinessShare, err := decimal.NewFromString(pkg.Invoice.Share.RoundupBusinessShare.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Share.RoundupBusinessShare.Amount invalid",
					"fn", "FactoryFromPkg",
					"roundupBusinessShare", pkg.Invoice.Share.RoundupBusinessShare.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.Share.RoundupBusinessShare = &roundupBusinessShare
		}

		if pkg.Invoice.Share.RawSellerShare != nil {
			rawSellerShare, err := decimal.NewFromString(pkg.Invoice.Share.RawSellerShare.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Share.RawSellerShare.Amount invalid",
					"fn", "FactoryFromPkg",
					"rawSellerShare", pkg.Invoice.Share.RawSellerShare.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.Share.RawSellerShare = &rawSellerShare
		}

		if pkg.Invoice.Share.RoundupSellerShare != nil {
			roundupSellerShare, err := decimal.NewFromString(pkg.Invoice.Share.RoundupSellerShare.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Share.RoundupSellerShare.Amount invalid",
					"fn", "FactoryFromPkg",
					"roundupSellerShare", pkg.Invoice.Share.RoundupSellerShare.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.Share.RoundupSellerShare = &roundupSellerShare
		}

		if pkg.Invoice.Share.RawSellerShippingNet != nil {
			rawSellerShippingNet, err := decimal.NewFromString(pkg.Invoice.Share.RawSellerShippingNet.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Share.RawSellerShippingNet.Amount invalid",
					"fn", "FactoryFromPkg",
					"rawSellerShippingNet", pkg.Invoice.Share.RawSellerShippingNet.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.Share.RawSellerShippingNet = &rawSellerShippingNet
		}

		if pkg.Invoice.Share.RoundupSellerShippingNet != nil {
			roundupSellerShippingNet, err := decimal.NewFromString(pkg.Invoice.Share.RoundupSellerShippingNet.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Share.RoundupSellerShippingNet.Amount invalid",
					"fn", "FactoryFromPkg",
					"roundupSellerShippingNet", pkg.Invoice.Share.RoundupSellerShippingNet.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.Share.RoundupSellerShippingNet = &roundupSellerShippingNet
		}

		pkgInvoiceFinance.Share.CreatedAt = pkg.Invoice.Share.CreatedAt
		pkgInvoiceFinance.Share.UpdatedAt = pkg.Invoice.Share.UpdatedAt
	}

	// Invoice Voucher
	if pkg.Invoice.Voucher != nil {
		pkgInvoiceFinance.Voucher = &PackageVoucherFinance{}

		if pkg.Invoice.Voucher.RawTotal != nil {
			rawTotal, err := decimal.NewFromString(pkg.Invoice.Voucher.RawTotal.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Voucher.RawTotal.Amount invalid",
					"fn", "FactoryFromPkg",
					"rawTotal", pkg.Invoice.Voucher.RawTotal.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.Voucher.RawTotal = &rawTotal
		}

		if pkg.Invoice.Voucher.RoundupTotal != nil {
			roundupTotal, err := decimal.NewFromString(pkg.Invoice.Voucher.RoundupTotal.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Voucher.RoundupTotal.Amount invalid",
					"fn", "FactoryFromPkg",
					"roundupTotal", pkg.Invoice.Voucher.RoundupTotal.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.Voucher.RoundupTotal = &roundupTotal
		}

		if pkg.Invoice.Voucher.RawCalcShipmentPrice != nil {
			rawCalcShipmentPrice, err := decimal.NewFromString(pkg.Invoice.Voucher.RawCalcShipmentPrice.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Voucher.RawCalcShipmentPrice.Amount invalid",
					"fn", "FactoryFromPkg",
					"rawCalcShipmentPrice", pkg.Invoice.Voucher.RawCalcShipmentPrice.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.Voucher.RawCalcShipmentPrice = &rawCalcShipmentPrice
		}

		if pkg.Invoice.Voucher.RoundupCalcShipmentPrice != nil {
			roundupCalcShipmentPrice, err := decimal.NewFromString(pkg.Invoice.Voucher.RoundupCalcShipmentPrice.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Voucher.RoundupCalcShipmentPrice.Amount invalid",
					"fn", "FactoryFromPkg",
					"roundupCalcShipmentPrice", pkg.Invoice.Voucher.RoundupCalcShipmentPrice.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.Voucher.RoundupCalcShipmentPrice = &roundupCalcShipmentPrice
		}

		pkgInvoiceFinance.Voucher.CreatedAt = pkg.Invoice.Voucher.CreatedAt
		pkgInvoiceFinance.Voucher.UpdatedAt = pkg.Invoice.Voucher.UpdatedAt
	}

	// Invoice SSO
	if pkg.Invoice.SSO != nil {
		pkgInvoiceFinance.SSO = &PackageSSOFinance{}
		pkgInvoiceFinance.SSO.Rate = pkg.Invoice.SSO.Rate
		pkgInvoiceFinance.SSO.IsObliged = pkg.Invoice.SSO.IsObliged

		if pkg.Invoice.SSO.RawTotal != nil {
			rawTotal, err := decimal.NewFromString(pkg.Invoice.SSO.RawTotal.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.SSO.RawTotal.Amount invalid",
					"fn", "FactoryFromPkg",
					"rawTotal", pkg.Invoice.SSO.RawTotal.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.SSO.RawTotal = &rawTotal
		}

		if pkg.Invoice.SSO.RoundupTotal != nil {
			roundupTotal, err := decimal.NewFromString(pkg.Invoice.SSO.RoundupTotal.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.SSO.RoundupTotal.Amount invalid",
					"fn", "FactoryFromPkg",
					"roundupTotal", pkg.Invoice.SSO.RoundupTotal.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.SSO.RoundupTotal = &roundupTotal
		}

		pkgInvoiceFinance.SSO.CreatedAt = pkg.Invoice.SSO.CreatedAt
		pkgInvoiceFinance.SSO.UpdatedAt = pkg.Invoice.SSO.UpdatedAt
	}

	// Invoice VAT
	if pkg.Invoice.VAT != nil {
		pkgInvoiceFinance.VAT = &PackageVATFinance{}
		if pkg.Invoice.VAT.SellerVAT != nil {
			pkgInvoiceFinance.VAT.SellerVAT = &PackageSellerVATFinance{}

			if pkg.Invoice.VAT.SellerVAT.RawTotal != nil {
				rawTotal, err := decimal.NewFromString(pkg.Invoice.VAT.SellerVAT.RawTotal.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.VAT.SellerVAT.RawTotal.Amount invalid",
						"fn", "FactoryFromPkg",
						"rawTotal", pkg.Invoice.VAT.SellerVAT.RawTotal.Amount,
						"pid", pkg.PId,
						"oid", pkg.OrderId,
						"error", err)
					return nil, err
				}
				pkgInvoiceFinance.VAT.SellerVAT.RawTotal = &rawTotal
			}

			if pkg.Invoice.VAT.SellerVAT.RoundupTotal != nil {
				roundupTotal, err := decimal.NewFromString(pkg.Invoice.VAT.SellerVAT.RoundupTotal.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.VAT.SellerVAT.RoundupTotal.Amount invalid",
						"fn", "FactoryFromPkg",
						"roundupTotal", pkg.Invoice.VAT.SellerVAT.RoundupTotal.Amount,
						"pid", pkg.PId,
						"oid", pkg.OrderId,
						"error", err)
					return nil, err
				}
				pkgInvoiceFinance.VAT.SellerVAT.RoundupTotal = &roundupTotal
			}

			pkgInvoiceFinance.VAT.SellerVAT.CreatedAt = pkg.Invoice.VAT.SellerVAT.CreatedAt
			pkgInvoiceFinance.VAT.SellerVAT.UpdatedAt = pkg.Invoice.VAT.SellerVAT.UpdatedAt
		}

		if pkg.Invoice.VAT.BusinessVAT != nil {
			pkgInvoiceFinance.VAT.BusinessVAT = &PackageBusinessVATFinance{}

			if pkg.Invoice.VAT.BusinessVAT.RawTotal != nil {
				rawTotal, err := decimal.NewFromString(pkg.Invoice.VAT.BusinessVAT.RawTotal.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.VAT.BusinessVAT.RawTotal.Amount invalid",
						"fn", "FactoryFromPkg",
						"rawTotal", pkg.Invoice.VAT.BusinessVAT.RawTotal.Amount,
						"pid", pkg.PId,
						"oid", pkg.OrderId,
						"error", err)
					return nil, err
				}
				pkgInvoiceFinance.VAT.BusinessVAT.RawTotal = &rawTotal
			}

			if pkg.Invoice.VAT.BusinessVAT.RoundupTotal != nil {
				roundupTotal, err := decimal.NewFromString(pkg.Invoice.VAT.BusinessVAT.RoundupTotal.Amount)
				if err != nil {
					applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.VAT.BusinessVAT.RoundupTotal.Amount invalid",
						"fn", "FactoryFromPkg",
						"roundupTotal", pkg.Invoice.VAT.BusinessVAT.RoundupTotal.Amount,
						"pid", pkg.PId,
						"oid", pkg.OrderId,
						"error", err)
					return nil, err
				}
				pkgInvoiceFinance.VAT.BusinessVAT.RoundupTotal = &roundupTotal
			}

			pkgInvoiceFinance.VAT.BusinessVAT.CreatedAt = pkg.Invoice.VAT.BusinessVAT.CreatedAt
			pkgInvoiceFinance.VAT.BusinessVAT.UpdatedAt = pkg.Invoice.VAT.BusinessVAT.UpdatedAt
		}
	}

	// invoice commission
	if pkg.Invoice.Commission != nil {
		pkgInvoiceFinance.Commission = &PackageCommissionFinance{}

		if pkg.Invoice.Commission.RawTotalPrice != nil {
			rawTotalPrice, err := decimal.NewFromString(pkg.Invoice.Commission.RawTotalPrice.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Commission.RawTotalPrice.Amount invalid",
					"fn", "FactoryFromPkg",
					"rawTotalPrice", pkg.Invoice.Commission.RawTotalPrice.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.Commission.RawTotalPrice = &rawTotalPrice
		}

		if pkg.Invoice.Commission.RoundupTotalPrice != nil {
			roundupTotalPrice, err := decimal.NewFromString(pkg.Invoice.Commission.RoundupTotalPrice.Amount)
			if err != nil {
				applog.GLog.Logger.FromContext(ctx).Error("pkg.Invoice.Commission.RoundupTotalPrice.Amount invalid",
					"fn", "FactoryFromPkg",
					"roundupTotalPrice", pkg.Invoice.Commission.RoundupTotalPrice.Amount,
					"pid", pkg.PId,
					"oid", pkg.OrderId,
					"error", err)
				return nil, err
			}
			pkgInvoiceFinance.Commission.RoundupTotalPrice = &roundupTotalPrice
		}

		pkgInvoiceFinance.Commission.CreatedAt = pkg.Invoice.Commission.CreatedAt
		pkgInvoiceFinance.Commission.UpdatedAt = pkg.Invoice.Commission.UpdatedAt
	}

	financeSubpackages := make([]*SubpackageFinance, 0, len(pkg.Subpackages))
	for i := 0; i < len(pkg.Subpackages); i++ {
		financeSubPkg, err := FactoryFromSubPkg(ctx, pkg.Subpackages[i])
		if err != nil {
			return nil, err
		}

		financeSubpackages = append(financeSubpackages, financeSubPkg)
	}

	return &PackageItemFinance{
		PId:         pkg.PId,
		Status:      pkg.Status,
		Invoice:     pkgInvoiceFinance,
		Subpackages: financeSubpackages,
	}, nil
}

func ConvertToPkg(ctx context.Context, finance *PackageItemFinance, pkg *entities.PackageItem) error {
	if finance.PId != pkg.PId {
		applog.GLog.Logger.FromContext(ctx).Error("package finance pid invalid",
			"fn", "ConvertToPkg",
			"finance pid", finance.PId,
			"pid", pkg.PId,
			"oid", pkg.OrderId)
		return errors.New("pkg finance pid not equal with pkg pid")
	}

	// Invoice.Share
	if finance.Invoice.Share != nil {
		if pkg.Invoice.Share == nil {
			pkg.Invoice.Share = &entities.PackageShare{}
		}

		if finance.Invoice.Share.RawBusinessShare != nil {
			pkg.Invoice.Share.RawBusinessShare = &entities.Money{
				Amount:   finance.Invoice.Share.RawBusinessShare.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.Share.RoundupBusinessShare != nil {
			pkg.Invoice.Share.RoundupBusinessShare = &entities.Money{
				Amount:   finance.Invoice.Share.RoundupBusinessShare.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.Share.RawSellerShare != nil {
			pkg.Invoice.Share.RawSellerShare = &entities.Money{
				Amount:   finance.Invoice.Share.RawSellerShare.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.Share.RoundupSellerShare != nil {
			pkg.Invoice.Share.RoundupSellerShare = &entities.Money{
				Amount:   finance.Invoice.Share.RoundupSellerShare.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.Share.RawSellerShippingNet != nil {
			pkg.Invoice.Share.RawSellerShippingNet = &entities.Money{
				Amount:   finance.Invoice.Share.RawSellerShippingNet.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.Share.RoundupSellerShippingNet != nil {
			pkg.Invoice.Share.RoundupSellerShippingNet = &entities.Money{
				Amount:   finance.Invoice.Share.RoundupSellerShippingNet.String(),
				Currency: "IRR",
			}
		}

		pkg.Invoice.Share.CreatedAt = finance.Invoice.Share.CreatedAt
		pkg.Invoice.Share.UpdatedAt = finance.Invoice.Share.UpdatedAt
	}

	// Invoice.Voucher
	if finance.Invoice.Voucher != nil {
		if pkg.Invoice.Voucher == nil {
			pkg.Invoice.Voucher = &entities.PackageVoucher{}
		}

		if finance.Invoice.Voucher.RawTotal != nil {
			pkg.Invoice.Voucher.RawTotal = &entities.Money{
				Amount:   finance.Invoice.Voucher.RawTotal.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.Voucher.RoundupTotal != nil {
			pkg.Invoice.Voucher.RoundupTotal = &entities.Money{
				Amount:   finance.Invoice.Voucher.RoundupTotal.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.Voucher.RawCalcShipmentPrice != nil {
			pkg.Invoice.Voucher.RawCalcShipmentPrice = &entities.Money{
				Amount:   finance.Invoice.Voucher.RawCalcShipmentPrice.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.Voucher.RoundupCalcShipmentPrice != nil {
			pkg.Invoice.Voucher.RoundupCalcShipmentPrice = &entities.Money{
				Amount:   finance.Invoice.Voucher.RoundupCalcShipmentPrice.String(),
				Currency: "IRR",
			}
		}

		pkg.Invoice.Voucher.CreatedAt = finance.Invoice.Voucher.CreatedAt
		pkg.Invoice.Voucher.UpdatedAt = finance.Invoice.Voucher.UpdatedAt
	}

	// Invoice.SSO
	if finance.Invoice.SSO != nil {
		//if pkg.Invoice.SSO == nil {
		//	pkg.Invoice.SSO = &entities.PackageSSO{}
		//}

		if finance.Invoice.SSO.RawTotal != nil {
			pkg.Invoice.SSO.RawTotal = &entities.Money{
				Amount:   finance.Invoice.SSO.RawTotal.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.SSO.RoundupTotal != nil {
			pkg.Invoice.SSO.RoundupTotal = &entities.Money{
				Amount:   finance.Invoice.SSO.RoundupTotal.String(),
				Currency: "IRR",
			}
		}

		pkg.Invoice.SSO.CreatedAt = finance.Invoice.SSO.CreatedAt
		pkg.Invoice.SSO.UpdatedAt = finance.Invoice.SSO.UpdatedAt
	}

	// Invoice.VAT
	if finance.Invoice.VAT != nil {
		if pkg.Invoice.VAT == nil {
			pkg.Invoice.VAT = &entities.PackageVAT{}
		}

		if finance.Invoice.VAT.SellerVAT != nil {
			if pkg.Invoice.VAT.SellerVAT == nil {
				pkg.Invoice.VAT.SellerVAT = &entities.PackageSellerVAT{}
			}

			if finance.Invoice.VAT.SellerVAT.RawTotal != nil {
				pkg.Invoice.VAT.SellerVAT.RawTotal = &entities.Money{
					Amount:   finance.Invoice.VAT.SellerVAT.RawTotal.String(),
					Currency: "IRR",
				}
			}

			if finance.Invoice.VAT.SellerVAT.RoundupTotal != nil {
				pkg.Invoice.VAT.SellerVAT.RoundupTotal = &entities.Money{
					Amount:   finance.Invoice.VAT.SellerVAT.RoundupTotal.String(),
					Currency: "IRR",
				}
			}

			pkg.Invoice.VAT.SellerVAT.CreatedAt = finance.Invoice.VAT.SellerVAT.CreatedAt
			pkg.Invoice.VAT.SellerVAT.UpdatedAt = finance.Invoice.VAT.SellerVAT.UpdatedAt
		}

		if finance.Invoice.VAT.BusinessVAT != nil {
			if pkg.Invoice.VAT.BusinessVAT == nil {
				pkg.Invoice.VAT.BusinessVAT = &entities.PackageBusinessVAT{}
			}

			if finance.Invoice.VAT.BusinessVAT.RawTotal != nil {
				pkg.Invoice.VAT.BusinessVAT.RawTotal = &entities.Money{
					Amount:   finance.Invoice.VAT.BusinessVAT.RawTotal.String(),
					Currency: "IRR",
				}
			}

			if finance.Invoice.VAT.BusinessVAT.RoundupTotal != nil {
				pkg.Invoice.VAT.BusinessVAT.RoundupTotal = &entities.Money{
					Amount:   finance.Invoice.VAT.BusinessVAT.RoundupTotal.String(),
					Currency: "IRR",
				}
			}

			pkg.Invoice.VAT.BusinessVAT.CreatedAt = finance.Invoice.VAT.BusinessVAT.CreatedAt
			pkg.Invoice.VAT.BusinessVAT.UpdatedAt = finance.Invoice.VAT.BusinessVAT.UpdatedAt
		}
	}

	// invoice commission
	if finance.Invoice.Commission != nil {
		if pkg.Invoice.Commission == nil {
			pkg.Invoice.Commission = &entities.PackageCommission{}
		}

		if finance.Invoice.Commission.RawTotalPrice != nil {
			pkg.Invoice.Commission.RawTotalPrice = &entities.Money{
				Amount:   finance.Invoice.Commission.RawTotalPrice.String(),
				Currency: "IRR",
			}
		}

		if finance.Invoice.Commission.RoundupTotalPrice != nil {
			pkg.Invoice.Commission.RoundupTotalPrice = &entities.Money{
				Amount:   finance.Invoice.Commission.RoundupTotalPrice.String(),
				Currency: "IRR",
			}
		}

		pkg.Invoice.Commission.CreatedAt = finance.Invoice.Commission.CreatedAt
		pkg.Invoice.Commission.UpdatedAt = finance.Invoice.Commission.UpdatedAt
	}

	for i := 0; i < len(finance.Subpackages); i++ {
		err := ConvertToSubPkg(ctx, finance.Subpackages[i], pkg.Subpackages[i])
		if err != nil {
			return err
		}
	}

	return nil
}
