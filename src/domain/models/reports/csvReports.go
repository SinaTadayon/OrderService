package reports

type ExportOrderItems struct {
	SId               uint64
	InventoryId       string
	ProductId         string
	BuyerId           uint64
	BuyerPhone        string
	SellerId          uint64
	SellerDisplayName string
	Price             string
	VoucherAmount     string
	ShippingCost      string
	Status            string
	CreatedAt         string
	UpdatedAt         string
}

type SellerExportOrders struct {
	OrderId     uint64
	ItemId      uint64
	ProductId   string
	InventoryId string
	PaidPrice   uint64
	Commission  float32
	Category    string
	Status      string
	CreatedAt   string
	UpdatedAt   string
}
