package calculate

import (
	"context"
	"errors"
	"github.com/shopspring/decimal"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"sort"
	"time"
)

type financeCalcFunc func(ctx context.Context, order *OrderFinance, mode FinanceMode) error

const (
	newStatus        string = "NEW"
	inProgressStatus string = "IN_PROGRESS"
	closedStatus     string = "CLOSED"
)

const (
	payToSellerState string = "Pay_To_Seller"
	payToBuyerState  string = "Pay_To_Buyer"
)

type baseItem struct {
	Sid         uint64
	InventoryId string
	UnitPrice   *decimal.Decimal
	Quantity    int32
}

type basePackage struct {
	Pid           uint64
	ShipmentPrice *decimal.Decimal
}

type SortBaseItem []*baseItem
type SortBasePkg []*basePackage

type financeCalculatorImpl struct {
	baseItems         SortBaseItem
	itemsMap          map[uint64]map[string]*ItemFinance
	baseVoucherRatio  decimal.Decimal
	baseShipmentRatio decimal.Decimal
	basePackages      SortBasePkg
	pkgMap            map[uint64]*PackageItemFinance
	timestamp         *time.Time
}

func (base SortBaseItem) Len() int { return len(base) }
func (base SortBaseItem) Less(i, j int) bool {
	return base[i].UnitPrice.LessThan(*base[j].UnitPrice)
}
func (base SortBaseItem) Swap(i, j int) {
	base[i], base[j] = base[j], base[i]
}

func (base SortBasePkg) Len() int { return len(base) }
func (base SortBasePkg) Less(i, j int) bool {
	return base[i].ShipmentPrice.LessThan(*base[j].ShipmentPrice)
}
func (base SortBasePkg) Swap(i, j int) {
	base[i], base[j] = base[j], base[i]
}

func New() FinanceCalculator {
	return &financeCalculatorImpl{}
}

