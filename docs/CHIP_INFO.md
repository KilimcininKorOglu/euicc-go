# Chip Information API

This document describes the chip information features added to the euicc-go library, which provide comprehensive access to eUICC chip details including memory/storage information, capabilities, and configuration.

## Overview

The chip info API allows you to retrieve detailed information about an eUICC (eSIM) chip, similar to lpac's `chip info` command. This is useful for:

- Checking available storage space before installing profiles
- Verifying chip capabilities and firmware versions
- Retrieving unique chip identifiers (EID)
- Understanding chip configuration and authorization rules

## Quick Start

### Get All Chip Information at Once

The simplest way to get chip information is using the `ChipInfo()` convenience function:

```go
package main

import (
    "fmt"
    "log"

    "github.com/KilimcininKorOglu/euicc-go/lpa"
    "github.com/KilimcininKorOglu/euicc-go/driver/qmi"
)

func main() {
    // Initialize your driver (QMI example)
    driver, err := qmi.New("/dev/cdc-wdm0", 1)
    if err != nil {
        log.Fatal(err)
    }

    // Create LPA client
    client, err := lpa.New(&lpa.Options{
        Channel: driver,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Get all chip information
    info, err := client.ChipInfo()
    if err != nil {
        log.Fatal(err)
    }

    // Display information
    fmt.Printf("EID: %s\n", info.EID)
    fmt.Printf("Free Storage: %d bytes (%.2f MB)\n",
        info.Info2.ExtCardResource.FreeNonVolatileMemory,
        float64(info.Info2.ExtCardResource.FreeNonVolatileMemory)/1024/1024)
    fmt.Printf("Free RAM: %d bytes\n",
        info.Info2.ExtCardResource.FreeVolatileMemory)
    fmt.Printf("Firmware Version: %s\n", info.Info2.EUICCFirmwareVer)
    fmt.Printf("SGP.22 Version: %s\n", info.Info2.SVN)
}
```

## API Reference

### ChipInfo Structure

The `ChipInfo` structure aggregates all chip information:

```go
type ChipInfo struct {
    // EID is the unique identifier of the eUICC (32 hex characters)
    EID string

    // ConfiguredAddresses contains the default SM-DP+ and root SM-DS addresses
    ConfiguredAddresses *EUICCConfiguredAddresses

    // Info2 contains detailed eUICC information
    Info2 *sgp22.EUICCInfo2

    // RulesAuthorisationTable contains profile policy authorization rules
    // May be nil if no RAT is configured
    RulesAuthorisationTable []*sgp22.RulesAuthorisationTable
}
```

### EUICCInfo2 Structure

The `EUICCInfo2` structure contains detailed chip information:

```go
type EUICCInfo2 struct {
    // Version Information
    ProfileVersion        string  // Profile specification version
    SVN                   string  // SGP.22 specification version
    EUICCFirmwareVer      string  // Chip firmware version
    TS102241Version       string  // JavaCard/ETSI version
    GlobalPlatformVersion string  // GlobalPlatform version
    PPVersion             string  // Protection Profile version

    // Memory/Storage Information (IMPORTANT!)
    ExtCardResource struct {
        InstalledApplication  uint32  // Number of installed applications
        FreeNonVolatileMemory uint32  // Free persistent storage in bytes
        FreeVolatileMemory    uint32  // Free RAM in bytes
    }

    // Capabilities
    UICCCapability []string  // Card capabilities (USIM, ISIM, etc.)
    RSPCapability  []string  // Remote SIM Provisioning capabilities

    // Security
    EUICCCiPKIdListForVerification []string  // Public key IDs for verification
    EUICCCiPKIdListForSigning      []string  // Public key IDs for signing
    ForbiddenProfilePolicyRules    []string  // Forbidden policy rules

    // Classification
    EUICCCategory string  // Category: "basicEuicc", "mediumEuicc", "contactlessEuicc", "other"

    // Certification
    SASAccreditationNumber string
    CertificationDataObject struct {
        PlatformLabel    string
        DiscoveryBaseURL string
    }
}
```

### Individual Functions

If you need specific information only, you can call individual functions:

#### Get Parsed EUICCInfo2

```go
info2, err := client.EUICCInfo2Parsed()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Available storage: %d bytes\n",
    info2.ExtCardResource.FreeNonVolatileMemory)
fmt.Printf("UICC Capabilities: %v\n", info2.UICCCapability)
```

#### Get EID

