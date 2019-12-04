package frame

type iFrameImpl struct {
	frameHeader IFrameHeader
	frameBody   IFrameBody
}

func New() IFrame {
	return &iFrameImpl{}
}

func NewOf(header IFrameHeader, body IFrameBody) IFrame {
	return &iFrameImpl{frameHeader: header, frameBody: body}
}

func NewFrom(frame IFrame) IFrame {
	return &iFrameImpl{frameHeader: frame.Header(), frameBody: frame.Body()}
}

func (iFrame iFrameImpl) Header() IFrameHeader {
	return iFrame.frameHeader
}

func (iFrame iFrameImpl) Body() IFrameBody {
	return iFrame.frameBody
}

func (iFrame iFrameImpl) Copy() IFrame {
	return NewOf(iFrame.Header().Copy(), iFrame.frameBody.Copy())
}

func (iFrame *iFrameImpl) CopyFrom(frame IFrame) {
	iFrame.frameHeader = frame.Header().Copy()
	iFrame.frameBody = frame.Body().Copy()
}
