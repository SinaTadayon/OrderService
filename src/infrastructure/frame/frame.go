package frame

type HeaderEnum string

const (
	HeaderOrder         HeaderEnum = "ORDER"
	HeaderSubpackages   HeaderEnum = "SUBPACKAGES"
	HeaderSubpackage    HeaderEnum = "SUBPACKAGE"
	HeaderPackage       HeaderEnum = "PACKAGE"
	HeaderItems         HeaderEnum = "ITEMS"
	HeaderOrderId       HeaderEnum = "ORDER_ID"
	HeaderSellerId      HeaderEnum = "SELLER_ID"
	HeaderItemId        HeaderEnum = "ITEM_ID"
	HeaderIPAddress     HeaderEnum = "IP_ADDRESS"
	HeaderInventoryId   HeaderEnum = "INVENTORY_ID"
	HeaderFuture        HeaderEnum = "FUTURE"
	HeaderFutureError   HeaderEnum = "FUTURE_ERROR"
	HeaderPaymentResult HeaderEnum = "PAYMENT_RESULT"
	HeaderEvent         HeaderEnum = "EVENT"
	HeaderNewOrder      HeaderEnum = "NEW_ORDER"
)

type IFrame interface {
	Header() IFrameHeader
	Body() IFrameBody
	Copy() IFrame
	CopyFrom(iFrame IFrame)
}

type IFrameHeader interface {
	KeyExists(key string) bool
	Value(key string) interface{}
	Copy() IFrameHeader
	CopyFrom(header IFrameHeader)
	CopyIfAbsent(header IFrameHeader)
}

type IFrameBody interface {
	SetContent(body interface{})
	Content() interface{}
	Copy() IFrameBody
	CopyFrom(body IFrameBody)
}
