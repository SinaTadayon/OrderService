package user_service

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.faza.io/go-framework/acl"
	"gitlab.faza.io/go-framework/logger"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
	userclient "gitlab.faza.io/services/user-app-client"
	"google.golang.org/grpc"
	"time"
)

const (
	// ISO8601 standard time format
	layout = "2006-01-02 15:04:05 +0000 MST"
)

type iUserServiceImpl struct {
	client        *userclient.Client
	serverAddress string
	serverPort    int
}

func NewUserService(serverAddress string, serverPort int) IUserService {
	return &iUserServiceImpl{serverAddress: serverAddress, serverPort: serverPort}
}

// TODO refactor fault-tolerant
func (userService *iUserServiceImpl) getUserService(ctx context.Context) error {

	if userService.client != nil {
		return nil
	}

	var err error
	config := &userclient.Config{
		Host:    userService.serverAddress,
		Port:    userService.serverPort,
		Timeout: 5 * time.Second,
	}
	userService.client, err = userclient.NewClient(ctx, config, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.Err("userclient.NewClient failed, address: %s, port: %d, error: %s", userService.serverAddress, userService.serverPort, err)
		return err
	}
	ctx, _ = context.WithTimeout(ctx, config.Timeout)
	//defer cancel()
	_, err = userService.client.Connect(ctx)
	if err != nil {
		logger.Err("userclient.NewClient failed, address: %s, port: %d, error: %s", userService.serverAddress, userService.serverPort, err)
		return err
	}

	return nil
}

func (userService *iUserServiceImpl) UserLogin(ctx context.Context, username, password string) future.IFuture {
	ctx1, _ := context.WithCancel(context.Background())
	if err := userService.getUserService(ctx1); err != nil {
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "UnknownError", errors.Wrap(err, "Connect to UserService Failed")).
			BuildAndSend()
	}

	result, err := userService.client.Login(username, password, ctx1)
	if err != nil {
		logger.Err("UserLogin() => userService.client.Login failed, username: %s, password: %s, error: %s", username, password, err)
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "UnknownError", errors.Wrap(err, "userService.client.Login Failed")).
			BuildAndSend()
	}

	if int(result.Code) != 200 {
		logger.Err("UserLogin() => userService.client.Login failed, username: %s, password: %s, error: %s", username, password, err)

		return future.Factory().SetCapacity(1).
			SetError(future.Forbidden, "User Login Failed", errors.Wrap(err, "User Login Failed")).
			BuildAndSend()
	}

	loginTokens := LoginTokens{
		AccessToken:  result.Data.AccessToken,
		RefreshToken: result.Data.RefreshToken,
	}

	return future.Factory().SetCapacity(1).SetData(loginTokens).BuildAndSend()
}

func (userService iUserServiceImpl) AuthenticateContextToken(ctx context.Context) (*acl.Acl, error) {
	ctx1, _ := context.WithCancel(context.Background())
	if err := userService.getUserService(ctx1); err != nil {
		return nil, err
	}
	access, err := userService.client.VerifyAndGetUserFromContextToken(ctx)
	return access, err
}

