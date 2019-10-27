package entities

import "time"

type PaymentRequest struct {
	Amount				uint64				`bson:"amount"`
	Currency			string				`bson:"currency"`
	Gateway 			string				`bson:"gateway"`
	FinanceInfo			FinanceInfo			`bson:"financeInfo"`
	CreatedAt   		time.Time			`bson:"createdAt"`
}

type PaymentResponse struct {
	Result 				bool				`bson:"result"`
	Reason				string				`bson:"reason"`
	Description 		string				`bson:"description"`
	CallBackUrl			string				`bson:"callbackUrl"`
	InvoiceId			int64				`bson:"invoiceId"`
	PaymentId			string				`bson:"paymentId"`
	CreatedAt   		time.Time			`bson:"createdAt"`
}

type PaymentResult struct {
	Result 				bool				`bson:"result"`
	Reason				string				`bson:"reason"`
	PaymentId  			string				`bson:"paymentId"`
	InvoiceId 			int64				`bson:"invoiceId"`
	Amount    			uint64				`bson:"amount"`
	ReqBody   			string				`bson:"reqBody"`
	ResBody   			string				`bson:"resBody"`
	CardNumMask			string				`bson:"cardNumMask"`
	CreatedAt   		time.Time			`bson:"createdAt"`
}

