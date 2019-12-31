package frame

import (
	"gitlab.faza.io/order-project/order-service/domain/events"
	"gitlab.faza.io/order-project/order-service/domain/models/entities"
	"gitlab.faza.io/order-project/order-service/infrastructure/future"
)

type Builder struct {
	body   IFrameBody
	header map[string]interface{}
}

func Factory() Builder {
	builder := &Builder{}
	builder.initBuilder(nil, nil, nil)
	return *builder
}

func FactoryOf(frame IFrame) Builder {
	builder := &Builder{}
	builder.initBuilder(frame, nil, nil)
	return *builder
}

func FactoryFromHeader(header IFrameHeader) Builder {
	builder := &Builder{}
	builder.initBuilder(nil, header, nil)
	return *builder
}

func FactoryFromBody(body IFrameBody) Builder {
	builder := &Builder{}
	builder.initBuilder(nil, nil, body)
	return *builder
}

func (builder *Builder) initBuilder(frame IFrame, header IFrameHeader, body IFrameBody) {
	if frame != nil {
		frameHeader := frame.Header().(*iFrameHeaderImpl)
		builder.header = deepCopy(frameHeader.header)
		builder.body = NewBodyFrom(frame.Body())
	} else if header != nil {
		frameHeader := header.(*iFrameHeaderImpl)
		builder.header = deepCopy(frameHeader.header)
		builder.body = NewBody()
	} else if body != nil {
		builder.header = make(map[string]interface{}, 16)
		builder.body = NewBodyFrom(body)
	} else {
		builder.header = make(map[string]interface{}, 16)
		builder.body = NewBody()
	}
}

func (builder Builder) SetHeader(key string, value interface{}) Builder {
	builder.header[key] = value
	return builder
}

func (builder Builder) SetDefaultHeader(key HeaderEnum, value interface{}) Builder {
	builder.header[string(key)] = value
	return builder
}

func (builder Builder) SetBody(body interface{}) Builder {
	builder.body.SetContent(body)
	return builder
}

func (builder Builder) SetSellerId(sellerId uint64) Builder {
	builder.header[string(HeaderPId)] = sellerId
	return builder
}

func (builder Builder) SetInventoryId(inventoryId string) Builder {
	builder.header[string(HeaderInventoryId)] = inventoryId
	return builder
}

func (builder Builder) SetOrderId(orderId uint64) Builder {
	builder.header[string(HeaderOrderId)] = orderId
	return builder
}

func (builder Builder) SetOrder(order *entities.Order) Builder {
	builder.header[string(HeaderOrder)] = order
	return builder
}

func (builder Builder) SetSIds(sid []uint64) Builder {
	builder.header[string(HeaderSIds)] = sid
	return builder
}

func (builder Builder) SetItem(item entities.Item) Builder {
	builder.header[string(HeaderItems)] = item
	return builder
}

func (builder Builder) SetSubpackages(subpackages []*entities.Subpackage) Builder {
	builder.header[string(HeaderSubpackages)] = subpackages
	return builder
}

func (builder Builder) SetSubpackage(subpackage *entities.Subpackage) Builder {
	builder.header[string(HeaderSubpackage)] = subpackage
	return builder
}

func (builder Builder) SetPackage(packageItem *entities.PackageItem) Builder {
	builder.header[string(HeaderPackage)] = packageItem
	return builder
}

func (builder Builder) SetEvent(event events.IEvent) Builder {
	builder.header[string(HeaderEvent)] = event
	return builder
}

func (builder Builder) SetFuture(iFuture future.IFuture) Builder {
	builder.header[string(HeaderFuture)] = iFuture
	return builder
}

func (builder Builder) Build() IFrame {
	return &iFrameImpl{NewHeader(builder.header), builder.body}
}
