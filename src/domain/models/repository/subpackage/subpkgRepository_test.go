package subpackage

import (
	"context"
	"github.com/stretchr/testify/require"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/go-framework/mongoadapter"
	"gitlab.faza.io/order-project/order-service/configs"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"strconv"
	"testing"
	"time"
)

var subPkgRepo ISubpackageRepository
var mongoAdapter *mongoadapter.Mongo

func TestMain(m *testing.M) {
	var path string
	if os.Getenv("APP_ENV") == "dev" {
		path = "../../../../testdata/.env"
	} else {
		path = ""
	}

	config, err := configs.LoadConfig(path)
	if err != nil {
		logger.Err("configs.LoadConfig failed, %s", err.Error())
		os.Exit(1)
	}

	// store in mongo
	mongoConf := &mongoadapter.MongoConfig{
		Host:     config.Mongo.Host,
		Port:     config.Mongo.Port,
		Username: config.Mongo.User,
		//Password:     App.Cfg.Mongo.Pass,
		ConnTimeout:     time.Duration(config.Mongo.ConnectionTimeout),
		ReadTimeout:     time.Duration(config.Mongo.ReadTimeout),
		WriteTimeout:    time.Duration(config.Mongo.WriteTimeout),
		MaxConnIdleTime: time.Duration(config.Mongo.MaxConnIdleTime),
		MaxPoolSize:     uint64(config.Mongo.MaxPoolSize),
		MinPoolSize:     uint64(config.Mongo.MinPoolSize),
	}

	mongoAdapter, err = mongoadapter.NewMongo(mongoConf)
	if err != nil {
		logger.Err("ISubpackageRepository Mongo: %v", err.Error())
		os.Exit(1)
	}

	subPkgRepo = NewSubPkgRepository(mongoAdapter)

	// Running Tests
	code := m.Run()
	removeCollection()
	os.Exit(code)
}

func TestSave(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	ctx, _ := context.WithCancel(context.Background())
	newSubPkg := order.Packages[0].Subpackages[1].DeepCopy()
	newSubPkg.ItemId = 0
	err = subPkgRepo.Save(ctx, newSubPkg)
	require.Nil(t, err)
	updateOrder, err := getOrder(order.OrderId)
	require.Nil(t, err)
	require.Equal(t, 3, len(updateOrder.Packages[0].Subpackages))
}

func TestUpdate(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	ctx, _ := context.WithCancel(context.Background())
	order.Packages[0].Subpackages[0].Status = "Payment_Pending"
	_, err = subPkgRepo.Update(ctx, order.Packages[0].Subpackages[0])
	require.Nil(t, err)
	updateOrder, err := getOrder(order.OrderId)
	require.Nil(t, err)
	require.Equal(t, "Payment_Pending", updateOrder.Packages[0].Subpackages[0].Status)
	require.Equal(t, uint64(1), updateOrder.Packages[0].Subpackages[0].Version)
	require.Equal(t, "New", updateOrder.Packages[0].Subpackages[1].Status)
	require.Equal(t, uint64(0), updateOrder.Packages[0].Subpackages[1].Version)
	require.Equal(t, "New", updateOrder.Packages[1].Subpackages[0].Status)
	require.Equal(t, uint64(0), updateOrder.Packages[1].Subpackages[0].Version)
	require.Equal(t, "New", updateOrder.Packages[1].Subpackages[1].Status)
	require.Equal(t, uint64(0), updateOrder.Packages[1].Subpackages[1].Version)
}

func TestFindByOrderAndItemId(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	ctx, _ := context.WithCancel(context.Background())
	subPkg, err := subPkgRepo.FindByOrderAndItemId(ctx, order.OrderId, order.Packages[1].Subpackages[1].ItemId)
	require.Nil(t, err)
	require.Equal(t, order.Packages[1].Subpackages[1].ItemId, subPkg.ItemId)
}

func TestFindByOrderAndSellerId(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	defer removeCollection()
	ctx, _ := context.WithCancel(context.Background())
	subPkgList, err := subPkgRepo.FindByOrderAndSellerId(ctx, order.OrderId, order.Packages[1].SellerId)
	require.Nil(t, err)
	require.Equal(t, order.Packages[1].SellerId, subPkgList[0].SellerId)
	require.Equal(t, 2, len(subPkgList))
}

