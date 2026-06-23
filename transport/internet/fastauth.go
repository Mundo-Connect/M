package internet

import (
	"context"
	stderrors "errors"
	"net/http"
	"strings"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
)

const (
	FastAuthRequestHeader  = "MundoConnect-Auth"
	FastAuthResponseHeader = "MundoConnect-Response"
)

type FastAuthRequest struct {
	Payload  *session.FastAuthPayload
	state    *session.FastAuthState
	encoded  string
}

func (r FastAuthRequest) Enabled() bool {
	return r.Payload != nil && (r.state == nil || r.state.State() != session.FastAuthCapabilityUnsupported)
}

func (r FastAuthRequest) EncodedPayload() (string, error) {
	if !r.Enabled() {
		return "", nil
	}
	if r.encoded != "" {
		return r.encoded, nil
	}
	return EncodeFastAuthPayload(r.Payload)
}

func (r FastAuthRequest) SetHTTPHeaders(header http.Header) error {
	if header == nil || !r.Enabled() {
		return nil
	}
	encoded, err := r.EncodedPayload()
	if err != nil {
		return err
	}
	if encoded != "" {
		header.Set(FastAuthRequestHeader, encoded)
	}
	return nil
}

func (r FastAuthRequest) ObserveHeader(header http.Header) *FastAuthResult {
	if header == nil {
		return r.ObserveValue("")
	}
	return r.ObserveValue(header.Get(FastAuthResponseHeader))
}

func (r FastAuthRequest) ObserveValue(value string) *FastAuthResult {
	if value == "" {
		if r.state != nil && r.state.State() == session.FastAuthCapabilityUnknown {
			r.state.MarkUnsupported()
		}
		return nil
	}

	switch value {
	case session.FastAuthStatusOK:
		if r.state != nil {
			r.state.MarkSupported()
		}
		return &FastAuthResult{Status: session.FastAuthStatusOK}
	case session.FastAuthStatusFail:
		if r.state != nil {
			r.state.MarkUnsupported()
		}
		return &FastAuthResult{Status: session.FastAuthStatusFail}
	}

	if r.state != nil && r.state.State() == session.FastAuthCapabilityUnknown {
		r.state.MarkUnsupported()
	}
	return nil
}

func (r FastAuthRequest) ConnectionResult(observed *FastAuthResult) *FastAuthResult {
	if observed != nil {
		return observed
	}
	if r.Enabled() && r.state != nil && r.state.State() == session.FastAuthCapabilitySupported {
		return &FastAuthResult{Status: session.FastAuthStatusOK}
	}
	return nil
}

type FastAuthResult struct {
	Status string
}

func PrepareFastAuthRequest(ctx context.Context) FastAuthRequest {
	payload := session.FastAuthPayloadFromContext(ctx)
	if payload == nil {
		return FastAuthRequest{}
	}

	state := session.FastAuthStateFromContext(ctx)
	if state != nil && state.State() == session.FastAuthCapabilityUnsupported {
		return FastAuthRequest{}
	}

	encoded, _ := EncodeFastAuthPayload(payload)
	return FastAuthRequest{
		Payload: payload,
		state:   state,
		encoded: encoded,
	}
}

func EncodeFastAuthPayload(payload *session.FastAuthPayload) (string, error) {
	if payload == nil {
		return "", nil
	}
	var builder strings.Builder
	builder.Grow(6 + len(payload.Protocol) + len(payload.AuthData) + len(payload.Network) + len(payload.TargetHost) + len(payload.TargetPort))

	if payload.Version == 0 {
		payload.Version = session.FastAuthVersionClassic
	}
	builder.WriteByte('1')
	writeNextFastAuthField(&builder, payload.Protocol)
	writeNextFastAuthField(&builder, payload.AuthData)
	writeNextFastAuthField(&builder, payload.Network)
	writeNextFastAuthField(&builder, payload.TargetHost)
	writeNextFastAuthField(&builder, payload.TargetPort)
	for _, arg := range payload.Args {
		writeNextFastAuthField(&builder, arg)
	}
	return builder.String(), nil
}

func DecodeFastAuthPayload(raw string) (*session.FastAuthPayload, error) {
	if raw == "" {
		return nil, nil
	}
	fields, err := splitFastAuthFields(raw)
	if err != nil {
		return nil, err
	}
	if len(fields) < 6 {
		return nil, stderrors.New("invalid fast auth payload")
	}

	payload := &session.FastAuthPayload{
		Protocol:   fields[1],
		AuthData:   fields[2],
		Network:    fields[3],
		TargetHost: fields[4],
		TargetPort: fields[5],
	}

	switch fields[0] {
	case "1":
		payload.Version = session.FastAuthVersionClassic
	default:
		return nil, stderrors.New("invalid fast auth version: " + fields[0])
	}

	if len(fields) > 6 {
		payload.Args = fields[6:]
	}
	return payload, nil
}

func writeNextFastAuthField(builder *strings.Builder, value string) {
	builder.WriteByte('|')
	writeFastAuthField(builder, value)
}

func writeFastAuthField(builder *strings.Builder, value string) {
	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '\\', '|':
			builder.WriteByte('\\')
		}
		builder.WriteByte(value[i])
	}
}

func splitFastAuthFields(raw string) ([]string, error) {
	if !strings.Contains(raw, "\\") {
		return strings.Split(raw, "|"), nil
	}
	fields := make([]string, 0, strings.Count(raw, "|")+1)
	var builder strings.Builder
	escaped := false
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if escaped {
			builder.WriteByte(ch)
			escaped = false
			continue
		}
		switch ch {
		case '\\':
			escaped = true
		case '|':
			fields = append(fields, builder.String())
			builder.Reset()
		default:
			builder.WriteByte(ch)
		}
	}
	if escaped {
		return nil, stderrors.New("invalid fast auth escape")
	}
	fields = append(fields, builder.String())
	return fields, nil
}

type FastAuthConnWrapper struct {
	Connection
	FastAuthProtocol string
	FastAuthUser     interface{}
	FastAuthTarget   net.Destination
	FastAuthResult   *FastAuthResult
}

func (w *FastAuthConnWrapper) Close() error {
	return w.Connection.Close()
}