func (finance financeCalculatorImpl) FinanceCalc(ctx context.Context, order entities.Order, calcType FinanceCalcType, mode FinanceMode) (*entities.Order, error) {

	timestamp := time.Now().UTC()
	finance.baseItems = make([]*baseItem, 0, 32)
	finance.itemsMap = make(map[uint64]map[string]*ItemFinance, 32)
	finance.basePackages = make([]*basePackage, 0, 16)
	finance.pkgMap = make(map[uint64]*PackageItemFinance, 16)
	finance.baseVoucherRatio = decimal.Zero
	finance.baseShipmentRatio = decimal.Zero
	finance.timestamp = &timestamp

	orderFinance, err := FactoryFromOrder(ctx, &order)
	if err != nil {
		return nil, err
	}

	// traverse order to get requirements of finance calculation
	for i := 0; i < len(orderFinance.Packages); i++ {
		finance.pkgMap[orderFinance.Packages[i].PId] = orderFinance.Packages[i]

		basePackage := &basePackage{
			Pid:           orderFinance.Packages[i].PId,
			ShipmentPrice: orderFinance.Packages[i].Invoice.ShipmentAmount,
		}

		finance.baseShipmentRatio = finance.baseShipmentRatio.Add(*orderFinance.Packages[i].Invoice.ShipmentAmount)
		finance.basePackages = append(finance.basePackages, basePackage)

		for j := 0; j < len(orderFinance.Packages[i].Subpackages); j++ {
			finance.itemsMap[orderFinance.Packages[i].Subpackages[j].SId] = make(map[string]*ItemFinance, len(orderFinance.Packages[i].Subpackages[j].Items))

			for k := 0; k < len(orderFinance.Packages[i].Subpackages[j].Items); k++ {
				finance.itemsMap[orderFinance.Packages[i].Subpackages[j].SId][orderFinance.Packages[i].Subpackages[j].Items[k].InventoryId] = orderFinance.Packages[i].Subpackages[j].Items[k]

				baseItem := &baseItem{
					Sid:         orderFinance.Packages[i].Subpackages[j].SId,
					InventoryId: orderFinance.Packages[i].Subpackages[j].Items[k].InventoryId,
					UnitPrice:   orderFinance.Packages[i].Subpackages[j].Items[k].Invoice.Unit,
					Quantity:    orderFinance.Packages[i].Subpackages[j].Items[k].Quantity,
				}

				finance.baseVoucherRatio = finance.baseVoucherRatio.Add(
					orderFinance.Packages[i].Subpackages[j].Items[k].Invoice.Unit.Mul(decimal.NewFromInt(int64(baseItem.Quantity))))
				finance.baseItems = append(finance.baseItems, baseItem)
			}
		}
	}

	sort.Sort(finance.baseItems)
	sort.Sort(finance.basePackages)

	base := func(ctx context.Context, order *OrderFinance, mode FinanceMode) error {
		return nil
	}

	var decorator financeCalcFunc

	if calcType != VOUCHER_CALC && Has(VOUCHER_CALC, calcType) {
		decorator = finance.voucherCalc(base)
		calcType = Clear(calcType, VOUCHER_CALC)
	} else {
		decorator = base
	}

	switch calcType {
	case VOUCHER_CALC:
		decorator = finance.voucherCalc(decorator)
		break

	case SELLER_VAT_CALC:
		decorator = finance.sellerVatCalc(decorator)
		break

	case NET_COMMISSION_CALC:
		decorator = finance.netCommissionCalc(finance.sellerVatCalc(decorator))
		break

	case BUSINESS_VAT_CALC:
		decorator = finance.businessVatCalc(finance.netCommissionCalc(finance.sellerVatCalc(decorator)))
		break

	case SELLER_SSO_CALC:
		decorator = finance.sellerSsoCalc(finance.businessVatCalc(finance.netCommissionCalc(finance.sellerVatCalc(decorator))))
		break

	case SHARE_CALC:
		decorator = finance.shareCalc(finance.sellerSsoCalc(finance.businessVatCalc(finance.netCommissionCalc(finance.sellerVatCalc(decorator)))))
		break

	default:
		return nil, errors.New("mode invalid")
	}

	err = decorator(ctx, orderFinance, mode)
	if err != nil {
		return nil, err
	}

	err = ConvertToOrder(ctx, orderFinance, &order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (finance financeCalculatorImpl) voucherCalc(decorator financeCalcFunc) financeCalcFunc {
	return func(ctx context.Context, order *OrderFinance, mode FinanceMode) error {
		if err := decorator(ctx, order, mode); err != nil {
			return err
		}

		// ORDER Finance only valid for voucher calculation
		if mode == SELLER_FINANCE || mode == BUYER_FINANCE {
			return nil
		}

		if mode == ORDER_FINANCE && order.Status != newStatus {
			return nil
		}

		if order.Invoice.Voucher == nil {
			return nil
		}

		var shipmentVoucherPrice = decimal.Zero
		if order.Invoice.GrandTotal.IsZero() {
			if order.Invoice.Voucher.Percent == 0 {
				shipmentVoucherPrice = *order.Invoice.ShipmentTotal
			}
		}

		if !shipmentVoucherPrice.IsZero() {
			order.Invoice.Voucher.RawShipmentAppliedPrice = &shipmentVoucherPrice
			roundupShipmentAppliedPrice := shipmentVoucherPrice.Ceil()
			order.Invoice.Voucher.RoundupShipmentAppliedPrice = &roundupShipmentAppliedPrice
		}

		var netVoucherAppliedPrice = decimal.Zero
		if order.Invoice.Voucher.RoundupAppliedPrice != nil {
			netVoucherAppliedPrice = (*order.Invoice.Voucher.RoundupAppliedPrice).Sub(shipmentVoucherPrice)
		} else {
			netVoucherAppliedPrice = (*order.Invoice.Voucher.AppliedPrice).Sub(shipmentVoucherPrice)
		}

		// calculate item voucher
		var roundupSum = decimal.Zero
		for i := 0; i < len(finance.baseItems); i++ {
			itemFinance := finance.itemsMap[finance.baseItems[i].Sid][finance.baseItems[i].InventoryId]
			if itemFinance.Invoice.Voucher == nil {
				itemFinance.Invoice.Voucher = &ItemVoucherFinance{}
				itemFinance.Invoice.Voucher.CreatedAt = finance.timestamp
			}

			if i == len(finance.baseItems)-1 {
				lastRawUnitPrice := (netVoucherAppliedPrice).
					Sub(roundupSum).
					Div(decimal.NewFromInt(int64(finance.baseItems[i].Quantity)))
				itemFinance.Invoice.Voucher.RawUnitPrice = &lastRawUnitPrice
				lastRoundupUnitPrice := lastRawUnitPrice.Ceil()
				itemFinance.Invoice.Voucher.RoundupUnitPrice = &lastRoundupUnitPrice

				lastRawTotalPrice := (netVoucherAppliedPrice).Sub(roundupSum)
				itemFinance.Invoice.Voucher.RawTotalPrice = &lastRawTotalPrice
				lastRoundupTotalPrice := lastRawTotalPrice.Ceil()
				itemFinance.Invoice.Voucher.RoundupTotalPrice = &lastRoundupTotalPrice

				itemFinance.Invoice.Voucher.UpdatedAt = finance.timestamp

			} else {
				rawUnit := finance.baseItems[i].UnitPrice.Mul(netVoucherAppliedPrice)
				rawUnit = rawUnit.Div(finance.baseVoucherRatio)
				itemFinance.Invoice.Voucher.RawUnitPrice = &rawUnit
				rawTotal := rawUnit.Mul(decimal.NewFromInt(int64(finance.baseItems[i].Quantity)))
				itemFinance.Invoice.Voucher.RawTotalPrice = &rawTotal

				roundupUnit := (*itemFinance.Invoice.Voucher.RawUnitPrice).Ceil()
				itemFinance.Invoice.Voucher.RoundupUnitPrice = &roundupUnit
				roundupTotal := roundupUnit.Mul(decimal.NewFromInt(int64(finance.baseItems[i].Quantity)))
				itemFinance.Invoice.Voucher.RoundupTotalPrice = &roundupTotal

				itemFinance.Invoice.Voucher.UpdatedAt = finance.timestamp

				roundupSum = roundupSum.Add(roundupTotal)
			}
		}

		// calculate package voucher and shipment voucher
		for i := 0; i < len(finance.basePackages); i++ {
			pkgFinance := finance.pkgMap[finance.basePackages[i].Pid]
			if pkgFinance.Invoice.Voucher == nil {
				pkgFinance.Invoice.Voucher = &PackageVoucherFinance{}
				pkgFinance.Invoice.Voucher.CreatedAt = finance.timestamp
			}

			var rawTotal = decimal.Zero
			var roundupTotal = decimal.Zero
			for j := 0; j < len(pkgFinance.Subpackages); j++ {
				for k := 0; k < len(pkgFinance.Subpackages[j].Items); k++ {
					rawTotal = rawTotal.Add(*pkgFinance.Subpackages[j].Items[k].Invoice.Voucher.RawTotalPrice)
					roundupTotal = roundupTotal.Add(*pkgFinance.Subpackages[j].Items[k].Invoice.Voucher.RoundupTotalPrice)
				}
			}
			pkgFinance.Invoice.Voucher.RawTotal = &rawTotal
			pkgFinance.Invoice.Voucher.RoundupTotal = &roundupTotal
			pkgFinance.Invoice.Voucher.UpdatedAt = finance.timestamp

			if !shipmentVoucherPrice.IsZero() {
				roundupSum = decimal.Zero

				if i == len(finance.basePackages)-1 {
					lastRawShipmentPrice := shipmentVoucherPrice.Sub(roundupSum)
					pkgFinance.Invoice.Voucher.RawCalcShipmentPrice = &lastRawShipmentPrice

					lastRoundupShipmentPrice := lastRawShipmentPrice.Ceil()
					pkgFinance.Invoice.Voucher.RoundupCalcShipmentPrice = &lastRoundupShipmentPrice

				} else {
					rawShipmentPrice := pkgFinance.Invoice.ShipmentAmount.Mul(shipmentVoucherPrice)
					rawShipmentPrice = rawShipmentPrice.Div(finance.baseShipmentRatio)
					pkgFinance.Invoice.Voucher.RawCalcShipmentPrice = &rawShipmentPrice

					roundupShipmentPrice := rawShipmentPrice.Ceil()
					pkgFinance.Invoice.Voucher.RawCalcShipmentPrice = &roundupShipmentPrice

					roundupSum = roundupSum.Add(roundupShipmentPrice)
				}
			}
		}

		return nil
	}
}

func (finance financeCalculatorImpl) sellerVatCalc(decorator financeCalcFunc) financeCalcFunc {
	return func(ctx context.Context, order *OrderFinance, mode FinanceMode) error {
		if err := decorator(ctx, order, mode); err != nil {
			return err
		}

		if mode == ORDER_FINANCE && order.Status == inProgressStatus {
			return nil
		}

		for i := 0; i < len(order.Packages); i++ {
			if mode == SELLER_FINANCE && order.Packages[i].Status != closedStatus {
				continue
			}

			if order.Packages[i].Invoice.VAT == nil {
				order.Packages[i].Invoice.VAT = &PackageVATFinance{}
			}

			if order.Packages[i].Invoice.VAT.SellerVAT == nil {
				order.Packages[i].Invoice.VAT.SellerVAT = &PackageSellerVATFinance{}
				order.Packages[i].Invoice.VAT.SellerVAT.CreatedAt = finance.timestamp
			}

			rawPkgTotalVat := decimal.Zero
			roundupPkgTotalVat := decimal.Zero

			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				if mode == SELLER_FINANCE && order.Packages[i].Subpackages[j].Status != payToSellerState {
					continue
				}

				for k := 0; k < len(order.Packages[i].Subpackages[j].Items); k++ {
					itemFinance := order.Packages[i].Subpackages[j].Items[k]

					if itemFinance.Invoice.VAT == nil || itemFinance.Invoice.VAT.SellerVat == nil {
						order.Packages[i].Invoice.VAT = nil
						order.Packages[i].Invoice.VAT.SellerVAT = nil
						continue
					}

					if itemFinance.Invoice.Share == nil {
						itemFinance.Invoice.Share = &ItemShareFinance{}
						itemFinance.Invoice.Share.CreatedAt = finance.timestamp
					}

					// calculate item gross
					if itemFinance.Invoice.Share.RawItemGross == nil {
						rawItemGross := *itemFinance.Invoice.Unit
						itemFinance.Invoice.Share.RawItemGross = &rawItemGross

						rawTotalGross := *itemFinance.Invoice.Total
						itemFinance.Invoice.Share.RawTotalGross = &rawTotalGross

						roundupItemGross := rawItemGross.Ceil()
						itemFinance.Invoice.Share.RoundupItemGross = &roundupItemGross

						roundupTotalGross := *itemFinance.Invoice.Total
						roundupTotalGross = roundupTotalGross.Ceil()
						itemFinance.Invoice.Share.RoundupTotalGross = &roundupTotalGross

						itemFinance.Invoice.Share.UpdatedAt = finance.timestamp
					}

					// calculation seller vat and item net
					if itemFinance.Invoice.VAT.SellerVat.IsObliged {
						rawItemVat := itemFinance.Invoice.Share.RawItemGross.
							Mul(decimal.NewFromFloat32(itemFinance.Invoice.VAT.SellerVat.Rate)).
							Div(decimal.NewFromInt(100))

						itemFinance.Invoice.VAT.SellerVat.RawUnitPrice = &rawItemVat
						rawTotalVat := rawItemVat.Mul(decimal.NewFromInt32(itemFinance.Quantity))
						itemFinance.Invoice.VAT.SellerVat.RawTotalPrice = &rawTotalVat

						roundupItemVat := itemFinance.Invoice.Share.RoundupItemGross.
							Mul(decimal.NewFromFloat32(itemFinance.Invoice.VAT.SellerVat.Rate)).
							Div(decimal.NewFromInt(100)).
							Ceil()
						itemFinance.Invoice.VAT.SellerVat.RoundupUnitPrice = &roundupItemVat
						roundupTotalVat := roundupItemVat.Mul(decimal.NewFromInt32(itemFinance.Quantity))
						itemFinance.Invoice.VAT.SellerVat.RoundupTotalPrice = &roundupTotalVat

						// calculate item net
						rawItemNet := itemFinance.Invoice.Share.RawItemGross.Sub(rawItemVat)
						itemFinance.Invoice.Share.RawItemNet = &rawItemNet
						rawTotalNet := rawItemNet.Mul(decimal.NewFromInt32(itemFinance.Quantity))
						itemFinance.Invoice.Share.RawTotalNet = &rawTotalNet

						roundupItemNet := itemFinance.Invoice.Share.RoundupItemGross.Sub(roundupItemVat)
						itemFinance.Invoice.Share.RoundupItemNet = &roundupItemNet
						roundupTotalNet := roundupItemNet.Mul(decimal.NewFromInt32(itemFinance.Quantity))
						itemFinance.Invoice.Share.RoundupTotalNet = &roundupTotalNet

						itemFinance.Invoice.VAT.SellerVat.UpdatedAt = finance.timestamp

					} else {
						//itemFinance.Invoice.VAT.SellerVat.RawUnitPrice = &decimal.Zero
						//itemFinance.Invoice.VAT.SellerVat.RawTotalPrice = &decimal.Zero
						//itemFinance.Invoice.VAT.SellerVat.RoundupUnitPrice = &decimal.Zero
						//itemFinance.Invoice.VAT.SellerVat.RoundupTotalPrice = &decimal.Zero

						rawItemNet := *itemFinance.Invoice.Share.RawItemGross
						itemFinance.Invoice.Share.RawItemNet = &rawItemNet
						rawTotalNet := *itemFinance.Invoice.Share.RawTotalGross
						itemFinance.Invoice.Share.RawTotalNet = &rawTotalNet

						roundupItemNet := *itemFinance.Invoice.Share.RoundupItemGross
						itemFinance.Invoice.Share.RoundupItemNet = &roundupItemNet
						roundupTotalNet := *itemFinance.Invoice.Share.RoundupTotalGross
						itemFinance.Invoice.Share.RoundupTotalNet = &roundupTotalNet

						itemFinance.Invoice.VAT.SellerVat.UpdatedAt = finance.timestamp
					}

					rawPkgTotalVat = rawPkgTotalVat.Add(*itemFinance.Invoice.VAT.SellerVat.RawTotalPrice)
					roundupPkgTotalVat = roundupPkgTotalVat.Add(*itemFinance.Invoice.VAT.SellerVat.RoundupTotalPrice)
				}
			}

			order.Packages[i].Invoice.VAT.SellerVAT.RawTotal = &rawPkgTotalVat
			order.Packages[i].Invoice.VAT.SellerVAT.RoundupTotal = &roundupPkgTotalVat
			order.Packages[i].Invoice.VAT.SellerVAT.UpdatedAt = finance.timestamp
		}
		return nil
	}
}

func (finance financeCalculatorImpl) netCommissionCalc(decorator financeCalcFunc) financeCalcFunc {
	return func(ctx context.Context, order *OrderFinance, mode FinanceMode) error {
		if err := decorator(ctx, order, mode); err != nil {
			return err
		}

		if mode == ORDER_FINANCE && order.Status == inProgressStatus {
			return nil
		}

		if mode == ORDER_FINANCE || mode == SELLER_FINANCE {
			if order.Status != inProgressStatus {
				if order.Invoice.Commission == nil {
					order.Invoice.Commission = &CommissionFinance{}
					order.Invoice.Commission.CreatedAt = finance.timestamp
				}
			}
		}

		rawTotalPrice := decimal.Zero
		roundupTotalPrice := decimal.Zero

		for i := 0; i < len(order.Packages); i++ {
			if mode == SELLER_FINANCE && order.Packages[i].Status != closedStatus {
				continue
			}

			if order.Packages[i].Invoice.Commission == nil {
				order.Packages[i].Invoice.Commission = &PackageCommissionFinance{}
				order.Packages[i].Invoice.Commission.CreatedAt = finance.timestamp
			}

			rawPkgTotalPrice := decimal.Zero
			roundupPkgTotalPrice := decimal.Zero

			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				if mode == SELLER_FINANCE && order.Packages[i].Subpackages[j].Status != payToSellerState {
					continue
				}

				for k := 0; k < len(order.Packages[i].Subpackages[j].Items); k++ {
					itemFinance := order.Packages[i].Subpackages[j].Items[k]

					if itemFinance.Invoice.Commission == nil {
						continue
					}

					rawUnitPrice := *itemFinance.Invoice.Share.RawItemNet
					rawUnitPrice = rawUnitPrice.
						Sub(decimal.NewFromFloat32(itemFinance.Invoice.Commission.ItemCommission)).
						Div(decimal.NewFromInt(100))
					itemFinance.Invoice.Commission.RawUnitPrice = &rawUnitPrice

					rawTotalPrice := rawUnitPrice.Mul(decimal.NewFromInt32(itemFinance.Quantity))
					itemFinance.Invoice.Commission.RawTotalPrice = &rawTotalPrice

					roundupUnitPrice := *itemFinance.Invoice.Share.RoundupItemNet
					roundupUnitPrice = roundupUnitPrice.
						Sub(decimal.NewFromFloat32(itemFinance.Invoice.Commission.ItemCommission)).
						Div(decimal.NewFromInt(100)).
						Ceil()
					itemFinance.Invoice.Commission.RoundupUnitPrice = &roundupUnitPrice

					roundupTotalPrice := roundupUnitPrice.Mul(decimal.NewFromInt32(itemFinance.Quantity))
					itemFinance.Invoice.Commission.RoundupTotalPrice = &roundupTotalPrice

					itemFinance.Invoice.Commission.UpdatedAt = finance.timestamp

					rawPkgTotalPrice = rawPkgTotalPrice.Add(*itemFinance.Invoice.Commission.RawTotalPrice)
					roundupPkgTotalPrice = roundupPkgTotalPrice.Add(*itemFinance.Invoice.Commission.RoundupTotalPrice)
				}
			}

			if rawPkgTotalPrice.IsZero() && roundupPkgTotalPrice.IsZero() {
				order.Packages[i].Invoice.Commission = nil
				continue
			}

			order.Packages[i].Invoice.Commission.RawTotalPrice = &rawPkgTotalPrice
			order.Packages[i].Invoice.Commission.RoundupTotalPrice = &roundupPkgTotalPrice
			order.Packages[i].Invoice.Commission.UpdatedAt = finance.timestamp

			// order commission
			rawTotalPrice = rawTotalPrice.Add(rawPkgTotalPrice)
			roundupTotalPrice = roundupTotalPrice.Add(roundupPkgTotalPrice)
		}

		if mode == ORDER_FINANCE || mode == SELLER_FINANCE {
			if order.Status != inProgressStatus {
				if rawTotalPrice.IsZero() && roundupTotalPrice.IsZero() {
					order.Invoice.Commission = nil
				} else {
					order.Invoice.Commission.UpdatedAt = finance.timestamp
					order.Invoice.Commission.RawTotalPrice = &rawTotalPrice
					order.Invoice.Commission.RoundupTotalPrice = &roundupTotalPrice
				}
			}
		}

		return nil
	}
}

