package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ShellySwitchPlus implements the Device interface for Shelly Plus switches
// using the Shelly Gen2 HTTP API.
type ShellySwitchPlus struct {
	info       DeviceInfo
	client     *http.Client
	connected  bool
	mu         sync.RWMutex
	lastStatus Status
	channel    int // Shelly devices can have multiple channels
	extInfo    ShellyExtendedInfo
}

// ShellyExtendedInfo contains additional device information from Shelly API.
type ShellyExtendedInfo struct {
	ID           string `json:"id"`
	MAC          string `json:"mac"`
	Model        string `json:"model"`
	Gen          int    `json:"gen"`
	FirmwareID   string `json:"fw_id"`
	Version      string `json:"ver"`
	App          string `json:"app"`
	Profile      string `json:"profile,omitempty"`
	AuthEnabled  bool   `json:"auth_en"`
	AuthDomain   string `json:"auth_domain,omitempty"`
	Discoverable bool   `json:"discoverable,omitempty"`
}

func NewShellySwitchPlus(id, name, address string, channel int) *ShellySwitchPlus {
	return &ShellySwitchPlus{
		info: DeviceInfo{
			ID:       id,
			Name:     name,
			Type:     TypeSwitch,
			Protocol: ProtocolShelly,
			Address:  address,
			Model:    "Shelly Gen2+",
		},
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		channel: channel,
	}
}

func (d *ShellySwitchPlus) Connect(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.connected {
		return nil
	}

	extInfo, err := d.fetchDeviceInfo(ctx)
	if err != nil {
		return NewDeviceError(d.info.ID, "connect", err)
	}

	d.extInfo = extInfo
	d.info.Firmware = extInfo.Version
	d.info.Model = extInfo.Model
	d.connected = true

	return nil
}

func (d *ShellySwitchPlus) Info() DeviceInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.info
}

func (d *ShellySwitchPlus) IsConnected() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.connected
}

func (d *ShellySwitchPlus) ExtendedInfo() ShellyExtendedInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.extInfo
}

func (d *ShellySwitchPlus) Disconnect(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.connected = false
	return nil
}

func (d *ShellySwitchPlus) fetchDeviceInfo(ctx context.Context) (ShellyExtendedInfo, error) {
	url := fmt.Sprintf("http://%s/rpc/Shelly.GetDeviceInfo", d.info.Address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ShellyExtendedInfo{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return ShellyExtendedInfo{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ShellyExtendedInfo{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result ShellyExtendedInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ShellyExtendedInfo{}, fmt.Errorf("parse response: %w", err)
	}

	return result, nil
}

func (d *ShellySwitchPlus) GetStatus(ctx context.Context) (Status, error) {
	d.mu.RLock()
	if !d.connected {
		d.mu.RUnlock()
		return Status{}, NewDeviceError(d.info.ID, "get_status", fmt.Errorf("device not connected"))
	}
	d.mu.RUnlock()

	status, err := d.fetchStatus(ctx)
	if err != nil {
		return Status{}, NewDeviceError(d.info.ID, "get_status", err)
	}

	d.mu.Lock()
	d.lastStatus = status
	d.mu.Unlock()

	return status, nil
}

func (d *ShellySwitchPlus) fetchStatus(ctx context.Context) (Status, error) {
	url := fmt.Sprintf("http://%s/rpc/Switch.GetStatus?id=%d", d.info.Address, d.channel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Status{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return Status{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Status{}, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID          int     `json:"id"`
		Source      string  `json:"source"`
		Output      bool    `json:"output"`
		APower      float64 `json:"apower"`
		Voltage     float64 `json:"voltage"`
		Current     float64 `json:"current"`
		Temperature struct {
			TC float64 `json:"tC"`
			TF float64 `json:"tF"`
		} `json:"temperature"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Status{}, fmt.Errorf("parse response: %w", err)
	}

	return Status{
		Online:      true,
		Power:       result.Output,
		Temperature: result.Temperature.TC,
		LastSeen:    time.Now(),
		Metadata: map[string]string{
			"source":  result.Source,
			"apower":  fmt.Sprintf("%.2f", result.APower),
			"voltage": fmt.Sprintf("%.2f", result.Voltage),
			"current": fmt.Sprintf("%.4f", result.Current),
		},
	}, nil
}

// Execute performs a command on the Shelly device.
func (d *ShellySwitchPlus) Execute(ctx context.Context, cmd Command) error {
	d.mu.RLock()
	if !d.connected {
		d.mu.RUnlock()
		return NewDeviceError(d.info.ID, "execute", fmt.Errorf("device not connected"))
	}
	d.mu.RUnlock()

	switch cmd.Action {
	case "on":
		return d.setSwitch(ctx, true)
	case "off":
		return d.setSwitch(ctx, false)
	case "toggle":
		return d.toggleSwitch(ctx)
	default:
		return NewDeviceError(d.info.ID, "execute", fmt.Errorf("unknown action: %s", cmd.Action))
	}
}

// setSwitch turns the switch on or off.
func (d *ShellySwitchPlus) setSwitch(ctx context.Context, on bool) error {
	url := fmt.Sprintf("http://%s/rpc/Switch.Set?id=%d&on=%t", d.info.Address, d.channel, on)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		WasOn bool `json:"was_on"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	d.mu.Lock()
	d.lastStatus.Power = on
	d.lastStatus.LastSeen = time.Now()
	d.mu.Unlock()

	return nil
}

// toggleSwitch inverts the current switch state.
func (d *ShellySwitchPlus) toggleSwitch(ctx context.Context) error {
	url := fmt.Sprintf("http://%s/rpc/Switch.Toggle?id=%d", d.info.Address, d.channel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		WasOn bool `json:"was_on"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	d.mu.Lock()
	d.lastStatus.Power = !result.WasOn
	d.lastStatus.LastSeen = time.Now()
	d.mu.Unlock()

	return nil
}
