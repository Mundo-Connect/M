package internet

import (
	"context"
	"testing"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	payload := &session.FastAuthPayload{
		Version:    session.FastAuthVersionClassic,
		Protocol:   "mx",
		AuthData:   "test-token",
		TargetHost: "example.com",
		TargetPort: "443",
		Network:    "tcp",
		Args:       []string{"tcp"},
	}

	encoded, err := EncodeFastAuthPayload(payload)
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	if encoded == "" {
		t.Fatal("encoded payload is empty")
	}

	decoded, err := DecodeFastAuthPayload(encoded)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	if decoded == nil {
		t.Fatal("decoded payload is nil")
	}
	if decoded.Protocol != payload.Protocol {
		t.Fatalf("protocol = %q, want %q", decoded.Protocol, payload.Protocol)
	}
	if decoded.AuthData != payload.AuthData {
		t.Fatalf("authData = %q, want %q", decoded.AuthData, payload.AuthData)
	}
	if decoded.TargetHost != payload.TargetHost {
		t.Fatalf("targetHost = %q, want %q", decoded.TargetHost, payload.TargetHost)
	}
	if decoded.TargetPort != payload.TargetPort {
		t.Fatalf("targetPort = %q, want %q", decoded.TargetPort, payload.TargetPort)
	}
	if decoded.Network != payload.Network {
		t.Fatalf("network = %q, want %q", decoded.Network, payload.Network)
	}
	if len(decoded.Args) != len(payload.Args) {
		t.Fatalf("args len = %d, want %d", len(decoded.Args), len(payload.Args))
	}
}

func TestEncodeDecodeWithEscape(t *testing.T) {
	payload := &session.FastAuthPayload{
		Version:    session.FastAuthVersionClassic,
		Protocol:   "mx",
		AuthData:   "token|with\\pipe",
		TargetHost: "example.com",
		TargetPort: "443",
		Network:    "tcp",
		Args:       []string{"tcp"},
	}

	encoded, err := EncodeFastAuthPayload(payload)
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	decoded, err := DecodeFastAuthPayload(encoded)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	if decoded.AuthData != payload.AuthData {
		t.Fatalf("authData = %q, want %q", decoded.AuthData, payload.AuthData)
	}
}

func TestEncodeDecodeManyArgs(t *testing.T) {
	payload := &session.FastAuthPayload{
		Version:    session.FastAuthVersionClassic,
		Protocol:   "mx",
		AuthData:   "token",
		TargetHost: "example.com",
		TargetPort: "443",
		Network:    "tcp",
		Args:       []string{"tcp", "arg1", "arg2", "arg3"},
	}

	encoded, err := EncodeFastAuthPayload(payload)
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	decoded, err := DecodeFastAuthPayload(encoded)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	if len(decoded.Args) != len(payload.Args) {
		t.Fatalf("args len = %d, want %d", len(decoded.Args), len(payload.Args))
	}
	for i := range payload.Args {
		if decoded.Arg(i) != payload.Arg(i) {
			t.Fatalf("arg[%d] = %q, want %q", i, decoded.Arg(i), payload.Arg(i))
		}
	}
}

func TestDecodeEmpty(t *testing.T) {
	decoded, err := DecodeFastAuthPayload("")
	if err != nil {
		t.Fatalf("decode empty: %v", err)
	}
	if decoded != nil {
		t.Fatal("expected nil payload for empty string")
	}
}

