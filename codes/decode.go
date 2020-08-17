package codec

func DecodeCompact32(data []byte) (U32, error) {
	var d = make([]byte, len(data))
	copy(d, data)
	sb, err := NewBytes(d)
	if err != nil {
		return 0, err
	}
	return sb.ToCompactUInt32()
}

func DecodeString(data []byte) (string, error) {
	var d = make([]byte, len(data))
	copy(d, data)
	sb, err := NewBytes(data)
	if err != nil {
		return "", err
	}
	return sb.ToString()
}

func DecodeUint64(data []byte) (U64, error) {
	var d = make([]byte, len(data))
	copy(d, data)
	sb, err := NewBytes(data)
	if err != nil {
		return 0, err
	}
	return sb.ToUint64()
}
