package reports

type BackOfficeExportItems struct {
	ItemId      uint64
	InventoryId string
	ProductId   string
	BuyerId     uint64
	BuyerPhone  string
	SellerId    uint64
	SellerName  string
	Price       uint64
	Status      string
	CreatedAt   string
	UpdatedAt   string
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
