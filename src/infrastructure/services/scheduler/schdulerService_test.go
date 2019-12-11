package scheduler_service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/app"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain"
	"gitlab.faza.io/order-project/order-service/domain/converter"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	order_repository "gitlab.faza.io/order-project/order-service/domain/models/repository/order"
	"gitlab.faza.io/order-project/order-service/domain/states"
	stock_service "gitlab.faza.io/order-project/order-service/infrastructure/services/stock"
	pb "gitlab.faza.io/protos/order"
	"os"
	"testing"
	"time"
)

var config *configs.Config
var schedulerService iSchedulerServiceImpl

func init() {
	var err error
	var path string
	if os.Getenv("APP_ENV") == "dev" {
		path = "../../../testdata/.env"
	} else {
		path = ""
	}

	config, err = configs.LoadConfig(path)
	if err != nil {
		logger.Err(err.Error())
		panic("configs.LoadConfig failed")
	}

	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		Host:     config.Mongo.Host,
		Port:     config.Mongo.Port,
		Username: config.Mongo.User,
		//Password:     App.Config.Mongo.Pass,
		ConnTimeout:     time.Duration(config.Mongo.ConnectionTimeout),
		ReadTimeout:     time.Duration(config.Mongo.ReadTimeout),
		WriteTimeout:    time.Duration(config.Mongo.WriteTimeout),
		MaxConnIdleTime: time.Duration(config.Mongo.MaxConnIdleTime),
		MaxPoolSize:     uint64(config.Mongo.MaxPoolSize),
		MinPoolSize:     uint64(config.Mongo.MinPoolSize),
	}

	mongoDriver, err := mongoadapter.NewMongo(mongoConf)
	if err != nil {
		logger.Err("mongoadapter.NewMongo Mongo: %v", err.Error())
		panic("mongo adapter creation failed, " + err.Error())
	}

	app.Globals.OrderRepository, err = order_repository.NewOrderRepository(mongoDriver)
	if err != nil {
		logger.Err("repository creation failed, %s ", err.Error())
		panic("order repository creation failed, " + err.Error())
	}

	app.Globals.StockService = stock_service.NewStockServiceMock()
	app.Globals.Converter = converter.NewConverter()

	// TODO create item repository
	flowManager, err := domain.NewFlowManager()
	if err != nil {
		logger.Err("flowManager creation failed, %s ", err.Error())
		panic("flowManager creation failed, " + err.Error())
	}

	schedulerService = iSchedulerServiceImpl{
		mongoAdapter: mongoDriver,
		flowManager:  flowManager,
	}
}

func createRequestNewOrder() *pb.RequestNewOrder {
	order := &pb.RequestNewOrder{
		Amount: &pb.Amount{},
		Buyer: &pb.Buyer{
			Finance:         &pb.FinanceInfo{},
			ShippingAddress: &pb.Address{},
		},
	}

	order.Amount.Total = 600000
	order.Amount.Subtotal = 550000
	order.Amount.Discount = 50000
	order.Amount.Currency = "IRR"
	order.Amount.PaymentMethod = "IPG"
	order.Amount.PaymentOption = "asanpardakht"
	order.Amount.ShipmentTotal = 700000
	order.Amount.Voucher = &pb.Voucher{
		Amount: 40000,
		Code:   "348",
	}

	order.Buyer.BuyerId = 123456
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

	item := pb.Item{
		Price:      &pb.PriceInfo{},
		Shipment:   &pb.ShippingSpec{},
		Attributes: make(map[string]string, 10),
		SellerId:   123456,
	}

	item.InventoryId = "11111-22222"
	item.Brand = "Asus"
	item.Category = "Electronic/laptop"
	item.Title = "Asus G503 i7, 256SSD, 32G Ram"
	item.Guaranty = "ضمانت سلامت کالا"
	item.Image = "http://baman.io/image/asus.png"
	item.Returnable = true
	item.Quantity = 5

	item.Attributes["Quantity"] = "10"
	item.Attributes["Width"] = "8cm"
	item.Attributes["Height"] = "10cm"
	item.Attributes["Length"] = "15cm"
	item.Attributes["Weight"] = "20kg"
	item.Attributes["Color"] = "blue"
	item.Attributes["Materials"] = "stone"

	item.Price.Special = 200000
	item.Price.Total = 22000000
	item.Price.Original = 20000000
	item.Price.SellerCommission = 10
	item.Price.Unit = 100000
	item.Price.Currency = "IRR"

	//Standard, Express, Economy or Sameday.
	item.Shipment.Details = "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد"
	item.Shipment.ShippingTime = 0
	item.Shipment.ReturnTime = 0
	item.Shipment.ReactionTime = 0
	item.Shipment.CarrierName = "Post"
	item.Shipment.CarrierProduct = "Post Express"
	item.Shipment.CarrierType = "standard"
	item.Shipment.ShippingCost = 100000
	item.Shipment.VoucherAmount = 0
	item.Shipment.Currency = "IRR"

	order.Items = append(order.Items, &item)

	item1 := pb.Item{
		Price:      &pb.PriceInfo{},
		Shipment:   &pb.ShippingSpec{},
		Attributes: make(map[string]string, 10),
		SellerId:   678912,
	}

	item1.InventoryId = "1111-33333"
	item1.Brand = "Lenovo"
	item1.Category = "Electronic/laptop"
	item1.Title = "Lenove G503 i7, 256SSD, 32G Ram"
	item1.Guaranty = "ضمانت سلامت کالا"
	item1.Image = "http://baman.io/image/asus.png"
	item1.Returnable = true
	item1.Quantity = 5

	item1.Attributes["Quantity"] = "10"
	item1.Attributes["Width"] = "8cm"
	item1.Attributes["Height"] = "10cm"
	item1.Attributes["Length"] = "15cm"
	item1.Attributes["Weight"] = "20kg"
	item1.Attributes["Color"] = "blue"
	item1.Attributes["Materials"] = "stone"

	item1.Price.Special = 250000
	item.Price.Total = 2500000
	item1.Price.Original = 200000
	item1.Price.SellerCommission = 10
	item1.Price.Unit = 200000
	item1.Price.Currency = "IRR"

	//Standard, Express, Economy or Sameday.
	item1.Shipment.Details = "پست پیشتاز و تیپاکس برای شهرستان ها و پیک برای تهران به صورت رایگان می باشد"
	item1.Shipment.ShippingTime = 0
	item1.Shipment.ReturnTime = 0
	item1.Shipment.ReactionTime = 0
	item1.Shipment.CarrierName = "Post"
	item1.Shipment.CarrierProduct = "Post Express"
	item1.Shipment.CarrierType = "standard"
	item1.Shipment.ShippingCost = 100000
	item1.Shipment.VoucherAmount = 0
	item1.Shipment.Currency = "IRR"

	order.Items = append(order.Items, &item1)
	return order
}

