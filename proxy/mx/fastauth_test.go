package mx

import (
	"context"
	"testing"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/session"
)

func TestBuildFastAuthPayload(t *testing.T) {
	target := net.TCPDestination(net.ParseAddress("example.com"), net.Port(443))
	payload := buildFastAuthPayload("test-token", target)

	if payload.Protocol != "mx" {
		t.Fatalf("protocol = %q, want mx", payload.Protocol)
	}
	if payload.AuthData != "test-token" {
		t.Fatalf("authData = %q, want test-token", payload.AuthData)
	}
	if payload.TargetHost != "example.com" {
		t.Fatalf("targetHost = %q, want example.com", payload.TargetHost)
	}
	if payload.TargetPort != "443" {
		t.Fatalf("targetPort = %q, want 443", payload.TargetPort)
	}
	if payload.Network != "tcp" {
		t.Fatalf("network = %q, want tcp", payload.Network)
	}
	if len(payload.Args) != 1 || payload.Args[0] != "tcp" {
		t.Fatalf("args = %v, want [tcp]", payload.Args)
	}
}

func TestBuildFastAuthPayloadUDP(t *testing.T) {
	target := net.UDPDestination(net.ParseAddress("8.8.8.8"), net.Port(53))
	payload := buildFastAuthPayload("dns-token", target)

	if payload.Network != "udp" {
		t.Fatalf("network = %q, want udp", payload.Network)
	}
	if payload.Args[0] != "udp" {
		t.Fatalf("args[0] = %q, want udp", payload.Args[0])
	}
}

func TestParseFastAuthPayloadTCP(t *testing.T) {
	payload := &session.FastAuthPayload{
		Protocol:   "mx",
		AuthData:   "test-token",
		TargetHost: "1.1.1.1",
		TargetPort: "443",
		Network:    "tcp",
		Args:       []string{"tcp"},
	}

	target, protocol, err := parseFastAuthPayload(payload)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if protocol != "tcp" {
		t.Fatalf("protocol = %q, want tcp", protocol)
	}
	if target.Network != net.Network_TCP {
		t.Fatalf("target network = %v, want TCP", target.Network)
	}
	if target.Address.String() != "1.1.1.1" {
		t.Fatalf("target address = %s", target.Address)
	}
	if target.Port != net.Port(443) {
		t.Fatalf("target port = %d", target.Port)
	}
}

func TestParseFastAuthPayloadUDP(t *testing.T) {
	payload := &session.FastAuthPayload{
		Protocol:   "mx",
		AuthData:   "token",
		TargetHost: "8.8.8.8",
		TargetPort: "53",
		Network:    "udp",
		Args:       []string{"udp"},
	}

	target, protocol, err := parseFastAuthPayload(payload)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if protocol != "udp" {
		t.Fatalf("protocol = %q, want udp", protocol)
	}
	if target.Network != net.Network_UDP {
		t.Fatalf("target network = %v, want UDP", target.Network)
	}
}

func TestParseFastAuthPayloadFallbackLegacy(t *testing.T) {
	payload := &session.FastAuthPayload{
		Protocol:   "mx",
		AuthData:   "token",
		TargetHost: "8.8.8.8",
		TargetPort: "53",
		Network:    "udp",
	}

	target, protocol, err := parseFastAuthPayload(payload)
	if err != nil {
		t.Fatalf("parse legacy: %v", err)
	}

	if protocol != "udp" {
		t.Fatalf("protocol = %q, want udp", protocol)
	}
	if target.Network != net.Network_UDP {
		t.Fatalf("target network = %v, want UDP", target.Network)
	}
}

func TestParseFastAuthPayloadInvalidProtocol(t *testing.T) {
	payload := &session.FastAuthPayload{
		Protocol: "vless",
	}

	_, _, err := parseFastAuthPayload(payload)
	if err == nil {
		t.Fatal("expected error for non-mx protocol")
	}
}

