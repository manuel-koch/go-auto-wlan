package service

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/manuel-koch/go-auto-wlan/utils"
)

type LidState int

const (
	LidUnknown LidState = iota
	LidOpen    LidState = iota
	LidClosed  LidState = iota
)

func LidStateToString(state LidState) string {
	switch state {
	case LidOpen:
		return "open"
	case LidClosed:
		return "closed"
	default:
		return "unknown"
	}
}

func getLidState() (LidState, error) {
	//ioreg -r -k AppleClamshellState -d 4 | grep AppleClamshellState | grep -i yes >/dev/null

	logger.Debug("Getting lid state...")
	lidState := LidUnknown

	cmd := exec.Command("ioreg", "-r", "-k", "AppleClamshellState", "-d", "4")
	if output, err := cmd.Output(); err != nil {
		logger.Error(fmt.Sprintf("Failed to get lid state: %v", err))
		return lidState, err
	} else {
		appleClamshellStateRe := regexp.MustCompile("\"AppleClamshellState\"\\s*=\\s*(?P<state>\\S+)")
		lines := strings.Split(string(output), "\n")
		for line := range lines {
			matches := utils.MatchNamedExpression(appleClamshellStateRe, lines[line])
			if matches != nil {
				switch strings.ToLower(matches["state"]) {
				case "yes":
					lidState = LidClosed
				case "no":
					lidState = LidOpen
				}
				break
			}
		}
	}

	logger.Debug(fmt.Sprintf("Got lid state %s", LidStateToString(lidState)))
	return lidState, nil
}
