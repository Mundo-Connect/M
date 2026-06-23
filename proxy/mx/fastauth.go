package mx

import (
	"context"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

const fastAuthProtocol = "mx"

const (
	argPayloadProtocol = 0
)

var (
	fastAuthStates   = make(map[string]*session.FastAuthState)
	fastAuthStatesMu sync.RWMutex
)

func proxyFastAuthState(token string, dest net.Destination) *session.FastAuthState {
	key := token + "@" + dest.NetAddr()
	fastAuthStatesMu.RLock()
	state, ok := fastAuthStates[key]
	fastAuthStatesMu.RUnlock()
	if ok {
		return state
	}

	fastAuthStatesMu.Lock()
	defer fastAuthStatesMu.Unlock()
	if state, ok = fastAuthStates[key]; ok {
		return state
	}
	state = &session.FastAuthState{}
	fastAuthStates[key] = state
	return state
}

func observeFastAuthResult(conn internet.Connection, state *session.FastAuthState) {
	if state == nil {
		return
	}
	if wrapper, ok := conn.(*internet.FastAuthConnWrapper); ok && wrapper != nil {
		if wrapper.FastAuthResult != nil && wrapper.FastAuthResult.Status == session.FastAuthStatusOK {
			state.MarkSupported()
		} else {
			state.MarkUnsupported()
		}
	} else {
		state.MarkUnsupported()
	}
}

func buildFastAuthPayload(token string, target net.Destination) *session.FastAuthPayload {
	return &session.FastAuthPayload{
		Version:    session.FastAuthVersionClassic,
		Protocol:   fastAuthProtocol,
		AuthData:   token,
		TargetHost: target.Address.String(),
		TargetPort: target.Port.String(),
		Network:    target.Network.SystemString(),
		Args:       []string{target.Network.SystemString()},
	}
}

func parseFastAuthPayload(payload *session.FastAuthPayload) (net.Destination, string, error) {
	if payload == nil || payload.Protocol != fastAuthProtocol {
		return net.Destination{}, "", errors.New("mx: invalid fast auth protocol")
	}
	if payload.TargetHost == "" || payload.TargetPort == "" {
		return net.Destination{}, "", errors.New("mx: missing fast auth target")
	}

	payloadProtocol := payload.Arg(argPayloadProtocol)
	if payloadProtocol == "" {
		payloadProtocol = payload.Network
	}

	port, err := net.PortFromString(payload.TargetPort)
	if err != nil {
		return net.Destination{}, "", errors.New("mx: invalid fast auth port").Base(err)
	}

	address := net.ParseAddress(payload.TargetHost)
	switch payloadProtocol {
	case "tcp":
		return net.TCPDestination(address, port), payloadProtocol, nil
	case "udp":
		return net.UDPDestination(address, port), payloadProtocol, nil
	default:
		return net.Destination{}, "", errors.New("mx: unsupported payload protocol: ", payloadProtocol)
	}
}

func (s *Servidor) ValidateFastAuth(_ context.Context, payload *session.FastAuthPayload) (*session.FastAuthInbound, error) {
	target, payloadProtocol, err := parseFastAuthPayload(payload)
	if err != nil {
		return nil, err
	}

	token := limpiarId(payload.AuthData)
	if token == "" {
		return nil, errors.New("mx: empty fast auth token")
	}

	user := s.validador.Buscar(token)
	if user == nil {
		return nil, errors.New("mx: invalid fast auth user")
	}

	return &session.FastAuthInbound{
		Protocol:             fastAuthProtocol,
		User:                 user,
		Target:               target,
		Args:                 []string{payloadProtocol},
		ConsumeClassicHeader: true,
	}, nil
}
