package app

import (
	"context"
	"fmt"

	"fyne.io/systray"
	"github.com/manuel-koch/go-auto-wlan/assets"
	"github.com/manuel-koch/go-auto-wlan/service"
)

const appName = "Auto WLAN"
const maxWlanDevices = 3

type wlanDeviceSettings struct {
	device          string
	enableOnLidOpen bool
	toggleMenuItem  *systray.MenuItem
}

type App struct {
	name        string
	versionInfo string
	buildInfo   string

	serviceCtx    context.Context
	serviceCancel func()
	service       *service.Service

	wlanDeviceSettings      []wlanDeviceSettings
	toggleWlanOnLidMenuItem *systray.MenuItem
	quitMenuItem            *systray.MenuItem
}

func NewApp(versionInfo, versionsSha1, buildInfo string) *App {
	logger.Info(fmt.Sprintf("%s, version v%s (%s), built %s", appName, versionInfo, versionsSha1, buildInfo))
	serviceCtx, serviceCancel := context.WithCancel(context.Background())
	return &App{
		name:        appName,
		versionInfo: versionInfo,
		buildInfo:   buildInfo,

		serviceCtx:    serviceCtx,
		serviceCancel: serviceCancel,
		service:       service.NewService(serviceCtx),

		wlanDeviceSettings: make([]wlanDeviceSettings, maxWlanDevices),
	}
}

func (a *App) Shutdown() {
	logger.Info("App shutdown")
	a.serviceCancel()
	systray.Quit()
}

func (a *App) handleServiceEvents(subscription *service.EventSubscription) {
	logger.Info("Starting to handle service events")
	for event := range subscription.Updates() {
		if lidEvent, ok := event.(service.LidStateChangedEvent); ok {
			a.handleLidEvent(lidEvent)
		} else if wlanEvent, ok := event.(service.WlanStateChangedEvent); ok {
			a.handleWlanEvent(wlanEvent)
		}
	}
	logger.Info("Stopped handling service events")
}

func (a *App) handleLidEvent(lidEvent service.LidStateChangedEvent) {
	logger.Info(fmt.Sprintf("App handling lid event %s", service.LidStateToString(lidEvent.LidState)))
	for i := range a.wlanDeviceSettings {
		setting := &a.wlanDeviceSettings[i]
		if len(setting.device) > 0 {
			switch lidEvent.LidState {
			case service.LidOpen:
				{
					if setting.enableOnLidOpen {
						a.service.SetWlanState(setting.device, service.WlanPowerOn)
						setting.enableOnLidOpen = false
					}
				}
			case service.LidClosed:
				{
					if setting.toggleMenuItem.Checked() {
						a.service.SetWlanState(setting.device, service.WlanPowerOff)
						setting.enableOnLidOpen = true
					}
				}
			}
		}
	}
}

func (a *App) handleWlanEvent(wlanEvent service.WlanStateChangedEvent) {
	logger.Info("App handling wlan event")
	a.updateWlanSettings(wlanEvent.Devices)
}

func (a *App) updateWlanSettings(devices []service.WlanDevice) {
	for i := range a.wlanDeviceSettings {
		if i < len(devices) {
			a.updateWlanMenuItem(&a.wlanDeviceSettings[i], devices[i])
		} else {
			a.wlanDeviceSettings[i].toggleMenuItem.Hide()
		}
	}
	a.updateIcon()
}

func (a *App) updateWlanMenuItem(setting *wlanDeviceSettings, device service.WlanDevice) {
	setting.toggleMenuItem.Show()

	var title string
	if len(device.Network) > 0 {
		title = fmt.Sprintf("WLAN %s (%s)", device.Name, device.Network)
	} else {
		title = fmt.Sprintf("WLAN %s", device.Name)
	}
	setting.toggleMenuItem.SetTitle(title)

	setting.device = device.Name
	if device.State == service.WlanPowerOn && !setting.toggleMenuItem.Checked() {
		setting.toggleMenuItem.Check()
	}
	if device.State == service.WlanPowerOff && setting.toggleMenuItem.Checked() {
		setting.toggleMenuItem.Uncheck()
	}
}

func (a *App) Run() {
	systray.Run(a.onSystrayReady, a.onSystrayExit)
}

func (a *App) onSystrayReady() {
	logger.Debug("App configure systray")

	systray.SetTooltip(fmt.Sprintf("Version v%s, built %s", a.versionInfo, a.buildInfo))

	for i := 0; i < maxWlanDevices; i++ {
		a.wlanDeviceSettings[i].toggleMenuItem = systray.AddMenuItemCheckbox(fmt.Sprintf("WLAN %d", i), "Toggle WLAN", false)
		a.wlanDeviceSettings[i].toggleMenuItem.Hide()
	}

	a.toggleWlanOnLidMenuItem = systray.AddMenuItemCheckbox("Toggle WLAN on Lid", "Toggle WLAN when lid closes / opens", true)

	systray.AddSeparator()

	a.quitMenuItem = systray.AddMenuItem("Quit", fmt.Sprintf("Quit %s", a.name))

	done := false

	for i := 0; i < maxWlanDevices; i++ {
		go func(setting *wlanDeviceSettings) {
			for !done {
				select {
				case <-setting.toggleMenuItem.ClickedCh:
					{
						if setting.toggleMenuItem.Checked() {
							a.service.SetWlanState(setting.device, service.WlanPowerOff)
							setting.toggleMenuItem.Uncheck()
						} else {
							a.service.SetWlanState(setting.device, service.WlanPowerOn)
							setting.toggleMenuItem.Check()
						}
					}
				}
			}
		}(&a.wlanDeviceSettings[i])
	}

	go func() {
		for !done {
			select {
			case <-a.toggleWlanOnLidMenuItem.ClickedCh:
				{
					if a.toggleWlanOnLidMenuItem.Checked() {
						a.toggleWlanOnLidMenuItem.Uncheck()
					} else {
						a.toggleWlanOnLidMenuItem.Check()
					}
				}
			case <-a.quitMenuItem.ClickedCh:
				{
					logger.Debug("Quit triggered")
					a.Shutdown()
					done = true
				}
			}
		}
	}()

	a.updateWlanSettings(a.service.GetWlanDevices())

	subscription := a.service.Subscripe()
	go a.handleServiceEvents(subscription)

	logger.Debug("App configure systray done")
}

func (a *App) anyWlanOn() bool {
	for _, setting := range a.wlanDeviceSettings {
		if setting.toggleMenuItem.Checked() {
			return true
		}
	}
	return false
}

func (a *App) updateIcon() {
	icon := assets.AutowlanBlackDisabled
	if a.anyWlanOn() {
		icon = assets.AutowlanBlack
	}
	systray.SetTemplateIcon(icon, icon)
}

func (a *App) onSystrayExit() {
	logger.Debug("App systray exit")
}
