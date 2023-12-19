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
	"context"
	"fmt"
	"slices"
	"time"
)

const (
	LidUpdateInterval  = 3 * time.Second
	WlanUpdateInterval = 3 * time.Second
)

type LidStateChangedEvent struct {
	LidState LidState
}

type WlanStateChangedEvent struct {
	Devices []WlanDevice
}

func NewWlanStateChangedEvent(devices []WlanDevice) WlanStateChangedEvent {
	return WlanStateChangedEvent{Devices: CopyWlanDevices(devices)}
}

type Service struct {
	ctx context.Context

	wlanDevices []WlanDevice
	lidState    LidState

	pendingEvtSubscriptions  chan *EventSubscription
	pendingEvtUnsubscription chan *EventSubscription
	evtSubscriptions         []*EventSubscription
	publishEvents            chan interface{}

	requestLidUpdate  chan interface{}
	requestWlanUpdate chan interface{}
}

type EventSubscription struct {
	service *Service
	updates chan interface{}
}

func (e *EventSubscription) Updates() <-chan interface{} {
	return e.updates
}

func (e *EventSubscription) Unsubscribe() {
	e.service.pendingEvtUnsubscription <- e
}

func NewService(ctx context.Context) *Service {
	s := &Service{
		ctx:                      ctx,
		pendingEvtSubscriptions:  make(chan *EventSubscription),
		pendingEvtUnsubscription: make(chan *EventSubscription),
		publishEvents:            make(chan interface{}),
		evtSubscriptions:         make([]*EventSubscription, 0),

		requestLidUpdate:  make(chan interface{}, 0),
		requestWlanUpdate: make(chan interface{}, 0),
	}

	if wifiDevices, err := getWlanDevices(); err == nil {
		s.wlanDevices = wifiDevices
	}
	if lidState, err := getLidState(); err == nil {
		s.lidState = lidState
	}

	for _, wlanDevice := range s.wlanDevices {
		logger.Info(fmt.Sprintf("WLAN device %s", wlanDevice.String()))
	}
	logger.Info(fmt.Sprintf("Lid is %s", LidStateToString(s.lidState)))

	go s.handleSubscriptions()
	go s.watchLid()
	go s.watchWlan()

	return s
}

func (s *Service) Subscripe() *EventSubscription {
	e := &EventSubscription{
		service: s,
		updates: make(chan interface{}),
	}
	s.pendingEvtSubscriptions <- e
	return e
}

func (s *Service) handleSubscriptions() {
	done := false
	for !done {
		select {
		case <-s.ctx.Done():
			{
				for _, subscription := range s.evtSubscriptions {
					close(subscription.updates)
				}
				s.evtSubscriptions = make([]*EventSubscription, 0)
				done = true
			}
		case pendingSubscribe := <-s.pendingEvtSubscriptions:
			s.evtSubscriptions = append(s.evtSubscriptions, pendingSubscribe)
		case pendingUnsubscribe := <-s.pendingEvtUnsubscription:
			for i, subscription := range s.evtSubscriptions {
				if subscription == pendingUnsubscribe {
					slices.Delete(s.evtSubscriptions, i, i)
					close(subscription.updates)
				}
			}
		case event := <-s.publishEvents:
			for i := range s.evtSubscriptions {
				s.evtSubscriptions[i].updates <- event
			}
		}
	}
}

func (s *Service) GetLidState() LidState {
	return s.lidState
}

func (s *Service) watchLid() {
	logger.Info("Start watching lid...")
	done := false
	for !done {
		select {
		case <-s.ctx.Done():
			done = true
		case <-s.requestLidUpdate:
			s.queryLid()
		case <-time.After(LidUpdateInterval):
			s.queryLid()
		}
	}
	logger.Info("Stopped watching lid")
}

func (s *Service) queryLid() {
	logger.Debug("Query lid")
	if lidState, err := getLidState(); err == nil {
		if lidState != s.lidState {
			logger.Info(fmt.Sprintf("New lid state: %s", LidStateToString(lidState)))
			s.lidState = lidState
			s.publishEvents <- LidStateChangedEvent{
				LidState: lidState,
			}
		}
	}
	logger.Debug("Queried lid")
}

func (s *Service) watchWlan() {
	logger.Info("Start watching wlan...")
	done := false
	for !done {
		select {
		case <-s.ctx.Done():
			done = true
		case <-s.requestWlanUpdate:
			s.queryWlan()
		case <-time.After(WlanUpdateInterval):
			s.queryWlan()
		}
	}
	logger.Info("Stopped watching wlan")
}

func (s *Service) queryWlan() {
	logger.Debug("Query wlan")
	if devices, err := getWlanDevices(); err == nil {
		if !slices.Equal(devices, s.wlanDevices) {
			for _, d := range devices {
				logger.Info(fmt.Sprintf("New wlan state: %s", d.String()))
			}
			s.wlanDevices = devices
			s.publishEvents <- NewWlanStateChangedEvent(devices)
		}
	}
	logger.Debug("Queried wlan")
}

func (s *Service) GetWlanDevices() []WlanDevice {
	return CopyWlanDevices(s.wlanDevices)
}

func (s *Service) SetWlanState(device string, state WlanState) {
	logger.Info(fmt.Sprintf("Setting WLAN device %s to %s", device, WlanStateToString(state)))
	if setWlanState(device, state) == nil {
		s.requestWlanUpdate <- true
	}
}
