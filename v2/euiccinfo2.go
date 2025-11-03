package sgp22

import (
	"encoding/hex"
	"fmt"

	"github.com/KilimcininKorOglu/euicc-go/bertlv"
	"github.com/KilimcininKorOglu/euicc-go/bertlv/primitive"
)

// EUICCInfo2 represents the parsed EUICCInfo2 response from the eUICC.
//
// See https://aka.pw/sgp22/v2.5#page=187 (Section 5.7.8, ES10b.GetEUICCInfo)
type EUICCInfo2 struct {
	ProfileVersion                  string
	SVN                             string
	EUICCFirmwareVer                string
	ExtCardResource                 ExtCardResource
	UICCCapability                  []string
	TS102241Version                 string
	GlobalPlatformVersion           string
	RSPCapability                   []string
	EUICCCiPKIdListForVerification  []string
	EUICCCiPKIdListForSigning       []string
	EUICCCategory                   string
	ForbiddenProfilePolicyRules     []string
	PPVersion                       string
	SASAccreditationNumber          string
	CertificationDataObject         CertificationDataObject
}

// ExtCardResource represents the extended card resource information.
// This contains memory and application information from the eUICC.
type ExtCardResource struct {
	InstalledApplication  uint32 // Number of installed applications
	FreeNonVolatileMemory uint32 // Free persistent memory in bytes (for profile installation)
	FreeVolatileMemory    uint32 // Free volatile (RAM) memory in bytes
}

// CertificationDataObject represents certification information for the eUICC.
type CertificationDataObject struct {
	PlatformLabel    string
	DiscoveryBaseURL string
}

// EUICCCategory represents the category of the eUICC.
type EUICCCategory int

const (
	EUICCCategoryOther        EUICCCategory = 0
	EUICCCategoryBasic        EUICCCategory = 1
	EUICCCategoryMedium       EUICCCategory = 2
	EUICCCategoryContactless  EUICCCategory = 3
)

func (c EUICCCategory) String() string {
	switch c {
	case EUICCCategoryBasic:
		return "basicEuicc"
	case EUICCCategoryMedium:
		return "mediumEuicc"
	case EUICCCategoryContactless:
		return "contactlessEuicc"
	default:
		return "other"
	}
}

// UnmarshalBERTLV parses the EUICCInfo2 response from BER-TLV format.
func (e *EUICCInfo2) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	// EUICCInfo2 is either tag 0xBF20 (32) or 0xBF22 (34)
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 32) &&
		!tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 34) {
		return ErrUnexpectedTag
	}

	for _, child := range tlv.Children {
		// Note: Tag.Value() returns only the tag number (bits 4-0), not the full byte.
		// For context-specific tags like 0x81, Tag.Value() returns 1, not 0x81.
		// We match against the full tag using Tag.If() for clarity.

		switch {
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Primitive, 1): // 0x81 profileVersion
			e.ProfileVersion = parseVersion(child.Value)
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Primitive, 2): // 0x82 svn
			e.SVN = parseVersion(child.Value)
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Primitive, 3): // 0x83 euiccFirmwareVer
			e.EUICCFirmwareVer = parseVersion(child.Value)
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Primitive, 4): // 0x84 extCardResource
			if err := e.ExtCardResource.UnmarshalBERTLV(child); err != nil {
				return err
			}
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Primitive, 5): // 0x85 uiccCapability
			e.UICCCapability = parseUICCCapability(child.Value)
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Primitive, 6): // 0x86 ts102241Version
			e.TS102241Version = parseVersion(child.Value)
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Primitive, 7): // 0x87 globalplatformVersion
			e.GlobalPlatformVersion = parseVersion(child.Value)
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Primitive, 8): // 0x88 rspCapability
			e.RSPCapability = parseRSPCapability(child.Value)
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 9): // 0xA9 euiccCiPKIdListForVerification
			e.EUICCCiPKIdListForVerification = parsePKIdList(child)
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 10): // 0xAA euiccCiPKIdListForSigning
			e.EUICCCiPKIdListForSigning = parsePKIdList(child)
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 11): // 0xAB euiccCategory
			if len(child.Value) > 0 {
				var cat int8
				child.UnmarshalValue(primitive.UnmarshalInt(&cat))
				e.EUICCCategory = EUICCCategory(cat).String()
			}
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Primitive, 25): // 0x99 forbiddenProfilePolicyRules
			e.ForbiddenProfilePolicyRules = parseForbiddenPPR(child.Value)
		case child.Tag.If(bertlv.Universal, bertlv.Primitive, 4): // 0x04 ppVersion (OCTET STRING)
			e.PPVersion = parseVersion(child.Value)
		case child.Tag.If(bertlv.Universal, bertlv.Primitive, 12): // 0x0C sasAccreditationNumber (UTF8String)
			e.SASAccreditationNumber = string(child.Value)
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 12): // 0xAC certificationDataObject
			if err := e.CertificationDataObject.UnmarshalBERTLV(child); err != nil {
				return err
			}
		}
	}

	return nil
}

