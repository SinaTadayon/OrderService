package entities

import "time"

// subpackage id same as sid
type Subpackage struct {
	SId       uint64                 `bson:"sid"`
	PId       uint64                 `bson:"pid"`
	OrderId   uint64                 `bson:"orderId"`
	Version   uint64                 `bson:"version"`
	Items     []Item                 `bson:"items"`
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
	CarrierName    string                 `bson:"carrierName"`
	ShippingMethod string                 `bson:"shippingMethod"`
	TrackingNumber string                 `bson:"trackingNumber"`
	Image          string                 `bson:"image"`
	Description    string                 `bson:"description"`
	ShippedAt      *time.Time             `bson:"shippedDate"`
	CreatedAt      time.Time              `bson:"createdAt"`
	Extended       map[string]interface{} `bson:"ext"`
}

type ReturnShippingDetail struct {
	CarrierName    string                 `bson:"carrierName"`
	ShippingMethod string                 `bson:"shippingMethod"`
	TrackingNumber string                 `bson:"trackingNumber"`
	Image          string                 `bson:"image"`
	Description    string                 `bson:"description"`
	ShippedAt      *time.Time             `bson:"shippedDate"`
	RequestedAt    *time.Time             `bson:"requestedAt"`
	CreatedAt      time.Time              `bson:"createdAt"`
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
	Reasons     []string               `bson:"reasons"`
	Attributes  map[string]string      `bson:"attributes"`
	Invoice     ItemInvoice            `bson:"invoice"`
	Extended    map[string]interface{} `bson:"ext"`
}

type ItemInvoice struct {
	Unit              Money                  `bson:"unit"`
	Total             Money                  `bson:"total"`
	Original          Money                  `bson:"original"`
	Special           Money                  `bson:"special"`
	Discount          Money                  `bson:"discount"`
	SellerCommission  float32                `bson:"sellerCommission"`
	ApplicableVoucher bool                   `bson:"applicableVoucher"`
	Voucher           *PackageVoucher        `bson:"voucher"`
	CartRule          *CartRule              `bson:"cartRule"`
	SSO               *SSO                   `bson:"sso"`
	VAT               *VAT                   `bson:"vat"`
	TAX               *TAX                   `bson:"tax"`
	Extended          map[string]interface{} `bson:"ext"`
}

type Progress struct {
	State    *State                 `bson:"state"`
	Action   *Action                `bson:"action"`
	History  []State                `bson:"history"`
	Extended map[string]interface{} `bson:"ext"`
}

type State struct {
	Name      string                 `bson:"name"`
	Index     int                    `bson:"index"`
	Data      map[string]interface{} `bson:"data"`
	Actions   []Action               `bson:"actions"`
	CreatedAt time.Time              `bson:"createdAt"`
	Extended  map[string]interface{} `bson:"ext"`
}

