package entities

import (
	"time"

	"gitlab.faza.io/order-project/order-service/domain/models"
)

// subpackage id same as sid
type Subpackage struct {
	SId       uint64                 `bson:"sid"`
	PId       uint64                 `bson:"pid"`
	OrderId   uint64                 `bson:"orderId"`
	Version   uint64                 `bson:"version"`
	Items     []*Item                `bson:"items"`
	Shipments *Shipment              `bson:"shipments"`
	Tracking  Progress               `bson:"tracking"`
	Status    string                 `bson:"status"`
	CreatedAt time.Time              `bson:"createdAt"`
	UpdatedAt time.Time              `bson:"updatedAt"`
	DeletedAt *time.Time             `bson:"deletedAt"`
	Extended  map[string]interface{} `bson:"ext"`
}

type Shipment struct {
	ShipmentDetail       *ShippingDetail       `bson:"shipmentDetail"`
	ReturnShipmentDetail *ReturnShippingDetail `bson:"returnShipmentDetail"`
}

type ShippingDetail struct {
	CourierName    string                 `bson:"courierName"`
	ShippingMethod string                 `bson:"shippingMethod"`
	TrackingNumber string                 `bson:"trackingNumber"`
	Image          string                 `bson:"image"`
	Description    string                 `bson:"description"`
	ShippedAt      *time.Time             `bson:"shippedDate"`
	CreatedAt      time.Time              `bson:"createdAt"`
	UpdatedAt      *time.Time             `bson:"updatedAt"`
	Extended       map[string]interface{} `bson:"ext"`
}

type ReturnShippingDetail struct {
	CourierName    string                 `bson:"courierName"`
	ShippingMethod string                 `bson:"shippingMethod"`
	TrackingNumber string                 `bson:"trackingNumber"`
	Image          string                 `bson:"image"`
	Description    string                 `bson:"description"`
	ShippedAt      *time.Time             `bson:"shippedDate"`
	RequestedAt    *time.Time             `bson:"requestedAt"`
	CreatedAt      time.Time              `bson:"createdAt"`
	UpdatedAt      *time.Time             `bson:"updatedAt"`
	Extended       map[string]interface{} `bson:"ext"`
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
	Reasons     []models.Reason        `bson:"reasons"`
	Attributes  map[string]*Attribute  `bson:"attributes"`
	Invoice     ItemInvoice            `bson:"invoice"`
	Extended    map[string]interface{} `bson:"ext"`
}

type Attribute struct {
	KeyTranslate   map[string]string `bson:"keyTranslate"`
	ValueTranslate map[string]string `bson:"valueTranslate"`
}

type ItemInvoice struct {
	Unit              Money                  `bson:"unit"`
	Total             Money                  `bson:"total"`
	Original          Money                  `bson:"original"`
	Special           Money                  `bson:"special"`
	Discount          Money                  `bson:"discount"`
	SellerCommission  float32                `bson:"sellerCommission"`
	Commission        *ItemCommission        `bson:"itemCommission"`
	Share             *ItemShare             `bson:"share"`
	ApplicableVoucher bool                   `bson:"applicableVoucher"`
	Voucher           *ItemVoucher           `bson:"voucher"`
	CartRule          *CartRule              `bson:"cartRule"`
	SSO               *ItemSSO               `bson:"sso"`
	VAT               *ItemVAT               `bson:"vat"`
	TAX               *TAX                   `bson:"tax"`
	Extended          map[string]interface{} `bson:"ext"`
}

type ItemShare struct {
	RawItemGross              *Money                 `bson:"rawItemGross"`
	RoundupItemGross          *Money                 `bson:"roundupItemGross"`
	RawTotalGross             *Money                 `bson:"rawTotalGross"`
	RoundupTotalGross         *Money                 `bson:"roundupTotalGross"`
	RawItemNet                *Money                 `bson:"rawItemNet"`
	RoundupItemNet            *Money                 `bson:"roundupItemNet"`
	RawTotalNet               *Money                 `bson:"rawTotalNet"`
	RoundupTotalNet           *Money                 `bson:"roundupTotalNet"`
	RawUnitBusinessShare      *Money                 `bson:"rawUnitBusinessShare"`
	RoundupUnitBusinessShare  *Money                 `bson:"roundupUnitBusinessShare"`
	RawTotalBusinessShare     *Money                 `bson:"rawTotalBusinessShare"`
	RoundupTotalBusinessShare *Money                 `bson:"roundupTotalBusinessShare"`
	RawUnitSellerShare        *Money                 `bson:"rawUnitSellerShare"`
	RoundupUnitSellerShare    *Money                 `bson:"roundupUnitSellerShare"`
	RawTotalSellerShare       *Money                 `bson:"rawTotalSellerShare"`
	RoundupTotalSellerShare   *Money                 `bson:"roundupTotalSellerShare"`
	CreatedAt                 *time.Time             `bson:"createdAt"`
	UpdatedAt                 *time.Time             `bson:"updatedAt"`
	Extended                  map[string]interface{} `bson:"ext"`
}