func updateOrderStatus(order *entities.Order, itemsId []uint64, orderStatus string, isUpdateOnlyOrderStatus bool, stepName string, stepIndex int) {

	if isUpdateOnlyOrderStatus == true {
		order.UpdatedAt = time.Now().UTC()
		order.Status = orderStatus
	} else {
		order.UpdatedAt = time.Now().UTC()
		order.Status = orderStatus
		findFlag := true
		if itemsId != nil && len(itemsId) > 0 {
			for _, id := range itemsId {
				findFlag = false
				for i := 0; i < len(order.Items); i++ {
					if order.Items[i].ItemId == id {
						doUpdateOrderStep(order, i, stepName, stepIndex)
						findFlag = true
						break
					}
				}
				if !findFlag {
					logger.Err("%s received sid %d not exist in order, orderId: %d", stepName, id, order.OrderId)
				}
			}
		} else {
			for i := 0; i < len(order.Items); i++ {
				doUpdateOrderStep(order, i, stepName, stepIndex)
			}
		}
	}
}

func doUpdateOrderStep(order *entities.Order, index int, stepName string, stepIndex int) {
	order.Items[index].Progress.CreatedAt = time.Now().UTC()
	order.Items[index].Progress.CurrentStepName = stepName
	order.Items[index].Progress.CurrentStepIndex = stepIndex

	stepHistory := entities.StateHistory{
		Name:      stepName,
		Index:     stepIndex,
		CreatedAt: order.Items[index].Progress.CreatedAt,
		//ActionHistory: make([]entities.Actions, 0, 1),
	}

	if order.Items[index].Progress.StepsHistory == nil || len(order.Items[index].Progress.StepsHistory) == 0 {
		order.Items[index].Progress.StepsHistory = make([]entities.StateHistory, 0, 5)
	}

	order.Items[index].Progress.StepsHistory = append(order.Items[index].Progress.StepsHistory, stepHistory)
}

func updateOrderItemsProgress(order *entities.Order, itemsId []uint64, action string, result bool, itemStatus string) {

	findFlag := false
	if itemsId != nil && len(itemsId) > 0 {
		for _, id := range itemsId {
			findFlag = false
			for i := 0; i < len(order.Items); i++ {
				if order.Items[i].ItemId == id {
					doUpdateOrderItemsProgress(order, i, action, result, itemStatus)
					findFlag = true
				}
			}

			if !findFlag {
				logger.Err("received sid %d not exist in order, orderId: %d", id, order.OrderId)
			}
		}
	} else {
		for i := 0; i < len(order.Items); i++ {
			doUpdateOrderItemsProgress(order, i, action, result, itemStatus)
		}
	}
}

