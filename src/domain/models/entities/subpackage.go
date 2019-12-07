package entities

import "time"

// subpackage id same as itemId
type Subpackage struct {
	ItemId    uint64     `bson:"itemId"`
	SellerId  uint64     `bson:"sellerId"`
	OrderId   uint64     `bson:"orderId"`
	Version   uint64     `bson:"version"`
	Items     []Item     `bson:"item"`
	Shipments *Shipment  `bson:"shipments"`
	Tracking  Progress   `bson:"tracking"`
	Status    string     `bson:"status"`
	CreatedAt time.Time  `bson:"createdAt"`
	UpdatedAt time.Time  `bson:"updatedAt"`
	DeletedAt *time.Time `bson:"deletedAt"`
}

type Shipment struct {
	ShipmentDetail       *ShippingDetail `bson:"shipmentDetail"`
	ReturnShipmentDetail *ShippingDetail `bson:"returnShipmentDetail"`
}

type ShippingDetail struct {
	CarrierName    string    `bson:"carrierName"`
	ShippingMethod string    `bson:"shippingMethod"`
	TrackingNumber string    `bson:"trackingNumber"`
	Image          string    `bson:"image"`
	Description    string    `bson:"description"`
	ShippedDate    time.Time `bson:"shippedDate"`
	CreatedAt      time.Time `bson:"createdAt"`
}

type Item struct {
	SKU         string            `bson:"sku"`
	InventoryId string            `bson:"inventoryId"`
	Title       string            `bson:"title"`
	Brand       string            `bson:"brand"`
	Guaranty    string            `bson:"guaranty"`
	Category    string            `bson:"category"`
	Image       string            `bson:"image"`
	Returnable  bool              `bson:"returnable"`
	Quantity    int32             `bson:"quantity"`
	Reasons     []string          `bson:"reasons"`
	Attributes  map[string]string `bson:"attributes"`
	Invoice     ItemInvoice       `bson:"invoice"`
}

type ItemInvoice struct {
	Unit              uint64  `bson:"unit"`
	Total             uint64  `bson:"total"`
	Original          uint64  `bson:"original"`
	Special           uint64  `bson:"special"`
	Discount          uint64  `bson:"discount"`
	SellerCommission  float32 `bson:"sellerCommission"`
	Currency          string  `bson:"currency"`
	ApplicableVoucher bool    `bson:"applicableVoucher"`
}

type Progress struct {
	State   *State  `bson:"state"`
	Action  *Action `bson:"action"`
	History []State `bson:"states"`
}

type State struct {
	Name      string                 `bson:"name"`
	Index     int                    `bson:"index"`
	Data      map[string]interface{} `bson:"data"`
	Actions   []Action               `bson:"action"`
	CreatedAt time.Time              `bson:"createdAt"`
}

/*
 Actions sample:
	ActionName: ApprovedAction
	Type: SellerInfoActor
	Get: "sample data"
*/
type Action struct {
	Name      string    `bson:"name"`
	Type      string    `bson:"type"`
	Result    string    `bson:"result"`
	Reasons   []string  `bson:"reasons"`
	CreatedAt time.Time `bson:"createdAt"`
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
		Attributes:  item.Attributes,
		Invoice: ItemInvoice{
			Unit:              item.Invoice.Unit,
			Total:             item.Invoice.Total,
			Original:          item.Invoice.Original,
			Special:           item.Invoice.Special,
			Discount:          item.Invoice.Discount,
			SellerCommission:  item.Invoice.SellerCommission,
			Currency:          item.Invoice.Currency,
			ApplicableVoucher: item.Invoice.ApplicableVoucher,
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
		ItemId:    subpackage.ItemId,
		SellerId:  subpackage.SellerId,
		OrderId:   subpackage.OrderId,
		Version:   subpackage.Version,
		Status:    subpackage.Status,
		CreatedAt: subpackage.CreatedAt,
		UpdatedAt: subpackage.UpdatedAt,
		DeletedAt: subpackage.DeletedAt,
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
			Invoice: ItemInvoice{
				Unit:              item.Invoice.Unit,
				Total:             item.Invoice.Total,
				Original:          item.Invoice.Original,
				Special:           item.Invoice.Special,
				Discount:          item.Invoice.Discount,
				SellerCommission:  item.Invoice.SellerCommission,
				Currency:          item.Invoice.Currency,
				ApplicableVoucher: item.Invoice.ApplicableVoucher,
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
				ShippedDate:    subpackage.Shipments.ShipmentDetail.ShippedDate,
				CreatedAt:      subpackage.Shipments.ShipmentDetail.CreatedAt,
			}
		}

		if subpackage.Shipments.ReturnShipmentDetail != nil {
			subPkg.Shipments.ReturnShipmentDetail = &ShippingDetail{
				CarrierName:    subpackage.Shipments.ReturnShipmentDetail.CarrierName,
				ShippingMethod: subpackage.Shipments.ReturnShipmentDetail.ShippingMethod,
				TrackingNumber: subpackage.Shipments.ReturnShipmentDetail.TrackingNumber,
				Image:          subpackage.Shipments.ReturnShipmentDetail.Image,
				Description:    subpackage.Shipments.ReturnShipmentDetail.Description,
				ShippedDate:    subpackage.Shipments.ReturnShipmentDetail.ShippedDate,
				CreatedAt:      subpackage.Shipments.ReturnShipmentDetail.CreatedAt,
			}
		}
	}

	subPkg.Tracking = Progress{
		State: &State{
			Name:      subpackage.Tracking.State.Name,
			Index:     subpackage.Tracking.State.Index,
			Data:      subpackage.Tracking.State.Data,
			Actions:   nil,
			CreatedAt: subpackage.Tracking.State.CreatedAt,
		},

		Action: &Action{
			Name:      subpackage.Tracking.Action.Name,
			Type:      subpackage.Tracking.Action.Type,
			Result:    subpackage.Tracking.Action.Result,
			Reasons:   nil,
			CreatedAt: subpackage.Tracking.Action.CreatedAt,
		},
	}

	if subPkg.Tracking.State.Actions != nil {
		subPkg.Tracking.State.Actions = make([]Action, 0, len(subpackage.Tracking.State.Actions))
		for _, action := range subpackage.Tracking.State.Actions {
			newAction := Action{
				Name:      action.Name,
				Type:      action.Type,
				Result:    action.Result,
				CreatedAt: action.CreatedAt,
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

	if subpackage.Tracking.Action.Reasons != nil {
		subPkg.Tracking.Action.Reasons = make([]string, 0, len(subpackage.Tracking.Action.Reasons))
		for _, reason := range subpackage.Tracking.Action.Reasons {
			subPkg.Tracking.Action.Reasons = append(subPkg.Tracking.Action.Reasons, reason)
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
