package planktoscope

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
)

func (c *Client) SegmenterStateBroadcasted() <-chan struct{} {
	return c.segmenterB.Broadcasted()
}

// Receive Updates

func (c *Client) updateSegmenterState(newState Segmenter) {
	c.stateL.Lock()
	defer c.stateL.Unlock()

	c.segmenter = newState
	c.segmenterB.BroadcastNext()
}

func (c *Client) handleSegmenterStatusUpdate(_ string, rawPayload []byte) (err error) {
	type SegmenterStatus struct {
		Status string `json:"status"`
	}
	var payload SegmenterStatus
	if err = json.Unmarshal(rawPayload, &payload); err != nil {
		return errors.Wrapf(err, "unparseable payload")
	}
	newState := Segmenter{
		StateKnown: true,
		Start:      c.segmenter.Start,
	}
	switch status := payload.Status; status {
	default:
		if !strings.HasPrefix(status, "Segmenting image") {
			// TODO: write the status to the segmenter state for display in the GUI
			c.Logger.Infof("unknown status %s", status)
			return nil
		}
		_, suffix, found := strings.Cut(status, ", image ")
		if !found {
			return errors.Errorf("couldn't parse status %s for segmenter progress", status)
		}
		frameRaw, _, found := strings.Cut(suffix, "/")
		if !found {
			return errors.Errorf("couldn't parse status %s for segmenter progress", status)
		}
		const (
			base  = 10
			width = 64 // bits
		)
		if newState.CurrentFrame, err = strconv.ParseUint(frameRaw, base, width); err != nil {
			return errors.Wrapf(err, "couldn't parse status %s for segmenter progress", status)
		}
		newState.Segmenting = true
		newState.LastObject = c.segmenter.LastObject
	case startedStatus:
		newState.Segmenting = true
		newState.Start = time.Now()
	case "Calculating flat":
		newState.Segmenting = true
		newState.CurrentFrame = c.segmenter.CurrentFrame
		newState.LastObject = c.segmenter.LastObject
		return nil
	case doneStatus:
		newState.Segmenting = false
		newState.CurrentFrame = c.segmenter.CurrentFrame
		newState.LastObject = c.segmenter.LastObject
	}

	// Commit changes
	c.updateSegmenterState(newState)
	c.Logger.Debugf("%s: %+v", c.Config.URL, newState)
	return nil
}

func (c *Client) handleSegmenterStatusObjectUpdate(_ string, rawPayload []byte) error {
	type SegmenterStatusObject struct {
		ID string `json:"object_id"`
	}
	var payload SegmenterStatusObject
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return errors.Wrapf(err, "unparseable payload")
	}
	const (
		base  = 10
		width = 64 // bits
	)
	id, err := strconv.ParseUint(payload.ID, base, width)
	if err != nil {
		return errors.Wrapf(err, "unparseable object ID %s", payload.ID)
	}
	newState := Segmenter{
		StateKnown:   true,
		Segmenting:   true,
		CurrentFrame: c.segmenter.CurrentFrame,
		LastObject:   id,
		Start:        c.segmenter.Start,
	}

	// Commit changes
	c.updateSegmenterState(newState)
	c.Logger.Debugf("%s: %+v", c.Config.URL, newState)
	return nil
}

func (c *Client) updateSegmenterSettings(newSettings SegmenterSettings) {
	c.stateL.Lock()
	defer c.stateL.Unlock()

	c.segmenterSettings = newSettings
	c.segmenterB.BroadcastNext()
}

const (
	segmentCommand = "segment"
)

func (c *Client) handleSegmenterSegmentingUpdate(_ string, rawPayload []byte) error {
	type SegmentSettings struct {
		EcoTaxa   bool   `json:"ecotaxa,omitempty"`
		Force     bool   `json:"force,omitempty"`
		Keep      bool   `json:"keep,omitempty"`
		ProcessID uint64 `json:"process_id,omitempty"`
		Recursive bool   `json:"recursive,omitempty"`
	}
	type SegmentCommand struct {
		Action   string
		Path     []string        `json:"path"`
		Settings SegmentSettings `json:"settings,omitempty"`
	}
	var payload SegmentCommand
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return errors.Wrapf(err, "unparseable payload")
	}
	newSettings := SegmenterSettings{}
	switch action := payload.Action; action {
	default:
		return errors.Errorf("unknown action %s", action)
	case segmentCommand:
		newSettings.ExportEcoTaxa = payload.Settings.EcoTaxa
		newSettings.ForceReprocessing = payload.Settings.Force
		newSettings.KeepObjects = payload.Settings.Keep
		newSettings.ProcessingID = payload.Settings.ProcessID
		newSettings.Recurse = payload.Settings.Recursive
		newSettings.Paths = payload.Path
	}

	// Commit changes
	c.updateSegmenterSettings(newSettings)
	c.Logger.Debugf("%s: %+v", c.Config.URL, newSettings)
	return nil
}

