package frame

type iFrameHeaderImpl struct {
	header map[string]interface{}
}

func NewHeader(header map[string]interface{}) IFrameHeader {
	return &iFrameHeaderImpl{deepCopy(header)}
}

func NewHeaderOf(frame iFrameHeaderImpl) IFrameHeader {
	return &iFrameHeaderImpl{deepCopy(frame.header)}
}

func (frame iFrameHeaderImpl) KeyExists(key string) bool {
	if _, ok := frame.header[key]; ok {
		return true
	} else {
		return false
	}
}

func (frame iFrameHeaderImpl) Value(key string) interface{} {
	return frame.header[key]
}

func (frame iFrameHeaderImpl) Copy() IFrameHeader {
	return NewHeader(frame.header)
}

func (frame *iFrameHeaderImpl) CopyFrom(header IFrameHeader) {
	rawHeader := header.(*iFrameHeaderImpl)
	frame.header = deepCopy(rawHeader.header)
}

func (frame iFrameHeaderImpl) CopyIfAbsent(header IFrameHeader) {
	rawHeader := header.(*iFrameHeaderImpl)
	for key, value := range rawHeader.header {
		if _, ok := frame.header[key]; !ok {
			frame.header[key] = value
		}
	}
}

func deepCopy(src map[string]interface{}) map[string]interface{} {
	target := make(map[string]interface{}, len(src))
	for key, value := range src {
		target[key] = value
	}
	return target
}