type SchedulerData struct {
	Name    string    `bson:"name"`
	Value   time.Time `bson:"value"`
	Action  string    `bson:"action"`
	Index   int32     `bson:"index"`
	Enabled bool      `bson:"enabled"`
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
	Reasons   []string               `bson:"reasons"`
	Data      map[string]interface{} `bson:"data"`
	CreatedAt time.Time              `bson:"createdAt"`
	Extended  map[string]interface{} `bson:"ext"`
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
		Attributes:  item.Attributes,
		Extended:    item.Extended,
		Invoice: ItemInvoice{
			Unit:              item.Invoice.Unit,
			Total:             item.Invoice.Total,
			Original:          item.Invoice.Original,
			Special:           item.Invoice.Special,
			Discount:          item.Invoice.Discount,
			SellerCommission:  item.Invoice.SellerCommission,
			ApplicableVoucher: item.Invoice.ApplicableVoucher,
			Voucher:           item.Invoice.Voucher,
			CartRule:          item.Invoice.CartRule,
			SSO:               item.Invoice.SSO,
			VAT:               item.Invoice.VAT,
			TAX:               item.Invoice.TAX,
			Extended:          item.Extended,
		},
	}
	if item.Reasons != nil {
		newItem.Reasons = make([]string, 0, len(item.Reasons))
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

	subPkg.Items = make([]Item, 0, len(subpackage.Items))
	for _, item := range subpackage.Items {
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
			Attributes:  item.Attributes,
			Extended:    item.Extended,
			Reasons:     nil,
			Invoice: ItemInvoice{
				Unit:              item.Invoice.Unit,
				Total:             item.Invoice.Total,
				Original:          item.Invoice.Original,
				Special:           item.Invoice.Special,
				Discount:          item.Invoice.Discount,
				SellerCommission:  item.Invoice.SellerCommission,
				ApplicableVoucher: item.Invoice.ApplicableVoucher,
				Voucher:           item.Invoice.Voucher,
				CartRule:          item.Invoice.CartRule,
				SSO:               item.Invoice.SSO,
				VAT:               item.Invoice.VAT,
				TAX:               item.Invoice.TAX,
				Extended:          item.Invoice.Extended,
			},
		}
		if item.Reasons != nil {
			newItem.Reasons = make([]string, 0, len(item.Reasons))
			for _, reason := range item.Reasons {
				newItem.Reasons = append(newItem.Reasons, reason)
			}
		}
		subPkg.Items = append(subPkg.Items, newItem)
	}

	if subpackage.Shipments != nil {
		subPkg.Shipments = &Shipment{}
		if subpackage.Shipments.ShipmentDetail != nil {
			subPkg.Shipments.ShipmentDetail = &ShippingDetail{
				CarrierName:    subpackage.Shipments.ShipmentDetail.CarrierName,
				ShippingMethod: subpackage.Shipments.ShipmentDetail.ShippingMethod,
				TrackingNumber: subpackage.Shipments.ShipmentDetail.TrackingNumber,
				Image:          subpackage.Shipments.ShipmentDetail.Image,
				Description:    subpackage.Shipments.ShipmentDetail.Description,
				ShippedAt:      subpackage.Shipments.ShipmentDetail.ShippedAt,
				CreatedAt:      subpackage.Shipments.ShipmentDetail.CreatedAt,
				Extended:       subpackage.Shipments.ShipmentDetail.Extended,
			}
		}

		if subpackage.Shipments.ReturnShipmentDetail != nil {
			subPkg.Shipments.ReturnShipmentDetail = &ReturnShippingDetail{
				CarrierName:    subpackage.Shipments.ReturnShipmentDetail.CarrierName,
				ShippingMethod: subpackage.Shipments.ReturnShipmentDetail.ShippingMethod,
				TrackingNumber: subpackage.Shipments.ReturnShipmentDetail.TrackingNumber,
				Image:          subpackage.Shipments.ReturnShipmentDetail.Image,
				Description:    subpackage.Shipments.ReturnShipmentDetail.Description,
				ShippedAt:      subpackage.Shipments.ReturnShipmentDetail.ShippedAt,
				CreatedAt:      subpackage.Shipments.ReturnShipmentDetail.CreatedAt,
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
			Name:      subpackage.Tracking.State.Name,
			Index:     subpackage.Tracking.State.Index,
			Data:      subpackage.Tracking.State.Data,
			Actions:   nil,
			CreatedAt: subpackage.Tracking.State.CreatedAt,
			Extended:  subpackage.Tracking.Extended,
		}
	}

	if subPkg.Tracking.Action != nil {
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
			Data:      subpackage.Tracking.Action.Data,
			CreatedAt: subpackage.Tracking.Action.CreatedAt,
			Extended:  subpackage.Tracking.Action.Extended,
		}
		if subpackage.Tracking.Action.Reasons != nil {
			subPkg.Tracking.Action.Reasons = make([]string, 0, len(subpackage.Tracking.Action.Reasons))
			for _, reason := range subpackage.Tracking.Action.Reasons {
				subPkg.Tracking.Action.Reasons = append(subPkg.Tracking.Action.Reasons, reason)
			}
		}
	}

	if subPkg.Tracking.State.Actions != nil {
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
				newAction.Reasons = make([]string, 0, len(action.Reasons))
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