func doUpdateOrderItemsProgress(order *entities.Order, index int,
	actionName string, result bool, itemStatus string) {

	order.Items[index].Status = itemStatus
	order.Items[index].UpdatedAt = time.Now().UTC()

	length := len(order.Items[index].Progress.StepsHistory) - 1

	if order.Items[index].Progress.StepsHistory[length].ActionHistory == nil || len(order.Items[index].Progress.StepsHistory[length].ActionHistory) == 0 {
		order.Items[index].Progress.StepsHistory[length].ActionHistory = make([]entities.Action, 0, 5)
	}

	expiredTime := order.Items[index].UpdatedAt.Add(time.Hour*
		time.Duration(0) +
		time.Minute*time.Duration(0) +
		time.Second*time.Duration(1))

	action := entities.Action{
		Name:   actionName,
		Result: result,
		Data: map[string]interface{}{
			"expiredTime": expiredTime,
		},
		CreatedAt: order.Items[index].UpdatedAt,
	}
	order.Items[index].Progress.StepsHistory[length].ActionHistory = append(order.Items[index].Progress.StepsHistory[length].ActionHistory, action)
}

func TestSchedulerSellerApprovalPending(t *testing.T) {

	ctx, _ := context.WithCancel(context.Background())
	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "20.Seller_Approval_Pending", 20)
	updateOrderItemsProgress(newOrder, nil, "ApprovalPending", true, states.OrderInProgressStatus)
	order, err := app.Globals.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	defer removeCollection()

	data := ScheduleModel{
		Step:   "20.Seller_Approval_Pending",
		Action: "ApprovalPending",
	}

	time.Sleep(3 * time.Second)
	schedulerService.doProcess(ctx, data)

	changedOrder, err := app.Globals.OrderRepository.FindById(order.OrderId)
	assert.Nil(t, err)

	length := len(changedOrder.Items[0].Progress.StepsHistory) - 1
	assert.Equal(t, changedOrder.Items[0].Progress.StepsHistory[length].Index, 80)
	assert.Equal(t, changedOrder.Items[0].Progress.StepsHistory[length].ActionHistory[len(changedOrder.Items[0].Progress.StepsHistory[length].ActionHistory)-1].Name, "CANCELED")
}

func TestSchedulerSellerShipmentPending(t *testing.T) {

	ctx, _ := context.WithCancel(context.Background())
	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "30.Shipment_Pending", 30)
	updateOrderItemsProgress(newOrder, nil, "SellerShipmentPending", true, states.OrderInProgressStatus)
	order, err := app.Globals.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	defer removeCollection()

	data := ScheduleModel{
		Step:   "30.Shipment_Pending",
		Action: "SellerShipmentPending",
	}

	time.Sleep(3 * time.Second)
	schedulerService.doProcess(ctx, data)

	changedOrder, err := app.Globals.OrderRepository.FindById(order.OrderId)
	assert.Nil(t, err)

	length := len(changedOrder.Items[0].Progress.StepsHistory) - 1
	assert.Equal(t, changedOrder.Items[0].Progress.StepsHistory[length].Index, 80)
	assert.Equal(t, changedOrder.Items[0].Progress.StepsHistory[length].ActionHistory[len(changedOrder.Items[0].Progress.StepsHistory[length].ActionHistory)-1].Name, "CANCELED")
}

func TestSchedulerShipmentDeliveredPending(t *testing.T) {

	ctx, _ := context.WithCancel(context.Background())
	requestNewOrder := createRequestNewOrder()
	value, err := app.Globals.Converter.Map(*requestNewOrder, entities.Order{})
	assert.Nil(t, err, "Converter failed")
	newOrder := value.(*entities.Order)

	updateOrderStatus(newOrder, nil, "IN_PROGRESS", false, "32.Shipment_Delivered", 32)
	updateOrderItemsProgress(newOrder, nil, "ShipmentDeliveredPending", true, states.OrderInProgressStatus)
	order, err := app.Globals.OrderRepository.Save(*newOrder)
	assert.Nil(t, err, "save failed")

	defer removeCollection()

	data := ScheduleModel{
		Step:   "32.Shipment_Delivered",
		Action: "ShipmentDeliveredPending",
	}

	time.Sleep(3 * time.Second)
	schedulerService.doProcess(ctx, data)

	changedOrder, err := app.Globals.OrderRepository.FindById(order.OrderId)
	assert.Nil(t, err)

	length := len(changedOrder.Items[0].Progress.StepsHistory) - 1
	assert.Equal(t, changedOrder.Items[0].Progress.StepsHistory[length].Index, 90)
	assert.Equal(t, changedOrder.Items[0].Progress.StepsHistory[length].ActionHistory[len(changedOrder.Items[0].Progress.StepsHistory[length].ActionHistory)-1].Name, "DELIVERED")
}

//func TestSchedulerHealthWorker(t *testing.T) {
//	ctx, _ := context.WithCancel(context.Background())
//	data := []ScheduleModel {
//		{
//			Step:   "32.Shipment_Delivered",
//			Actions: "ShipmentDeliveredPending",
//		},
//	}
//
//	err := schedulerService.Scheduler(ctx, data)
//	assert.Nil(t, err)
//	time.Sleep(1 * time.Minute)
//}

func removeCollection() {
	if err := app.Globals.OrderRepository.RemoveAll(); err != nil {
	}
}
