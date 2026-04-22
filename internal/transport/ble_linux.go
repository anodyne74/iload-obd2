//go:build linux

package transport

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/rzetterberg/elmobd"
)

const (
	defaultBLEServiceUUID    = "6E400001-B5A3-F393-E0A9-E50E24DCCA9E"
	defaultBLEWriteCharUUID  = "6E400002-B5A3-F393-E0A9-E50E24DCCA9E"
	defaultBLENotifyCharUUID = "6E400003-B5A3-F393-E0A9-E50E24DCCA9E"
	defaultScanTimeout       = 12 * time.Second
	defaultCommandTimeout    = 1500 * time.Millisecond
)

// BLEDevice talks to an ELM327-compatible adapter over BLE GATT.
type BLEDevice struct {
	client         ble.Client
	writeChar      *ble.Characteristic
	commandTimeout time.Duration
	debug          bool

	commandMu sync.Mutex
	notifyMu  sync.Mutex
	notifyBuf bytes.Buffer
	promptCh  chan struct{}
}

// NewBLEDevice creates a BLE GATT-backed OBD device.
func NewBLEDevice(cfg *Config) (Device, error) {
	if strings.TrimSpace(cfg.BLE.DeviceName) == "" && strings.TrimSpace(cfg.BLE.MACAddress) == "" {
		return nil, fmt.Errorf("ble transport requires either ble.deviceName or ble.macAddress")
	}

	scanTimeout := defaultScanTimeout
	if cfg.BLE.ScanTimeoutSeconds > 0 {
		scanTimeout = time.Duration(cfg.BLE.ScanTimeoutSeconds) * time.Second
	}

	commandTimeout := defaultCommandTimeout
	if cfg.BLE.CommandTimeoutMS > 0 {
		commandTimeout = time.Duration(cfg.BLE.CommandTimeoutMS) * time.Millisecond
	}

	serviceUUID := cfg.BLE.ServiceUUID
	if serviceUUID == "" {
		serviceUUID = defaultBLEServiceUUID
	}
	writeUUID := cfg.BLE.WriteCharUUID
	if writeUUID == "" {
		writeUUID = defaultBLEWriteCharUUID
	}
	notifyUUID := cfg.BLE.NotifyCharUUID
	if notifyUUID == "" {
		notifyUUID = defaultBLENotifyCharUUID
	}

	bleDev, err := linux.NewDevice()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize linux BLE device: %w", err)
	}
	ble.SetDefaultDevice(bleDev)

	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()

	client, err := ble.Connect(ctx, func(a ble.Advertisement) bool {
		if cfg.BLE.MACAddress != "" && strings.EqualFold(a.Addr().String(), cfg.BLE.MACAddress) {
			return true
		}
		if cfg.BLE.DeviceName != "" && strings.EqualFold(strings.TrimSpace(a.LocalName()), strings.TrimSpace(cfg.BLE.DeviceName)) {
			return true
		}
		return false
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to BLE adapter: %w", err)
	}

	profile, err := client.DiscoverProfile(true)
	if err != nil {
		client.CancelConnection()
		return nil, fmt.Errorf("failed to discover BLE profile: %w", err)
	}

	if found := profile.FindService(ble.NewService(ble.MustParse(serviceUUID))); found == nil {
		client.CancelConnection()
		return nil, fmt.Errorf("BLE service %s not found on adapter", serviceUUID)
	}

	writeChar := profile.FindCharacteristic(ble.NewCharacteristic(ble.MustParse(writeUUID)))
	if writeChar == nil {
		client.CancelConnection()
		return nil, fmt.Errorf("BLE write characteristic %s not found", writeUUID)
	}

	notifyChar := profile.FindCharacteristic(ble.NewCharacteristic(ble.MustParse(notifyUUID)))
	if notifyChar == nil {
		client.CancelConnection()
		return nil, fmt.Errorf("BLE notify characteristic %s not found", notifyUUID)
	}

	dev := &BLEDevice{
		client:         client,
		writeChar:      writeChar,
		commandTimeout: commandTimeout,
		debug:          cfg.Debug,
		promptCh:       make(chan struct{}, 1),
	}

	if err := client.Subscribe(notifyChar, false, dev.handleNotification); err != nil {
		client.CancelConnection()
		return nil, fmt.Errorf("failed to subscribe to BLE notifications: %w", err)
	}

	if err := dev.bootstrapAdapter(); err != nil {
		client.CancelConnection()
		return nil, err
	}

	return dev, nil
}

