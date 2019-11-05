package entities

type BackOfficeExportItems struct {
	ItemId 			string
	InventoryId		string
	ProductId		string
	BuyerId			string
	BuyerPhone		string
	SellerId		string
	SellerName		string
	Price			uint64
	Status          string
	CreatedAt		string
	UpdatedAt		string
}

type SellerExportOrders struct {
	OrderId 			string
	ItemId				string
	ProductId			string
	InventoryId			string
	PaidPrice			uint64
	CommissionAmount	uint64
	Category			string
	Status 				string
	CreatedAt			string
	UpdatedAt			string
}


