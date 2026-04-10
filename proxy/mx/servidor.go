package mx

import (
	"context"
	"io"
	"time"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	udp_proto "github.com/v2fly/v2ray-core/v5/common/protocol/udp"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/features/policy"
	"github.com/v2fly/v2ray-core/v5/features/routing"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	udp "github.com/v2fly/v2ray-core/v5/transport/internet/udp"
)

type Servidor struct {
	politicas policy.Manager
	validador *Validador
}

func NuevoServidor(ctx context.Context, configuracion *ConfiguracionServidor) (*Servidor, error) {
	validador := new(Validador)
	for _, usuario := range configuracion.Users {
		memoria, err := usuario.ToMemoryUser()
		if err != nil {
			return nil, errors.New("mx: usuario invalido").Base(err)
		}
		if err := validador.Add(memoria); err != nil {
			return nil, err
		}
	}

	v := core.MustFromContext(ctx)
	return &Servidor{
		politicas: v.GetFeature(policy.ManagerType()).(policy.Manager),
		validador: validador,
	}, nil
}

func (s *Servidor) AddUser(ctx context.Context, usuario *protocol.MemoryUser) error {
	return s.validador.Add(usuario)
}

func (s *Servidor) RemoveUser(ctx context.Context, correo string) error {
	return s.validador.Del(correo)
}

func (*Servidor) Network() []net.Network {
	return []net.Network{net.Network_TCP, net.Network_UNIX}
}

func (s *Servidor) Process(ctx context.Context, _ net.Network, conexion internet.Connection, despachador routing.Dispatcher) error {
	sesionPolitica := s.politicas.ForLevel(0)
	if err := conexion.SetReadDeadline(time.Now().Add(sesionPolitica.Timeouts.Handshake)); err != nil {
		return errors.New("mx: no se pudo fijar timeout").Base(err)
	}

	solicitud, err := leerCabecera(conexion)
	if err != nil {
		return err
	}

	usuario := s.validador.Buscar(solicitud.Id)
	if usuario == nil {
		return errors.New("mx: usuario invalido")
	}

	if err := conexion.SetReadDeadline(time.Time{}); err != nil {
		return errors.New("mx: no se pudo limpiar timeout").Base(err)
	}

	inbound := session.InboundFromContext(ctx)
	if inbound != nil {
		inbound.User = usuario
	}

	sesionPolitica = s.politicas.ForLevel(usuario.Level)
	ctx = policy.ContextWithBufferPolicy(ctx, sesionPolitica.Buffer)

	if solicitud.Destino.Network == net.Network_UDP {
		if _, err := packetaddr.GetDestinationSubsetOf(solicitud.Destino); err == nil {
			return s.procesarUDPDirigido(ctx, conexion, despachador)
		}
		return s.procesarUDPFijo(ctx, conexion, despachador, solicitud.Destino)
	}

	enlace, err := despachador.Dispatch(ctx, solicitud.Destino)
	if err != nil {
		return errors.New("mx: no se pudo despachar").Base(err)
	}

	return relevarConexion(
		ctx,
		sesionPolitica,
		ladoCopia{lector: buf.NewReader(conexion), escritor: enlace.Writer, cerrarSalida: true},
		ladoCopia{lector: enlace.Reader, escritor: buf.NewWriter(conexion)},
	)
}

func (s *Servidor) procesarUDPFijo(
	ctx context.Context,
	conexion internet.Connection,
	despachador routing.Dispatcher,
	destino net.Destination,
) error {
	lector := &lectorPaqueteTamano{Reader: conexion}
	escritor := &escritorPaqueteTamano{Writer: conexion}
	servidor := udp.NewSplitDispatcher(despachador, func(ctx context.Context, paquete *udp_proto.Packet) {
		_ = escritor.WriteMultiBuffer(buf.MultiBuffer{paquete.Payload})
	})
	defer servidor.Close()

	for {
		mb, err := lector.ReadMultiBuffer()
		if err != nil {
			if errors.Cause(err) == io.EOF {
				return nil
			}
			return err
		}
		for _, b := range mb {
			servidor.Dispatch(ctx, destino, b)
		}
	}
}

func (s *Servidor) procesarUDPDirigido(
	ctx context.Context,
	conexion internet.Connection,
	despachador routing.Dispatcher,
) error {
	lector := &lectorPaqueteDireccionado{Reader: conexion}
	escritor := &escritorPaqueteDireccionado{Writer: conexion}
	creador := udp.NewPacketAddrDispatcherCreator(ctx)
	servidor := creador.NewPacketAddrDispatcher(despachador, func(ctx context.Context, paquete *udp_proto.Packet) {
		direccion := &net.UDPAddr{
			IP:   paquete.Source.Address.IP(),
			Port: int(paquete.Source.Port),
		}
		_, _ = escritor.WriteTo(paquete.Payload.Bytes(), direccion)
		paquete.Payload.Release()
	})
	defer servidor.Close()

	for {
		paquete, err := lector.leer()
		if err != nil {
			if errors.Cause(err) == io.EOF {
				return nil
			}
			return err
		}
		for _, b := range paquete.buffer {
			servidor.Dispatch(ctx, paquete.destino, b)
		}
	}
}

func init() {
	common.Must(common.RegisterConfig((*ConfiguracionServidor)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NuevoServidor(ctx, config.(*ConfiguracionServidor))
	}))
}
