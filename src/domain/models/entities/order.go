package entities

import (
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"time"
)

const (
	randomMin int = 100
	randomMax int = 999
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//type ObjectId struct {
//	ID 	primitive.ObjectID `bson:"_id"`
//}

//Order Status: New, InProgress, Closed
type Order struct {
	ID             primitive.ObjectID `bson:"-"`
	OrderId        uint64             `bson:"orderId"`
	Version        uint64             `bson:"version"`
	Platform       string             `bson:"platform"`
	PaymentService []PaymentService   `bson:"paymentService"`
	SystemPayment  SystemPayment      `bson:"systemPayment"`
	Status         string             `bson:"status"`
	BuyerInfo      BuyerInfo          `bson:"buyerInfo"`
	Invoice        Invoice            `bson:"invoice"`
	Packages       []PackageItem      `bson:"packages"`
	CreatedAt      time.Time          `bson:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt"`
	DeletedAt      *time.Time         `bson:"deletedAt"`
}

type PaymentService struct {
	PaymentRequest  *PaymentRequest  `bson:"paymentRequest"`
	PaymentResponse *PaymentResponse `bson:"paymentResponse"`
	PaymentResult   *PaymentResult   `bson:"paymentResult"`
}

// TODO get configs of pay to market from siavash
type SystemPayment struct {
	PayToBuyer  []PayToBuyerInfo  `bson:"payToBuyer"`
	PayToSeller []PayToSellerInfo `bson:"payToSeller"`
	PayToMarket []PayToMarket     `bson:"payToMarket"`
}

type PayToBuyerInfo struct {
	PaymentRequest  *PaymentRequest  `bson:"paymentRequest"`
	PaymentResponse *PaymentResponse `bson:"paymentResponse"`
	PaymentResult   *PaymentResult   `bson:"paymentResult"`
}

type PayToSellerInfo struct {
	PaymentRequest  *PaymentRequest  `bson:"paymentRequest"`
	PaymentResponse *PaymentResponse `bson:"paymentResponse"`
	PaymentResult   *PaymentResult   `bson:"paymentResult"`
}

type PayToMarket struct {
	PaymentRequest  *PaymentRequest  `bson:"paymentRequest"`
	PaymentResponse *PaymentResponse `bson:"paymentResponse"`
	PaymentResult   *PaymentResult   `bson:"paymentResult"`
}

type Invoice struct {
	GrandTotal     uint64         `bson:"grandTotal"`
	Subtotal       uint64         `bson:"subtotal"`
	Discount       uint64         `bson:"discount"`
	ShipmentTotal  uint64         `bson:"shipmentTotal"`
	Currency       string         `bson:"currency"`
	PaymentMethod  string         `bson:"paymentMethod"`
	PaymentGateway string         `bson:"paymentGateway"`
	PaymentOption  *PaymentOption `bson:"paymentOption"`
	Voucher        *Voucher       `bson:"voucher"`
	CartRule       *CartRule      `bson:"cartRule"`
}

type PaymentOption struct {
}

type Voucher struct {
	Amount  float64         `bson:"amount"`
	Code    string          `bson:"code"`
	Details *VoucherDetails `bson:"details"`
}

type CartRule struct {
	Amount uint64 `bson:"amount"`
}

type VoucherDetails struct {
	StartDate        time.Time `bson:"startDate"`
	EndDate          time.Time `bson:"endDate"`
	Type             string    `bson:"type"`
	MaxDiscountValue uint64    `bson:"maxDiscountValue"`
	MinBasketValue   uint64    `bson:"minBasketValue"`
}

type PaymentRequest struct {
	Amount    uint64    `bson:"amount"`
	Currency  string    `bson:"currency"`
	Gateway   string    `bson:"gateway"`
	CreatedAt time.Time `bson:"createdAt"`
}

type PaymentResponse struct {
	Result      bool      `bson:"result"`
	Reason      string    `bson:"reason"`
	Description string    `bson:"description"`
	CallBackUrl string    `bson:"callbackUrl"`
	InvoiceId   int64     `bson:"invoiceId"`
	PaymentId   string    `bson:"paymentId"`
	CreatedAt   time.Time `bson:"createdAt"`
}

type PaymentResult struct {
	Result      bool      `bson:"result"`
	Reason      string    `bson:"reason"`
	PaymentId   string    `bson:"paymentId"`
	InvoiceId   int64     `bson:"invoiceId"`
	Amount      uint64    `bson:"amount"`
	CardNumMask string    `bson:"cardNumMask"`
	CreatedAt   time.Time `bson:"createdAt"`
}

func (order Order) IsIdEmpty() bool {
	for _, v := range order.ID {
		if v != 0 {
			return false
		}
	}
	return true
}

// TODO concurrency check
func GenerateOrderId() uint64 {
	var err error
	var bytes []byte
	var orderId uint32
	for {
		bytes, err = uuid.New().MarshalBinary()
		if err == nil {
			orderId = byteToHash(bytes)
			break
		}
	}
	//return strconv.FormatUint(uint64(orderId), 10)
	return uint64(orderId)
}

func byteToHash(bytes []byte) uint32 {
	var h uint32 = 0
	for _, val := range bytes {
		h = 31*h + uint32(val&0xff)
	}
	return h
}

func GenerateRandomNumber() uint32 {
	return uint32(rand.Intn(randomMax-randomMin+1) + randomMin)
}
