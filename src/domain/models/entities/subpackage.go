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
	StateName     string         `bson:"stateName"`
	StateIndex    int            `bson:"stateIndex"`
	Action        Action         `bson:"action"`
	StatesHistory []StateHistory `bson:"statesHistory"`
}

type StateHistory struct {
	Name          string    `bson:"name"`
	Index         int       `bson:"index"`
	ActionHistory []Action  `bson:"actionHistory"`
	CreatedAt     time.Time `bson:"createdAt"`
	UpdatedAt     time.Time `bson:"updatedAt"`
}

/*
 Action sample:
	ActionName: ApprovedAction
	Type: SellerInfoActor
	Get: "sample data"
*/
type Action struct {
	Name      string                 `bson:"name"`
	Type      string                 `bson:"type"`
	Data      map[string]interface{} `bson:"data"`
	Result    string                 `bson:"result"`
	Reasons   []string               `bson:"reasons"`
	CreatedAt time.Time              `bson:"createdAt"`
}

func (subpackage Subpackage) DeepCopy() *Subpackage {
	var subPkg = Subpackage{
		ItemId:    subpackage.ItemId,
		SellerId:  subpackage.SellerId,
		OrderId:   subpackage.OrderId,
		Version:   subpackage.Version,
		Status:    subpackage.Status,
		Items:     subpackage.Items,
		CreatedAt: subpackage.CreatedAt,
		UpdatedAt: subpackage.UpdatedAt,
		DeletedAt: subpackage.DeletedAt,
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
		StateName:  subpackage.Tracking.StateName,
		StateIndex: subpackage.Tracking.StateIndex,
		Action:     subpackage.Tracking.Action,
	}

	if subpackage.Tracking.StatesHistory != nil {
		subPkg.Tracking.StatesHistory = make([]StateHistory, 0, len(subpackage.Tracking.StatesHistory))
		for _, state := range subpackage.Tracking.StatesHistory {
			var newState StateHistory
			newState.ActionHistory = make([]Action, 0, len(state.ActionHistory))
			for _, action := range state.ActionHistory {
				newState.ActionHistory = append(newState.ActionHistory, action)
			}
			subPkg.Tracking.StatesHistory = append(subPkg.Tracking.StatesHistory, state)
		}
	}
	return &subPkg
}
