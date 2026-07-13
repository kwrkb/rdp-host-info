package hostinfo

import (
	"errors"
	"testing"
)

func TestCollect_RecommendedPrefersTailscale(t *testing.T) {
	p := Providers{
		LocalIPv4:   func() ([]string, error) { return []string{"192.168.1.20"}, nil },
		TailscaleIP: func() (string, error) { return "100.80.10.5", nil },
	}
	info := Collect(p)
	if info.Recommended != "100.80.10.5" {
		t.Errorf("Recommended = %q, want tailscale IP", info.Recommended)
	}
}

func TestCollect_RecommendedFallsBackToLocalIPv4(t *testing.T) {
	p := Providers{
		LocalIPv4:   func() ([]string, error) { return []string{"192.168.1.20"}, nil },
		TailscaleIP: func() (string, error) { return "", errors.New("not found") },
	}
	info := Collect(p)
	if info.Recommended != "192.168.1.20" {
		t.Errorf("Recommended = %q, want local IPv4", info.Recommended)
	}
}

func TestCollect_RecommendedEmptyWhenNothingAvailable(t *testing.T) {
	info := Collect(Providers{})
	if info.Recommended != "" {
		t.Errorf("Recommended = %q, want empty", info.Recommended)
	}
}

func TestCollect_FailedProviderLeavesFieldEmpty(t *testing.T) {
	p := Providers{
		ComputerName: func() (string, error) { return "", errors.New("boom") },
		UserName:     func() (string, error) { return "OMEN16\\yugo", nil },
	}
	info := Collect(p)
	if info.PCName != "" {
		t.Errorf("PCName = %q, want empty on error", info.PCName)
	}
	if info.UserName != "OMEN16\\yugo" {
		t.Errorf("UserName = %q, want OMEN16\\yugo", info.UserName)
	}
}

func TestCollect_NilProvidersDoNotPanic(t *testing.T) {
	info := Collect(Providers{})
	if info.PCName != "" || info.Recommended != "" {
		t.Errorf("expected zero-value HostInfo, got %+v", info)
	}
}
