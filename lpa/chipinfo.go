package lpa

import sgp22 "github.com/KilimcininKorOglu/euicc-go/v2"

// ChipInfo contains comprehensive information about the eUICC chip.
// This is a convenience structure that aggregates data from multiple
// ES10 functions, similar to lpac's "chip info" command.
type ChipInfo struct {
	// EID is the unique identifier of the eUICC
	EID string

	// ConfiguredAddresses contains the default SM-DP+ and root SM-DS addresses
	ConfiguredAddresses *EUICCConfiguredAddresses

	// Info2 contains detailed eUICC information including:
	// - Memory/storage information (ExtCardResource)
	// - Version information
	// - Capabilities (UICC and RSP)
	// - Security and certification data
	Info2 *sgp22.EUICCInfo2

	// RulesAuthorisationTable contains profile policy authorization rules.
	// This may be nil if no RAT is configured on the eUICC.
	RulesAuthorisationTable []*sgp22.RulesAuthorisationTable
}

// ChipInfo retrieves comprehensive information about the eUICC chip.
// This is a convenience function that calls multiple ES10 functions
// and aggregates the results into a single structure.
//
// The function attempts to retrieve all available information, but will
// not fail if some optional information (like RAT) is unavailable.
// Only EID retrieval failure will cause the function to return an error.
//
// Example usage:
//
//	client, _ := lpa.New(&lpa.Options{Channel: driver})
//	info, err := client.ChipInfo()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("EID: %s\n", info.EID)
//	fmt.Printf("Free NV Memory: %d bytes\n", info.Info2.ExtCardResource.FreeNonVolatileMemory)
//	fmt.Printf("Free Volatile Memory: %d bytes\n", info.Info2.ExtCardResource.FreeVolatileMemory)
func (c *Client) ChipInfo() (*ChipInfo, error) {
	var info ChipInfo
	var err error

	// EID is required - fail if we can't get it
	eidBytes, err := c.EID()
	if err != nil {
		return nil, err
	}
	info.EID = hexEncodeEID(eidBytes)

	// ConfiguredAddresses is optional - ignore errors
	info.ConfiguredAddresses, _ = c.EUICCConfiguredAddresses()

	// EUICCInfo2 is highly recommended - ignore errors but try to get it
	info.Info2, _ = c.EUICCInfo2Parsed()

	// RulesAuthorisationTable is optional - ignore errors
	info.RulesAuthorisationTable, _ = c.GetRAT()

	return &info, nil
}

// hexEncodeEID converts EID bytes to a hex string
func hexEncodeEID(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	result := make([]byte, len(data)*2)
	const hexChars = "0123456789ABCDEF"
	for i, b := range data {
		result[i*2] = hexChars[b>>4]
		result[i*2+1] = hexChars[b&0x0f]
	}
	return string(result)
}
