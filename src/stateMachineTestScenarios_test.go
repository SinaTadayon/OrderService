package main

//import (
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//)
//
//func TestPaymentFailed_AssertTrue(t *testing.T) {
//	path := []string{PaymentPending, PaymentFailed}
//	foundedRoutes, err := checkJourney(path, true, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPaymentSuccessRejectBySystem_AssertTrue(t *testing.T) {
//	path := []string{PaymentPending, PaymentSuccess, PaymentRejected, PayToBuyer}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPaymentSuccessRejectedByController_AssertTrue(t *testing.T) {
//	path := []string{PaymentPending, PaymentSuccess, PaymentControl, PaymentRejected}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPaymentAcceptedByController_AssertTrue(t *testing.T) {
//	path := []string{PaymentPending, PaymentSuccess, PaymentControl, SellerApprovalPending}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPaymentSuccess_AssertTrue(t *testing.T) {
//	path := []string{PaymentPending, PaymentSuccess, SellerApprovalPending}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestSellerApprovalRejected_AssertTrue(t *testing.T) {
//	path := []string{SellerApprovalPending, ShipmentRejectedBySeller, PayToBuyer}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPaymentSellerApprovalAccepted_AssertTrue(t *testing.T) {
//	path := []string{PaymentSuccess, SellerApprovalPending, ShipmentPending}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestShipmentShipped_AssertTrue(t *testing.T) {
//	path := []string{SellerApprovalPending, ShipmentPending, Shipped}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestShipmentDetailDelayed_AssertTrue(t *testing.T) {
//	path := []string{ShipmentPending, ShipmentDetailDelayed, Shipped}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestShipmentDetailDelayedCanceled_AssertTrue(t *testing.T) {
//	path := []string{ShipmentPending, ShipmentDetailDelayed, ShipmentCanceled, PayToBuyer}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestShipmentDelivered_AssertTrue(t *testing.T) {
//	path := []string{Shipped, ShipmentDelivered}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestShipmentDeliveryPending_AssertTrue(t *testing.T) {
//	path := []string{Shipped, ShipmentDeliveryPending, ShipmentDelivered}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestShipmentDeliveryDelayedCanceled_AssertTrue(t *testing.T) {
//	path := []string{Shipped, ShipmentDeliveryPending, ShipmentDeliveryDelayed, ShipmentCanceled}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestShipmentDeliveryDelayedDelivered_AssertTrue(t *testing.T) {
//	path := []string{Shipped, ShipmentDeliveryPending, ShipmentDeliveryDelayed, ShipmentDelivered}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestShipmentApproved_AssertTrue(t *testing.T) {
//	path := []string{ShipmentDelivered, ShipmentSuccess}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestShipmentReturnShipped_AssertTrue(t *testing.T) {
//	path := []string{ShipmentDelivered, ReturnShipmentPending, ReturnShipped}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestReturnShipmentDeliveryProblemShipped_AssertTrue(t *testing.T) {
//	path := []string{ShipmentDelivered, ShipmentDeliveryProblem, ReturnShipmentPending, ReturnShipmentDetailDelayed, ReturnShipped}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestReturnDeliveryProblemCanceled_AssertTrue(t *testing.T) {
//	path := []string{ShipmentDelivered, ShipmentDeliveryProblem, ReturnShipmentPending, ReturnShipmentDetailDelayed, ShipmentSuccess}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestShipmentSuccess_AssertTrue(t *testing.T) {
//	path := []string{ShipmentSuccess, PayToSeller}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestReturnDelivered_AssertTrue(t *testing.T) {
//	path := []string{ReturnShipped, ReturnShipmentDelivered}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestReturnDeliveryPending_AssertTrue(t *testing.T) {
//	path := []string{ReturnShipped, ReturnShipmentDeliveryPending, ReturnShipmentDelivered}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestReturnDeliveryDelayed_AssertTrue(t *testing.T) {
//	path := []string{ReturnShipped, ReturnShipmentDeliveryPending, ReturnShipmentDeliveryDelayed, ReturnShipmentDelivered}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestReturnCanceled_AssertTrue(t *testing.T) {
//	path := []string{ReturnShipped, ReturnShipmentDeliveryPending, ReturnShipmentDeliveryDelayed, ReturnShipmentCanceled}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestReturnDelayedSuccess_AssertTrue(t *testing.T) {
//	path := []string{ReturnShipped, ReturnShipmentDeliveryPending, ReturnShipmentDeliveryDelayed, ReturnShipmentDelivered, ReturnShipmentSuccess}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestReturnDeliveryProblem_AssertTrue(t *testing.T) {
//	path := []string{ReturnShipped, ReturnShipmentDeliveryPending, ReturnShipmentDeliveryDelayed, ReturnShipmentDelivered, ReturnShipmentDeliveryProblem, ReturnShipmentSuccess}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestReturnDeliveryCanceled_AssertTrue(t *testing.T) {
//	path := []string{ReturnShipped, ReturnShipmentDeliveryPending, ReturnShipmentDeliveryDelayed, ReturnShipmentDelivered, ReturnShipmentDeliveryProblem, ReturnShipmentCanceled}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestReturnCanceledPayment_AssertTrue(t *testing.T) {
//	path := []string{ReturnShipmentCanceled, PayToSeller}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestReturnSuccessPayment_AssertTrue(t *testing.T) {
//	path := []string{ReturnShipmentSuccess, PayToBuyer}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPayToSellerSuccess_AssertTrue(t *testing.T) {
//	path := []string{PayToSeller, PayToSellerSuccess}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPayToSellerFailed_AssertTrue(t *testing.T) {
//	path := []string{PayToSeller, PayToSellerFailed, PayToSellerSuccess}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPayToSellerRetry_AssertTrue(t *testing.T) {
//	path := []string{PayToSeller, PayToSellerFailed, PayToSeller, PayToSellerSuccess}
//	foundedRoutes, err := checkJourney(path, false, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPayToBuyerSuccess_AssertTrue(t *testing.T) {
//	path := []string{PayToBuyer, PayToBuyerSuccess}
//	foundedRoutes, err := checkJourney(path, true, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPayToBuyerFailed_AssertTrue(t *testing.T) {
//	path := []string{PayToBuyer, PayToBuyerFailed, PayToBuyerSuccess}
//	foundedRoutes, err := checkJourney(path, true, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPayToBuyerRetry_AssertTrue(t *testing.T) {
//	path := []string{PayToBuyer, PayToBuyerFailed, PayToBuyer, PayToBuyerSuccess}
//	foundedRoutes, err := checkJourney(path, true, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPayToMarketSuccess_AssertTrue(t *testing.T) {
//	path := []string{PayToMarket, PayToMarketSuccess}
//	foundedRoutes, err := checkJourney(path, true, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPayToMarketFailed_AssertTrue(t *testing.T) {
//	path := []string{PayToMarket, PayToMarketFailed, PayToMarketSuccess}
//	foundedRoutes, err := checkJourney(path, true, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
//func TestPayToMarketRetry_AssertTrue(t *testing.T) {
//	path := []string{PayToMarket, PayToMarketFailed, PayToMarket, PayToMarketSuccess}
//	foundedRoutes, err := checkJourney(path, true, false)
//	assert.Nil(t, err)
//	assert.Equal(t, len(path), foundedRoutes)
//}
