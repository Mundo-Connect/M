package proxyman

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

type ResultadoAutenticacionRapida struct {
	Protocolo string
	Usuario   *protocol.MemoryUser
	Destino   net.Destination
	Enlace    *transport.Link
}

type IAutenticacionRapidaEntrada interface {
	ResolverAutenticacionRapida(context.Context, internet.Connection) (*ResultadoAutenticacionRapida, error)
}

type IAutenticacionRapidaSalida interface {
	PrepararAutenticacionRapida(context.Context, net.Destination, *protocol.MemoryUser) context.Context
}

type SesionAutenticacionRapida struct {
	Protocolo      string
	Usuario        *protocol.MemoryUser
	Destino        net.Destination
	omitirCabecera bool
}

func NuevaSesionAutenticacionRapida(protocolo string, usuario *protocol.MemoryUser, destino net.Destination) *SesionAutenticacionRapida {
	return &SesionAutenticacionRapida{
		Protocolo: protocolo,
		Usuario:   usuario,
		Destino:   destino,
	}
}

func (s *SesionAutenticacionRapida) Activar() {
	s.omitirCabecera = true
}

func (s *SesionAutenticacionRapida) OmiteCabecera() bool {
	return s != nil && s.omitirCabecera
}

type claveResultadoAutenticacionRapida struct{}
type claveSesionAutenticacionRapida struct{}

func ContextWithResultadoAutenticacionRapida(ctx context.Context, resultado *ResultadoAutenticacionRapida) context.Context {
	return context.WithValue(ctx, claveResultadoAutenticacionRapida{}, resultado)
}

func ResultadoAutenticacionRapidaFromContext(ctx context.Context) *ResultadoAutenticacionRapida {
	valor := ctx.Value(claveResultadoAutenticacionRapida{})
	if valor == nil {
		return nil
	}
	return valor.(*ResultadoAutenticacionRapida)
}

func ContextWithSesionAutenticacionRapida(ctx context.Context, sesion *SesionAutenticacionRapida) context.Context {
	return context.WithValue(ctx, claveSesionAutenticacionRapida{}, sesion)
}

func SesionAutenticacionRapidaFromContext(ctx context.Context) *SesionAutenticacionRapida {
	valor := ctx.Value(claveSesionAutenticacionRapida{})
	if valor == nil {
		return nil
	}
	return valor.(*SesionAutenticacionRapida)
}
