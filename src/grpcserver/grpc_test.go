package grpcserver

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	pb "gitlab.faza.io/protos/order"
	message "gitlab.faza.io/protos/order/general"
	"google.golang.org/grpc/status"
	"strconv"

	//"github.com/golang
	// /protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	//"log"
	"testing"

	//"errors"
	//"net"
	//"net/http"
	"time"

	//"github.com/rs/xid"

	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/configs"

	"google.golang.org/grpc"
)

var config *configs.Cfg

func init() {
	var err error
	config, err = configs.LoadConfig("")
	if err != nil {
		logger.Err(err.Error())
		return
	}
	go startGrpc(config.App.Port)
}

func createNewOrderRequest() *pb.NewOrderRequest {
	order := &pb.NewOrderRequest{
		Amount: &pb.Amount{},
		Buyer: &pb.Buyer{
			Finance: &pb.FinanceInfo{},
			ShippingAddress: &pb.Address{},
		},
	}

	order.Amount.Total = 600000
	order.Amount.Payable = 550000
	order.Amount.Discount = 50000
	order.Amount.Currency = "RR"
	order.Amount.PaymentMethod = "IPG"
	order.Amount.PaymentOption = "AAP"
	order.Amount.ShipmentTotal = 700000
	order.Amount.Voucher = &pb.Voucher{
		Amount: 40000,
		Code: "348",
	}

	order.Buyer.LastName = "Tadayon"
	order.Buyer.FirstName = "Sina"
	order.Buyer.Email = "Sina.Tadayon@baman.io"
	order.Buyer.Mobile = "09124566788"
	order.Buyer.NationalId = "005938404734"
	order.Buyer.Ip = "127.0.0.1"
	order.Buyer.Gender = "male"

	order.Buyer.Finance.Iban = "IR165411211001514313143545"
	order.Buyer.Finance.AccountNumber = "303.100.1269574.1"
	order.Buyer.Finance.CardNumber = "4345345423533453"
	order.Buyer.Finance.BankName = "pasargad"

	order.Buyer.ShippingAddress.Address = "Sheikh bahaee, p 5"
	order.Buyer.ShippingAddress.Province = "Tehran"
	order.Buyer.ShippingAddress.Phone = "+98912193870"
	order.Buyer.ShippingAddress.ZipCode = "1651764614"
	order.Buyer.ShippingAddress.City = "Tehran"
	order.Buyer.ShippingAddress.Country = "Iran"
	order.Buyer.ShippingAddress.Neighbourhood = "Seool"
	order.Buyer.ShippingAddress.Lat = "10.1345664"
	order.Buyer.ShippingAddress.Long = "22.1345664"

	item := pb.Item {
		Price:    &pb.PriceInfo{},
		Shipment: &pb.ShippingSpec{},
		Attributes: &pb.Attributes{},
		SellerId: "6546345",
	}

	item.InventoryId = "453564554435345"
	item.Brand = "Asus"
	item.Categories = "Electronic/laptop"
	item.Title = "Asus G503 i7, 256SSD, 32G Ram"
	item.Warranty = "ضمانت سلامت کالا"
	item.Quantity = 10
	item.Image = "http://baman.io/image/asus.png"
	item.Returnable = true

	item.Price.Discount = 200000
	item.Price.Payable = 20000000
	item.Price.Total = 1600000
	item.Price.SellerCommission = 10
	item.Price.Unit = 100000
	item.Price.Currency = "RR"

	//Standard, Express, Economy or Sameday.
	item.Shipment.Details = "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد"
	item.Shipment.ShippingTime = 72
	item.Shipment.ReturnTime = 72
	item.Shipment.ReactionTime = 24
	item.Shipment.CarrierName = "Post"
	item.Shipment.CarrierProduct = "Post Express"
	item.Shipment.CarrierType = "standard"
	item.Shipment.ShippingAmount = 100000
	item.Shipment.VoucherAmount = 0
	item.Shipment.Currency = "RR"

	item.SellerId = "345346343"
	//item. Seller.FirstName = "farzan"
	//item.Seller.LastName = "dalaii"
	//item.Seller.Title = "Digikala"
	//item.Seller.NationalId = "4353435343"
	//item.Seller.Email = "farzan.dalaii@bamilo.com"
	//item.Seller.Mobile = "0912329389"
	//item.Seller.CompanyName = "Digi"
	//item.Seller.RegistrationName = "Digikala gostaran e shargh"
	//item.Seller.EconomicCode = "13211"

	//item.Seller.Address.Address = "address"
	//item.Seller.Address.Lan = "23031121"
	//item.Seller.Address.Lat = "03221211"
	//item.Seller.Address.Country = "Iran"
	//item.Seller.Address.City = "Tehran"
	//item.Seller.Address.ZipCode = "1651145864"
	//item.Seller.Address.Phone = "0212222222"
	//item.Seller.Address.State = "Tehran"
	//item.Seller.Address.Title = "office"

	//item.Seller.Finance.Iban = "IR165411211001514313143545354134"
	//item.Seller.Finance.CardNumber = "1234123412341234"
	//item.Seller.Finance.AccountNumber = "234.545.12342344.4"
	//item.Seller.Finance.BankName = "melli"
	//item.Seller.Finance.Gateway = ""

	order.Items = append(order.Items, &item)
	return order
}

func createMetaDataRequest() *message.RequestMetadata {
	var metadata = &message.RequestMetadata{
		Page:                 1,
		PerPage:              25,
		Sorts:                []*message.MetaSorts{
			{
				Name:      "mobile",
				Direction: 0,
			}, {
				Name:      "name",
				Direction: 1,
			},
		},
		Filters:              []*message.MetaFilter{
			{
				Name: "mobile",
				Opt: "eq",
				Value: "012933434",
			},
		},
	}

	return metadata
}

// Grpc test
func TestNewOrder(t *testing.T) {

	//time.Sleep(3 * time.Second)
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	grpcConnNewOrder, err := grpc.DialContext(ctx, ":" + config.App.Port, grpc.WithInsecure())
	assert.Nil(t, err)
	OrderService := pb.NewOrderServiceClient(grpcConnNewOrder)

	newOrderRequest := createNewOrderRequest()
	metadata := createMetaDataRequest()

	serializedOrder, err := proto.Marshal(newOrderRequest)
	if err != nil {
		logger.Err("could not serialize timestamp")
	}

	orderId := entities.GenerateOrderId()
	request := message.Request {
		OrderId: orderId,
		ItemId: orderId + strconv.Itoa(int(entities.GenerateRandomNumber())),
		Time: ptypes.TimestampNow(),
		Meta: metadata,
		Data: &any.Any{
			TypeUrl: "baman.io/" + proto.MessageName(newOrderRequest),
			Value:   serializedOrder,
		},
	}

	resOrder, err := OrderService.OrderRequestsHandler(ctx, &request)

	if err != nil {
		st := status.Convert(err)
		for _, detail := range st.Details() {
			switch t := detail.(type) {
			case *message.ErrorDetails:
				fmt.Println("Oops! Your request was rejected by the server.")
				for _, validate := range t.Validation {
					fmt.Printf("The %q field was wrong:\n", validate.GetField())
					fmt.Printf("\t%s\n", validate.GetDesc())
				}
			}
		}
	}

	//assert.Nil(t, err)
	assert.NotNil(t, resOrder)
}