func (d *BLEDevice) handleNotification(payload []byte) {
	d.notifyMu.Lock()
	defer d.notifyMu.Unlock()

	d.notifyBuf.Write(payload)
	if bytes.Contains(d.notifyBuf.Bytes(), []byte(">")) {
		select {
		case d.promptCh <- struct{}{}:
		default:
		}
	}
}

func (d *BLEDevice) resetNotificationBuffer() {
	d.notifyMu.Lock()
	defer d.notifyMu.Unlock()

	d.notifyBuf.Reset()
	for len(d.promptCh) > 0 {
		<-d.promptCh
	}
}

func (d *BLEDevice) readNotificationBuffer() string {
	d.notifyMu.Lock()
	defer d.notifyMu.Unlock()
	return d.notifyBuf.String()
}

func (d *BLEDevice) runRawCommand(command string) ([]string, error) {
	d.commandMu.Lock()
	defer d.commandMu.Unlock()

	d.resetNotificationBuffer()

	if err := d.client.WriteCharacteristic(d.writeChar, []byte(command+"\r"), true); err != nil {
		return nil, fmt.Errorf("failed to write BLE command %q: %w", command, err)
	}

	select {
	case <-d.promptCh:
	case <-time.After(d.commandTimeout):
		return nil, fmt.Errorf("timeout waiting for BLE response to %q", command)
	}

	raw := d.readNotificationBuffer()
	lines := parseAdapterLines(raw, command)
	if d.debug && len(lines) == 0 {
		return nil, fmt.Errorf("empty BLE response for command %q", command)
	}

	return lines, nil
}

func (d *BLEDevice) bootstrapAdapter() error {
	initCommands := []string{"ATZ", "ATE0", "ATL0", "ATS0", "ATH0", "ATSP0"}
	for _, cmd := range initCommands {
		if _, err := d.runRawCommand(cmd); err != nil {
			return fmt.Errorf("BLE adapter bootstrap failed at %s: %w", cmd, err)
		}
	}
	return nil
}

func (d *BLEDevice) RunOBDCommand(cmd elmobd.OBDCommand) (elmobd.OBDCommand, error) {
	outputs, err := d.runRawCommand(cmd.ToCommand())
	if err != nil {
		return cmd, err
	}

	result, err := parseOBDResponse(cmd, outputs)
	if err != nil {
		return cmd, err
	}
	if result == nil {
		return cmd, nil
	}

	if err := result.Validate(cmd); err != nil {
		return cmd, err
	}
	if err := cmd.SetValue(result); err != nil {
		return cmd, err
	}
	return cmd, nil
}

func parseAdapterLines(raw string, command string) []string {
	raw = strings.ReplaceAll(raw, "\r", "\n")
	parts := strings.Split(raw, "\n")
	out := make([]string, 0, len(parts))

	command = strings.ToUpper(strings.TrimSpace(command))
	for _, p := range parts {
		line := strings.TrimSpace(strings.TrimSuffix(p, ">"))
		if line == "" {
			continue
		}

		up := strings.ToUpper(line)
		if up == command {
			continue
		}

		if isHexLiteral(up) {
			up = spacedHex(up)
		}

		out = append(out, up)
	}
	return out
}

func parseOBDResponse(cmd elmobd.OBDCommand, outputs []string) (*elmobd.Result, error) {
	payload := ""

	for _, out := range outputs {
		switch {
		case strings.HasPrefix(out, "UNABLE TO CONNECT"):
			return nil, fmt.Errorf("'UNABLE TO CONNECT' received, is the ignition on?")
		case strings.HasPrefix(out, "NO DATA"):
			return nil, fmt.Errorf("'NO DATA' received, timeout from elm device?")
		case strings.HasPrefix(out, "SEARCHING"):
			continue
		case strings.HasPrefix(out, "BUS INIT"):
			continue
		case strings.HasPrefix(out, "OK"):
			continue
		}

		payload = out
		break
	}

	if payload == "" {
		return nil, nil
	}

	result, err := elmobd.NewResult(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OBD response %q: %w", payload, err)
	}
	return result, nil
}

func isHexLiteral(value string) bool {
	if len(value) < 6 || len(value)%2 != 0 {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

func spacedHex(value string) string {
	parts := make([]string, 0, len(value)/2)
	for i := 0; i+1 < len(value); i += 2 {
		parts = append(parts, value[i:i+2])
	}
	return strings.Join(parts, " ")
}
