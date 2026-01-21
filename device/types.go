package device

import "time"

type DeviceType string

const (
	TypeSwitch DeviceType = "switch"
)

type Protocol string

const (
	ProtocolShelly Protocol = "shelly"
	ProtocolHTTP   Protocol = "http"
)

type DeviceInfo struct {
	ID       string
	Name     string
	Type     DeviceType
	Protocol Protocol
	Address  string
	Model    string
	Firmware string
}

type Status struct {
	Online      bool
	Power       bool
	Temperature float64
	LastSeen    time.Time
	Metadata    map[string]string
}

type Command struct {
	Action string
	Params map[string]interface{}
}

type DeviceError struct {
	DeviceID  string
	Operation string
	Err       error
}

func (e *DeviceError) Error() string {
	return e.DeviceID + ": " + e.Operation + ": " + e.Err.Error()
}

func NewDeviceError(deviceID, operation string, err error) *DeviceError {
	return &DeviceError{
		DeviceID:  deviceID,
		Operation: operation,
		Err:       err,
	}
}
