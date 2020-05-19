//
// this package contains reasons of actions like
// cancellation action or return action which will
// show corresponding reason of buyer for seller
//
package utils

import (
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
)

const (
	ReasonResponsibleBuyer  = "BUYER"
	ReasonResponsibleSeller = "SELLER"
	ReasonResponsibleNone   = "NONE"
)

func Responsible(responsible string) int32 {
	switch responsible {
	case ReasonResponsibleNone:
		return 0;
	case ReasonResponsibleBuyer:
		return 1;
	case ReasonResponsibleSeller:
		return 2;
	default:
		return 0;
	}
}

type ReasonConfigs map[string]entities.ReasonConfig

func InitialReasonConfig() (mp ReasonConfigs) {
	list := []entities.ReasonConfig{
		{
			Key:            "change_of_mind",
			Translation:    "نظرم درباره خرید این کالا تغییر کرد",
			HasDescription: false,
			Cancel:         true,
			Return:         true,
			IsActive:       true,
			Responsible:    ReasonResponsibleBuyer,
		},
		{
			Key:            "forgot_voucher",
			Translation:    "فراموش کردم کد تخفیف را اعمال کنم",
			HasDescription: false,
			Cancel:         true,
			Return:         false,
			IsActive:       true,
			Responsible:    ReasonResponsibleBuyer,
		},
		{
			Key:            "delivery_too_long",
			Translation:    "زمان ارسال طولانی است",
			HasDescription: false,
			Cancel:         true,
			Return:         false,
			IsActive:       true,
			Responsible:    ReasonResponsibleNone,
		},
		{
			Key:            "found_better_price",
			Translation:    "در فروشگاهی دیگر این محصول را با قیمت پایین‌تر پیدا کردم",
			HasDescription: false,
			Cancel:         true,
			Return:         true,
			IsActive:       true,
			Responsible:    ReasonResponsibleBuyer,
		},
		{
			Key:            "defective_or_damaged",
			Translation:    "محصول معیوب است و یا کار نمی‌کند",
			HasDescription: false,
			Cancel:         false,
			Return:         true,
			IsActive:       true,
			Responsible:    ReasonResponsibleSeller,
		},
		{
			Key:            "differs_from_content",
			Translation:    "کالا با عکس یا مشخصات درج شده روی سایت مطابقت ندارد",
			HasDescription: false,
			Cancel:         false,
			Return:         true,
			IsActive:       true,
			Responsible:    ReasonResponsibleSeller,
		},
		{
			Key:            "fake_or_counterfeit",
			Translation:    "محصول ارسال شده اصل نیست",
			HasDescription: false,
			Cancel:         false,
			Return:         true,
			IsActive:       true,
			Responsible:    ReasonResponsibleSeller,
		},
		{
			Key:            "low_quality",
			Translation:    "محصول کیفیت پایینی دارد",
			HasDescription: false,
			Cancel:         false,
			Return:         true,
			IsActive:       true,
			Responsible:    ReasonResponsibleBuyer,
		},
		{
			Key:            "missing_parts",
			Translation:    "محصول ناقص ارسال شده است",
			HasDescription: false,
			Cancel:         false,
			Return:         true,
			IsActive:       true,
			Responsible:    ReasonResponsibleSeller,
		},
		{
			Key:            "does_not_fit",
			Translation:    "سایز محصول مناسب نبود (برای کالاهای دسته پوشاک)",
			HasDescription: false,
			Cancel:         false,
			Return:         true,
			IsActive:       true,
			Responsible:    ReasonResponsibleBuyer,
		},
		{
			Key:            "other",
			Translation:    "سایر (با توضیحات کامل)",
			HasDescription: true,
			Cancel:         true,
			Return:         true,
			IsActive:       true,
			Responsible:    ReasonResponsibleNone,
		},
	}
	mp = make(ReasonConfigs, 0)
	for _, r := range list {
		mp[r.Key] = r
	}
	return
}
