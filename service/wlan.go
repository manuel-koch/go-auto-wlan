// This file is part of go-auto-wlan.
//
// go-auto-wlan is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-auto-wlan is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-auto-wlan. If not, see <http://www.gnu.org/licenses/>.
//
// Copyright 2023 Manuel Koch
package service

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/manuel-koch/go-auto-wlan/utils"
)

type WlanState int

const (
	WlanUnknown  WlanState = iota
	WlanPowerOn  WlanState = iota
	WlanPowerOff WlanState = iota
)

type WlanDevice struct {
	Name    string
	State   WlanState
	Network string
}

type InvalidWlanStateError struct {
	state WlanState
}

func (e InvalidWlanStateError) Error() string {
	return fmt.Sprintf("Invalid wlan state: %d (%s)", e.state, WlanStateToString(e.state))
}

func (d *WlanDevice) String() string {
	s := fmt.Sprintf("%s is %s", d.Name, WlanStateToString(d.State))
	if len(d.Network) > 0 {
		s += fmt.Sprintf(" (%s)", d.Network)
	}
	return s
}

func WlanStateToString(state WlanState) string {
	switch state {
	case WlanPowerOn:
		return "on"
	case WlanPowerOff:
		return "off"
	default:
		return "unknown"
	}
}

func CopyWlanDevices(devices []WlanDevice) []WlanDevice {
	copyDevices := make([]WlanDevice, len(devices))
	copy(copyDevices, devices)
	return copyDevices
}

func getWlanDevices() ([]WlanDevice, error) {
	logger.Debug("Searching wlan devices...")

	devices := make([]WlanDevice, 0)

	cmd := exec.Command("networksetup", "-listallhardwareports")
	if outputBytes, err := cmd.Output(); err != nil {
		logger.Error(fmt.Sprintf("Failed to get network hardware ports: %v", err))
		return devices, err
	} else {
		// merge non-empty lines into one line
		outputStr := string(outputBytes)
		mergeLinesRe := regexp.MustCompile("(\\S+)\\n")
		outputStr = mergeLinesRe.ReplaceAllString(outputStr, "$1 ")

		wifiRe := regexp.MustCompile("Hardware Port:\\s+Wi-Fi")
		deviceRe := regexp.MustCompile("Device:\\s+(?P<name>\\S+)")

		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			wifiMatch := utils.MatchNamedExpression(wifiRe, line)
			if wifiMatch != nil {
				deviceMatch := utils.MatchNamedExpression(deviceRe, line)
				if deviceMatch != nil {
					deviceName := deviceMatch["name"]
					if state, err := getWlanState(deviceName); err == nil {
						device := WlanDevice{Name: deviceName, State: state}
						if device.State == WlanPowerOn {
							if network, err := getWlanNetwork(deviceName); err == nil {
								device.Network = network
							}
						}
						devices = append(devices, device)
						logger.Debug(fmt.Sprintf("Found wlan device %s", device.String()))
					}
				}
			}
		}
	}

	return devices, nil
}

func getWlanState(device string) (WlanState, error) {
	cmd := exec.Command("networksetup", "-getairportpower", device)
	if output, err := cmd.Output(); err != nil {
		logger.Error(fmt.Sprintf("Failed to get network airport power: %v", err))
		return WlanUnknown, err
	} else {
		stateRe := regexp.MustCompile(fmt.Sprintf("Wi-Fi\\s+Power\\s+\\(%s\\):\\s+(?P<state>\\S+)", device))
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			stateMatch := utils.MatchNamedExpression(stateRe, line)
			if stateMatch != nil {
				switch strings.ToLower(stateMatch["state"]) {
				case "on":
					{
						return WlanPowerOn, nil
					}
				case "off":
					{
						return WlanPowerOff, nil
					}
				}
			}
		}
		return WlanUnknown, nil
	}
}

func setWlanState(device string, state WlanState) error {
	var power string
	switch state {
	case WlanPowerOn:
		power = "on"
	case WlanPowerOff:
		power = "off"
	default:
		return InvalidWlanStateError{state: WlanUnknown}
	}
	cmd := exec.Command("networksetup", "-setairportpower", device, power)
	if _, err := cmd.Output(); err != nil {
		logger.Error(fmt.Sprintf("Failed to set network airport power: %v", err))
		return err
	}
	return nil
}

func getWlanNetwork(device string) (string, error) {
	cmd := exec.Command("networksetup", "-getairportnetwork", device)
	if output, err := cmd.Output(); err != nil {
		logger.Error(fmt.Sprintf("Failed to get network airport network: %v", err))
		return "", err
	} else {
		networkRe := regexp.MustCompile("Current\\s+Wi-Fi\\s+Network:\\s+(?P<network>\\S+)")
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			networkMatch := utils.MatchNamedExpression(networkRe, line)
			if networkMatch != nil {
				return networkMatch["network"], nil
			}
		}
		return "", nil
	}
}
