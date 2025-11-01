package sgp22

import "github.com/KilimcininKorOglu/euicc-go/bertlv"

func mustMarshalValue(tlv *bertlv.TLV, err error) *bertlv.TLV {
	if err != nil {
		panic(err)
	}
	return tlv
}
