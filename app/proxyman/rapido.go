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
