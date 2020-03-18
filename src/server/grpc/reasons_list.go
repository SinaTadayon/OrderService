package grpc_server

import (
	"gitlab.faza.io/order-project/order-service/domain/models"
	pb "gitlab.faza.io/protos/order"
)

type reasonsMap map[string]models.ReasonConfig

func (rm reasonsMap) toGRPC() (p []*pb.ReasonDetail) {
	p = make([]*pb.ReasonDetail, 0)
	for _, r := range rm {
		i := &pb.ReasonDetail{
			Key:            r.Key,
			Translation:    r.Translation,
			HasDescription: r.HasDescription,
			Cancel:         r.Cancel,
			Return:         r.Return,
			IsActive:       r.IsActive,
		}
		switch r.Responsible {
		case models.ReasonResponsibleBuyer:
			i.Responsible = pb.ReasonDetail_BUYER
		case models.ReasonResponsibleSeller:
			i.Responsible = pb.ReasonDetail_SELLER
		case models.ReasonResponsibleNone:
			i.Responsible = pb.ReasonDetail_NONE
		default:
			i.Responsible = pb.ReasonDetail_NONE
		}
		p = append(p, i)

	}
	return
}

func createReasonsMap() (mp reasonsMap) {
	list := []models.ReasonConfig{
		{
			Key:            "change_of_mind",
			Translation:    "نظرم درباره خرید این کالا تغییر کرد",
			HasDescription: false,
			Cancel:         true,
			Return:         true,
			IsActive:       true,
			Responsible:    models.ReasonResponsibleBuyer,
		},
		{
			Key:            "forgot_voucher",
			Translation:    "فراموش کردم کد تخفیف را اعمال کنم",
			HasDescription: false,
			Cancel:         true,
			Return:         false,
			IsActive:       true,
			Responsible:    models.ReasonResponsibleBuyer,
		},
		{
			Key:            "delivery_too_long",
			Translation:    "زمان ارسال طولانی است",
			HasDescription: false,
			Cancel:         true,
			Return:         false,
			IsActive:       true,
			Responsible:    models.ReasonResponsibleNone,
		},
		{
			Key:            "found_better_price",
			Translation:    "در فروشگاهی دیگر این محصول را با قیمت پایین‌تر پیدا کردم",
			HasDescription: false,
			Cancel:         true,
			Return:         true,
			IsActive:       true,
			Responsible:    models.ReasonResponsibleBuyer,
		},
		{
			Key:            "defective_or_damaged",
			Translation:    "محصول معیوب است و یا کار نمی‌کند",
			HasDescription: false,
			Cancel:         false,
			Return:         true,
			IsActive:       true,
			Responsible:    models.ReasonResponsibleSeller,
		},
		{
			Key:            "differs_from_content",
			Translation:    "کالا با عکس یا مشخصات درج شده روی سایت مطابقت ندارد",
			HasDescription: false,
			Cancel:         false,
			Return:         true,
			IsActive:       true,
			Responsible:    models.ReasonResponsibleSeller,
		},
		{
			Key:            "fake_or_counterfeit",
			Translation:    "محصول ارسال شده اصل نیست",
			HasDescription: false,
			Cancel:         false,
			Return:         true,
			IsActive:       true,
			Responsible:    models.ReasonResponsibleSeller,
		},
		{
			Key:            "low_quality",
			Translation:    "محصول کیفیت پایینی دارد",
			HasDescription: false,
			Cancel:         false,
			Return:         true,
			IsActive:       true,
			Responsible:    models.ReasonResponsibleBuyer,
		},
		{
			Key:            "missing_parts",
			Translation:    "محصول ناقص ارسال شده است",
			HasDescription: false,
			Cancel:         false,
			Return:         true,
			IsActive:       true,
			Responsible:    models.ReasonResponsibleSeller,
		},
		{
			Key:            "does_not_fit",
			Translation:    "سایز محصول مناسب نبود (برای کالاهای دسته پوشاک)",
			HasDescription: false,
			Cancel:         false,
			Return:         true,
			IsActive:       true,
			Responsible:    models.ReasonResponsibleBuyer,
		},
		{
			Key:            "other",
			Translation:    "سایر (با توضیحات کامل)",
			HasDescription: true,
			Cancel:         true,
			Return:         true,
			IsActive:       true,
			Responsible:    models.ReasonResponsibleNone,
		},
	}
	mp = make(reasonsMap, 0)
	for _, r := range list {
		mp[r.Key] = r
	}
	return
}
