package frame

type iFrameBodyImpl struct {
	body interface{}
}

func NewBody() IFrameBody {
	return &iFrameBodyImpl{}
}

func NewBodyOf(body interface{}) IFrameBody {
	return &iFrameBodyImpl{body}
}

func NewBodyFrom(body IFrameBody) IFrameBody {
	return &iFrameBodyImpl{body.Content()}
}

func (frame *iFrameBodyImpl) SetContent(body interface{}) {
	frame.body = body
}

func (frame iFrameBodyImpl) Content() interface{} {
	return frame.body
}

func (frame *iFrameBodyImpl) Copy() IFrameBody {
	return NewBodyFrom(frame)
}

func (frame *iFrameBodyImpl) CopyFrom(body IFrameBody) {
	frame.body = body.Content()
}