func (userService iUserServiceImpl) GetSellerProfile(ctx context.Context, sellerId string) future.IFuture {
	ctx1, _ := context.WithCancel(context.Background())
	if err := userService.getUserService(ctx1); err != nil {
		return future.Factory().SetCapacity(1).
			SetError(future.InternalError, "UnknownError", errors.Wrap(err, "Connect to UserService Failed")).
			BuildAndSend()
	}

	userProfile, err := userService.client.InternalUserGetOne("userId", sellerId, "", ctx)
	if err != nil {
		logger.Err("userService.client.InternalUserGetOne failed, pid: %s, error: %s", sellerId, err)
		return future.Factory().SetCapacity(1).
			SetError(future.NotFound, "sellerId Not Found", errors.Wrap(err, "sellerId Not Found")).
			BuildAndSend()
	}

	sellerProfile := &entities.SellerProfile{
		SellerId: userProfile.Data.UserId,
	}

	if userProfile.Data.Seller.GeneralInfo != nil {
		sellerProfile.GeneralInfo = &entities.GeneralSellerInfo{
			ShopDisplayName:          userProfile.Data.Seller.GeneralInfo.ShopDisplayName,
			Type:                     userProfile.Data.Seller.GeneralInfo.Type,
			Email:                    userProfile.Data.Seller.GeneralInfo.Email,
			LandPhone:                userProfile.Data.Seller.GeneralInfo.LandPhone,
			MobilePhone:              userProfile.Data.Seller.GeneralInfo.MobilePhone,
			Website:                  userProfile.Data.Seller.GeneralInfo.Website,
			PostalAddress:            userProfile.Data.Seller.GeneralInfo.PostalAddress,
			PostalCode:               userProfile.Data.Seller.GeneralInfo.PostalCode,
			IsVATObliged:             userProfile.Data.Seller.GeneralInfo.IsVATObliged,
			VATCertificationImageURL: userProfile.Data.Seller.GeneralInfo.VATCertificationImageURL,
		}

		if userProfile.Data.Seller.GeneralInfo.Province != nil {
			sellerProfile.GeneralInfo.Province = userProfile.Data.Seller.GeneralInfo.Province.Name
		}

		if userProfile.Data.Seller.GeneralInfo.City != nil {
			sellerProfile.GeneralInfo.City = userProfile.Data.Seller.GeneralInfo.City.Name
		}

		if userProfile.Data.Seller.GeneralInfo.Neighborhood != nil {
			sellerProfile.GeneralInfo.Neighborhood = userProfile.Data.Seller.GeneralInfo.Neighborhood.Name
		}
	}

	if userProfile.Data.Seller.CorpInfo != nil {
		sellerProfile.CorporationInfo = &entities.CorporateSellerInfo{
			CompanyRegisteredName:     userProfile.Data.Seller.CorpInfo.CompanyRegisteredName,
			CompanyRegistrationNumber: userProfile.Data.Seller.CorpInfo.CompanyRegistrationNumber,
			CompanyRationalId:         userProfile.Data.Seller.CorpInfo.CompanyRationalID,
			TradeNumber:               userProfile.Data.Seller.CorpInfo.TradeNumber,
		}
	}

	if userProfile.Data.Seller.IndivInfo != nil {
		sellerProfile.IndividualInfo = &entities.IndividualSellerInfo{
			FirstName:          userProfile.Data.Seller.IndivInfo.FirstName,
			FamilyName:         userProfile.Data.Seller.IndivInfo.FamilyName,
			NationalId:         userProfile.Data.Seller.IndivInfo.NationalID,
			NationalIdFrontURL: userProfile.Data.Seller.IndivInfo.NationalIDfrontURL,
			NationalIdBackURL:  userProfile.Data.Seller.IndivInfo.NationalIDbackURL,
		}
	}

	if userProfile.Data.Seller.ReturnInfo != nil {
		sellerProfile.ReturnInfo = &entities.ReturnInfo{

			PostalAddress: userProfile.Data.Seller.ReturnInfo.PostalAddress,
			PostalCode:    userProfile.Data.Seller.ReturnInfo.PostalCode,
		}

		if userProfile.Data.Seller.ReturnInfo.Country != nil {
			sellerProfile.ReturnInfo.Country = userProfile.Data.Seller.ReturnInfo.Country.Name
		}

		if userProfile.Data.Seller.ReturnInfo.Province != nil {
			sellerProfile.ReturnInfo.Province = userProfile.Data.Seller.ReturnInfo.Province.Name
		}

		if userProfile.Data.Seller.ReturnInfo.City != nil {
			sellerProfile.ReturnInfo.City = userProfile.Data.Seller.ReturnInfo.City.Name
		}

		if userProfile.Data.Seller.ReturnInfo.Neighborhood != nil {
			sellerProfile.ReturnInfo.Neighborhood = userProfile.Data.Seller.ReturnInfo.Neighborhood.Name
		}
	}

	if userProfile.Data.Seller.ContactPerson != nil {
		sellerProfile.ContactPerson = &entities.SellerContactPerson{
			FirstName:   userProfile.Data.Seller.ContactPerson.FirstName,
			FamilyName:  userProfile.Data.Seller.ContactPerson.FamilyName,
			MobilePhone: userProfile.Data.Seller.ContactPerson.MobilePhone,
			Email:       userProfile.Data.Seller.ContactPerson.Email,
		}
	}

	if userProfile.Data.Seller.ShipmentInfo != nil {
		sellerProfile.ShipmentInfo = &entities.SellerShipmentInfo{}
		if userProfile.Data.Seller.ShipmentInfo.SameCity != nil {
			sellerProfile.ShipmentInfo.SameCity = &entities.PricePlan{
				Threshold:        userProfile.Data.Seller.ShipmentInfo.SameCity.Threshold,
				BelowPrice:       userProfile.Data.Seller.ShipmentInfo.SameCity.BelowPrice,
				ReactionTimeDays: userProfile.Data.Seller.ShipmentInfo.SameCity.ReactionTimeDays,
			}
		}

		if userProfile.Data.Seller.ShipmentInfo.DifferentCity != nil {
			sellerProfile.ShipmentInfo.DifferentCity = &entities.PricePlan{
				Threshold:        userProfile.Data.Seller.ShipmentInfo.DifferentCity.Threshold,
				BelowPrice:       userProfile.Data.Seller.ShipmentInfo.DifferentCity.BelowPrice,
				ReactionTimeDays: userProfile.Data.Seller.ShipmentInfo.DifferentCity.ReactionTimeDays,
			}
		}
	}

	if userProfile.Data.Seller.FinanceData != nil {
		sellerProfile.FinanceData = &entities.SellerFinanceData{
			Iban:                    userProfile.Data.Seller.FinanceData.Iban,
			AccountHolderFirstName:  userProfile.Data.Seller.FinanceData.AccountHolderFirstName,
			AccountHolderFamilyName: userProfile.Data.Seller.FinanceData.AccountHolderFamilyName,
		}
	}

	timestamp, err := time.Parse(layout, userProfile.Data.CreatedAt)
	if err != nil {
		logger.Err("GetSellerProfile() => createdAt time parse failed, pid: %s, error: %s", sellerId, err)
		timestamp = time.Now()
	}

	sellerProfile.CreatedAt = timestamp
	timestamp, err = time.Parse(layout, userProfile.Data.UpdatedAt)
	if err != nil {
		logger.Err("GetSellerProfile() => updatedAt time parse failed, pid: %s, error: %s", sellerId, err)
		timestamp = time.Now()
	}
	sellerProfile.UpdatedAt = timestamp
	return future.Factory().SetCapacity(1).SetData(sellerProfile).BuildAndSend()
}
