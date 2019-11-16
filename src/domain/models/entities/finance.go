package entities

type FinanceInfo struct {
	Iban          string `bson:"iban"`
	CardNumber    string `bson:"cardNumber"`
	AccountNumber string `bson:"accountNumber"`
	BankName      string `bson:"backName"`
}
