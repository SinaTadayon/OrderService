package entities

type BuyerInfo struct {
	FirstName  			string					`bson:"firstName"`
	LastName   			string					`bson:"lastName"`
	Mobile     			string					`bson:"mobile"`
	Email      			string					`bson:"email"`
	NationalId 			string					`bson:"nationalId"`
	Gender				string					`bson:"gender"`
	IP         			string					`bson:"ip"`
	FinanceInfo    		FinanceInfo				`bson:"financeInfo"`
	ShippingAddress    	AddressInfo				`bson:"shippingAddress"`
}

