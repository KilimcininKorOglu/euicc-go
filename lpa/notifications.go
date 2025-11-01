package lpa

import (
	"fmt"

	sgp22 "github.com/KilimcininKorOglu/euicc-go/v2"
)

// NotificationProcessResult contains the result of processing a single notification.
type NotificationProcessResult struct {
	SequenceNumber sgp22.SequenceNumber
	Success        bool
	Error          error
	Removed        bool
}

// ProcessNotificationsOptions provides configuration for notification processing.
type ProcessNotificationsOptions struct {
	// AutoRemove automatically removes notifications from the eUICC after
	// successful processing (sending to SM-DP+).
	AutoRemove bool

	// ContinueOnError continues processing remaining notifications even if
	// one fails. Results for each notification are returned in the slice.
	ContinueOnError bool
}

// ProcessNotifications retrieves and processes pending notifications by sending them
// to their respective SM-DP+ servers. This is a high-level convenience function
// that automates the notification handling workflow.
//
// The function processes notifications identified by their sequence numbers. If no
// sequence numbers are provided, no notifications are processed (use ProcessAllNotifications
// to process all pending notifications).
//
// Workflow for each notification:
//  1. Retrieve notification content via ES10b.RetrieveNotificationsList
//  2. Send notification to SM-DP+ server via ES9p.HandleNotification
//  3. Optionally remove from eUICC via ES10b.RemoveNotificationFromList (if AutoRemove is true)
//
// Example usage:
//
//	results, err := client.ProcessNotifications(
//	    &lpa.ProcessNotificationsOptions{
//	        AutoRemove: true,
//	        ContinueOnError: true,
//	    },
//	    sgp22.SequenceNumber(1),
//	    sgp22.SequenceNumber(2),
//	)
//	for _, result := range results {
//	    if result.Success {
//	        fmt.Printf("Processed notification %d\n", result.SequenceNumber)
//	    } else {
//	        fmt.Printf("Failed to process %d: %v\n", result.SequenceNumber, result.Error)
//	    }
//	}
func (c *Client) ProcessNotifications(opts *ProcessNotificationsOptions, sequenceNumbers ...sgp22.SequenceNumber) ([]*NotificationProcessResult, error) {
	if opts == nil {
		opts = &ProcessNotificationsOptions{}
	}

	if len(sequenceNumbers) == 0 {
		return nil, nil
	}

	results := make([]*NotificationProcessResult, 0, len(sequenceNumbers))

	for _, seqNum := range sequenceNumbers {
		result := &NotificationProcessResult{
			SequenceNumber: seqNum,
			Success:        false,
			Removed:        false,
		}

		// Process this single notification
		removed, err := c.processSingleNotification(seqNum, opts.AutoRemove)
		result.Removed = removed

		if err != nil {
			result.Error = err
			results = append(results, result)

			if !opts.ContinueOnError {
				return results, fmt.Errorf("failed to process notification %d: %w", seqNum, err)
			}
			continue
		}

		result.Success = true
		results = append(results, result)
	}

	return results, nil
}

// ProcessAllNotifications retrieves and processes all pending notifications on the eUICC.
// This function first lists all notifications, then processes each one.
//
// The function is equivalent to calling ListNotification() followed by ProcessNotifications()
// for all returned notifications.
//
// Example usage:
//
//	results, err := client.ProcessAllNotifications(&lpa.ProcessNotificationsOptions{
//	    AutoRemove: true,
//	    ContinueOnError: true,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Processed %d notifications\n", len(results))
func (c *Client) ProcessAllNotifications(opts *ProcessNotificationsOptions) ([]*NotificationProcessResult, error) {
	if opts == nil {
		opts = &ProcessNotificationsOptions{}
	}

	// List all pending notifications
	notifications, err := c.ListNotification()
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}

	if len(notifications) == 0 {
		return nil, nil
	}

	// Extract sequence numbers
	sequenceNumbers := make([]sgp22.SequenceNumber, len(notifications))
	for i, notif := range notifications {
		sequenceNumbers[i] = notif.SequenceNumber
	}

	// Process all notifications
	return c.ProcessNotifications(opts, sequenceNumbers...)
}

// processSingleNotification processes a single notification identified by its sequence number.
// It returns true if the notification was removed from the eUICC, and any error encountered.
func (c *Client) processSingleNotification(seqNum sgp22.SequenceNumber, autoRemove bool) (bool, error) {
	// Step 1: Retrieve notification content from eUICC
	notifications, err := c.RetrieveNotificationList(seqNum)
	if err != nil {
		return false, fmt.Errorf("retrieve notification: %w", err)
	}

	if len(notifications) == 0 {
		return false, fmt.Errorf("notification with sequence number %d not found", seqNum)
	}

	notification := notifications[0]

	// Step 2: Send notification to SM-DP+ server
	// HandleNotification internally uses notification.Notification.Address
	if err := c.HandleNotification(notification); err != nil {
		return false, fmt.Errorf("handle notification: %w", err)
	}

	// Step 3: Optionally remove notification from eUICC
	if autoRemove {
		if err := c.RemoveNotificationFromList(seqNum); err != nil {
			// Notification was sent successfully but removal failed
			// This is not a critical error, so we log it but don't fail
			return false, fmt.Errorf("remove notification: %w", err)
		}
		return true, nil
	}

	return false, nil
}
