package encoder

import "layeh.com/gopus"

type OpusEncoder struct {
	encoder *gopus.Encoder
}

func NewOpusEncoder() (*OpusEncoder, error) {
	enc, err := gopus.NewEncoder(48000, 2, gopus.Audio)
	if err != nil {
		return nil, err
	}
	return &OpusEncoder{encoder: enc}, nil
}

func (o *OpusEncoder) Encode(pcm []int16, frameSize, maxBytes int) ([]byte, error) {
	return o.encoder.Encode(pcm, frameSize, maxBytes)
}