func TestFindAll(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	order, err = createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	defer removeCollection()
	ctx, _ := context.WithCancel(context.Background())
	subPkgList, err := subPkgRepo.FindAll(ctx, order.Packages[1].SellerId)
	require.Nil(t, err)
	require.Equal(t, order.Packages[1].SellerId, subPkgList[0].SellerId)
	require.Equal(t, order.Packages[1].SellerId, subPkgList[1].SellerId)
	require.Equal(t, 4, len(subPkgList))
}

func TestFindAllWithSort(t *testing.T) {
	defer removeCollection()
	order := createOrder()
	firstTime := time.Now().UTC().Add(time.Duration(1 * time.Hour))
	order.Packages[0].Subpackages[0].CreatedAt = firstTime
	order.Packages[0].Subpackages[1].CreatedAt = time.Now().UTC().Add(time.Duration(1 * time.Minute))
	_, err := insertWithoutChangeTime(order)
	require.Nil(t, err, "insert order failed")
	require.NotEmpty(t, order.OrderId, "insert order failed, order id not generated")

	order = createOrder()
	secondTime := time.Now().UTC().Add(time.Duration(3 * time.Minute))
	order.Packages[0].Subpackages[0].CreatedAt = secondTime
	_, err = insertWithoutChangeTime(order)
	require.Nil(t, err, "insert order failed")
	require.NotEmpty(t, order.OrderId, "insert order failed, order id not generated")

	defer removeCollection()
	ctx, _ := context.WithCancel(context.Background())
	subPkgList, err := subPkgRepo.FindAllWithSort(ctx, order.Packages[0].SellerId, "createdAt", -1)
	require.Nil(t, err)
	require.Equal(t, 4, len(subPkgList))
	require.Equal(t, firstTime.Unix(), subPkgList[0].CreatedAt.Unix())
	require.Equal(t, secondTime.Unix(), subPkgList[1].CreatedAt.Unix())
}

func TestFindAllWithPage(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	order, err = createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	order, err = createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	defer removeCollection()
	ctx, _ := context.WithCancel(context.Background())
	subPkgList, totalPage, err := subPkgRepo.FindAllWithPage(ctx, order.Packages[1].SellerId, 1, 2)
	require.Nil(t, err)
	require.Equal(t, order.Packages[1].SellerId, subPkgList[0].SellerId)
	require.Equal(t, order.Packages[1].SellerId, subPkgList[1].SellerId)
	require.Equal(t, 2, len(subPkgList))
	require.Equal(t, int64(3), totalPage)
}

func TestFindAllWithPageAndSort(t *testing.T) {
	defer removeCollection()
	order := createOrder()
	firstTime := time.Now().UTC().Add(time.Duration(1 * time.Hour))
	order.Packages[0].Subpackages[0].CreatedAt = firstTime
	order.Packages[0].Subpackages[1].CreatedAt = time.Now().UTC().Add(time.Duration(1 * time.Minute))
	_, err := insertWithoutChangeTime(order)
	require.Nil(t, err, "insert order failed")
	require.NotEmpty(t, order.OrderId, "insert order failed, order id not generated")

	order = createOrder()
	secondTime := time.Now().UTC().Add(time.Duration(3 * time.Minute))
	order.Packages[0].Subpackages[0].CreatedAt = secondTime
	_, err = insertWithoutChangeTime(order)
	require.Nil(t, err, "insert order failed")
	require.NotEmpty(t, order.OrderId, "insert order failed, order id not generated")

	order, err = createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")

	defer removeCollection()
	ctx, _ := context.WithCancel(context.Background())
	subPkgList, totalPage, err := subPkgRepo.FindAllWithPageAndSort(ctx, order.Packages[0].SellerId, 1, 2, "createdAt", -1)
	require.Nil(t, err)
	require.Equal(t, 2, len(subPkgList))
	require.Equal(t, int64(3), totalPage)
	require.Equal(t, firstTime.Unix(), subPkgList[0].CreatedAt.Unix())
	require.Equal(t, secondTime.Unix(), subPkgList[1].CreatedAt.Unix())
}

