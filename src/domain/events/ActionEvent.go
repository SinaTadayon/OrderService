package events

type ActionData struct {
	SubPackages    []ActionSubpackage
	Carrier        string
	TrackingNumber string
}

type ActionSubpackage struct {
	ItemId uint64
	Items  []ActionItem
}

type ActionItem struct {
	InventoryId string
	Quantity    int32
	Reasons     []string
}

type ActionResponse struct {
	OrderId uint64
	ItemsId []uint64
}
