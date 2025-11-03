package sgp22

import (
	"encoding/hex"
	"testing"

	"github.com/KilimcininKorOglu/euicc-go/bertlv"
	"github.com/stretchr/testify/assert"
)

func TestExtCardResourceUnmarshal(t *testing.T) {
	// This is a sample ExtCardResource TLV from a real eUICC
	// Tag 0x84 (context-specific primitive 4)
	// The value contains nested TLVs:
	// - 0x81: installedApplication
	// - 0x82: freeNonVolatileMemory
	// - 0x83: freeVolatileMemory

	// Example: 30 30 0C (tag 0x84, length 0x0C, value contains nested TLVs)
	// Nested: 81 01 05 (installedApplication = 5)
	//         82 03 0F 42 40 (freeNonVolatileMemory = 1000000)
	//         83 02 07 D0 (freeVolatileMemory = 2000)
	data := "840C810105820301E24083020FA0"
	rawBytes, err := hex.DecodeString(data)
	assert.NoError(t, err)

	var tlv bertlv.TLV
	err = tlv.UnmarshalBinary(rawBytes)
	assert.NoError(t, err)

	t.Logf("TLV Tag: 0x%X, Class: %v, Constructed: %v", tlv.Tag.Value(), tlv.Tag.Class(), tlv.Tag.Constructed())
	t.Logf("TLV Value (hex): %X", tlv.Value)
	t.Logf("TLV Children count: %d", len(tlv.Children))

	var resource ExtCardResource
	err = resource.UnmarshalBERTLV(&tlv)
	assert.NoError(t, err)

	t.Logf("InstalledApplication: %d", resource.InstalledApplication)
	t.Logf("FreeNonVolatileMemory: %d", resource.FreeNonVolatileMemory)
	t.Logf("FreeVolatileMemory: %d", resource.FreeVolatileMemory)

	assert.Equal(t, uint32(5), resource.InstalledApplication)
	assert.Equal(t, uint32(123456), resource.FreeNonVolatileMemory)
	assert.Equal(t, uint32(4000), resource.FreeVolatileMemory)
}

func TestEUICCInfo2UnmarshalWithExtCardResource(t *testing.T) {
	// Minimal EUICCInfo2 with just extCardResource
	// BF22: tag (context-specific constructed 34)
	// Content: just extCardResource
	data := "BF220E" + // tag + length (14 bytes total)
		"840C810105820301E24083020FA0" // extCardResource

	rawBytes, err := hex.DecodeString(data)
	assert.NoError(t, err)

	var tlv bertlv.TLV
	err = tlv.UnmarshalBinary(rawBytes)
	assert.NoError(t, err)

	t.Logf("Root TLV Tag Value: 0x%X", tlv.Tag.Value())
	t.Logf("Root TLV Children: %d", len(tlv.Children))

	if len(tlv.Children) > 0 {
		for i, child := range tlv.Children {
			t.Logf("Child[%d] Tag Value: 0x%X, Class: %v, Constructed: %v",
				i, child.Tag.Value(), child.Tag.Class(), child.Tag.Constructed())
			t.Logf("Child[%d] Value (hex): %X", i, child.Value)
		}
	}

	var info EUICCInfo2
	err = info.UnmarshalBERTLV(&tlv)
	assert.NoError(t, err)

	t.Logf("InstalledApplication: %d", info.ExtCardResource.InstalledApplication)
	t.Logf("FreeNonVolatileMemory: %d", info.ExtCardResource.FreeNonVolatileMemory)
	t.Logf("FreeVolatileMemory: %d", info.ExtCardResource.FreeVolatileMemory)

	assert.Equal(t, uint32(5), info.ExtCardResource.InstalledApplication)
	assert.Equal(t, uint32(123456), info.ExtCardResource.FreeNonVolatileMemory)
	assert.Equal(t, uint32(4000), info.ExtCardResource.FreeVolatileMemory)
}
