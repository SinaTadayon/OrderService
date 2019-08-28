package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckNextStep_AssertTrue(t *testing.T) {
	currentStep := PaymentControl
	nextStep := SellerApprovalPending
	assert.True(t, CheckNextState(currentStep, nextStep))
}

func TestCheckNextStep_AssertFalse(t *testing.T) {
	currentStep := PaymentPending
	nextStep := ShipmentPending
	assert.False(t, CheckNextState(currentStep, nextStep))
}