func (finance financeCalculatorImpl) businessVatCalc(decorator financeCalcFunc) financeCalcFunc {
	return func(ctx context.Context, order *OrderFinance, mode FinanceMode) error {
		if err := decorator(ctx, order, mode); err != nil {
			return err
		}

		if mode == ORDER_FINANCE && order.Status == inProgressStatus {
			return nil
		}

		if mode == ORDER_FINANCE || mode == SELLER_FINANCE {
			if order.Status != inProgressStatus {
				if order.Invoice.VAT == nil {
					order.Invoice.VAT = &VATFinance{}
					order.Invoice.VAT.CreatedAt = finance.timestamp
				}
			}
		}

		rawTotal := decimal.Zero
		roundupTotal := decimal.Zero

		for i := 0; i < len(order.Packages); i++ {
			if mode == SELLER_FINANCE && order.Packages[i].Status != closedStatus {
				continue
			}

			if order.Packages[i].Invoice.VAT == nil {
				continue
			}

			if order.Packages[i].Invoice.VAT.BusinessVAT == nil {
				order.Packages[i].Invoice.VAT.BusinessVAT = &PackageBusinessVATFinance{}
				order.Packages[i].Invoice.VAT.BusinessVAT.CreatedAt = finance.timestamp
			}

			rawPkgTotalVat := decimal.Zero
			roundupPkgTotalVat := decimal.Zero

			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				if mode == SELLER_FINANCE && order.Packages[i].Subpackages[j].Status != payToSellerState {
					continue
				}

				for k := 0; k < len(order.Packages[i].Subpackages[j].Items); k++ {
					itemFinance := order.Packages[i].Subpackages[j].Items[k]

					rawItemVat := itemFinance.Invoice.Commission.RawUnitPrice.
						Mul(decimal.NewFromFloat32(itemFinance.Invoice.VAT.BusinessVat.Rate)).
						Div(decimal.NewFromInt(100))

					itemFinance.Invoice.VAT.BusinessVat.RawUnitPrice = &rawItemVat
					rawTotalVat := rawItemVat.Mul(decimal.NewFromInt32(itemFinance.Quantity))
					itemFinance.Invoice.VAT.BusinessVat.RawTotalPrice = &rawTotalVat

					roundupItemVat := itemFinance.Invoice.Commission.RoundupUnitPrice.
						Mul(decimal.NewFromFloat32(itemFinance.Invoice.VAT.BusinessVat.Rate)).
						Div(decimal.NewFromInt(100)).
						Ceil()
					itemFinance.Invoice.VAT.BusinessVat.RoundupUnitPrice = &roundupItemVat
					roundupTotalVat := roundupItemVat.Mul(decimal.NewFromInt32(itemFinance.Quantity))
					itemFinance.Invoice.VAT.BusinessVat.RoundupTotalPrice = &roundupTotalVat

					itemFinance.Invoice.VAT.BusinessVat.UpdatedAt = finance.timestamp

					rawPkgTotalVat = rawPkgTotalVat.Add(*itemFinance.Invoice.VAT.BusinessVat.RawTotalPrice)
					roundupPkgTotalVat = roundupPkgTotalVat.Add(*itemFinance.Invoice.VAT.BusinessVat.RoundupTotalPrice)
				}
			}

			order.Packages[i].Invoice.VAT.BusinessVAT.RawTotal = &rawPkgTotalVat
			order.Packages[i].Invoice.VAT.BusinessVAT.RoundupTotal = &roundupPkgTotalVat
			order.Packages[i].Invoice.VAT.BusinessVAT.UpdatedAt = finance.timestamp

			// order business vat
			rawTotal = rawTotal.Add(*order.Packages[i].Invoice.VAT.BusinessVAT.RawTotal)
			roundupTotal = roundupTotal.Add(*order.Packages[i].Invoice.VAT.BusinessVAT.RoundupTotal)
		}

		if mode == ORDER_FINANCE || mode == SELLER_FINANCE {
			if order.Status != inProgressStatus {
				if rawTotal.IsZero() && roundupTotal.IsZero() {
					order.Invoice.VAT = nil
				} else {
					order.Invoice.VAT.UpdatedAt = finance.timestamp
					order.Invoice.VAT.RawTotal = &rawTotal
					order.Invoice.VAT.RoundupTotal = &roundupTotal
				}
			}
		}

		return nil
	}
}

