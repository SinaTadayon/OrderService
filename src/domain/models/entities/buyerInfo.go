package entities

type BuyerInfo struct {
	BuyerId         uint64      `bson:"buyerId"`
	FirstName       string      `bson:"firstName"`
	LastName        string      `bson:"lastName"`
	Phone           string      `bson:"phone"`
	Mobile          string      `bson:"mobile"`
	Email           string      `bson:"email"`
	NationalId      string      `bson:"nationalId"`
	Gender          string      `bson:"gender"`
	IP              string      `bson:"ip"`
	FinanceInfo     FinanceInfo `bson:"financeInfo"`
	ShippingAddress AddressInfo `bson:"shippingAddress"`
}