// UnmarshalBERTLV parses the ExtCardResource from BER-TLV format.
func (e *ExtCardResource) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Primitive, 4) {
		return ErrUnexpectedTag
	}

	// ExtCardResource is encoded as an OCTET STRING (primitive tag 0x84)
	// containing nested TLVs. We need to manually parse these nested TLVs
	// from the value bytes.

	value := tlv.Value
	offset := 0

	for offset < len(value) {
		// Parse tag
		if offset >= len(value) {
			break
		}

		tag := value[offset]
		offset++

		// Parse length
		if offset >= len(value) {
			break
		}

		length := int(value[offset])
		offset++

		// Handle multi-byte length (if bit 7 is set)
		if length&0x80 != 0 {
			numLengthBytes := length & 0x7F
			length = 0
			for i := 0; i < numLengthBytes && offset < len(value); i++ {
				length = (length << 8) | int(value[offset])
				offset++
			}
		}

		// Extract value
		if offset+length > len(value) {
			break
		}

		fieldValue := value[offset : offset+length]
		offset += length

		// Parse the field based on tag
		switch tag {
		case 0x81: // installedApplication
			if len(fieldValue) > 0 {
				var val int32
				// Create temporary TLV for unmarshaling
				tempTLV := bertlv.NewValue(bertlv.ContextSpecific.Primitive(1), fieldValue)
				tempTLV.UnmarshalValue(primitive.UnmarshalInt(&val))
				e.InstalledApplication = uint32(val)
			}
		case 0x82: // freeNonVolatileMemory
			if len(fieldValue) > 0 {
				var val int32
				tempTLV := bertlv.NewValue(bertlv.ContextSpecific.Primitive(2), fieldValue)
				tempTLV.UnmarshalValue(primitive.UnmarshalInt(&val))
				e.FreeNonVolatileMemory = uint32(val)
			}
		case 0x83: // freeVolatileMemory
			if len(fieldValue) > 0 {
				var val int32
				tempTLV := bertlv.NewValue(bertlv.ContextSpecific.Primitive(3), fieldValue)
				tempTLV.UnmarshalValue(primitive.UnmarshalInt(&val))
				e.FreeVolatileMemory = uint32(val)
			}
		}
	}

	return nil
}

// UnmarshalBERTLV parses the CertificationDataObject from BER-TLV format.
func (c *CertificationDataObject) UnmarshalBERTLV(tlv *bertlv.TLV) error {
	if !tlv.Tag.If(bertlv.ContextSpecific, bertlv.Constructed, 12) {
		return ErrUnexpectedTag
	}

	for _, child := range tlv.Children {
		switch {
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Primitive, 0): // 0x80 platformLabel
			c.PlatformLabel = string(child.Value)
		case child.Tag.If(bertlv.ContextSpecific, bertlv.Primitive, 1): // 0x81 discoveryBaseURL
			c.DiscoveryBaseURL = string(child.Value)
		}
	}

	return nil
}

// parseVersion converts BER-encoded version bytes to a dotted string (e.g., "2.1.0")
func parseVersion(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	// Version is typically 3 bytes: major.minor.revision
	if len(data) == 3 {
		return fmt.Sprintf("%d.%d.%d", data[0], data[1], data[2])
	}
	// If not 3 bytes, join all bytes with dots
	result := fmt.Sprintf("%d", data[0])
	for i := 1; i < len(data); i++ {
		result += fmt.Sprintf(".%d", data[i])
	}
	return result
}

// parseUICCCapability parses UICC capability bit string
func parseUICCCapability(data []byte) []string {
	capabilities := []string{
		"contactlessSupport",
		"usimSupport",
		"isimSupport",
		"csimSupport",
		"akaMilenage",
		"akaCave",
		"akaTuak128",
		"akaTuak256",
		"rfu1",
		"rfu2",
		"gbaAuthenUsim",
		"gbaAuthenISim",
		"mbmsAuthenUsim",
		"eapClient",
		"javacard",
		"multos",
		"multipleUsimSupport",
		"multipleIsimSupport",
		"multipleCsimSupport",
		"berTlvFileSupport",
		"dfLinkSupport",
		"catTp",
		"getIdentity",
		"profile-a-x25519",
		"profile-b-p256",
		"suciCalculatorApi",
	}

	return parseBitString(data, capabilities)
}

// parseRSPCapability parses RSP capability bit string
func parseRSPCapability(data []byte) []string {
	capabilities := []string{
		"additionalProfile",
		"crlSupport",
		"rpmSupport",
		"testProfileSupport",
		"deviceInfoExtensibilitySupport",
	}

	return parseBitString(data, capabilities)
}

// parseForbiddenPPR parses forbidden profile policy rules bit string
func parseForbiddenPPR(data []byte) []string {
	rules := []string{
		"pprUpdateControl",
		"ppr1",
		"ppr2",
		"ppr3",
	}

	return parseBitString(data, rules)
}

// parseBitString converts a BER bit string to a list of capability names
func parseBitString(data []byte, names []string) []string {
	if len(data) == 0 {
		return nil
	}

	// First byte is the number of unused bits in the last byte
	unusedBits := int(data[0])
	if len(data) < 2 {
		return nil
	}

	var result []string
	bitData := data[1:]

	bitIndex := 0
	totalBits := len(bitData)*8 - unusedBits

	for _, b := range bitData {
		for bitPos := 7; bitPos >= 0; bitPos-- {
			if bitIndex >= totalBits {
				break
			}
			if bitIndex < len(names) {
				// Check if bit is set
				if (b & (1 << bitPos)) != 0 {
					result = append(result, names[bitIndex])
				}
			}
			bitIndex++
		}
	}

	return result
}

// parsePKIdList parses a list of Public Key Identifiers (as hex strings)
func parsePKIdList(tlv *bertlv.TLV) []string {
	if tlv == nil || len(tlv.Children) == 0 {
		return nil
	}

	var result []string
	for _, child := range tlv.Children {
		if len(child.Value) > 0 {
			result = append(result, hex.EncodeToString(child.Value))
		}
	}

	return result
}
