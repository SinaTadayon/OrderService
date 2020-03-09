package entities

import "time"

type PackageItem struct {
	PId             uint64                 `bson:"pid"`
	OrderId         uint64                 `bson:"orderId"`
	Version         uint64                 `bson:"version"`
	Invoice         PackageInvoice         `bson:"invoice"`
	SellerInfo      *SellerProfile         `bson:"sellerInfo"`
	ShopName        string                 `bson:"shopName"`
	ShippingAddress AddressInfo            `bson:"shippingAddress"`
	ShipmentSpec    ShipmentSpec           `bson:"shipmentSpec"`
	PayToSeller     []PayToSellerInfo      `bson:"payToSeller"`
	Subpackages     []*Subpackage          `bson:"subpackages"`
	Status          string                 `bson:"status"`
	CreatedAt       time.Time              `bson:"createdAt"`
	UpdatedAt       time.Time              `bson:"updatedAt"`
	DeletedAt       *time.Time             `bson:"deletedAt"`
	Extended        map[string]interface{} `bson:"ext"`
}

type PayToSellerInfo struct {
	PaymentRequest  *PaymentRequest        `bson:"paymentRequest"`
	PaymentResponse *PaymentResponse       `bson:"paymentResponse"`
	PaymentResult   *PaymentResult         `bson:"paymentResult"`
	Extended        map[string]interface{} `bson:"ext"`
}

type PackageInvoice struct {
	Subtotal       Money                  `bson:"subtotal"`
	Discount       Money                  `bson:"discount"`
	ShipmentAmount Money                  `bson:"shipmentAmount"`
	Share          *PackageShare          `bson:"share"`
	Commission     *PackageCommission     `bson:"packageCommission"`
	Voucher        *PackageVoucher        `bson:"voucher"`
	CartRule       *CartRule              `bson:"cartRule"`
	SSO            *PackageSSO            `bson:"sso"`
	VAT            *PackageVAT            `bson:"vat"`
	TAX            *TAX                   `bson:"tax"`
	Extended       map[string]interface{} `bson:"ext"`
}

type PackageCommission struct {
	RawTotalPrice     *Money                 `bson:"rawTotalPrice"`
	RoundupTotalPrice *Money                 `bson:"roundupTotalPrice"`
	CreatedAt         *time.Time             `bson:"createdAt"`
	UpdatedAt         *time.Time             `bson:"updatedAt"`
	Extended          map[string]interface{} `bson:"ext"`
}

type PackageShare struct {
	RawBusinessShare     *Money                 `bson:"rawBusinessShare"`
	RoundupBusinessShare *Money                 `bson:"roundupBusinessShare"`
	RawSellerShare       *Money                 `bson:"rawSellerShare"`
	RoundupSellerShare   *Money                 `bson:"roundupSellerShare"`
	CreatedAt            *time.Time             `bson:"createdAt"`
	UpdatedAt            *time.Time             `bson:"updatedAt"`
	Extended             map[string]interface{} `bson:"ext"`
}

type PackageVoucher struct {
	RawTotal                 *Money                 `bson:"rawTotal"`
	RoundupTotal             *Money                 `bson:"roundupTotal"`
	RawCalcShipmentPrice     *Money                 `bson:"rawCalcShipmentPrice"`
	RoundupCalcShipmentPrice *Money                 `bson:"roundupCalcShipmentPrice"`
	CreatedAt                *time.Time             `bson:"createdAt"`
	UpdatedAt                *time.Time             `bson:"updatedAt"`
	Extended                 map[string]interface{} `bson:"ext"`
}

type PackageSSO struct {
	Rate         float32                `bson:"rate"`
	IsObliged    bool                   `bson:"isObliged"`
	RawTotal     *Money                 `bson:"rawTotal"`
	RoundupTotal *Money                 `bson:"roundupTotal"`
	CreatedAt    *time.Time             `bson:"createdAt"`
	UpdatedAt    *time.Time             `bson:"updatedAt"`
	Extended     map[string]interface{} `bson:"ext"`
}

type PackageVAT struct {
	SellerVAT   *PackageSellerVAT      `bson:"sellerVat"`
	BusinessVAT *PackageBusinessVAT    `bson:"businessVat"`
	Extended    map[string]interface{} `bson:"ext"`
}

type PackageSellerVAT struct {
	RawTotal     *Money                 `bson:"rawTotal"`
	RoundupTotal *Money                 `bson:"roundupTotal"`
	CreatedAt    *time.Time             `bson:"createdAt"`
	UpdatedAt    *time.Time             `bson:"updatedAt"`
	Extended     map[string]interface{} `bson:"ext"`
}

type PackageBusinessVAT struct {
	RawTotal     *Money                 `bson:"rawTotal"`
	RoundupTotal *Money                 `bson:"roundupTotal"`
	CreatedAt    *time.Time             `bson:"createdAt"`
	UpdatedAt    *time.Time             `bson:"updatedAt"`
	Extended     map[string]interface{} `bson:"ext"`
}

// Time unit hours
type ShipmentSpec struct {
	CarrierNames   []string               `bson:"carrierNames"`
	CarrierProduct string                 `bson:"carrierProduct"`
	CarrierType    string                 `bson:"carrierType"`
	ShippingCost   *Money                 `bson:"shippingCost"`
	ReactionTime   int32                  `bson:"reactionTime"`
	ShippingTime   int32                  `bson:"shippingTime"`
	ReturnTime     int32                  `bson:"returnTime"`
	Details        string                 `bson:"details"`
	Extended       map[string]interface{} `bson:"ext"`
}
