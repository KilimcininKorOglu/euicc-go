# Profile Discovery API Documentation

This document provides comprehensive documentation for the profile discovery features in euicc-go.

**Last Updated:** 2025-11-01

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
- [Common Use Cases](#common-use-cases)
- [SM-DS Servers](#sm-ds-servers)
- [Error Handling](#error-handling)
- [Comparison with lpac](#comparison-with-lpac)

---

## Overview

The Profile Discovery API enables automatic discovery of available eSIM profiles from SM-DS (Subscription Manager Discovery Service) servers without requiring activation codes. This is particularly useful when mobile operators have pre-configured profiles for a device but the user doesn't have the activation code.

### What is SM-DS?

SM-DS (Subscription Manager - Discovery Service) is a server that maintains a registry of pending profile downloads for eUICCs. When a mobile operator wants to provision a profile to a device, they can register it with an SM-DS server, and the device can discover it automatically using the eUICC's EID.

### How Does Discovery Work?

The discovery process follows SGP.22 v2.5 Section 5.8:

1. **Get eUICC Challenge** - Retrieve cryptographic challenge from eUICC
2. **Initiate Authentication** - Start authentication with SM-DS server
3. **Authenticate Server** - Verify SM-DS server to eUICC
4. **Retrieve Events** - Get list of pending profile downloads

---

## Quick Start

### Basic Usage - Discover from Default SM-DS

```go
package main

import (
    "fmt"
    "log"

    "github.com/KilimcininKorOglu/euicc-go/lpa"
    "github.com/KilimcininKorOglu/euicc-go/driver/qmi"
)

func main() {
    // Initialize driver (example with QMI)
    driver, err := qmi.New("/dev/cdc-wdm0", 1)
    if err != nil {
        log.Fatal(err)
    }
    defer driver.Disconnect()

    // Create LPA client
    client, err := lpa.New(&lpa.Options{
        Channel: driver,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Discover profiles from default GSMA SM-DS
    profiles, err := client.DiscoverProfiles(nil)
    if err != nil {
        log.Fatal(err)
    }

    // Display discovered profiles
    if len(profiles) == 0 {
        fmt.Println("No profiles available for download")
        return
    }

    fmt.Printf("Found %d profile(s):\n", len(profiles))
    for i, profile := range profiles {
        fmt.Printf("%d. Event ID: %s\n", i+1, profile.EventID)
        fmt.Printf("   SM-DP+ Server: %s\n", profile.SMDPAddress)
    }
}
```

### Discover and Download in One Step

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/KilimcininKorOglu/euicc-go/lpa"
    "github.com/KilimcininKorOglu/euicc-go/driver/qmi"
)

func main() {
    driver, err := qmi.New("/dev/cdc-wdm0", 1)
    if err != nil {
        log.Fatal(err)
    }
    defer driver.Disconnect()

    client, err := lpa.New(&lpa.Options{
        Channel: driver,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Discover and download first available profile
    result, err := client.DiscoverAndDownload(context.Background(), nil, nil)
    if err != nil {
        log.Fatal(err)
    }

    if result != nil {
        fmt.Println("✓ Profile downloaded successfully!")
    } else {
        fmt.Println("No profiles available for download")
    }
}
```

---

## API Reference

### Types

#### `DiscoveredProfile`

Represents a profile discovered from SM-DS server.

```go
type DiscoveredProfile struct {
    // EventID is the unique identifier for this discovery event
    EventID string

    // SMDPAddress is the SM-DP+ server address where the profile can be downloaded
    SMDPAddress string
}
```

#### `DiscoverProfilesOptions`

Configuration options for profile discovery.

```go
type DiscoverProfilesOptions struct {
    // SMDSAddress is the SM-DS server address to query for available profiles.
    // If not specified, defaults to GSMA's production SM-DS: "lpa.ds.gsma.com"
    //
    // Other well-known SM-DS servers:
    // - Google: "prod.smds.rsp.goog"
    // - eSIM Discovery: "lpa.live.esimdiscovery.com"
    SMDSAddress string

    // IMEI is the device IMEI to use during authentication.
    // This is optional and may be nil for devices without IMEI.
    IMEI []byte
}
```

### Functions

#### `DiscoverProfiles`

```go
func (c *Client) DiscoverProfiles(
    opts *DiscoverProfilesOptions,
) ([]*DiscoveredProfile, error)
```

Discovers available profiles from an SM-DS server.

**Parameters:**
- `opts`: Configuration options (nil for defaults)

**Returns:**
- Slice of `DiscoveredProfile` containing available profiles
- Error if discovery fails

**Default Behavior (opts=nil):**
- Uses GSMA SM-DS: `lpa.ds.gsma.com`
- No IMEI authentication

**Example:**
```go
// Discover from default GSMA SM-DS
profiles, err := client.DiscoverProfiles(nil)
if err != nil {
    log.Fatal(err)
}

for _, profile := range profiles {
    fmt.Printf("Event: %s, SM-DP+: %s\n",
        profile.EventID, profile.SMDPAddress)
}
```

#### `DiscoverAndDownload`

```go
func (c *Client) DiscoverAndDownload(
    ctx context.Context,
    discoveryOpts *DiscoverProfilesOptions,
    downloadOpts *DownloadOptions,
) (*sgp22.LoadBoundProfilePackageResponse, error)
```

Discovers profiles from SM-DS and automatically downloads the first available profile.

**Parameters:**
- `ctx`: Context for cancellation and timeout
- `discoveryOpts`: Discovery configuration (nil for defaults)
- `downloadOpts`: Download configuration (nil for defaults)

**Returns:**
- `LoadBoundProfilePackageResponse` if profile was downloaded
- `nil` if no profiles were found
- Error if discovery or download fails

**Behavior:**
- Discovers all available profiles
- Returns `nil` (no error) if no profiles found
- Downloads only the **first** profile if multiple are found

**Example:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result, err := client.DiscoverAndDownload(ctx, nil, &lpa.DownloadOptions{
    OnProgress: func(stage lpa.DownloadStage) {
        fmt.Printf("Progress: %s\n", stage)
    },
})

if err != nil {
    log.Fatal(err)
}

if result != nil {
    fmt.Println("✓ Profile downloaded successfully!")
} else {
    fmt.Println("No profiles available")
}
```

---

## Common Use Cases

### 1. Discover from Custom SM-DS Server

```go
// Use Google's SM-DS instead of GSMA
profiles, err := client.DiscoverProfiles(&lpa.DiscoverProfilesOptions{
    SMDSAddress: "prod.smds.rsp.goog",
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d profile(s) from Google SM-DS\n", len(profiles))
```

### 2. Discover with IMEI Authentication

```go
import "encoding/hex"

// Some SM-DS servers require IMEI for authentication
imei, _ := hex.DecodeString("313233343536373839303132333435") // "12345678901234" in hex

profiles, err := client.DiscoverProfiles(&lpa.DiscoverProfilesOptions{
    SMDSAddress: "lpa.ds.gsma.com",
    IMEI:        imei,
})
if err != nil {
    log.Fatal(err)
}
```

### 3. Try Multiple SM-DS Servers

```go
func discoverFromMultipleServers(client *lpa.Client) ([]*lpa.DiscoveredProfile, error) {
    servers := []string{
        "lpa.ds.gsma.com",           // GSMA
        "prod.smds.rsp.goog",        // Google
        "lpa.live.esimdiscovery.com", // eSIM Discovery
    }

    for _, server := range servers {
        fmt.Printf("Trying %s...\n", server)

        profiles, err := client.DiscoverProfiles(&lpa.DiscoverProfilesOptions{
            SMDSAddress: server,
        })

        if err != nil {
            fmt.Printf("  Failed: %v\n", err)
            continue
        }

        if len(profiles) > 0 {
            fmt.Printf("  ✓ Found %d profile(s)\n", len(profiles))
            return profiles, nil
        }

        fmt.Printf("  No profiles available\n")
    }

    return nil, fmt.Errorf("no profiles found on any SM-DS server")
}
```

### 4. Display Profile Details Before Download

```go
// First, discover available profiles
profiles, err := client.DiscoverProfiles(nil)
if err != nil {
    log.Fatal(err)
}

if len(profiles) == 0 {
    fmt.Println("No profiles available")
    return
}

// Display profiles and let user choose
fmt.Printf("Available profiles:\n")
for i, profile := range profiles {
    fmt.Printf("%d. SM-DP+: %s (Event: %s)\n",
        i+1, profile.SMDPAddress, profile.EventID)
}

// Get user selection
fmt.Print("Select profile to download (1-n): ")
var selection int
fmt.Scanln(&selection)

if selection < 1 || selection > len(profiles) {
    log.Fatal("Invalid selection")
}

selectedProfile := profiles[selection-1]

// Download selected profile
ac := &lpa.ActivationCode{
    SMDP: &url.URL{
        Scheme: "https",
        Host:   selectedProfile.SMDPAddress,
    },
}

result, err := client.DownloadProfile(context.Background(), ac, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Println("✓ Profile downloaded successfully!")
```

### 5. Automatic Provisioning Workflow

```go
func autoProvision(client *lpa.Client) error {
    fmt.Println("Starting automatic provisioning...")

    // Step 1: Discover profiles
    fmt.Println("1. Discovering available profiles...")
    profiles, err := client.DiscoverProfiles(nil)
    if err != nil {
        return fmt.Errorf("discovery failed: %w", err)
    }

    if len(profiles) == 0 {
        return fmt.Errorf("no profiles available for this device")
    }

    fmt.Printf("   Found %d profile(s)\n", len(profiles))

    // Step 2: Download first profile
    fmt.Println("2. Downloading profile...")
    ctx := context.Background()
    result, err := client.DiscoverAndDownload(ctx, nil, &lpa.DownloadOptions{
        OnProgress: func(stage lpa.DownloadStage) {
            fmt.Printf("   Progress: %s\n", stage)
        },
    })
    if err != nil {
        return fmt.Errorf("download failed: %w", err)
    }

    // Step 3: Enable the downloaded profile
    fmt.Println("3. Enabling profile...")
    profileList, err := client.ListProfile()
    if err != nil {
        return fmt.Errorf("failed to list profiles: %w", err)
    }

    // Find the newly downloaded profile (last in list)
    if len(profileList) > 0 {
        lastProfile := profileList[len(profileList)-1]
        err = client.EnableProfile(lastProfile.ICCID, true)
        if err != nil {
            return fmt.Errorf("failed to enable profile: %w", err)
        }
    }

    // Step 4: Process notifications
    fmt.Println("4. Processing notifications...")
    _, err = client.ProcessAllNotifications(&lpa.ProcessNotificationsOptions{
        AutoRemove:      true,
        ContinueOnError: true,
    })
    if err != nil {
        fmt.Printf("   Warning: Some notifications failed: %v\n", err)
    }

    fmt.Println("✓ Automatic provisioning completed!")
    return nil
}
```

### 6. Discovery with Timeout

```go
import "time"

func discoverWithTimeout(client *lpa.Client, timeout time.Duration) ([]*lpa.DiscoveredProfile, error) {
    type result struct {
        profiles []*lpa.DiscoveredProfile
        err      error
    }

    ch := make(chan result, 1)

    go func() {
        profiles, err := client.DiscoverProfiles(nil)
        ch <- result{profiles, err}
    }()

    select {
    case res := <-ch:
        return res.profiles, res.err
    case <-time.After(timeout):
        return nil, fmt.Errorf("discovery timeout after %v", timeout)
    }
}

// Usage
profiles, err := discoverWithTimeout(client, 30*time.Second)
if err != nil {
    log.Fatal(err)
}
```

---

## SM-DS Servers

### Well-Known SM-DS Servers

| Provider | Address | Notes |
|----------|---------|-------|
| **GSMA** | `lpa.ds.gsma.com` | Default, most widely supported |
| **Google** | `prod.smds.rsp.goog` | Used by Google Fi and partners |
| **eSIM Discovery** | `lpa.live.esimdiscovery.com` | Third-party discovery service |

### Default SM-DS Server

euicc-go uses **GSMA's production SM-DS** (`lpa.ds.gsma.com`) as the default server. This is the most widely supported SM-DS server and is recommended for most use cases.

```go
const DefaultSMDSAddress = "lpa.ds.gsma.com"
```

### Choosing an SM-DS Server

1. **Use default (GSMA)** - Best compatibility
2. **Operator-specific** - If your operator specifies a SM-DS
3. **Try multiple** - Some devices may be registered on different SM-DS servers

---

## Error Handling

### Common Errors

#### 1. No Profiles Found

```go
profiles, err := client.DiscoverProfiles(nil)
if err != nil {
    log.Fatal(err)
}

if len(profiles) == 0 {
    fmt.Println("No profiles available for this device")
    fmt.Println("Possible reasons:")
    fmt.Println("  - No operator has provisioned profiles for this EID")
    fmt.Println("  - Profiles may be on a different SM-DS server")
    fmt.Println("  - Device may not be registered with SM-DS")
}
```

#### 2. SM-DS Server Unreachable

```go
profiles, err := client.DiscoverProfiles(&lpa.DiscoverProfilesOptions{
    SMDSAddress: "invalid.smds.server",
})

if err != nil {
    if strings.Contains(err.Error(), "SM-DS discovery failed") {
        fmt.Println("Cannot reach SM-DS server")
        fmt.Println("Check:")
        fmt.Println("  - Internet connection")
        fmt.Println("  - SM-DS server address")
        fmt.Println("  - Firewall/proxy settings")
    }
}
```

#### 3. Authentication Failure

```go
profiles, err := client.DiscoverProfiles(nil)
if err != nil {
    if strings.Contains(err.Error(), "authenticate") {
        fmt.Println("Authentication with SM-DS failed")
        fmt.Println("This may indicate:")
        fmt.Println("  - EID not recognized by SM-DS")
        fmt.Println("  - IMEI required but not provided")
        fmt.Println("  - Server-side authentication issue")
    }
}
```

### Best Practices

1. **Always check for empty results**
   ```go
   profiles, err := client.DiscoverProfiles(nil)
   if err != nil {
       // Handle error
   }
   if len(profiles) == 0 {
       // Handle no profiles case
   }
   ```

2. **Use DiscoverAndDownload with caution**
   ```go
   // Good: Check what will be downloaded
   profiles, _ := client.DiscoverProfiles(nil)
   if len(profiles) > 1 {
       fmt.Println("Multiple profiles found, please choose manually")
   }

   // Use DiscoverAndDownload only when appropriate
   result, _ := client.DiscoverAndDownload(ctx, nil, nil)
   ```

3. **Provide user feedback during discovery**
   ```go
   fmt.Println("Discovering available profiles...")
   profiles, err := client.DiscoverProfiles(nil)
   if err != nil {
       fmt.Printf("Discovery failed: %v\n", err)
       return
   }
   fmt.Printf("Discovery complete: %d profile(s) found\n", len(profiles))
   ```

---

## Comparison with lpac

| Feature | lpac | euicc-go | Notes |
|---------|------|----------|-------|
| **Discover Profiles** | `lpac profile discovery` | `DiscoverProfiles()` | Identical functionality |
| **Custom SM-DS** | `-s <address>` | `SMDSAddress: "..."` | euicc-go uses options struct |
| **IMEI Parameter** | `-i <imei>` | `IMEI: []byte{...}` | euicc-go uses byte slice |
| **Default SM-DS** | `lpa.ds.gsma.com` | `lpa.ds.gsma.com` | Same default |
| **Output Format** | JSON array | `[]*DiscoveredProfile` slice | euicc-go provides programmatic access |
| **One-Step Download** | Not available | `DiscoverAndDownload()` | euicc-go provides convenience function |

### lpac vs euicc-go Examples

**lpac:**
```bash
# Discover from default SM-DS
lpac profile discovery

# Discover from custom SM-DS
lpac profile discovery -s prod.smds.rsp.goog

# Discover with IMEI
lpac profile discovery -i 123456789012345
```

**euicc-go:**
```go
// Discover from default SM-DS
profiles, err := client.DiscoverProfiles(nil)

// Discover from custom SM-DS
profiles, err := client.DiscoverProfiles(&lpa.DiscoverProfilesOptions{
    SMDSAddress: "prod.smds.rsp.goog",
})

// Discover with IMEI
profiles, err := client.DiscoverProfiles(&lpa.DiscoverProfilesOptions{
    IMEI: []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x5},
})
```

---

## Performance Considerations

### Discovery Time

A typical discovery operation takes:
1. **eUICC communication** (~200-300ms) - Get challenge and info
2. **SM-DS authentication** (~500-1000ms) - HTTPS round trip
3. **Event retrieval** (~300-500ms) - HTTPS round trip

**Total:** ~1-2 seconds per discovery attempt

### Caching

Discovery results are **not cached** by euicc-go. If you need to cache results, implement it at the application level:

```go
type discoveryCache struct {
    profiles  []*lpa.DiscoveredProfile
    timestamp time.Time
    ttl       time.Duration
}

func (c *discoveryCache) Get(client *lpa.Client) ([]*lpa.DiscoveredProfile, error) {
    if time.Since(c.timestamp) < c.ttl && c.profiles != nil {
        return c.profiles, nil
    }

    profiles, err := client.DiscoverProfiles(nil)
    if err != nil {
        return nil, err
    }

    c.profiles = profiles
    c.timestamp = time.Now()
    return profiles, nil
}
```

---

## Thread Safety

The `DiscoverProfiles` and `DiscoverAndDownload` functions are **not thread-safe**. Do not call them concurrently from multiple goroutines on the same `lpa.Client` instance.

**Safe:**
```go
// Sequential discovery
profiles1, _ := client.DiscoverProfiles(nil)
profiles2, _ := client.DiscoverProfiles(opts)
```

**Unsafe:**
```go
// Concurrent discovery - DO NOT DO THIS
go client.DiscoverProfiles(nil)
go client.DiscoverProfiles(opts)
```

---

## See Also

- [SGP.22 v2.5 Specification](https://aka.pw/sgp22/v2.5) - Section 5.8 (SM-DS Discovery)
- [ES10b.GetEUICCChallengeAndInfo](https://aka.pw/sgp22/v2.5#page=187)
- [ES11.AuthenticateClient](https://aka.pw/sgp22/v2.5#page=212)
- [lpac profile discovery](https://github.com/estkme-group/lpac)

---

## Changelog

### v1.4.0 (Current)
- ✅ Added `DiscoverProfiles()` function
- ✅ Added `DiscoverAndDownload()` convenience function
- ✅ Added `DiscoveredProfile` structure
- ✅ Added `DiscoverProfilesOptions` configuration
- ✅ Default GSMA SM-DS server support
- ✅ Custom SM-DS server support
- ✅ Optional IMEI authentication support
