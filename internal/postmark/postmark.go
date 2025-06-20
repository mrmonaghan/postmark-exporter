package postmark

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var PostmarkBaseUrl = "https://api.postmarkapp.com"

type OutboundStats struct {
	Sent                  int64
	Bounced               int64
	SMTPApiErrors         int64
	BounceRate            float64
	SpamComplaints        int64
	SpamComplaintsRate    float64
	Opens                 int64
	UniqueOpens           int64
	Tracked               int64
	WithLinkTracking      int64
	WithOpenTracking      int64
	TotalTrackedLinksSent int64
	UniqueLinksClicked    int64
	TotalClicks           int64
	WithClientRecorded    int64
	WithPlatformRecorded  int64
}

type BounceStats struct {
	Days []struct {
		Date         string
		HardBounce   int
		SoftBounce   int
		Transient    int
		SMTPAPIError int
	}
	HardBounce   int
	SMTPAPIError int
	SoftBounce   int
	Transient    int
}

type Client struct {
	client      *http.Client
	serverToken string
}

func New(serverToken string) *Client {
	return &Client{
		client:      &http.Client{},
		serverToken: serverToken,
	}
}

func (c *Client) GetOutboundStats() (OutboundStats, error) {
	req, err := http.NewRequest("GET", PostmarkBaseUrl+"/stats/outbound", nil)
	if err != nil {
		return OutboundStats{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Postmark-Server-Token", c.serverToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return OutboundStats{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return OutboundStats{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stats OutboundStats
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&stats); err != nil {
		return OutboundStats{}, err
	}

	return stats, nil
}

func (c *Client) GetBounceStats() (BounceStats, error) {
	req, err := http.NewRequest("GET", PostmarkBaseUrl+"/stats/outbound/bounces", nil)
	if err != nil {
		return BounceStats{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Postmark-Server-Token", c.serverToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return BounceStats{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return BounceStats{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stats BounceStats
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&stats); err != nil {
		return BounceStats{}, err
	}

	return stats, nil
}
