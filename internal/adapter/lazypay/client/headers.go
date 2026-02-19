package client

import (
	lpConstants "lending-hub-service/internal/adapter/lazypay/constants"
	sharedContext "lending-hub-service/internal/shared/context"
)

// headersWithDevice returns all 6 required Lazypay headers.
// Used by: Onboarding, Eligibility, Create Order, Refund
func headersWithDevice(accessKey, sig string, rc *sharedContext.RequestContext) map[string]string {
	return map[string]string{
		lpConstants.HeaderContentType:   lpConstants.ContentTypeJSON,
		lpConstants.HeaderAccessKey:     accessKey,
		lpConstants.HeaderSignature:     sig,
		lpConstants.HeaderDeviceID:      rc.DeviceID,
		lpConstants.HeaderPlatform:      rc.Platform,
		lpConstants.HeaderUserIPAddress: rc.UserIP,
	}
}

// headersWithoutDevice returns 5 headers (no deviceId).
// Used by: Customer Status, Order Enquiry
func headersWithoutDevice(accessKey, sig string, rc *sharedContext.RequestContext) map[string]string {
	return map[string]string{
		lpConstants.HeaderContentType:   lpConstants.ContentTypeJSON,
		lpConstants.HeaderAccessKey:     accessKey,
		lpConstants.HeaderSignature:     sig,
		lpConstants.HeaderPlatform:      rc.Platform,
		lpConstants.HeaderUserIPAddress: rc.UserIP,
	}
}