func (finance financeCalculatorImpl) sellerSsoCalc(decorator financeCalcFunc) financeCalcFunc {
	return func(ctx context.Context, order *OrderFinance, mode FinanceMode) error {
		if err := decorator(ctx, order, mode); err != nil {
			return err
		}

		if mode == ORDER_FINANCE && order.Status == inProgressStatus {
			return nil
		}

		if mode == ORDER_FINANCE || mode == SELLER_FINANCE {
			if order.Status != inProgressStatus {
				if order.Invoice.SSO == nil {
					order.Invoice.SSO = &SSOFinance{}
					order.Invoice.SSO.CreatedAt = finance.timestamp
				}
			}
		}

		rawTotal := decimal.Zero
		roundupTotal := decimal.Zero

		for i := 0; i < len(order.Packages); i++ {
			if mode == SELLER_FINANCE && order.Packages[i].Status != closedStatus {
				continue
			}

			if order.Packages[i].Invoice.SSO == nil || !order.Packages[i].Invoice.SSO.IsObliged {
				continue
			}

			pkgRawTotal := decimal.Zero
			pkgRoundupTotal := decimal.Zero

			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				if mode == SELLER_FINANCE && order.Packages[i].Subpackages[j].Status != payToSellerState {
					continue
				}

				for k := 0; k < len(order.Packages[i].Subpackages[j].Items); k++ {
					itemFinance := order.Packages[i].Subpackages[j].Items[k]
					if itemFinance.Invoice.SSO == nil {
						itemFinance.Invoice.SSO = &ItemSSOFinance{}
						itemFinance.Invoice.SSO.CreatedAt = finance.timestamp
					}

					rawUnitPrice := (*itemFinance.Invoice.Commission.RawUnitPrice).
						Mul(decimal.NewFromFloat32(order.Packages[i].Invoice.SSO.Rate)).
						Div(decimal.NewFromInt(100))
					itemFinance.Invoice.SSO.RawUnitPrice = &rawUnitPrice

					rawTotalPrice := rawUnitPrice.Mul(decimal.NewFromInt32(itemFinance.Quantity))
					itemFinance.Invoice.SSO.RawTotalPrice = &rawTotalPrice

					roundupUnitPrice := (*itemFinance.Invoice.Commission.RoundupUnitPrice).
						Mul(decimal.NewFromFloat32(order.Packages[i].Invoice.SSO.Rate)).
						Div(decimal.NewFromInt(100)).
						Ceil()
					itemFinance.Invoice.SSO.RoundupUnitPrice = &roundupUnitPrice

					roundupTotalPrice := roundupUnitPrice.Mul(decimal.NewFromInt32(itemFinance.Quantity))
					itemFinance.Invoice.SSO.RoundupTotalPrice = &roundupTotalPrice

					itemFinance.Invoice.SSO.UpdatedAt = finance.timestamp

					pkgRawTotal = pkgRawTotal.Add(*itemFinance.Invoice.SSO.RawTotalPrice)
					pkgRoundupTotal = pkgRoundupTotal.Add(*itemFinance.Invoice.SSO.RoundupTotalPrice)
				}
			}

			order.Packages[i].Invoice.SSO.RawTotal = &pkgRawTotal
			order.Packages[i].Invoice.SSO.RoundupTotal = &pkgRoundupTotal
			order.Packages[i].Invoice.SSO.UpdatedAt = finance.timestamp

			// order sso
			rawTotal = rawTotal.Add(*order.Packages[i].Invoice.SSO.RawTotal)
			roundupTotal = roundupTotal.Add(*order.Packages[i].Invoice.SSO.RoundupTotal)
		}

		if mode == ORDER_FINANCE || mode == SELLER_FINANCE {
			if order.Status != inProgressStatus {
				if rawTotal.IsZero() && roundupTotal.IsZero() {
					order.Invoice.SSO = nil
				} else {
					order.Invoice.SSO.UpdatedAt = finance.timestamp
					order.Invoice.SSO.RawTotal = &rawTotal
					order.Invoice.SSO.RoundupTotal = &roundupTotal
				}
			}
		}

		return nil
	}
}

