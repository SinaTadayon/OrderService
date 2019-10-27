package entities

import "time"

type SellerProfile struct {
	Title            	string					`bson:"title"`
	FirstName        	string					`bson:"firstName"`
	LastName         	string					`bson:"lastName"`
	Mobile           	string					`bson:"mobile"`
	Email            	string					`bson:"email"`
	NationalId       	string					`bson:"nationalId"`
	CompanyName      	string					`bson:"companyName"`
	RegistrationName 	string					`bson:"registrationName"`
	EconomicCode     	string					`bson:"economicCode"`
	Finance          	FinanceInfo				`bson:"finance"`
	ShippingAddress     AddressInfo				`bson:"shippingAddress"`
	CreatedAt			time.Time				`bson:"createdAt"`
}

