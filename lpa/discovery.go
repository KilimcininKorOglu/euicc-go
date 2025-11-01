package lpa

import (
	"context"
	"fmt"
	"net/url"

	sgp22 "github.com/KilimcininKorOglu/euicc-go/v2"
)

// DiscoveredProfile represents a profile discovered from SM-DS server.
type DiscoveredProfile struct {
	// EventID is the unique identifier for this discovery event
	EventID string

	// SMDPAddress is the SM-DP+ server address where the profile can be downloaded
	SMDPAddress string
}

// DiscoverProfilesOptions provides configuration for profile discovery from SM-DS.
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

const (
	// DefaultSMDSAddress is the default GSMA SM-DS server address
	DefaultSMDSAddress = "lpa.ds.gsma.com"
)

// DiscoverProfiles discovers available profiles from an SM-DS (Subscription Manager Discovery Service) server.
//
// This function automates the profile discovery workflow specified in SGP.22 section 5.8:
//  1. Retrieves eUICC challenge and info
//  2. Initiates authentication with SM-DS server
//  3. Authenticates the SM-DS server to the eUICC
//  4. Retrieves the list of pending profile downloads
//
// The function returns a list of discovered profiles, each containing an event ID
// and the SM-DP+ server address where the profile can be downloaded from.
//
// Discovery is particularly useful when you don't have an activation code but want
// to check if the eUICC has any pending profile downloads configured by mobile operators.
//
// Example usage:
//
//	// Discover from default GSMA SM-DS
//	profiles, err := client.DiscoverProfiles(nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, profile := range profiles {
//	    fmt.Printf("Found profile: Event=%s, SM-DP+=%s\n",
//	        profile.EventID, profile.SMDPAddress)
//	}
//
//	// Discover from custom SM-DS with IMEI
//	profiles, err := client.DiscoverProfiles(&lpa.DiscoverProfilesOptions{
//	    SMDSAddress: "prod.smds.rsp.goog",
//	    IMEI: []byte("123456789012345"),
//	})
//
// See https://aka.pw/sgp22/v2.5#page=212 (Section 5.8, SM-DS Discovery)
func (c *Client) DiscoverProfiles(opts *DiscoverProfilesOptions) ([]*DiscoveredProfile, error) {
	if opts == nil {
		opts = &DiscoverProfilesOptions{}
	}

	// Use default SM-DS if not specified
	smdsAddress := opts.SMDSAddress
	if smdsAddress == "" {
		smdsAddress = DefaultSMDSAddress
	}

	// Build SM-DS URL
	smdsURL := &url.URL{
		Scheme: "https",
		Host:   smdsAddress,
	}

	// Call the lower-level Discovery function
	eventEntries, err := c.Discovery(smdsURL, opts.IMEI)
	if err != nil {
		return nil, fmt.Errorf("SM-DS discovery failed: %w", err)
	}

	// Convert EventEntry list to DiscoveredProfile list
	profiles := make([]*DiscoveredProfile, len(eventEntries))
	for i, entry := range eventEntries {
		profiles[i] = &DiscoveredProfile{
			EventID:     entry.EventID,
			SMDPAddress: entry.Address,
		}
	}

	return profiles, nil
}

// DiscoverAndDownload is a convenience function that discovers profiles from SM-DS
// and automatically downloads the first available profile.
//
// This combines DiscoverProfiles() and DownloadProfile() into a single operation,
// which is useful for automated provisioning scenarios where you want to download
// any available profile without manual intervention.
//
// If no profiles are found, returns nil with no error.
// If multiple profiles are found, only the first one is downloaded.
//
// Example usage:
//
//	// Discover and download first available profile
//	result, err := client.DiscoverAndDownload(context.Background(), nil, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if result != nil {
//	    fmt.Printf("Downloaded profile\n")
//	} else {
//	    fmt.Println("No profiles available for download")
//	}
func (c *Client) DiscoverAndDownload(ctx context.Context, discoveryOpts *DiscoverProfilesOptions, downloadOpts *DownloadOptions) (*sgp22.LoadBoundProfilePackageResponse, error) {
	// Discover available profiles
	profiles, err := c.DiscoverProfiles(discoveryOpts)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	// No profiles found
	if len(profiles) == 0 {
		return nil, nil
	}

	// Download the first available profile
	firstProfile := profiles[0]

	// Build activation code from discovered SM-DP+ address
	ac := &ActivationCode{
		SMDP: &url.URL{
			Scheme: "https",
			Host:   firstProfile.SMDPAddress,
		},
	}

	// Download the profile
	result, err := c.DownloadProfile(ctx, ac, downloadOpts)
	if err != nil {
		return nil, fmt.Errorf("download from %s failed: %w", firstProfile.SMDPAddress, err)
	}

	return result, nil
}