```go
eid, err := client.EID()
if err != nil {
    log.Fatal(err)
}

// Convert to hex string
eidHex := hex.EncodeToString(eid)
fmt.Printf("EID: %s\n", strings.ToUpper(eidHex))
```

#### Get Configured Addresses

```go
addresses, err := client.EUICCConfiguredAddresses()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Default SM-DP+: %s\n", addresses.DefaultSMDPAddress)
fmt.Printf("Root SM-DS: %s\n", addresses.RootSMDSAddress)
```

#### Get Rules Authorisation Table

```go
ratList, err := client.GetRAT()
if err != nil {
    log.Fatal(err)
}

for _, rat := range ratList {
    fmt.Printf("PPR IDs: %v\n", rat.PPRIds)
    for _, op := range rat.AllowedOperators {
        fmt.Printf("  Allowed Operator - PLMN: %s, GID1: %s, GID2: %s\n",
            op.PLMN, op.GID1, op.GID2)
    }
}
```

## Common Use Cases

### Check if Enough Space for Profile Installation

```go
info, err := client.ChipInfo()
if err != nil {
    log.Fatal(err)
}

const requiredSpace = 500 * 1024 // 500 KB
if info.Info2.ExtCardResource.FreeNonVolatileMemory < requiredSpace {
    fmt.Printf("Warning: Low storage! Only %d bytes available\n",
        info.Info2.ExtCardResource.FreeNonVolatileMemory)
} else {
    fmt.Println("Sufficient storage available for profile installation")
}
```

### Display Chip Capabilities

```go
info, err := client.ChipInfo()
if err != nil {
    log.Fatal(err)
}

fmt.Println("Chip Capabilities:")
fmt.Println("UICC Capabilities:")
for _, cap := range info.Info2.UICCCapability {
    fmt.Printf("  - %s\n", cap)
}

fmt.Println("\nRSP Capabilities:")
for _, cap := range info.Info2.RSPCapability {
    fmt.Printf("  - %s\n", cap)
}
```

### Check JavaCard/Multiple USIM Support

```go
info, err := client.ChipInfo()
if err != nil {
    log.Fatal(err)
}

hasJavaCard := false
hasMultipleUSIM := false

for _, cap := range info.Info2.UICCCapability {
    if cap == "javacard" {
        hasJavaCard = true
    }
    if cap == "multipleUsimSupport" {
        hasMultipleUSIM = true
    }
}

fmt.Printf("JavaCard Support: %v\n", hasJavaCard)
fmt.Printf("Multiple USIM Support: %v\n", hasMultipleUSIM)
```

### Export Chip Info as JSON

```go
import "encoding/json"

info, err := client.ChipInfo()
if err != nil {
    log.Fatal(err)
}

jsonData, err := json.MarshalIndent(info, "", "  ")
if err != nil {
    log.Fatal(err)
}

fmt.Println(string(jsonData))
```

## Field Descriptions

### ExtCardResource Fields

- **`InstalledApplication`**: Number of applications currently installed on the eUICC
- **`FreeNonVolatileMemory`**: Available persistent storage space in bytes. This is the space available for installing new eSIM profiles
- **`FreeVolatileMemory`**: Available RAM in bytes. This affects runtime operations

### UICC Capability Values

Common capability values include:

- `contactlessSupport` - NFC/contactless support
- `usimSupport` - 3G/UMTS SIM support
- `isimSupport` - IMS SIM support
- `csimSupport` - CDMA SIM support
- `akaMilenage`, `akaCave`, `akaTuak128`, `akaTuak256` - Authentication algorithms
- `gbaAuthenUsim`, `gbaAuthenISim` - Generic Bootstrapping Architecture
- `eapClient` - Extensible Authentication Protocol
- `javacard` - JavaCard support
- `multipleUsimSupport` - Multiple USIM profiles support
- `multipleIsimSupport` - Multiple ISIM profiles support

### RSP Capability Values

- `additionalProfile` - Support for multiple profiles
- `crlSupport` - Certificate Revocation List support
- `rpmSupport` - Remote Profile Management support
- `testProfileSupport` - Test profile support
- `deviceInfoExtensibilitySupport` - Extended device info support

### eUICC Categories

- `basicEuicc` - Basic eUICC functionality
- `mediumEuicc` - Medium feature set
- `contactlessEuicc` - Contactless/NFC enabled
- `other` - Other/unspecified category

## Error Handling

The `ChipInfo()` function is designed to be fault-tolerant:

