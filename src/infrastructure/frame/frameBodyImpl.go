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
	if body != nil {
		return &iFrameBodyImpl{body.Content()}
	}
	return &iFrameBodyImpl{}
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
	if body != nil {
		frame.body = body.Content()
	}
}