type ItemCommission struct {
	ItemCommission    float32                `bson:"itemCommission"`
	RawUnitPrice      *Money                 `bson:"rawUnitPrice"`
	RoundupUnitPrice  *Money                 `bson:"roundupUnitPrice"`
	RawTotalPrice     *Money                 `bson:"rawTotalPrice"`
	RoundupTotalPrice *Money                 `bson:"roundupTotalPrice"`
	CreatedAt         *time.Time             `bson:"createdAt"`
	UpdatedAt         *time.Time             `bson:"updatedAt"`
	Extended          map[string]interface{} `bson:"ext"`
}

type ItemSSO struct {
	RawUnitPrice      *Money                 `bson:"rawUnitPrice"`
	RoundupUnitPrice  *Money                 `bson:"roundupUnitPrice"`
	RawTotalPrice     *Money                 `bson:"rawTotalPrice"`
	RoundupTotalPrice *Money                 `bson:"roundupTotalPrice"`
	CreatedAt         *time.Time             `bson:"createdAt"`
	UpdatedAt         *time.Time             `bson:"updatedAt"`
	Extended          map[string]interface{} `bson:"ext"`
}

type ItemVoucher struct {
	RawUnitPrice      *Money                 `bson:"rawUnitPrice"`
	RoundupUnitPrice  *Money                 `bson:"roundupUnitPrice"`
	RawTotalPrice     *Money                 `bson:"rawTotalPrice"`
	RoundupTotalPrice *Money                 `bson:"roundupTotalPrice"`
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
	RawUnitPrice      *Money                 `bson:"rawUnitPrice"`
	RoundupUnitPrice  *Money                 `bson:"roundupUnitPrice"`
	RawTotalPrice     *Money                 `bson:"rawTotalPrice"`
	RoundupTotalPrice *Money                 `bson:"roundupTotalPrice"`
	CreatedAt         *time.Time             `bson:"createdAt"`
	UpdatedAt         *time.Time             `bson:"updatedAt"`
	Extended          map[string]interface{} `bson:"ext"`
}

type BusinessVAT struct {
	Rate              float32                `bson:"rate"`
	RawUnitPrice      *Money                 `bson:"rawUnitPrice"`
	RoundupUnitPrice  *Money                 `bson:"roundupUnitPrice"`
	RawTotalPrice     *Money                 `bson:"rawTotalPrice"`
	RoundupTotalPrice *Money                 `bson:"roundupTotalPrice"`
	CreatedAt         *time.Time             `bson:"createdAt"`
	UpdatedAt         *time.Time             `bson:"updatedAt"`
	Extended          map[string]interface{} `bson:"ext"`
}

type Progress struct {
	State    *State                 `bson:"state"`
	Action   *Action                `bson:"action"`
	History  []State                `bson:"history"`
	Extended map[string]interface{} `bson:"ext"`
}

type State struct {
	Name       string                 `bson:"name"`
	Index      int                    `bson:"index"`
	Schedulers []*SchedulerData       `bson:"schedulers"`
	Data       map[string]interface{} `bson:"data"`
	Actions    []Action               `bson:"actions"`
	CreatedAt  time.Time              `bson:"createdAt"`
	UpdatedAt  time.Time              `bson:"updatedAt"`
	Extended   map[string]interface{} `bson:"ext"`
}

type SchedulerData struct {
	OId        uint64                 `bson:"oid"`
	PId        uint64                 `bson:"pid"`
	SId        uint64                 `bson:"sid"`
	StateName  string                 `bson:"stateName"`
	StateIndex int                    `bson:"stateIndex"`
	Name       string                 `bson:"name"`
	Group      string                 `bson:"group"`
	Action     string                 `bson:"action"`
	Index      int32                  `bson:"index"`
	Retry      int32                  `bson:"retry"`
	Cron       string                 `bson:"cron"`
	Start      *time.Time             `bson:"start"`
	End        *time.Time             `bson:"end"`
	Type       string                 `bson:"type"`
	Mode       string                 `bson:"mode"`
	Policy     interface{}            `bson:"policy"`
	Enabled    bool                   `bson:"enabled"`
	Data       interface{}            `bson:"data"`
	CreatedAt  time.Time              `bson:"createdAt"`
	UpdatedAt  time.Time              `bson:"updatedAt"`
	DeletedAt  *time.Time             `bson:"deletedAt"`
	Extended   map[string]interface{} `bson:"ext"`
}