func (finance financeCalculatorImpl) shareCalc(decorator financeCalcFunc) financeCalcFunc {
	return func(ctx context.Context, order *OrderFinance, mode FinanceMode) error {
		if err := decorator(ctx, order, mode); err != nil {
			return err
		}

		if mode == ORDER_FINANCE && order.Status == inProgressStatus {
			return nil
		}

		if order.Invoice.VAT == nil &&
			order.Invoice.Commission == nil &&
			order.Invoice.SSO == nil {
			return nil
		}

		if mode == ORDER_FINANCE || mode == SELLER_FINANCE {
			if order.Status != inProgressStatus {
				if order.Invoice.Share == nil {
					order.Invoice.Share = &ShareFinance{}
					order.Invoice.Share.CreatedAt = finance.timestamp
				}
			}
		}

		rawTotalShare := decimal.Zero
		roundupTotalShare := decimal.Zero

		for i := 0; i < len(order.Packages); i++ {
			if mode == SELLER_FINANCE && order.Packages[i].Status != closedStatus {
				continue
			}

			if order.Packages[i].Invoice.Share == nil {
				order.Packages[i].Invoice.Share = &PackageShareFinance{}
				order.Packages[i].Invoice.Share.CreatedAt = finance.timestamp
			}

			rawPkgBusinessShare := decimal.Zero
			roundupPkgBusinessShare := decimal.Zero
			rawPkgSellerShare := decimal.Zero
			roundupPkgSellerShare := decimal.Zero

			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				if mode == SELLER_FINANCE && order.Packages[i].Subpackages[j].Status != payToSellerState {
					continue
				}

				for k := 0; k < len(order.Packages[i].Subpackages[j].Items); k++ {
					itemFinance := order.Packages[i].Subpackages[j].Items[k]

					// seller share raw
					var rawUnitSellerShare = decimal.Zero
					if itemFinance.Invoice.SSO.RawUnitPrice != nil {
						rawUnitSellerShare = *itemFinance.Invoice.SSO.RawUnitPrice
					}

					if itemFinance.Invoice.VAT.SellerVat.RawUnitPrice != nil {
						rawUnitSellerShare = rawUnitSellerShare.Add(*itemFinance.Invoice.VAT.SellerVat.RawUnitPrice)
					}

					rawUnitSellerShare = rawUnitSellerShare.Add(*itemFinance.Invoice.Share.RawItemNet).
						Sub(*itemFinance.Invoice.Commission.RawUnitPrice).
						Sub(*itemFinance.Invoice.VAT.BusinessVat.RawUnitPrice)

					itemFinance.Invoice.Share.RawUnitSellerShare = &rawUnitSellerShare
					rawTotalSellerShare := rawUnitSellerShare.Mul(decimal.NewFromInt32(itemFinance.Quantity))
					itemFinance.Invoice.Share.RawTotalSellerShare = &rawTotalSellerShare

					// seller share roundup
					var roundupUnitSellerShare = decimal.Zero
					if itemFinance.Invoice.SSO.RoundupUnitPrice != nil {
						roundupUnitSellerShare = *itemFinance.Invoice.SSO.RoundupUnitPrice
					}

					if itemFinance.Invoice.VAT.SellerVat.RoundupUnitPrice != nil {
						roundupUnitSellerShare = roundupUnitSellerShare.Add(*itemFinance.Invoice.VAT.SellerVat.RoundupUnitPrice)
					}

					roundupUnitSellerShare = roundupUnitSellerShare.Add(*itemFinance.Invoice.Share.RoundupItemNet).
						Sub(*itemFinance.Invoice.Commission.RoundupUnitPrice).
						Sub(*itemFinance.Invoice.VAT.BusinessVat.RoundupUnitPrice)

					//roundupUnitSellerShare := rawUnitSellerShare.Ceil()
					itemFinance.Invoice.Share.RoundupUnitSellerShare = &roundupUnitSellerShare

					roundupTotalSellerShare := roundupUnitSellerShare.Mul(decimal.NewFromInt32(itemFinance.Quantity))
					itemFinance.Invoice.Share.RoundupTotalSellerShare = &roundupTotalSellerShare

					// business share raw
					rawUnitBusinessShare := (*itemFinance.Invoice.Commission.RawUnitPrice).
						Add(*itemFinance.Invoice.VAT.BusinessVat.RawUnitPrice)

					if itemFinance.Invoice.SSO.RawUnitPrice != nil {
						rawUnitBusinessShare = rawUnitBusinessShare.Sub(*itemFinance.Invoice.SSO.RawUnitPrice)
					}

					itemFinance.Invoice.Share.RawUnitBusinessShare = &rawUnitBusinessShare
					rawTotalBusinessShare := rawUnitBusinessShare.Mul(decimal.NewFromInt32(itemFinance.Quantity))
					itemFinance.Invoice.Share.RawTotalBusinessShare = &rawTotalBusinessShare

					// business share roundup
					roundupUnitBusinessShare := (*itemFinance.Invoice.Commission.RoundupUnitPrice).
						Add(*itemFinance.Invoice.VAT.BusinessVat.RoundupUnitPrice)

					if itemFinance.Invoice.SSO.RoundupUnitPrice != nil {
						roundupUnitBusinessShare = roundupUnitBusinessShare.Sub(*itemFinance.Invoice.SSO.RoundupUnitPrice)
					}

					//roundupUnitBusinessShare := rawUnitBusinessShare.Ceil()
					itemFinance.Invoice.Share.RoundupUnitBusinessShare = &roundupUnitBusinessShare

					roundupTotalBusinessShare := roundupUnitBusinessShare.Mul(decimal.NewFromInt32(itemFinance.Quantity))
					itemFinance.Invoice.Share.RoundupTotalBusinessShare = &roundupTotalBusinessShare

					itemFinance.Invoice.Share.UpdatedAt = finance.timestamp

					rawPkgSellerShare = rawPkgSellerShare.Add(*itemFinance.Invoice.Share.RawTotalSellerShare)
					roundupPkgSellerShare = roundupPkgSellerShare.Add(*itemFinance.Invoice.Share.RoundupTotalSellerShare)

					rawPkgBusinessShare = rawPkgBusinessShare.Add(*itemFinance.Invoice.Share.RawTotalBusinessShare)
					roundupPkgBusinessShare = roundupPkgBusinessShare.Add(*itemFinance.Invoice.Share.RoundupTotalBusinessShare)
				}
			}

			order.Packages[i].Invoice.Share.RawBusinessShare = &rawPkgBusinessShare
			order.Packages[i].Invoice.Share.RoundupBusinessShare = &roundupPkgBusinessShare

			order.Packages[i].Invoice.Share.RawSellerShare = &rawPkgSellerShare
			order.Packages[i].Invoice.Share.RoundupSellerShare = &roundupPkgSellerShare

			order.Packages[i].Invoice.Share.UpdatedAt = finance.timestamp

			// calculate order business share
			rawTotalShare = rawTotalShare.Add(*order.Packages[i].Invoice.Share.RawBusinessShare)
			roundupTotalShare = roundupTotalShare.Add(*order.Packages[i].Invoice.Share.RoundupBusinessShare)
		}

		if mode == ORDER_FINANCE || mode == SELLER_FINANCE {
			if order.Status != inProgressStatus {
				order.Invoice.Share.UpdatedAt = finance.timestamp
				order.Invoice.Share.RawTotalShare = &rawTotalShare
				order.Invoice.Share.RoundupTotalShare = &roundupTotalShare
			}
		}
		return nil
	}
}