- **Required**: Only EID retrieval failure causes an error
- **Optional**: ConfiguredAddresses, Info2, and RAT failures are silently ignored
- Individual fields may be `nil` or empty if not available

For more strict error handling, use individual functions:

```go
// This will fail if Info2 cannot be retrieved
info2, err := client.EUICCInfo2Parsed()
if err != nil {
    log.Printf("Failed to get EUICCInfo2: %v", err)
    return
}

// This will fail if RAT cannot be retrieved
ratList, err := client.GetRAT()
if err != nil {
    log.Printf("Failed to get RAT: %v", err)
    return
}
```

## Comparison with lpac

This implementation provides equivalent functionality to lpac's `chip info` command:

| lpac | euicc-go | Description |
|------|----------|-------------|
| `eidValue` | `ChipInfo.EID` | Chip identifier |
| `EuiccConfiguredAddresses` | `ChipInfo.ConfiguredAddresses` | Server addresses |
| `EUICCInfo2.extCardResource.freeNonVolatileMemory` | `Info2.ExtCardResource.FreeNonVolatileMemory` | Available storage |
| `EUICCInfo2.extCardResource.freeVolatileMemory` | `Info2.ExtCardResource.FreeVolatileMemory` | Available RAM |
| `EUICCInfo2.uiccCapability` | `Info2.UICCCapability` | Card capabilities |
| `EUICCInfo2.rspCapability` | `Info2.RSPCapability` | RSP capabilities |
| `rulesAuthorisationTable` | `ChipInfo.RulesAuthorisationTable` | Authorization rules |

## Performance Considerations

- `ChipInfo()` makes multiple APDU calls to the eUICC (EID, addresses, Info2, RAT)
- Typical execution time: 500ms - 2s depending on the hardware
- For better performance when you only need specific information, use individual functions
- Results can be cached as chip information rarely changes

## Thread Safety

The LPA client is not thread-safe. Do not call chip info functions from multiple goroutines simultaneously on the same client instance. Create separate client instances for concurrent access.

## Known Issues and Fixes

### Version History

#### v1.2.2 (2025-11-03) - Critical Bug Fixes

**Fixed: ExtCardResource fields returning zero values**

Prior to this version, the `ExtCardResource` fields (memory information) were incorrectly parsed and returned zero values:
- `InstalledApplication`: Always returned 0
- `FreeNonVolatileMemory`: Always returned 0
- `FreeVolatileMemory`: Always returned 0

**Root Cause:**
Two bugs were discovered and fixed:

1. **ExtCardResource Nested TLV Parsing**: The tag `0x84` (ExtCardResource) is a primitive tag containing nested TLVs. The previous implementation attempted to use `bertlv.TLV.UnmarshalBinary()` on the value bytes, which failed to properly parse the nested structure. The fix implements manual TLV parsing similar to lpac's implementation.

2. **Tag Comparison Bug**: The code was using `Tag.Value()` in switch statements, comparing against full byte values like `0x84`. However, `Tag.Value()` only returns the tag number (bits 4-0), not the full byte. For tag `0x84`, it returns `4`, not `0x84`. The fix uses `Tag.If(class, form, number)` for proper tag matching.

**Changes Made:**
- `v2/euiccinfo2.go`: Rewrote `ExtCardResource.UnmarshalBERTLV()` with manual TLV parsing (lines 124-199)
- `v2/euiccinfo2.go`: Changed all tag comparisons from `switch Tag.Value()` to `switch` with `Tag.If()` conditions (lines 77-122)
- Added comprehensive unit tests in `v2/euiccinfo2_test.go`

**Verification:**
```bash
# Before fix:
"ext_card_resource": {
  "installed_application": 0,
  "free_non_volatile_memory": 0,
  "free_volatile_memory": 0
}

# After fix:
"ext_card_resource": {
  "installed_application": 5,
  "free_non_volatile_memory": 524288,  // Example: 512 KB
  "free_volatile_memory": 65536        // Example: 64 KB
}
```

This fix ensures compatibility with lpac's chip info implementation and correctly parses all eUICC memory and resource information.

## Related Documentation

- [SGP.22 v2.5 Specification](https://aka.pw/sgp22/v2.5)
- [ETSI TS 102 226](https://www.etsi.org/deliver/etsi_ts/102200_102299/102226/) - Remote APDU structure for UICC
- [Main README](../README.md) - Library overview and setup
- [CLAUDE.md](../CLAUDE.md) - Project architecture and development guide

## License

This implementation follows the same license as the main euicc-go project.