/*
 Actions sample:
	ActionName: ApprovedAction
	UTP: SellerInfoActor
	Get: "sample data"
*/
type Action struct {
	Name      string                 `bson:"name"`
	Type      string                 `bson:"type"`
	UId       uint64                 `bson:"uid"`
	UTP       string                 `bson:"utp"`
	Perm      string                 `bson:"perm"`
	Priv      string                 `bson:"priv"`
	Policy    string                 `bson:"policy"`
	Result    string                 `bson:"result"`
	Reasons   []models.Reason        `bson:"reasons"`
	Note      string                 `bson:"note"`
	Data      map[string]interface{} `bson:"data"`
	CreatedAt time.Time              `bson:"createdAt"`
	Extended  map[string]interface{} `bson:"ext"`
}

type StockActionData struct {
	InventoryId string `bson:"inventoryId"`
	Quantity    int    `bson:"quantity"`
	Result      bool   `bson:"result"`
}

func (item Item) DeepCopy() *Item {
	newItem := Item{
		SKU:         item.SKU,
		InventoryId: item.InventoryId,
		Title:       item.Title,
		Brand:       item.Brand,
		Guaranty:    item.Guaranty,
		Category:    item.Category,
		Image:       item.Image,
		Returnable:  item.Returnable,
		Quantity:    item.Quantity,
		Reasons:     nil,
		Attributes:  nil,
		Extended:    item.Extended,
		Invoice: ItemInvoice{
			Unit:              item.Invoice.Unit,
			Total:             item.Invoice.Total,
			Original:          item.Invoice.Original,
			Special:           item.Invoice.Special,
			Discount:          item.Invoice.Discount,
			SellerCommission:  item.Invoice.SellerCommission,
			Commission:        nil,
			Share:             nil,
			ApplicableVoucher: item.Invoice.ApplicableVoucher,
			Voucher:           nil,
			CartRule:          nil,
			SSO:               nil,
			VAT:               nil,
			TAX:               nil,
			Extended:          item.Extended,
		},
	}

	if item.Invoice.Commission != nil {
		newItem.Invoice.Commission = &ItemCommission{
			ItemCommission:    item.Invoice.Commission.ItemCommission,
			RawUnitPrice:      nil,
			RoundupUnitPrice:  nil,
			RawTotalPrice:     nil,
			RoundupTotalPrice: nil,
			CreatedAt:         item.Invoice.Commission.CreatedAt,
			UpdatedAt:         item.Invoice.Commission.UpdatedAt,
			Extended:          item.Invoice.Commission.Extended,
		}

		if item.Invoice.Commission.RawTotalPrice != nil {
			newItem.Invoice.Commission.RawUnitPrice = &Money{
				Amount:   item.Invoice.Commission.RawTotalPrice.Amount,
				Currency: item.Invoice.Commission.RawTotalPrice.Currency,
			}
		}

		if item.Invoice.Commission.RoundupUnitPrice != nil {
			newItem.Invoice.Commission.RoundupUnitPrice = &Money{
				Amount:   item.Invoice.Commission.RoundupUnitPrice.Amount,
				Currency: item.Invoice.Commission.RoundupUnitPrice.Currency,
			}
		}

		if item.Invoice.Commission.RawTotalPrice != nil {
			newItem.Invoice.Commission.RawTotalPrice = &Money{
				Amount:   item.Invoice.Commission.RawTotalPrice.Amount,
				Currency: item.Invoice.Commission.RawTotalPrice.Currency,
			}
		}

		if item.Invoice.Commission.RoundupTotalPrice != nil {
			newItem.Invoice.Commission.RoundupTotalPrice = &Money{
				Amount:   item.Invoice.Commission.RoundupTotalPrice.Amount,
				Currency: item.Invoice.Commission.RoundupTotalPrice.Currency,
			}
		}
	}

	if item.Invoice.Share != nil {
		newItem.Invoice.Share = &ItemShare{
			RawItemGross:              nil,
			RoundupItemGross:          nil,
			RawTotalGross:             nil,
			RoundupTotalGross:         nil,
			RawItemNet:                nil,
			RoundupItemNet:            nil,
			RawTotalNet:               nil,
			RoundupTotalNet:           nil,
			RawUnitBusinessShare:      nil,
			RoundupUnitBusinessShare:  nil,
			RawTotalBusinessShare:     nil,
			RoundupTotalBusinessShare: nil,
			RawUnitSellerShare:        nil,
			RoundupUnitSellerShare:    nil,
			RawTotalSellerShare:       nil,
			RoundupTotalSellerShare:   nil,
			CreatedAt:                 item.Invoice.Share.CreatedAt,
			UpdatedAt:                 item.Invoice.Share.UpdatedAt,
			Extended:                  item.Invoice.Share.Extended,
		}

		if item.Invoice.Share.RawItemGross != nil {
			newItem.Invoice.Share.RawItemGross = &Money{
				Amount:   item.Invoice.Share.RawItemGross.Amount,
				Currency: item.Invoice.Share.RawItemGross.Currency,
			}
		}

		if item.Invoice.Share.RoundupItemGross != nil {
			newItem.Invoice.Share.RoundupItemGross = &Money{
				Amount:   item.Invoice.Share.RoundupItemGross.Amount,
				Currency: item.Invoice.Share.RoundupItemGross.Currency,
			}
		}

		if item.Invoice.Share.RawTotalGross != nil {
			newItem.Invoice.Share.RawTotalGross = &Money{
				Amount:   item.Invoice.Share.RawTotalGross.Amount,
				Currency: item.Invoice.Share.RawTotalGross.Currency,
			}
		}

		if item.Invoice.Share.RoundupTotalGross != nil {
			newItem.Invoice.Share.RoundupTotalGross = &Money{
				Amount:   item.Invoice.Share.RoundupTotalGross.Amount,
				Currency: item.Invoice.Share.RoundupTotalGross.Currency,
			}
		}

		if item.Invoice.Share.RawItemNet != nil {
			newItem.Invoice.Share.RawItemNet = &Money{
				Amount:   item.Invoice.Share.RawItemNet.Amount,
				Currency: item.Invoice.Share.RawItemNet.Currency,
			}
		}

		if item.Invoice.Share.RoundupItemNet != nil {
			newItem.Invoice.Share.RoundupItemNet = &Money{
				Amount:   item.Invoice.Share.RoundupItemNet.Amount,
				Currency: item.Invoice.Share.RoundupItemNet.Currency,
			}
		}

		if item.Invoice.Share.RawTotalNet != nil {
			newItem.Invoice.Share.RawTotalNet = &Money{
				Amount:   item.Invoice.Share.RawTotalNet.Amount,
				Currency: item.Invoice.Share.RawTotalNet.Currency,
			}
		}

		if item.Invoice.Share.RoundupTotalNet != nil {
			newItem.Invoice.Share.RoundupTotalNet = &Money{
				Amount:   item.Invoice.Share.RoundupTotalNet.Amount,
				Currency: item.Invoice.Share.RoundupTotalNet.Currency,
			}
		}

		if item.Invoice.Share.RawUnitBusinessShare != nil {
			newItem.Invoice.Share.RawUnitBusinessShare = &Money{
				Amount:   item.Invoice.Share.RawUnitBusinessShare.Amount,
				Currency: item.Invoice.Share.RawUnitBusinessShare.Currency,
			}
		}

		if item.Invoice.Share.RoundupUnitBusinessShare != nil {
			newItem.Invoice.Share.RoundupUnitBusinessShare = &Money{
				Amount:   item.Invoice.Share.RoundupUnitBusinessShare.Amount,
				Currency: item.Invoice.Share.RoundupUnitBusinessShare.Currency,
			}
		}

		if item.Invoice.Share.RawTotalBusinessShare != nil {
			newItem.Invoice.Share.RawTotalBusinessShare = &Money{
				Amount:   item.Invoice.Share.RawTotalBusinessShare.Amount,
				Currency: item.Invoice.Share.RawTotalBusinessShare.Currency,
			}
		}

		if item.Invoice.Share.RoundupTotalBusinessShare != nil {
			newItem.Invoice.Share.RoundupTotalBusinessShare = &Money{
				Amount:   item.Invoice.Share.RoundupTotalBusinessShare.Amount,
				Currency: item.Invoice.Share.RoundupTotalBusinessShare.Currency,
			}
		}

		if item.Invoice.Share.RawUnitSellerShare != nil {
			newItem.Invoice.Share.RawUnitSellerShare = &Money{
				Amount:   item.Invoice.Share.RawUnitSellerShare.Amount,
				Currency: item.Invoice.Share.RawUnitSellerShare.Currency,
			}
		}

		if item.Invoice.Share.RoundupUnitSellerShare != nil {
			newItem.Invoice.Share.RoundupUnitSellerShare = &Money{
				Amount:   item.Invoice.Share.RoundupUnitSellerShare.Amount,
				Currency: item.Invoice.Share.RoundupUnitSellerShare.Currency,
			}
		}

		if item.Invoice.Share.RawTotalSellerShare != nil {
			newItem.Invoice.Share.RawTotalSellerShare = &Money{
				Amount:   item.Invoice.Share.RawTotalSellerShare.Amount,
				Currency: item.Invoice.Share.RawTotalSellerShare.Currency,
			}
		}

		if item.Invoice.Share.RoundupTotalSellerShare != nil {
			newItem.Invoice.Share.RoundupTotalSellerShare = &Money{
				Amount:   item.Invoice.Share.RoundupTotalSellerShare.Amount,
				Currency: item.Invoice.Share.RoundupTotalSellerShare.Currency,
			}
		}
	}

	if item.Invoice.Voucher != nil {
		newItem.Invoice.Voucher = &ItemVoucher{
			RawUnitPrice:      nil,
			RoundupUnitPrice:  nil,
			RawTotalPrice:     nil,
			RoundupTotalPrice: nil,
			CreatedAt:         item.Invoice.Voucher.CreatedAt,
			UpdatedAt:         item.Invoice.Voucher.UpdatedAt,
			Extended:          item.Invoice.Voucher.Extended,
		}

		if item.Invoice.Voucher.RawUnitPrice != nil {
			newItem.Invoice.Voucher.RawUnitPrice = &Money{
				Amount:   item.Invoice.Voucher.RawUnitPrice.Amount,
				Currency: item.Invoice.Voucher.RawUnitPrice.Currency,
			}
		}

		if item.Invoice.Voucher.RoundupUnitPrice != nil {
			newItem.Invoice.Voucher.RoundupUnitPrice = &Money{
				Amount:   item.Invoice.Voucher.RoundupUnitPrice.Amount,
				Currency: item.Invoice.Voucher.RoundupUnitPrice.Currency,
			}
		}

		if item.Invoice.Voucher.RawTotalPrice != nil {
			newItem.Invoice.Voucher.RawTotalPrice = &Money{
				Amount:   item.Invoice.Voucher.RawTotalPrice.Amount,
				Currency: item.Invoice.Voucher.RawTotalPrice.Currency,
			}
		}

		if item.Invoice.Voucher.RoundupTotalPrice != nil {
			newItem.Invoice.Voucher.RoundupTotalPrice = &Money{
				Amount:   item.Invoice.Voucher.RoundupTotalPrice.Amount,
				Currency: item.Invoice.Voucher.RoundupTotalPrice.Currency,
			}
		}
	}

	if item.Invoice.SSO != nil {
		newItem.Invoice.SSO = &ItemSSO{
			RawUnitPrice:      nil,
			RoundupUnitPrice:  nil,
			RawTotalPrice:     nil,
			RoundupTotalPrice: nil,
			CreatedAt:         item.Invoice.SSO.CreatedAt,
			UpdatedAt:         item.Invoice.SSO.UpdatedAt,
			Extended:          item.Invoice.SSO.Extended,
		}

		if item.Invoice.SSO.RawUnitPrice != nil {
			newItem.Invoice.SSO.RawUnitPrice = &Money{
				Amount:   item.Invoice.SSO.RawUnitPrice.Amount,
				Currency: item.Invoice.SSO.RawUnitPrice.Currency,
			}
		}

		if item.Invoice.SSO.RoundupUnitPrice != nil {
			newItem.Invoice.SSO.RoundupUnitPrice = &Money{
				Amount:   item.Invoice.SSO.RoundupUnitPrice.Amount,
				Currency: item.Invoice.SSO.RoundupUnitPrice.Currency,
			}
		}

		if item.Invoice.SSO.RawTotalPrice != nil {
			newItem.Invoice.SSO.RawTotalPrice = &Money{
				Amount:   item.Invoice.SSO.RawTotalPrice.Amount,
				Currency: item.Invoice.SSO.RawTotalPrice.Currency,
			}
		}

		if item.Invoice.SSO.RoundupTotalPrice != nil {
			newItem.Invoice.SSO.RoundupTotalPrice = &Money{
				Amount:   item.Invoice.SSO.RoundupTotalPrice.Amount,
				Currency: item.Invoice.SSO.RoundupTotalPrice.Currency,
			}
		}
	}

	if item.Invoice.VAT != nil {
		newItem.Invoice.VAT = &ItemVAT{
			SellerVat:   nil,
			BusinessVat: nil,
			Extended:    item.Invoice.VAT.Extended,
		}

		if item.Invoice.VAT.SellerVat != nil {
			newItem.Invoice.VAT.SellerVat = &SellerVAT{
				Rate:              item.Invoice.VAT.SellerVat.Rate,
				IsObliged:         item.Invoice.VAT.SellerVat.IsObliged,
				RawUnitPrice:      nil,
				RoundupUnitPrice:  nil,
				RawTotalPrice:     nil,
				RoundupTotalPrice: nil,
				CreatedAt:         item.Invoice.VAT.SellerVat.CreatedAt,
				UpdatedAt:         item.Invoice.VAT.SellerVat.UpdatedAt,
				Extended:          item.Invoice.VAT.SellerVat.Extended,
			}

			if item.Invoice.VAT.SellerVat.RawUnitPrice != nil {
				newItem.Invoice.VAT.SellerVat.RawUnitPrice = &Money{
					Amount:   item.Invoice.VAT.SellerVat.RawUnitPrice.Amount,
					Currency: item.Invoice.VAT.SellerVat.RawUnitPrice.Currency,
				}
			}

			if item.Invoice.VAT.SellerVat.RoundupUnitPrice != nil {
				newItem.Invoice.VAT.SellerVat.RoundupUnitPrice = &Money{
					Amount:   item.Invoice.VAT.SellerVat.RoundupUnitPrice.Amount,
					Currency: item.Invoice.VAT.SellerVat.RoundupUnitPrice.Currency,
				}
			}

			if item.Invoice.VAT.SellerVat.RawTotalPrice != nil {
				newItem.Invoice.VAT.SellerVat.RawTotalPrice = &Money{
					Amount:   item.Invoice.VAT.SellerVat.RawTotalPrice.Amount,
					Currency: item.Invoice.VAT.SellerVat.RawTotalPrice.Currency,
				}
			}

			if item.Invoice.VAT.SellerVat.RoundupTotalPrice != nil {
				newItem.Invoice.VAT.SellerVat.RoundupTotalPrice = &Money{
					Amount:   item.Invoice.VAT.SellerVat.RoundupTotalPrice.Amount,
					Currency: item.Invoice.VAT.SellerVat.RoundupTotalPrice.Currency,
				}
			}
		}

		if item.Invoice.VAT.BusinessVat != nil {
			newItem.Invoice.VAT.BusinessVat = &BusinessVAT{
				Rate:              item.Invoice.VAT.BusinessVat.Rate,
				RawUnitPrice:      nil,
				RoundupUnitPrice:  nil,
				RawTotalPrice:     nil,
				RoundupTotalPrice: nil,
				CreatedAt:         item.Invoice.VAT.BusinessVat.CreatedAt,
				UpdatedAt:         item.Invoice.VAT.BusinessVat.UpdatedAt,
				Extended:          item.Invoice.VAT.BusinessVat.Extended,
			}

			if item.Invoice.VAT.BusinessVat.RawUnitPrice != nil {
				newItem.Invoice.VAT.BusinessVat.RawUnitPrice = &Money{
					Amount:   item.Invoice.VAT.BusinessVat.RawUnitPrice.Amount,
					Currency: item.Invoice.VAT.BusinessVat.RawUnitPrice.Currency,
				}
			}

			if item.Invoice.VAT.BusinessVat.RoundupUnitPrice != nil {
				newItem.Invoice.VAT.BusinessVat.RoundupUnitPrice = &Money{
					Amount:   item.Invoice.VAT.BusinessVat.RoundupUnitPrice.Amount,
					Currency: item.Invoice.VAT.BusinessVat.RoundupUnitPrice.Currency,
				}
			}

			if item.Invoice.VAT.BusinessVat.RawTotalPrice != nil {
				newItem.Invoice.VAT.BusinessVat.RawTotalPrice = &Money{
					Amount:   item.Invoice.VAT.BusinessVat.RawTotalPrice.Amount,
					Currency: item.Invoice.VAT.BusinessVat.RawTotalPrice.Currency,
				}
			}

			if item.Invoice.VAT.BusinessVat.RoundupTotalPrice != nil {
				newItem.Invoice.VAT.BusinessVat.RoundupTotalPrice = &Money{
					Amount:   item.Invoice.VAT.BusinessVat.RoundupTotalPrice.Amount,
					Currency: item.Invoice.VAT.BusinessVat.RoundupTotalPrice.Currency,
				}
			}
		}
	}

	if item.Attributes != nil {
		newItem.Attributes = make(map[string]*Attribute, len(item.Attributes))
		for attrKey, attribute := range item.Attributes {
			keyTranslates := make(map[string]string, len(attribute.KeyTranslate))
			for keyTran, value := range attribute.KeyTranslate {
				keyTranslates[keyTran] = value
			}
			valTranslates := make(map[string]string, len(attribute.ValueTranslate))
			for valTran, value := range attribute.ValueTranslate {
				valTranslates[valTran] = value
			}
			newItem.Attributes[attrKey] = &Attribute{
				KeyTranslate:   keyTranslates,
				ValueTranslate: valTranslates,
			}
		}
	}

	if item.Reasons != nil {
		newItem.Reasons = make([]models.Reason, 0, len(item.Reasons))
		for _, reason := range item.Reasons {
			newItem.Reasons = append(newItem.Reasons, reason)
		}
	}

	return &newItem
}