func TestFindByFilter(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	order, err = createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	defer removeCollection()
	ctx, _ := context.WithCancel(context.Background())
	totalPipeline := []bson.M{
		{"$match": bson.M{"packages.sellerId": order.Packages[0].SellerId, "packages.deletedAt": nil}},
		{"$project": bson.M{"subSize": bson.M{"$size": "$packages.subpackages"}}},
		{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": "$subSize"}}},
		{"$project": bson.M{"_id": 0, "count": 1}},
	}
	pipeline := []bson.M{
		{"$match": bson.M{"packages.sellerId": order.Packages[0].SellerId, "packages.deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.sellerId": order.Packages[0].SellerId, "packages.deletedAt": nil}},
		{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
		{"$unwind": "$packages.subpackages"},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		{"$replaceRoot": bson.M{"newRoot": "$subpackages"}},
	}

	subPkgList, err := subPkgRepo.FindByFilter(ctx, func() (filter interface{}) { return totalPipeline }, func() (filter interface{}) { return pipeline })
	require.Nil(t, err)
	require.Equal(t, order.Packages[0].SellerId, subPkgList[0].SellerId)
	require.Equal(t, order.Packages[0].SellerId, subPkgList[1].SellerId)
	require.Equal(t, 4, len(subPkgList))
}

func TestFindByFilterWithPage(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	order, err = createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	defer removeCollection()
	ctx, _ := context.WithCancel(context.Background())
	totalPipeline := []bson.M{
		{"$match": bson.M{"packages.sellerId": order.Packages[0].SellerId, "packages.deletedAt": nil}},
		{"$project": bson.M{"subSize": bson.M{"$size": "$packages.subpackages"}}},
		{"$group": bson.M{"_id": nil, "count": bson.M{"$sum": "$subSize"}}},
		{"$project": bson.M{"_id": 0, "count": 1}},
	}

	pipeline := []bson.M{
		{"$match": bson.M{"packages.sellerId": order.Packages[0].SellerId, "packages.deletedAt": nil}},
		{"$unwind": "$packages"},
		{"$match": bson.M{"packages.sellerId": order.Packages[0].SellerId, "packages.deletedAt": nil}},
		{"$project": bson.M{"_id": 0, "packages.subpackages": 1}},
		{"$unwind": "$packages.subpackages"},
		{"$skip": 0},
		{"$limit": 2},
		{"$replaceRoot": bson.M{"newRoot": "$packages"}},
		{"$replaceRoot": bson.M{"newRoot": "$subpackages"}},
	}
	subPkgList, totalPages, err := subPkgRepo.FindByFilterWithPage(ctx, func() (filter interface{}) { return totalPipeline }, func() (filter interface{}) { return pipeline }, 1, 2)
	require.Nil(t, err)
	require.Equal(t, order.Packages[0].SellerId, subPkgList[0].SellerId)
	require.Equal(t, order.Packages[0].SellerId, subPkgList[1].SellerId)
	require.Equal(t, 2, len(subPkgList))
	require.Equal(t, int64(2), totalPages)
}

func TestExitsById_Success(t *testing.T) {
	defer removeCollection()
	order, err := createOrderAndSave()
	require.Nil(t, err, "createOrderAndSave failed")
	require.NotEmpty(t, order.OrderId, "createOrderAndSave failed, order id not generated")
	ctx, _ := context.WithCancel(context.Background())
	result, err := subPkgRepo.ExistsById(ctx, order.Packages[1].Subpackages[0].ItemId)
	require.Nil(t, err, "subPkgRepo.ExistsById failed")
	require.True(t, result)
}

func removeCollection() {
	if _, err := mongoAdapter.DeleteMany(databaseName, collectionName, bson.M{}); err != nil {
	}
}

func insert(order *entities.Order) (*entities.Order, error) {

	if order.OrderId == 0 {
		order.OrderId = entities.GenerateOrderId()
		mapItemIds := make(map[int]uint64, 64)

		for i := 0; i < len(order.Packages); i++ {
			order.Packages[i].OrderId = order.OrderId
			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				for {
					random := int(entities.GenerateRandomNumber())
					if _, ok := mapItemIds[random]; ok {
						continue
					}
					mapItemIds[random] = order.Packages[i].SellerId
					itemId, _ := strconv.Atoi(strconv.Itoa(int(order.OrderId)) + strconv.Itoa(random))
					order.Packages[i].Subpackages[j].ItemId = uint64(itemId)
					order.Packages[i].Subpackages[j].SellerId = order.Packages[i].SellerId
					order.Packages[i].Subpackages[j].OrderId = order.OrderId
					order.Packages[i].Subpackages[j].CreatedAt = time.Now().UTC()
					order.Packages[i].Subpackages[j].UpdatedAt = time.Now().UTC()
					break
				}
			}
		}

		order.CreatedAt = time.Now().UTC()
		var insertOneResult, err = mongoAdapter.InsertOne(databaseName, collectionName, &order)
		if err != nil {
			if mongoAdapter.IsDupError(err) {
				for mongoAdapter.IsDupError(err) {
					insertOneResult, err = mongoAdapter.InsertOne(databaseName, collectionName, &order)
				}
			} else {
				return nil, err
			}
		}
		order.ID = insertOneResult.InsertedID.(primitive.ObjectID)
	} else {
		order.CreatedAt = time.Now().UTC()
		var insertOneResult, err = mongoAdapter.InsertOne(databaseName, collectionName, &order)
		if err != nil {
			return nil, err
		}
		order.ID = insertOneResult.InsertedID.(primitive.ObjectID)
	}
	return order, nil
}

func insertWithoutChangeTime(order *entities.Order) (*entities.Order, error) {

	if order.OrderId == 0 {
		order.OrderId = entities.GenerateOrderId()
		mapItemIds := make(map[int]uint64, 64)

		for i := 0; i < len(order.Packages); i++ {
			order.Packages[i].OrderId = order.OrderId
			for j := 0; j < len(order.Packages[i].Subpackages); j++ {
				for {
					random := int(entities.GenerateRandomNumber())
					if _, ok := mapItemIds[random]; ok {
						continue
					}
					mapItemIds[random] = order.Packages[i].SellerId
					itemId, _ := strconv.Atoi(strconv.Itoa(int(order.OrderId)) + strconv.Itoa(random))
					order.Packages[i].Subpackages[j].ItemId = uint64(itemId)
					order.Packages[i].Subpackages[j].SellerId = order.Packages[i].SellerId
					order.Packages[i].Subpackages[j].OrderId = order.OrderId
					break
				}
			}
		}

		order.CreatedAt = time.Now().UTC()
		var insertOneResult, err = mongoAdapter.InsertOne(databaseName, collectionName, &order)
		if err != nil {
			if mongoAdapter.IsDupError(err) {
				for mongoAdapter.IsDupError(err) {
					insertOneResult, err = mongoAdapter.InsertOne(databaseName, collectionName, &order)
				}
			} else {
				return nil, err
			}
		}
		order.ID = insertOneResult.InsertedID.(primitive.ObjectID)
	} else {
		order.CreatedAt = time.Now().UTC()
		var insertOneResult, err = mongoAdapter.InsertOne(databaseName, collectionName, &order)
		if err != nil {
			return nil, err
		}
		order.ID = insertOneResult.InsertedID.(primitive.ObjectID)
	}
	return order, nil
}

func createOrderAndSave() (*entities.Order, error) {
	return insert(createOrder())
}

func createOrder() *entities.Order {

	paymentRequest := entities.PaymentRequest{
		Amount:    75400000,
		Currency:  "IRR",
		Gateway:   "AAP",
		CreatedAt: time.Now().UTC(),
	}

	paymentResponse := entities.PaymentResponse{
		Result:      true,
		Reason:      "",
		Description: "",
		CallBackUrl: "http://baman.io/payment-service",
		InvoiceId:   12345678946,
		PaymentId:   "r3r434ef45d",
		CreatedAt:   time.Now().UTC(),
	}

	paymentResult := entities.PaymentResult{
		Result:      true,
		Reason:      "",
		PaymentId:   "r3r434ef45d",
		InvoiceId:   12345678946,
		Amount:      75400000,
		CardNumMask: "545498******4553",
		CreatedAt:   time.Now().UTC(),
	}

	buyerInfo := entities.BuyerInfo{
		BuyerId:    6453563,
		FirstName:  "Sina",
		LastName:   "Tadayon",
		Phone:      "09124234234",
		Mobile:     "09123343534",
		Email:      "sina.tadayon@baman.io",
		NationalId: "00598342521",
		Gender:     "male",
		IP:         "127.0.0.1",
		FinanceInfo: entities.FinanceInfo{
			Iban:          "IR9450345802934803",
			CardNumber:    "4444555533332222",
			AccountNumber: "293.6000.9439283.1",
			BankName:      "passargad",
		},
		ShippingAddress: entities.AddressInfo{
			FirstName:     "Ali Reza",
			LastName:      "Rastegar",
			Address:       "Tehran, Narmak, Golestan.st",
			Phone:         "0217734873",
			Mobile:        "091284345",
			Country:       "Iran",
			City:          "Tehran",
			Province:      "Tehran",
			Neighbourhood: "Chizar",
			Location: &entities.Location{
				Type:        "Point",
				Coordinates: []float64{-72.7738706, 41.6332836},
			},
			ZipCode: "1645630586",
		},
	}

	newOrder := entities.Order{
		OrderId: 0,
		Version: 0,
		PaymentService: []entities.PaymentService{{
			PaymentRequest:  &paymentRequest,
			PaymentResponse: &paymentResponse,
			PaymentResult:   &paymentResult,
		}},
		SystemPayment: entities.SystemPayment{
			PayToBuyer: []entities.PayToBuyerInfo{{
				PaymentRequest:  &paymentRequest,
				PaymentResponse: &paymentResponse,
				PaymentResult:   &paymentResult,
			}},
			PayToSeller: []entities.PayToSellerInfo{{
				PaymentRequest:  &paymentRequest,
				PaymentResponse: &paymentResponse,
				PaymentResult:   &paymentResult,
			}},
			PayToMarket: []entities.PayToMarket{{
				PaymentRequest:  &paymentRequest,
				PaymentResponse: &paymentResponse,
				PaymentResult:   &paymentResult,
			}},
		},
		Status:    "New",
		BuyerInfo: buyerInfo,
		Invoice: entities.Invoice{
			GrandTotal:     75400000,
			Subtotal:       73000000,
			Discount:       15600000,
			Currency:       "IRR",
			ShipmentTotal:  5700000,
			PaymentMethod:  "IPG",
			PaymentGateway: "APP",
			CartRule:       nil,
			Voucher: &entities.Voucher{
				Amount: 230000,
				Code:   "Market",
				Details: &entities.VoucherDetails{
					StartDate:        time.Now().UTC(),
					EndDate:          time.Now().UTC(),
					Type:             "Value",
					MaxDiscountValue: 1000,
					MinBasketValue:   13450,
				},
			},
		},
		Packages: []entities.PackageItem{
			{
				SellerId: 129384234,
				OrderId:  0,
				Version:  0,
				Invoice: entities.PackageInvoice{
					Subtotal:       2873423,
					Discount:       9283443,
					ShipmentAmount: 98734,
				},
				SellerInfo: &entities.SellerProfile{
					SellerId: 129384234,
					GeneralInfo: &entities.GeneralSellerInfo{
						ShopDisplayName:          "Sazgar",
						Type:                     "",
						Email:                    "info@sazgar.com",
						LandPhone:                "02834709",
						MobilePhone:              "1836491827346",
						Website:                  "www.sazgar.com",
						Province:                 "tehran",
						City:                     "tehran",
						Neighborhood:             "joradan",
						PostalAddress:            "jordan, shaghayegh",
						PostalCode:               "1254754",
						IsVATObliged:             false,
						VATCertificationImageURL: "http://test.faza.io",
					},
					CorporationInfo: &entities.CorporateSellerInfo{
						CompanyRegisteredName:     "avazhang",
						CompanyRegistrationNumber: "10237128366",
						CompanyRationalId:         "1823128434",
						TradeNumber:               "19293712937",
					},
					IndividualInfo: &entities.IndividualSellerInfo{
						FirstName:          "Sazgar",
						FamilyName:         "Sazgar",
						NationalId:         "3254534334",
						NationalIdFrontURL: "http://adkuhfadlf",
						NationalIdBackURL:  "http://adkuhfadlf",
					},
					ReturnInfo: &entities.ReturnInfo{
						Country:       "Iran",
						Province:      "Tehran",
						City:          "Tehran",
						Neighborhood:  "Tehran",
						PostalAddress: "joradan",
						PostalCode:    "28349394332",
					},
					ContactPerson: &entities.SellerContactPerson{
						FirstName:   "Sazgar",
						FamilyName:  "Sazgar",
						MobilePhone: "9324729348",
						Email:       "sazgar@sazgar.com",
					},
					ShipmentInfo: &entities.SellerShipmentInfo{
						SameCity: &entities.PricePlan{
							Threshold:        934858,
							BelowPrice:       92384729,
							ReactionTimeDays: 98293484,
						},
						DifferentCity: &entities.PricePlan{
							Threshold:        934858,
							BelowPrice:       92384729,
							ReactionTimeDays: 98293484,
						},
					},
					FinanceData: &entities.SellerFinanceData{
						Iban:                    "405872058724850",
						AccountHolderFirstName:  "sazgar",
						AccountHolderFamilyName: "sazgar",
					},
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				},
				ShipmentSpec: entities.ShipmentSpec{
					CarrierNames:   []string{"Post", "Snap"},
					CarrierProduct: "Post Express",
					CarrierType:    "Standard",
					ShippingCost:   1249348,
					VoucherAmount:  3242344,
					Currency:       "IRR",
					ReactionTime:   2,
					ShippingTime:   8,
					ReturnTime:     24,
					Details:        "no return",
				},
				Subpackages: []entities.Subpackage{
					{
						ItemId:   0,
						SellerId: 129384234,
						OrderId:  0,
						Version:  0,
						Items: []entities.Item{
							{
								SKU:         "yt545-34",
								InventoryId: "1111111111",
								Title:       "Mobile",
								Brand:       "Nokia",
								Guaranty:    "Sazegar",
								Category:    "Electronic",
								Image:       "",
								Returnable:  false,
								Quantity:    5,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "0",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit:              1270000,
									Total:             7450000,
									Original:          1270000,
									Special:           1000000,
									Discount:          23000,
									SellerCommission:  5334444,
									Currency:          "IRR",
									ApplicableVoucher: false,
								},
							},
							{
								SKU:         "ye-564634",
								InventoryId: "11111999999",
								Title:       "TV",
								Brand:       "Nokia",
								Guaranty:    "Sazegar",
								Category:    "Electronic",
								Image:       "",
								Returnable:  false,
								Quantity:    2,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "2",
									"Width":     "120cm",
									"Height":    "110cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit:              3270000,
									Total:             87450000,
									Original:          21270000,
									Special:           10000000,
									Discount:          230000,
									SellerCommission:  5334444,
									Currency:          "IRR",
									ApplicableVoucher: false,
								},
							},
						},
						Shipments: &entities.Shipment{
							ShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
						},
						Tracking: entities.Progress{
							StateName:  "0.NewOrder",
							StateIndex: 0,
							Action: &entities.Action{
								Name:      "BuyerCancel",
								Type:      "OrderBuyerCancel",
								Data:      nil,
								Result:    "Success",
								Reasons:   nil,
								CreatedAt: time.Now().UTC(),
							},
							States: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											Type:      "OrderBuyerCancel",
											Data:      nil,
											Result:    "Success",
											Reasons:   nil,
											CreatedAt: time.Now().UTC(),
										},
									},
									CreatedAt: time.Now().UTC(),
								},
							},
						},
						Status:    "New",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
						DeletedAt: nil,
					},
					{
						ItemId:   0,
						SellerId: 129384234,
						OrderId:  0,
						Version:  0,
						Items: []entities.Item{
							{
								SKU:         "gd534-34344",
								InventoryId: "2222222222",
								Title:       "Laptop",
								Brand:       "Lenovo",
								Guaranty:    "Iranargham",
								Category:    "Electronic",
								Image:       "",
								Returnable:  true,
								Quantity:    5,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "0",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit:              1270000,
									Total:             8750000,
									Original:          1270000,
									Special:           1000000,
									Discount:          2355434,
									SellerCommission:  5334444,
									Currency:          "IRR",
									ApplicableVoucher: true,
								},
							},
							{
								SKU:         "5645-yer434",
								InventoryId: "22222888888",
								Title:       "AllInOne",
								Brand:       "Lazada",
								Guaranty:    "Iranargham",
								Category:    "Electronic",
								Image:       "",
								Returnable:  true,
								Quantity:    2,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "2",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit:              3270000,
									Total:             12750000,
									Original:          2270000,
									Special:           100000,
									Discount:          2355434,
									SellerCommission:  5334444,
									Currency:          "IRR",
									ApplicableVoucher: true,
								},
							},
						},
						Shipments: &entities.Shipment{
							ShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
						},
						Tracking: entities.Progress{
							StateName:  "0.NewOrder",
							StateIndex: 0,
							Action: &entities.Action{
								Name:      "BuyerCancel",
								Type:      "OrderBuyerCancel",
								Data:      nil,
								Result:    "Success",
								Reasons:   nil,
								CreatedAt: time.Now().UTC(),
							},
							States: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											Type:      "OrderBuyerCancel",
											Data:      nil,
											Result:    "Success",
											Reasons:   nil,
											CreatedAt: time.Now().UTC(),
										},
									},
									CreatedAt: time.Now().UTC(),
								},
							},
						},
						Status:    "New",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
						DeletedAt: nil,
					},
				},
				Status:    "NEW",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				DeletedAt: nil,
			},
			{
				SellerId: 99988887777,
				OrderId:  0,
				Version:  0,
				Invoice: entities.PackageInvoice{
					Subtotal:       2873423,
					Discount:       9283443,
					ShipmentAmount: 98734,
				},
				SellerInfo: &entities.SellerProfile{
					SellerId: 99988887777,
					GeneralInfo: &entities.GeneralSellerInfo{
						ShopDisplayName:          "Sazgar",
						Type:                     "",
						Email:                    "info@sazgar.com",
						LandPhone:                "02834709",
						MobilePhone:              "1836491827346",
						Website:                  "www.sazgar.com",
						Province:                 "tehran",
						City:                     "tehran",
						Neighborhood:             "joradan",
						PostalAddress:            "jordan, shaghayegh",
						PostalCode:               "1254754",
						IsVATObliged:             false,
						VATCertificationImageURL: "http://test.faza.io",
					},
					CorporationInfo: &entities.CorporateSellerInfo{
						CompanyRegisteredName:     "avazhang",
						CompanyRegistrationNumber: "10237128366",
						CompanyRationalId:         "1823128434",
						TradeNumber:               "19293712937",
					},
					IndividualInfo: &entities.IndividualSellerInfo{
						FirstName:          "Sazgar",
						FamilyName:         "Sazgar",
						NationalId:         "3254534334",
						NationalIdFrontURL: "http://adkuhfadlf",
						NationalIdBackURL:  "http://adkuhfadlf",
					},
					ReturnInfo: &entities.ReturnInfo{
						Country:       "Iran",
						Province:      "Tehran",
						City:          "Tehran",
						Neighborhood:  "Tehran",
						PostalAddress: "joradan",
						PostalCode:    "28349394332",
					},
					ContactPerson: &entities.SellerContactPerson{
						FirstName:   "Sazgar",
						FamilyName:  "Sazgar",
						MobilePhone: "9324729348",
						Email:       "sazgar@sazgar.com",
					},
					ShipmentInfo: &entities.SellerShipmentInfo{
						SameCity: &entities.PricePlan{
							Threshold:        934858,
							BelowPrice:       92384729,
							ReactionTimeDays: 98293484,
						},
						DifferentCity: &entities.PricePlan{
							Threshold:        934858,
							BelowPrice:       92384729,
							ReactionTimeDays: 98293484,
						},
					},
					FinanceData: &entities.SellerFinanceData{
						Iban:                    "405872058724850",
						AccountHolderFirstName:  "sazgar",
						AccountHolderFamilyName: "sazgar",
					},
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				},
				ShipmentSpec: entities.ShipmentSpec{
					CarrierNames:   []string{"Post", "Snap"},
					CarrierProduct: "Post Express",
					CarrierType:    "Standard",
					ShippingCost:   1249348,
					VoucherAmount:  3242344,
					Currency:       "IRR",
					ReactionTime:   2,
					ShippingTime:   8,
					ReturnTime:     24,
					Details:        "no return",
				},
				Subpackages: []entities.Subpackage{
					{
						ItemId:   0,
						SellerId: 99988887777,
						OrderId:  0,
						Version:  0,
						Items: []entities.Item{
							{
								SKU:         "trrer-5343fdf",
								InventoryId: "55555555555",
								Title:       "Mobile",
								Brand:       "Nokia",
								Guaranty:    "Sazegar",
								Category:    "Electronic",
								Image:       "",
								Returnable:  false,
								Quantity:    5,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "0",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit:              1270000,
									Total:             7340000,
									Original:          1270000,
									Special:           1000000,
									Discount:          23000,
									SellerCommission:  5334444,
									Currency:          "IRR",
									ApplicableVoucher: false,
								},
							},
							{
								SKU:         "uer5434-5343",
								InventoryId: "555554444444",
								Title:       "MobileMini",
								Brand:       "Mac",
								Guaranty:    "Sazegar",
								Category:    "Electronic",
								Image:       "",
								Returnable:  false,
								Quantity:    3,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "3",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit:              2270000,
									Total:             6340000,
									Original:          4270000,
									Special:           100000,
									Discount:          2343000,
									SellerCommission:  533444,
									Currency:          "IRR",
									ApplicableVoucher: false,
								},
							},
						},
						Shipments: &entities.Shipment{
							ShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
						},
						Tracking: entities.Progress{
							StateName:  "0.NewOrder",
							StateIndex: 0,
							Action: &entities.Action{
								Name:      "BuyerCancel",
								Type:      "OrderBuyerCancel",
								Data:      nil,
								Result:    "Success",
								Reasons:   nil,
								CreatedAt: time.Now().UTC(),
							},
							States: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											Type:      "OrderBuyerCancel",
											Data:      nil,
											Result:    "Success",
											Reasons:   nil,
											CreatedAt: time.Now().UTC(),
										},
									},
									CreatedAt: time.Now().UTC(),
								},
							},
						},
						Status:    "New",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
						DeletedAt: nil,
					},
					{
						ItemId:   0,
						SellerId: 99988887777,
						OrderId:  0,
						Version:  0,
						Items: []entities.Item{
							{
								SKU:         "5456",
								InventoryId: "3333333333333",
								Title:       "PC",
								Brand:       "HP",
								Guaranty:    "Iranargham",
								Category:    "Electronic",
								Image:       "",
								Returnable:  true,
								Quantity:    3,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "3",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit:              1270000,
									Total:             5646700,
									Original:          7340000,
									Special:           1000000,
									Discount:          2355434,
									SellerCommission:  5334444,
									Currency:          "IRR",
									ApplicableVoucher: true,
								},
							},
							{
								SKU:         "uet-54634",
								InventoryId: "33333666666",
								Title:       "ZeroClient",
								Brand:       "Lardan",
								Guaranty:    "Iranargham",
								Category:    "Electronic",
								Image:       "",
								Returnable:  true,
								Quantity:    3,
								Reasons:     nil,
								Attributes: map[string]string{
									"Quantity":  "3",
									"Width":     "5cm",
									"Height":    "7cm",
									"Length":    "2m",
									"Weight":    "5kg",
									"Color":     "Blue",
									"Materials": "Stone",
								},
								Invoice: entities.ItemInvoice{
									Unit:              7270000,
									Total:             4646700,
									Original:          2340000,
									Special:           1000000,
									Discount:          45355434,
									SellerCommission:  5334444,
									Currency:          "IRR",
									ApplicableVoucher: true,
								},
							},
						},
						Shipments: &entities.Shipment{
							ShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
							ReturnShipmentDetail: &entities.ShippingDetail{
								CarrierName:    "Post",
								ShippingMethod: "Normal",
								TrackingNumber: "545349534958349",
								Image:          "",
								Description:    "",
								ShippedDate:    time.Now().UTC(),
								CreatedAt:      time.Now().UTC(),
							},
						},
						Tracking: entities.Progress{
							StateName:  "0.NewOrder",
							StateIndex: 0,
							Action: &entities.Action{
								Name:      "BuyerCancel",
								Type:      "OrderBuyerCancel",
								Data:      nil,
								Result:    "Success",
								Reasons:   nil,
								CreatedAt: time.Now().UTC(),
							},
							States: []entities.State{
								{
									Name:  "1.New",
									Index: 1,
									Actions: []entities.Action{
										{
											Name:      "BuyerCancel",
											Type:      "OrderBuyerCancel",
											Data:      nil,
											Result:    "Success",
											Reasons:   nil,
											CreatedAt: time.Now().UTC(),
										},
									},
									CreatedAt: time.Now().UTC(),
								},
							},
						},
						Status:    "New",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
						DeletedAt: nil,
					},
				},
				Status:    "NEW",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				DeletedAt: nil,
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		DeletedAt: nil,
	}

	return &newOrder
}

func getOrder(orderId uint64) (*entities.Order, error) {
	var order entities.Order
	singleResult := mongoAdapter.FindOne(databaseName, collectionName, bson.D{{"orderId", orderId}, {"deletedAt", nil}})
	if err := singleResult.Err(); err != nil {
		return nil, err
	}

	if err := singleResult.Decode(&order); err != nil {
		return nil, err
	}

	return &order, nil
}
