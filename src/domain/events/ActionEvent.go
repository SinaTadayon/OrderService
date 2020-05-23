package events

import (
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
)

type ActionData struct {
	SubPackages    []ActionSubpackage
	Carrier        string
	TrackingNumber string
}

type ActionSubpackage struct {
	SId   uint64
	Items []ActionItem
}

type ActionItem struct {
	InventoryId string
	Quantity    int32
	Reasons     []entities.Reason
}

type ActionResponse struct {
	OrderId uint64
	SIds    []uint64
}