func (subpackage Subpackage) DeepCopy() *Subpackage {
	var subPkg = Subpackage{
		SId:       subpackage.SId,
		PId:       subpackage.PId,
		OrderId:   subpackage.OrderId,
		Version:   subpackage.Version,
		Items:     nil,
		Shipments: nil,
		Tracking:  Progress{},
		Status:    subpackage.Status,
		CreatedAt: subpackage.CreatedAt,
		UpdatedAt: subpackage.UpdatedAt,
		DeletedAt: subpackage.DeletedAt,
		Extended:  subpackage.Extended,
	}

	subPkg.Items = make([]*Item, 0, len(subpackage.Items))
	for _, item := range subpackage.Items {
		newItem := item.DeepCopy()
		subPkg.Items = append(subPkg.Items, newItem)
	}

	if subpackage.Shipments != nil {
		subPkg.Shipments = &Shipment{}
		if subpackage.Shipments.ShipmentDetail != nil {
			subPkg.Shipments.ShipmentDetail = &ShippingDetail{
				CourierName:    subpackage.Shipments.ShipmentDetail.CourierName,
				ShippingMethod: subpackage.Shipments.ShipmentDetail.ShippingMethod,
				TrackingNumber: subpackage.Shipments.ShipmentDetail.TrackingNumber,
				Image:          subpackage.Shipments.ShipmentDetail.Image,
				Description:    subpackage.Shipments.ShipmentDetail.Description,
				ShippedAt:      subpackage.Shipments.ShipmentDetail.ShippedAt,
				CreatedAt:      subpackage.Shipments.ShipmentDetail.CreatedAt,
				UpdatedAt:      subpackage.Shipments.ShipmentDetail.UpdatedAt,
				Extended:       subpackage.Shipments.ShipmentDetail.Extended,
			}
		}

		if subpackage.Shipments.ReturnShipmentDetail != nil {
			subPkg.Shipments.ReturnShipmentDetail = &ReturnShippingDetail{
				CourierName:    subpackage.Shipments.ReturnShipmentDetail.CourierName,
				ShippingMethod: subpackage.Shipments.ReturnShipmentDetail.ShippingMethod,
				TrackingNumber: subpackage.Shipments.ReturnShipmentDetail.TrackingNumber,
				Image:          subpackage.Shipments.ReturnShipmentDetail.Image,
				Description:    subpackage.Shipments.ReturnShipmentDetail.Description,
				ShippedAt:      subpackage.Shipments.ReturnShipmentDetail.ShippedAt,
				CreatedAt:      subpackage.Shipments.ReturnShipmentDetail.CreatedAt,
				UpdatedAt:      subpackage.Shipments.ReturnShipmentDetail.UpdatedAt,
				Extended:       subpackage.Shipments.ReturnShipmentDetail.Extended,
			}
		}
	}

	subPkg.Tracking = Progress{
		State:    nil,
		Action:   nil,
		Extended: subpackage.Tracking.Extended,
		History:  nil,
	}

	if subpackage.Tracking.State != nil {
		subPkg.Tracking.State = &State{
			Name:       subpackage.Tracking.State.Name,
			Index:      subpackage.Tracking.State.Index,
			Data:       subpackage.Tracking.State.Data,
			Schedulers: subpackage.Tracking.State.Schedulers,
			Actions:    nil,
			CreatedAt:  subpackage.Tracking.State.CreatedAt,
			Extended:   subpackage.Tracking.Extended,
		}
	}

	if subpackage.Tracking.State.Schedulers != nil {
		subPkg.Tracking.State.Schedulers = make([]*SchedulerData, 0, len(subpackage.Tracking.State.Schedulers))
		for _, schedulerData := range subpackage.Tracking.State.Schedulers {
			newSchedulerData := &SchedulerData{
				OId:        schedulerData.OId,
				PId:        schedulerData.PId,
				SId:        schedulerData.SId,
				StateName:  schedulerData.StateName,
				StateIndex: schedulerData.StateIndex,
				Name:       schedulerData.Name,
				Group:      schedulerData.Group,
				Action:     schedulerData.Action,
				Index:      schedulerData.Index,
				Retry:      schedulerData.Retry,
				Cron:       schedulerData.Cron,
				Start:      schedulerData.Start,
				End:        schedulerData.End,
				Type:       schedulerData.Type,
				Mode:       schedulerData.Mode,
				Policy:     schedulerData.Policy,
				Enabled:    schedulerData.Enabled,
				Data:       schedulerData.Data,
				CreatedAt:  schedulerData.CreatedAt,
				UpdatedAt:  schedulerData.UpdatedAt,
				DeletedAt:  schedulerData.DeletedAt,
				Extended:   schedulerData.Extended,
			}

			subPkg.Tracking.State.Schedulers = append(subPkg.Tracking.State.Schedulers, newSchedulerData)
		}
	}

	if subpackage.Tracking.Action != nil {
		subPkg.Tracking.Action = &Action{
			Name:      subpackage.Tracking.Action.Name,
			Type:      subpackage.Tracking.Action.Type,
			UId:       subpackage.Tracking.Action.UId,
			UTP:       subpackage.Tracking.Action.UTP,
			Perm:      subpackage.Tracking.Action.Perm,
			Priv:      subpackage.Tracking.Action.Priv,
			Policy:    subpackage.Tracking.Action.Policy,
			Result:    subpackage.Tracking.Action.Result,
			Reasons:   subpackage.Tracking.Action.Reasons,
			Note:      subpackage.Tracking.Action.Note,
			Data:      subpackage.Tracking.Action.Data,
			CreatedAt: subpackage.Tracking.Action.CreatedAt,
			Extended:  subpackage.Tracking.Action.Extended,
		}
		if subpackage.Tracking.Action.Reasons != nil {
			subPkg.Tracking.Action.Reasons = make([]models.Reason, 0, len(subpackage.Tracking.Action.Reasons))
			for _, reason := range subpackage.Tracking.Action.Reasons {
				subPkg.Tracking.Action.Reasons = append(subPkg.Tracking.Action.Reasons, reason)
			}
		}
	}

	if subpackage.Tracking.State.Actions != nil {
		subPkg.Tracking.State.Actions = make([]Action, 0, len(subpackage.Tracking.State.Actions))
		for _, action := range subpackage.Tracking.State.Actions {
			newAction := Action{
				Name:      action.Name,
				Type:      action.Type,
				UId:       action.UId,
				UTP:       action.UTP,
				Perm:      action.Perm,
				Priv:      action.Priv,
				Policy:    action.Policy,
				Result:    action.Result,
				Reasons:   action.Reasons,
				Data:      action.Data,
				CreatedAt: action.CreatedAt,
				Extended:  action.Extended,
			}
			if action.Reasons != nil {
				newAction.Reasons = make([]models.Reason, 0, len(action.Reasons))
				for _, reason := range action.Reasons {
					newAction.Reasons = append(action.Reasons, reason)
				}
			}
			subPkg.Tracking.State.Actions = append(subPkg.Tracking.State.Actions, newAction)
		}
	}

	if subpackage.Tracking.History != nil {
		subPkg.Tracking.History = make([]State, 0, len(subpackage.Tracking.History))
		for _, state := range subpackage.Tracking.History {
			var newState State
			newState.Actions = make([]Action, 0, len(state.Actions))
			for _, action := range state.Actions {
				newState.Actions = append(newState.Actions, action)
			}
			subPkg.Tracking.History = append(subPkg.Tracking.History, state)
		}
	}
	return &subPkg
}
