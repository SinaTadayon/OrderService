package entities

type BuyerInfo struct {
	BuyerId         uint64                 `bson:"buyerId"`
	FirstName       string                 `bson:"firstName"`
	LastName        string                 `bson:"lastName"`
	Phone           string                 `bson:"phone"`
	Mobile          string                 `bson:"mobile"`
	Email           string                 `bson:"email"`
	NationalId      string                 `bson:"nationalId"`
	Gender          string                 `bson:"gender"`
	IP              string                 `bson:"ip"`
	FinanceInfo     FinanceInfo            `bson:"financeInfo"`
	ShippingAddress AddressInfo            `bson:"shippingAddress"`
	Extended        map[string]interface{} `bson:"ext"`
}

type FinanceInfo struct {
	Iban          string                 `bson:"iban"`
	CardNumber    string                 `bson:"cardNumber"`
	AccountNumber string                 `bson:"accountNumber"`
	BankName      string                 `bson:"backName"`
	Extended      map[string]interface{} `bson:"ext"`
}

type AddressInfo struct {
	FirstName     string                 `bson:"firstName"`
	LastName      string                 `bson:"lastName"`
	Address       string                 `bson:"address"`
	Phone         string                 `bson:"phone"`
	Mobile        string                 `bson:"mobile"`
	Country       string                 `bson:"country"`
	City          string                 `bson:"city"`
	Province      string                 `bson:"province"`
	Neighbourhood string                 `bson:"neighbourhood"`
	Location      *Location              `bson:"location"`
	ZipCode       string                 `bson:"zipCode"`
	Extended      map[string]interface{} `bson:"ext"`
}

type Location struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}