func (c *Client) handleSegmenterUpdate(topic string, rawPayload []byte) error {
	type SegmenterBaseCommand struct {
		Action string `json:"action"`
	}
	var basePayload SegmenterBaseCommand
	if err := json.Unmarshal(rawPayload, &basePayload); err != nil {
		return errors.Wrapf(err, "unparseable base payload")
	}
	broker := c.Config.URL
	switch action := basePayload.Action; action {
	default:
		var payload interface{}
		if err := json.Unmarshal(rawPayload, &payload); err != nil {
			c.Logger.Errorf("%s/%s: unknown payload %s", broker, topic, rawPayload)
			return nil
		}
		c.Logger.Infof("%s/%s: %v", broker, topic, payload)
	case segmentCommand:
		if err := c.handleSegmenterSegmentingUpdate(topic, rawPayload); err != nil {
			return errors.Wrap(err, "invalid segmenter config update command")
		}
	}
	return nil
}

func (c *Client) handleSegmenterMessage(topic string, rawPayload []byte) error {
	broker := c.Config.URL

	switch topic {
	default:
		var payload interface{}
		if err := json.Unmarshal(rawPayload, &payload); err != nil {
			return errors.Wrapf(err, "%s/%s: unparseable payload %s", broker, topic, rawPayload)
		}
		c.Logger.Infof("%s/%s: %v", broker, topic, payload)
	case "segmenter/segment":
		if err := c.handleSegmenterUpdate(topic, rawPayload); err != nil {
			return errors.Wrapf(err, "%s/%s: invalid payload %s", broker, topic, rawPayload)
		}
	case "status/segmenter":
		if err := c.handleSegmenterStatusUpdate(topic, rawPayload); err != nil {
			return errors.Wrapf(err, "%s/%s: invalid payload %s", broker, topic, rawPayload)
		}
	case "status/segmenter/object_id":
		if err := c.handleSegmenterStatusObjectUpdate(topic, rawPayload); err != nil {
			return errors.Wrapf(err, "%s/%s: invalid payload %s", broker, topic, rawPayload)
		}
	case "status/segmenter/metric":
		// We ignore these messages because they aren't useful to us right now
	}

	return nil
}

// Send Commands

func (c *Client) StartSegmenting(
	paths []string, processingID uint64,
	recurse bool, forceReprocessing bool, keepObjects bool, exportEcoTaxa bool,
) (mqtt.Token, error) {
	type CommandSettings struct {
		ExportEcoTaxa     bool   `json:"ecotaxa"`
		ForceReprocessing bool   `json:"force"`
		KeepObjects       bool   `json:"keep"`
		ProcessingID      uint64 `json:"process_id"`
		Recurse           bool   `json:"recursive"`
	}
	command := struct {
		Action   string          `json:"action"`
		Paths    []string        `json:"path"`
		Settings CommandSettings `json:"settings"`
	}{
		Action: segmentCommand,
		Paths:  paths,
		Settings: CommandSettings{
			ExportEcoTaxa:     exportEcoTaxa,
			ForceReprocessing: forceReprocessing,
			KeepObjects:       keepObjects,
			ProcessingID:      processingID,
			Recurse:           recurse,
		},
	}
	marshaled, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}

	c.stateL.Lock()
	defer c.stateL.Unlock()

	c.segmenterSettings.Paths = paths
	c.segmenterSettings.ProcessingID = processingID
	c.segmenterSettings.Recurse = recurse
	c.segmenterSettings.ForceReprocessing = forceReprocessing
	c.segmenterSettings.KeepObjects = keepObjects
	c.segmenterSettings.ExportEcoTaxa = exportEcoTaxa

	token := c.MQTT.Publish("segmenter/segment", mqttExactlyOnce, false, marshaled)
	return token, nil
}
