package main

import (
	"testing"
)

func createPaymentRequestSample() PaymentPendingRequest {
	var pr = PaymentPendingRequest{}

	pr.orderNumber = "102102"
	// buyer info
	pr.buyer.info.firstName = "farzan"
	pr.buyer.info.lastName = "dalaee"
	pr.buyer.info.email = "farzan.dalaee@gmail.com"
	pr.buyer.info.gender = "male"
	pr.buyer.info.mobile = "+98912193870"
	pr.buyer.info.nationalId = "0012345678"
	// buyer address
	pr.buyer.address.title = "Home"
	pr.buyer.address.phone = "+98912193870"
	pr.buyer.address.zipCode = "1651764614"
	pr.buyer.address.country = "Iran"
	pr.buyer.address.state = "Tehran"
	pr.buyer.address.city = "Tehran"
	pr.buyer.address.address = "Sheikh bahaee, p 5"
	pr.buyer.address.lat = "10.1345664"
	pr.buyer.address.lan = "22.1345664"
	// buyer finance
	pr.buyer.finance.iban = "IR165411211001514313143545354134"
	pr.buyer.finance.bankName = "saman"
	pr.buyer.finance.cartNumber = "6014 1111 2222 3333"
	// amount
	pr.amount.total = 200000
	pr.amount.discount = 40000
	pr.amount.paid = 160000
	// items
	pr.items[0].sku = "aaa000"
	pr.items[0].amount.total = 200000
	pr.items[0].amount.paid = 160000
	pr.items[0].amount.discount = 40000
	pr.items[0].amount.sellerCommission = 10000
	pr.items[0].amount.systemCommission = 5000

	return pr
}

func TestPaymentPendingMessageValidate(t *testing.T) {

}

/*func PaymentPendingMessageValidate(message *sarama.ConsumerMessage) (*sarama.ConsumerMessage, error) {
	return message, nil
}
func PaymentPendingAction(message *sarama.ConsumerMessage) error {

	err := PaymentPendingProduce("", []byte{})
	if err != nil {
		return err
	}
	return nil
}
func PaymentPendingProduce(topic string, payload []byte) error {
	return nil
}
*/
