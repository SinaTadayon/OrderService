package entities

type AddressInfo struct {
	Address 			string					`bson:"address"`
	Phone   			string					`bson:"phone"`
	Mobile   			string					`bson:"mobile"`
	Country 			string					`bson:"country"`
	City    			string					`bson:"city"`
	Province   			string					`bson:"province"`
	Neighbourhood		string					`bson:"neighbourhood"`
	Location			Location				`bson:"location"`
	ZipCode 			string					`bson:"zipCode"`
}

type Location struct {
	Type string    				`bson:"type"`
	Coordinates []float64 		`bson:"coordinates"`
}

