package reports

import (
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"time"
)

type FinanceOrderItem struct {
	SId                      uint64          `bson:"sid"`
	PId                      uint64          `bson:"pid"`
	OrderId                  uint64          `bson:"orderId"`
	ShipmentAmount           entities.Money  `bson:"shipmentAmount"`
	RawSellerShippingNet     *entities.Money `bson:"rawSellerShippingNet"`
	RoundupSellerShippingNet *entities.Money `bson:"roundupSellerShippingNet"`
	Items                    []*Item         `bson:"items"`
	Status                   string          `bson:"status"`
	CreatedAt                time.Time       `bson:"createdAt"`
	UpdatedAt                time.Time       `bson:"updatedAt"`
	OrderCreatedAt           time.Time       `bson:"orderCreatedAt"`
}

type Item struct {
	SKU         string                 `bson:"sku"`
	InventoryId string                 `bson:"inventoryId"`
	Title       string                 `bson:"title"`
	Brand       string                 `bson:"brand"`
	Guaranty    string                 `bson:"guaranty"`
	Category    string                 `bson:"category"`
	Image       string                 `bson:"image"`
	Returnable  bool                   `bson:"returnable"`
	Quantity    int32                  `bson:"quantity"`
	Reasons     []entities.Reason      `bson:"reasons"`
	Attributes  map[string]*entities.Attribute  `bson:"attributes"`
	Invoice     ItemInvoice            `bson:"invoice"`
	Extended    map[string]interface{} `bson:"ext"`
}

type ItemInvoice struct {
	Unit              entities.Money         `bson:"unit"`
	Total             entities.Money         `bson:"total"`
	Original          entities.Money         `bson:"original"`
	Special           entities.Money         `bson:"special"`
	Discount          entities.Money         `bson:"discount"`
	SellerCommission  float32                `bson:"sellerCommission"`
	Commission        *ItemCommission        `bson:"itemCommission"`
	Share             *ItemShare             `bson:"share"`
	ApplicableVoucher bool                   `bson:"applicableVoucher"`
	Voucher           *ItemVoucher           `bson:"voucher"`
	CartRule          *entities.CartRule     `bson:"cartRule"`
	SSO               *ItemSSO               `bson:"sso"`
	VAT               *ItemVAT               `bson:"vat"`
	TAX               *entities.TAX          `bson:"tax"`
	Extended          map[string]interface{} `bson:"ext"`
}

type ItemShare struct {
	RawItemGross              *entities.Money        `bson:"rawItemGross"`
	RoundupItemGross          *entities.Money        `bson:"roundupItemGross"`
	RawTotalGross             *entities.Money        `bson:"rawTotalGross"`
	RoundupTotalGross         *entities.Money        `bson:"roundupTotalGross"`
	RawItemNet                *entities.Money        `bson:"rawItemNet"`
	RoundupItemNet            *entities.Money        `bson:"roundupItemNet"`
	RawTotalNet               *entities.Money        `bson:"rawTotalNet"`
	RoundupTotalNet           *entities.Money        `bson:"roundupTotalNet"`
	RawUnitBusinessShare      *entities.Money        `bson:"rawUnitBusinessShare"`
	RoundupUnitBusinessShare  *entities.Money        `bson:"roundupUnitBusinessShare"`
	RawTotalBusinessShare     *entities.Money        `bson:"rawTotalBusinessShare"`
	RoundupTotalBusinessShare *entities.Money        `bson:"roundupTotalBusinessShare"`
	RawUnitSellerShare        *entities.Money        `bson:"rawUnitSellerShare"`
	RoundupUnitSellerShare    *entities.Money        `bson:"roundupUnitSellerShare"`
	RawTotalSellerShare       *entities.Money        `bson:"rawTotalSellerShare"`
	RoundupTotalSellerShare   *entities.Money        `bson:"roundupTotalSellerShare"`
	CreatedAt                 *time.Time             `bson:"createdAt"`
	UpdatedAt                 *time.Time             `bson:"updatedAt"`
	Extended                  map[string]interface{} `bson:"ext"`
}

type ItemCommission struct {
	ItemCommission    float32                `bson:"itemCommission"`
	RawUnitPrice      *entities.Money        `bson:"rawUnitPrice"`
	RoundupUnitPrice  *entities.Money        `bson:"roundupUnitPrice"`
	RawTotalPrice     *entities.Money        `bson:"rawTotalPrice"`
	RoundupTotalPrice *entities.Money        `bson:"roundupTotalPrice"`
	CreatedAt         *time.Time             `bson:"createdAt"`
	UpdatedAt         *time.Time             `bson:"updatedAt"`
	Extended          map[string]interface{} `bson:"ext"`
}

type ItemSSO struct {
	Rate              float32                `bson:"rate"`
	IsObliged         bool                   `bson:"isObliged"`
	RawUnitPrice      *entities.Money        `bson:"rawUnitPrice"`
	RoundupUnitPrice  *entities.Money        `bson:"roundupUnitPrice"`
	RawTotalPrice     *entities.Money        `bson:"rawTotalPrice"`
	RoundupTotalPrice *entities.Money        `bson:"roundupTotalPrice"`
	CreatedAt         *time.Time             `bson:"createdAt"`
	UpdatedAt         *time.Time             `bson:"updatedAt"`
	Extended          map[string]interface{} `bson:"ext"`
}

type ItemVoucher struct {
	RawUnitPrice      *entities.Money        `bson:"rawUnitPrice"`
	RoundupUnitPrice  *entities.Money        `bson:"roundupUnitPrice"`
	RawTotalPrice     *entities.Money        `bson:"rawTotalPrice"`
	RoundupTotalPrice *entities.Money        `bson:"roundupTotalPrice"`
	CreatedAt         *time.Time             `bson:"createdAt"`
	UpdatedAt         *time.Time             `bson:"updatedAt"`
	Extended          map[string]interface{} `bson:"ext"`
}

type ItemVAT struct {
	SellerVat   *SellerVAT             `bson:"sellerVat"`
	BusinessVat *BusinessVAT           `bson:"businessVat"`
	Extended    map[string]interface{} `bson:"ext"`
}

type SellerVAT struct {
	Rate              float32                `bson:"rate"`
	IsObliged         bool                   `bson:"isObliged"`
	RawUnitPrice      *entities.Money        `bson:"rawUnitPrice"`
	RoundupUnitPrice  *entities.Money        `bson:"roundupUnitPrice"`
	RawTotalPrice     *entities.Money        `bson:"rawTotalPrice"`
	RoundupTotalPrice *entities.Money        `bson:"roundupTotalPrice"`
	CreatedAt         *time.Time             `bson:"createdAt"`
	UpdatedAt         *time.Time             `bson:"updatedAt"`
	Extended          map[string]interface{} `bson:"ext"`
}

type BusinessVAT struct {
	Rate              float32                `bson:"rate"`
	RawUnitPrice      *entities.Money        `bson:"rawUnitPrice"`
	RoundupUnitPrice  *entities.Money        `bson:"roundupUnitPrice"`
	RawTotalPrice     *entities.Money        `bson:"rawTotalPrice"`
	RoundupTotalPrice *entities.Money        `bson:"roundupTotalPrice"`
	CreatedAt         *time.Time             `bson:"createdAt"`
	UpdatedAt         *time.Time             `bson:"updatedAt"`
	Extended          map[string]interface{} `bson:"ext"`
}
