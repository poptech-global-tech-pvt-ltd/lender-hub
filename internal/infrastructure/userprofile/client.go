package userprofile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client handles calls to external User Profile Service
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new User Profile Service client
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Response DTOs matching external service contract
type profileResponse struct {
	IsSuccess bool        `json:"is_success"`
	Message   string      `json:"message"`
	Data      profileData `json:"data"`
}

type profileData struct {
	ID           int     `json:"id"`
	Email        *string `json:"email"`        // can be null
	PhoneNumber  string  `json:"phone_number"` // "+919686019629"
	GlobalUserID string  `json:"global_user_id"`
}

// UserContactInfo represents resolved user contact information
type UserContactInfo struct {
	Mobile   string // 10-digit, no prefix
	Email    string // empty = no email
	RawPhone string // original "+91..."
}

// GetUserContact fetches user contact info from external profile service
func (c *Client) GetUserContact(ctx context.Context, userID string) (*UserContactInfo, error) {
	url := fmt.Sprintf("%s/api/v2/users/%s", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("profile service call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("profile service returned %d", resp.StatusCode)
	}

	var pr profileResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("decode profile response: %w", err)
	}

	if !pr.IsSuccess {
		return nil, fmt.Errorf("profile service error: %s", pr.Message)
	}

	if pr.Data.PhoneNumber == "" {
		return nil, fmt.Errorf("phone_number is empty for user %s", userID)
	}

	// Strip country prefix: "+919686019629" → "9686019629"
	mobile := stripIndianPrefix(pr.Data.PhoneNumber)

	email := ""
	if pr.Data.Email != nil && *pr.Data.Email != "" {
		email = *pr.Data.Email
	}

	return &UserContactInfo{
		Mobile:   mobile,
		Email:    email,
		RawPhone: pr.Data.PhoneNumber,
	}, nil
}

// LenderProfileUpdateRequest is used to push lender profile updates upstream
type LenderProfileUpdateRequest struct {
	UserID         string  `json:"userId"`
	Lender         string  `json:"lender"`
	AvailableLimit float64 `json:"availableLimit"`
	PreApproved    bool    `json:"preApproved"`
	OnboardingDone bool    `json:"onboardingDone"`
	IsBlocked      bool    `json:"isBlocked"`
	CurrentStatus  string  `json:"currentStatus"`
}

// UpdateLenderProfile pushes lender profile data to User Profile Service (TODO)
func (c *Client) UpdateLenderProfile(ctx context.Context, req LenderProfileUpdateRequest) error {
	// TODO: POST {profileServiceBaseURL}/api/v2/users/{userId}/lender-profile
	_ = ctx
	_ = req
	return nil
}

// stripIndianPrefix removes +91, 91, 0 prefixes to get 10-digit mobile
func stripIndianPrefix(phone string) string {
	phone = strings.TrimSpace(phone)
	if strings.HasPrefix(phone, "+91") {
		phone = phone[3:]
	} else if strings.HasPrefix(phone, "91") && len(phone) == 12 {
		phone = phone[2:]
	} else if strings.HasPrefix(phone, "0") && len(phone) == 11 {
		phone = phone[1:]
	}
	return phone
}