func TestParseFastAuthPayloadEmptyHost(t *testing.T) {
	payload := &session.FastAuthPayload{
		Protocol:   "mx",
		TargetHost: "",
		TargetPort: "443",
	}

	_, _, err := parseFastAuthPayload(payload)
	if err == nil {
		t.Fatal("expected error for empty host")
	}
}

func TestParseFastAuthPayloadInvalidPort(t *testing.T) {
	payload := &session.FastAuthPayload{
		Protocol:   "mx",
		TargetHost: "example.com",
		TargetPort: "invalid",
	}

	_, _, err := parseFastAuthPayload(payload)
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestValidateFastAuth(t *testing.T) {
	validador := &Validador{}
	usuario := &protocol.MemoryUser{
		Account: &CuentaMemoria{Id: "test-token"},
		Level:   1,
	}
	if err := validador.Add(usuario); err != nil {
		t.Fatalf("add user: %v", err)
	}

	servidor := &Servidor{validador: validor}

	payload := &session.FastAuthPayload{
		Protocol:   "mx",
		AuthData:   "test-token",
		TargetHost: "example.com",
		TargetPort: "443",
		Network:    "tcp",
		Args:       []string{"tcp"},
	}

	inbound, err := servidor.ValidateFastAuth(context.Background(), payload)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}

	if inbound == nil {
		t.Fatal("inbound is nil")
	}
	if inbound.Protocol != "mx" {
		t.Fatalf("protocol = %q, want mx", inbound.Protocol)
	}
	if !inbound.ConsumeClassicHeader {
		t.Fatal("expected ConsumeClassicHeader to be true")
	}
	if user, ok := inbound.User.(*protocol.MemoryUser); !ok || user != usuario {
		t.Fatal("user mismatch")
	}
	if target, ok := inbound.Target.(net.Destination); !ok || target.Address.String() != "example.com" {
		t.Fatal("target mismatch")
	}
}

func TestValidateFastAuthInvalidToken(t *testing.T) {
	validador := &Validador{}
	servidor := &Servidor{validador: validor}

	payload := &session.FastAuthPayload{
		Protocol:   "mx",
		AuthData:   "invalid-token",
		TargetHost: "example.com",
		TargetPort: "443",
		Network:    "tcp",
		Args:       []string{"tcp"},
	}

	_, err := servidor.ValidateFastAuth(context.Background(), payload)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestValidateFastAuthEmptyToken(t *testing.T) {
	validador := &Validador{}
	servidor := &Servidor{validador: validor}

	payload := &session.FastAuthPayload{
		Protocol:   "mx",
		AuthData:   "",
		TargetHost: "example.com",
		TargetPort: "443",
		Network:    "tcp",
		Args:       []string{"tcp"},
	}

	_, err := servidor.ValidateFastAuth(context.Background(), payload)
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestValidateFastAuthNonMX(t *testing.T) {
	validador := &Validador{}
	servidor := &Servidor{validador: validor}

	payload := &session.FastAuthPayload{
		Protocol:   "vmess",
		AuthData:   "token",
		TargetHost: "example.com",
		TargetPort: "443",
	}

	_, err := servidor.ValidateFastAuth(context.Background(), payload)
	if err == nil {
		t.Fatal("expected error for non-mx protocol")
	}
}

func TestProxyFastAuthState(t *testing.T) {
	dest := net.TCPDestination(net.ParseAddress("server.example.com"), net.Port(443))
	state := proxyFastAuthState("token-1", dest)
	if state == nil {
		t.Fatal("state is nil")
	}
	if state.State() != session.FastAuthCapabilityUnknown {
		t.Fatalf("initial state = %v, want unknown", state.State())
	}

	state2 := proxyFastAuthState("token-1", dest)
	if state != state2 {
		t.Fatal("expected same state instance for same key")
	}

	state3 := proxyFastAuthState("token-2", dest)
	if state == state3 {
		t.Fatal("expected different state instance for different key")
	}
}
