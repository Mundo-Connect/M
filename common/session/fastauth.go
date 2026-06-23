package session

import (
	"context"
	"sync/atomic"
)

type FastAuthPayload struct {
	Version    uint8
	Protocol   string
	AuthData   string
	TargetHost string
	TargetPort string
	Network    string
	Args       []string
}

type FastAuthInbound struct {
	Protocol             string
	User                 interface{}
	Target               interface{}
	Args                 []string
	Prepared             *FastAuthPrepared
	ConsumeClassicHeader bool
}

type FastAuthValidator interface {
	ValidateFastAuth(context.Context, *FastAuthPayload) (*FastAuthInbound, error)
}

type FastAuthPrepared struct {
	link  interface{}
	err   error
	ready chan struct{}
}

func NewFastAuthPrepared() *FastAuthPrepared {
	return &FastAuthPrepared{
		ready: make(chan struct{}),
	}
}

func (p *FastAuthPrepared) Complete(link interface{}, err error) {
	if p == nil {
		return
	}
	p.link = link
	p.err = err
	close(p.ready)
}

func (p *FastAuthPrepared) Wait() (interface{}, error) {
	if p == nil {
		return nil, nil
	}
	<-p.ready
	return p.link, p.err
}

const (
	FastAuthVersionClassic uint8 = 1
)

type FastAuthCapabilityState uint32

const (
	FastAuthCapabilityUnknown FastAuthCapabilityState = iota
	FastAuthCapabilityUnsupported
	FastAuthCapabilitySupported
)

const (
	FastAuthStatusOK   = "ok"
	FastAuthStatusFail = "fail"
)

type FastAuthState struct {
	state atomic.Uint32
}

func (s *FastAuthState) State() FastAuthCapabilityState {
	if s == nil {
		return FastAuthCapabilityUnknown
	}
	return FastAuthCapabilityState(s.state.Load())
}

func (s *FastAuthState) MarkSupported() {
	if s == nil {
		return
	}
	s.state.Store(uint32(FastAuthCapabilitySupported))
}

func (s *FastAuthState) MarkUnsupported() {
	if s == nil {
		return
	}
	s.state.Store(uint32(FastAuthCapabilityUnsupported))
}

func (p *FastAuthPayload) Arg(index int) string {
	if index < 0 || index >= len(p.Args) {
		return ""
	}
	return p.Args[index]
}

func (p *FastAuthPayload) SetArg(index int, value string) {
	if index < 0 {
		return
	}
	if index >= len(p.Args) {
		next := make([]string, index+1)
		copy(next, p.Args)
		p.Args = next
	}
	p.Args[index] = value
}

func (i *FastAuthInbound) Arg(index int) string {
	if index < 0 || index >= len(i.Args) {
		return ""
	}
	return i.Args[index]
}

func (i *FastAuthInbound) SetArg(index int, value string) {
	if index < 0 {
		return
	}
	if index >= len(i.Args) {
		next := make([]string, index+1)
		copy(next, i.Args)
		i.Args = next
	}
	i.Args[index] = value
}

type fastAuthPayloadKey struct{}
type fastAuthValidatorKey struct{}
type fastAuthInboundKey struct{}
type fastAuthStateKey struct{}

func ContextWithFastAuthPayload(ctx context.Context, payload *FastAuthPayload) context.Context {
	if payload == nil {
		return ctx
	}
	return context.WithValue(ctx, fastAuthPayloadKey{}, payload)
}

func FastAuthPayloadFromContext(ctx context.Context) *FastAuthPayload {
	payload, _ := ctx.Value(fastAuthPayloadKey{}).(*FastAuthPayload)
	return payload
}

func ContextWithFastAuthValidator(ctx context.Context, validator FastAuthValidator) context.Context {
	if validator == nil {
		return ctx
	}
	return context.WithValue(ctx, fastAuthValidatorKey{}, validator)
}

func FastAuthValidatorFromContext(ctx context.Context) FastAuthValidator {
	validator, _ := ctx.Value(fastAuthValidatorKey{}).(FastAuthValidator)
	return validator
}

func ContextWithFastAuthInbound(ctx context.Context, inbound *FastAuthInbound) context.Context {
	if inbound == nil {
		return ctx
	}
	return context.WithValue(ctx, fastAuthInboundKey{}, inbound)
}

func FastAuthInboundFromContext(ctx context.Context) *FastAuthInbound {
	inbound, _ := ctx.Value(fastAuthInboundKey{}).(*FastAuthInbound)
	return inbound
}

func ContextWithFastAuthState(ctx context.Context, state *FastAuthState) context.Context {
	if state == nil {
		return ctx
	}
	return context.WithValue(ctx, fastAuthStateKey{}, state)
}

func FastAuthStateFromContext(ctx context.Context) *FastAuthState {
	state, _ := ctx.Value(fastAuthStateKey{}).(*FastAuthState)
	return state
}