func TestDecodeInvalid(t *testing.T) {
	_, err := DecodeFastAuthPayload("1|mx")
	if err == nil {
		t.Fatal("expected error for short payload")
	}

	_, err = DecodeFastAuthPayload("3|mx|token|tcp|host|port")
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestPrepareFastAuthRequestUnsupported(t *testing.T) {
	state := &session.FastAuthState{}
	state.MarkUnsupported()
	ctx := session.ContextWithFastAuthState(
		session.ContextWithFastAuthPayload(context.Background(),
			&session.FastAuthPayload{Protocol: "mx", AuthData: "token", TargetHost: "host", TargetPort: "443", Network: "tcp"}),
		state,
	)

	request := PrepareFastAuthRequest(ctx)
	if request.Enabled() {
		t.Fatal("request should be disabled when state is unsupported")
	}
}

func TestPrepareFastAuthRequestEnabled(t *testing.T) {
	ctx := session.ContextWithFastAuthPayload(context.Background(),
		&session.FastAuthPayload{Protocol: "mx", AuthData: "token", TargetHost: "host", TargetPort: "443", Network: "tcp"})

	request := PrepareFastAuthRequest(ctx)
	if !request.Enabled() {
		t.Fatal("request should be enabled")
	}

	encoded, err := request.EncodedPayload()
	if err != nil {
		t.Fatalf("encoded payload: %v", err)
	}
	if encoded == "" {
		t.Fatal("encoded payload is empty")
	}
}

func TestObserveValueOK(t *testing.T) {
	state := &session.FastAuthState{}
	ctx := session.ContextWithFastAuthState(
		session.ContextWithFastAuthPayload(context.Background(),
			&session.FastAuthPayload{Protocol: "mx"}),
		state,
	)

	request := PrepareFastAuthRequest(ctx)
	result := request.ObserveValue(session.FastAuthStatusOK)
	if result == nil || result.Status != session.FastAuthStatusOK {
		t.Fatalf("ObserveValue(ok) = %#v, want ok", result)
	}
	if state.State() != session.FastAuthCapabilitySupported {
		t.Fatalf("state = %v, want supported", state.State())
	}
}

func TestObserveValueFail(t *testing.T) {
	state := &session.FastAuthState{}
	ctx := session.ContextWithFastAuthState(
		session.ContextWithFastAuthPayload(context.Background(),
			&session.FastAuthPayload{Protocol: "mx"}),
		state,
	)

	request := PrepareFastAuthRequest(ctx)
	result := request.ObserveValue(session.FastAuthStatusFail)
	if result == nil || result.Status != session.FastAuthStatusFail {
		t.Fatalf("ObserveValue(fail) = %#v, want fail", result)
	}
	if state.State() != session.FastAuthCapabilityUnsupported {
		t.Fatalf("state = %v, want unsupported", state.State())
	}
}

func TestObserveValueEmpty(t *testing.T) {
	state := &session.FastAuthState{}
	ctx := session.ContextWithFastAuthState(
		session.ContextWithFastAuthPayload(context.Background(),
			&session.FastAuthPayload{Protocol: "mx", TargetHost: "host", TargetPort: "443", Network: "tcp"}),
		state,
	)

	request := PrepareFastAuthRequest(ctx)
	result := request.ObserveValue("")
	if result != nil {
		t.Fatal("expected nil result for empty value")
	}
	if state.State() != session.FastAuthCapabilityUnsupported {
		t.Fatalf("state = %v, want unsupported", state.State())
	}
}

func TestFastAuthPayloadSetArg(t *testing.T) {
	payload := &session.FastAuthPayload{
		Protocol:   "mx",
		AuthData:   "token",
		TargetHost: "host",
		TargetPort: "443",
		Network:    "tcp",
	}

	if payload.Arg(0) != "" {
		t.Fatal("expected empty arg 0")
	}

	payload.SetArg(0, "tcp")
	if payload.Arg(0) != "tcp" {
		t.Fatalf("arg[0] = %q, want tcp", payload.Arg(0))
	}

	payload.SetArg(5, "far")
	if payload.Arg(5) != "far" {
		t.Fatalf("arg[5] = %q, want far", payload.Arg(5))
	}
	if payload.Arg(3) != "" {
		t.Fatalf("arg[3] = %q, want empty", payload.Arg(3))
	}
}

func TestEncodeFastAuthPayloadWithTarget(t *testing.T) {
	target := net.TCPDestination(net.ParseAddress("1.2.3.4"), net.Port(8080))
	payload := &session.FastAuthPayload{
		Version:    session.FastAuthVersionClassic,
		Protocol:   "mx",
		AuthData:   "my-token",
		TargetHost: target.Address.String(),
		TargetPort: target.Port.String(),
		Network:    target.Network.SystemString(),
		Args:       []string{target.Network.SystemString()},
	}

	encoded, err := EncodeFastAuthPayload(payload)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	expected := "1|mx|my-token|tcp|1.2.3.4|8080|tcp"
	if encoded != expected {
		t.Fatalf("encoded = %q, want %q", encoded, expected)
	}
}

func TestFastAuthConnWrapper(t *testing.T) {
	wrapper := &FastAuthConnWrapper{
		FastAuthProtocol: "mx",
		FastAuthResult:   &FastAuthResult{Status: session.FastAuthStatusOK},
	}

	if wrapper.FastAuthProtocol != "mx" {
		t.Fatalf("protocol = %q, want mx", wrapper.FastAuthProtocol)
	}
	if wrapper.FastAuthResult.Status != session.FastAuthStatusOK {
		t.Fatalf("status = %q, want ok", wrapper.FastAuthResult.Status)
	}
}
