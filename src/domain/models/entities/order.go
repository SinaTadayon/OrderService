package entities

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	randomMin int = 100
	randomMax int = 999
)

const (
	DocumentVersion string = "1.0.8"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//type ObjectId struct {
//	ID 	primitive.ObjectID `bson:"_id"`
//}

//Order Status: New, InProgress, Closed
type Order struct {
	ID            primitive.ObjectID     `bson:"-"`
	OrderId       uint64                 `bson:"orderId"`
	Version       uint64                 `bson:"version"`
	DocVersion    string                 `bson:"docVersion"`
	Platform      string                 `bson:"platform"`
	OrderPayment  []PaymentService       `bson:"orderPayment"`
	SystemPayment SystemPayment          `bson:"systemPayment"`
	Status        string                 `bson:"status"`
	BuyerInfo     BuyerInfo              `bson:"buyerInfo"`
	Invoice       Invoice                `bson:"invoice"`
	Packages      []*PackageItem         `bson:"packages"`
	CreatedAt     time.Time              `bson:"createdAt"`
	UpdatedAt     time.Time              `bson:"updatedAt"`
	DeletedAt     *time.Time             `bson:"deletedAt"`
	Extended      map[string]interface{} `bson:"ext"`
}

type PaymentService struct {
	PaymentRequest  *PaymentRequest        `bson:"paymentRequest"`
	PaymentResponse *PaymentResponse       `bson:"paymentResponse"`
	PaymentResult   *PaymentResult         `bson:"paymentResult"`
	Extended        map[string]interface{} `bson:"ext"`
}

// TODO get configs of pay to market from back-office
type SystemPayment struct {
	PayToBuyer  []PayToBuyerInfo       `bson:"payToBuyer"`
	PayToMarket []PayToMarket          `bson:"payToMarket"`
	Extended    map[string]interface{} `bson:"ext"`
}

type PayToBuyerInfo struct {
	PaymentRequest  *PaymentRequest        `bson:"paymentRequest"`
	PaymentResponse *PaymentResponse       `bson:"paymentResponse"`
	PaymentResult   *PaymentResult         `bson:"paymentResult"`
	Extended        map[string]interface{} `bson:"ext"`
}

type PayToMarket struct {
	PaymentRequest  *PaymentRequest        `bson:"paymentRequest"`
	PaymentResponse *PaymentResponse       `bson:"paymentResponse"`
	PaymentResult   *PaymentResult         `bson:"paymentResult"`
	Extended        map[string]interface{} `bson:"ext"`
}

type Invoice struct {
	GrandTotal     Money                  `bson:"grandTotal"`
	Subtotal       Money                  `bson:"subtotal"`
	Discount       Money                  `bson:"discount"`
	ShipmentTotal  Money                  `bson:"shipmentTotal"`
	PaymentMethod  string                 `bson:"paymentMethod"`
	PaymentGateway string                 `bson:"paymentGateway"`
	PaymentOption  *PaymentOption         `bson:"paymentOption"`
	Share          *OrderShare            `bson:"share"`
	Commission     *Commission            `bson:"commission"`
	Voucher        *Voucher               `bson:"voucher"`
	CartRule       *CartRule              `bson:"cartRule"`
	SSO            *SSO                   `bson:"sso"`
	VAT            *VAT                   `bson:"vat"`
	TAX            *TAX                   `bson:"tax"`
	Extended       map[string]interface{} `bson:"ext"`
}

type PaymentOption struct {
}

type Commission struct {
	RawTotalPrice     *Money                 `bson:"rawTotalPrice"`
	RoundupTotalPrice *Money                 `bson:"roundupTotalPrice"`
	CreatedAt         *time.Time             `bson:"createdAt"`
	UpdatedAt         *time.Time             `bson:"updatedAt"`
	Extended          map[string]interface{} `bson:"ext"`
}

type OrderShare struct {
	RawTotalShare     *Money                 `bson:"rawTotalShare"`
	RoundupTotalShare *Money                 `bson:"roundupTotalShare"`
	CreatedAt         *time.Time             `bson:"createdAt"`
	UpdatedAt         *time.Time             `bson:"updatedAt"`
	Extended          map[string]interface{} `bson:"ext"`
}

type Voucher struct {
	Percent                     float64                `bson:"percent"`
	AppliedPrice                *Money                 `bson:"appliedPrice"`
	RoundupAppliedPrice         *Money                 `bson:"roundupAppliedPrice"`
	RawShipmentAppliedPrice     *Money                 `bson:"rawShipmentAppliedPrice"`
	RoundupShipmentAppliedPrice *Money                 `bson:"roundupShipmentAppliedPrice"`
	Price                       *Money                 `bson:"price"`
	Code                        string                 `bson:"code"`
	Details                     *VoucherDetails        `bson:"details"`
	Settlement                  string                 `bson:"settlement"`
	SettlementAt                *time.Time             `bson:"settlementAt"`
	Reserved                    string                 `bson:"reserved"`
	ReservedAt                  *time.Time             `bson:"reservedAt"`
	Extended                    map[string]interface{} `bson:"ext"`
}

type CartRule struct {
}

type SSO struct {
	RawTotal     *Money                 `bson:"rawTotal"`
	RoundupTotal *Money                 `bson:"roundupTotal"`
	CreatedAt    *time.Time             `bson:"createdAt"`
	UpdatedAt    *time.Time             `bson:"updatedAt"`
	Extended     map[string]interface{} `bson:"ext"`
}

type VAT struct {
	Rate         float32                `bson:"rate"`
	RawTotal     *Money                 `bson:"rawTotal"`
	RoundupTotal *Money                 `bson:"roundupTotal"`
	CreatedAt    *time.Time             `bson:"createdAt"`
	UpdatedAt    *time.Time             `bson:"updatedAt"`
	Extended     map[string]interface{} `bson:"ext"`
}

type TAX struct {
}

type VoucherDetails struct {
	Title            string                 `bson:"title"`
	Prefix           string                 `bson:"prefix"`
	UseLimit         int32                  `bson:"useLimit"`
	Count            int32                  `bson:"count"`
	Length           int32                  `bson:"length"`
	Categories       []string               `bson:"categories"`
	Products         []string               `bson:"products"`
	Users            []string               `bson:"users"`
	Sellers          []string               `bson:"sellers"`
	IsFirstPurchase  bool                   `bson:"isFirstPurchase"`
	StartDate        time.Time              `bson:"startDate"`
	EndDate          time.Time              `bson:"endDate"`
	Type             string                 `bson:"type"`
	MaxDiscountValue uint64                 `bson:"maxDiscountValue"`
	MinBasketValue   uint64                 `bson:"minBasketValue"`
	Extended         map[string]interface{} `bson:"ext"`
}

type PaymentRequest struct {
	Price     *Money                 `bson:"price"`
	Gateway   string                 `bson:"gateway"`
	CreatedAt time.Time              `bson:"createdAt"`
	Mobile    string                 `bson:"mobile"`
	Data      interface{}            `bson:"data"`
	Extended  map[string]interface{} `bson:"ext"`
}

type PaymentResponse struct {
	Result      bool                   `bson:"result"`
	Reason      string                 `bson:"reason"`
	Description string                 `bson:"description"`
	Response    interface{}            `bson:"response"`
	CreatedAt   time.Time              `bson:"createdAt"`
	Extended    map[string]interface{} `bson:"ext"`
}

type PaymentIPGResponse struct {
	CallBackUrl string                 `bson:"callbackUrl"`
	InvoiceId   int64                  `bson:"invoiceId"`
	PaymentId   string                 `bson:"paymentId"`
	Extended    map[string]interface{} `bson:"ext"`
}

type PaymentMPGResponse struct {
	HostRequest     string                 `bson:"hostRequest"`
	HostRequestSign string                 `bson:"hostRequestSign"`
	PaymentId       string                 `bson:"paymentId"`
	Extended        map[string]interface{} `bson:"ext"`
}

type PaymentResult struct {
	Result      bool                   `bson:"result"`
	Reason      string                 `bson:"reason"`
	PaymentId   string                 `bson:"paymentId"`
	InvoiceId   int64                  `bson:"invoiceId"`
	Price       *Money                 `bson:"price"`
	CardNumMask string                 `bson:"cardNumMask"`
	Data        interface{}            `bson:"data"`
	CreatedAt   time.Time              `bson:"createdAt"`
	Extended    map[string]interface{} `bson:"ext"`
}

type Money struct {
	Amount   string `bson:"amount"`
	Currency string `bson:"cur"`
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
