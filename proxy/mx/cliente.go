package mx

import (
	"context"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/app/proxyman"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/net/packetaddr"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/signal"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/policy"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
	"github.com/v2fly/v2ray-core/v5/transport/internet/udp"
)

type Cliente struct {
	selector  protocol.ServerPicker
	politicas policy.Manager
}

func NuevoCliente(ctx context.Context, configuracion *ConfiguracionCliente) (*Cliente, error) {
	if len(configuracion.Vnext) == 0 {
		return nil, errors.New("mx: vnext vacio")
	}

	lista := protocol.NewServerList()
	for _, destino := range configuracion.Vnext {
		servidor, err := protocol.NewServerSpecFromPB(destino)
		if err != nil {
			return nil, errors.New("mx: servidor invalido").Base(err)
		}
		usuario := servidor.PickUser()
		if usuario == nil {
			return nil, errors.New("mx: usuario vacio")
		}
		if _, ok := usuario.Account.(*CuentaMemoria); !ok {
			return nil, errors.New("mx: cuenta invalida")
		}
		lista.AddServer(servidor)
	}

	v := core.MustFromContext(ctx)
	return &Cliente{
		selector:  protocol.NewRoundRobinServerPicker(lista),
		politicas: v.GetFeature(policy.ManagerType()).(policy.Manager),
	}, nil
}

func (c *Cliente) Process(ctx context.Context, enlace *transport.Link, marcador internet.Dialer) error {
	salida := session.OutboundFromContext(ctx)
	if salida == nil || !salida.Target.IsValid() {
		return errors.New("mx: destino no especificado")
	}
	if salida.Target.Network != net.Network_TCP && salida.Target.Network != net.Network_UDP {
		return errors.New("mx: red no soportada")
	}

	servidor := c.selector.PickServer()
	usuario := servidor.PickUser()
	cuenta, ok := usuario.Account.(*CuentaMemoria)
	if !ok {
		return errors.New("mx: cuenta invalida")
	}

	sesionRapida := proxyman.NuevaSesionAutenticacionRapida("mx", usuario, salida.Target)
	ctxConexion := proxyman.ContextWithSesionAutenticacionRapida(ctx, sesionRapida)
	if preparador, ok := marcador.(proxyman.IAutenticacionRapidaSalida); ok {
		ctxConexion = preparador.PrepararAutenticacionRapida(ctxConexion, salida.Target, usuario)
	}

	conexion, err := marcador.Dial(ctxConexion, servidor.Destination())
	if err != nil {
		return errors.New("mx: no se pudo conectar al servidor").Base(err)
	}
	defer conexion.Close()

	sesionPolitica := c.politicas.ForLevel(usuario.Level)
	if flujo, err := packetaddr.ToPacketAddrConn(enlace, salida.Target); err == nil {
		return c.procesarUDPDirigido(ctx, sesionPolitica, conexion, flujo, enlace, cuenta.Id, salida.Target, sesionRapida.OmiteCabecera())
	}

	if salida.Target.Network == net.Network_UDP {
		return c.procesarUDPFijo(ctx, sesionPolitica, conexion, enlace, cuenta.Id, salida.Target, sesionRapida.OmiteCabecera())
	}

	if !sesionRapida.OmiteCabecera() {
		if err := escribirCabecera(conexion, cuenta.Id, salida.Target); err != nil {
			return err
		}
	}
	return relevarConexion(
		ctx,
		sesionPolitica,
		ladoCopia{lector: enlace.Reader, escritor: buf.NewWriter(conexion)},
		ladoCopia{lector: buf.NewReader(conexion), escritor: enlace.Writer, cerrarSalida: true},
	)
}

func (c *Cliente) procesarUDPFijo(
	ctx context.Context,
	sesionPolitica policy.Session,
	conexion internet.Connection,
	enlace *transport.Link,
	id string,
	destino net.Destination,
	omitirCabecera bool,
) error {
	if !omitirCabecera {
		if err := escribirCabecera(conexion, id, destino); err != nil {
			return err
		}
	}
	return relevarConexion(
		ctx,
		sesionPolitica,
		ladoCopia{lector: enlace.Reader, escritor: &escritorPaqueteTamano{Writer: conexion}},
		ladoCopia{lector: &lectorPaqueteTamano{Reader: conexion}, escritor: enlace.Writer, cerrarSalida: true},
	)
}

func (c *Cliente) procesarUDPDirigido(
	ctx context.Context,
	sesionPolitica policy.Session,
	conexion internet.Connection,
	flujo net.PacketConn,
	enlace *transport.Link,
	id string,
	destino net.Destination,
	omitirCabecera bool,
) error {
	ctx, cancelar := context.WithCancel(ctx)
	defer cancelar()

	temporizador := signal.CancelAfterInactivity(ctx, cancelar, sesionPolitica.Timeouts.ConnectionIdle)

	subida := func() error {
		defer temporizador.SetTimeout(sesionPolitica.Timeouts.DownlinkOnly)

		var primero [2048]byte
		n, direccion, err := flujo.ReadFrom(primero[:])
		if err != nil {
			return errors.New("mx: no se pudo leer primer paquete").Base(err)
		}

		escritorBuffer := buf.NewBufferedWriter(buf.NewWriter(conexion))
		if !omitirCabecera {
			if err := escribirCabecera(escritorBuffer, id, destino); err != nil {
				return err
			}
		}

		escritor := &escritorPaqueteDireccionado{Writer: escritorBuffer}
		if _, err := escritor.WriteTo(primero[:n], direccion); err != nil {
			return err
		}
		if err := escritorBuffer.SetBuffered(false); err != nil {
			return errors.New("mx: no se pudo vaciar buffer").Base(err)
		}

		return udp.CopyPacketConn(escritor, flujo, udp.UpdateActivity(temporizador))
	}

	bajada := func() error {
		defer temporizador.SetTimeout(sesionPolitica.Timeouts.UplinkOnly)
		lector := &lectorConexionPaquete{lector: &lectorPaqueteDireccionado{Reader: conexion}}
		return udp.CopyPacketConn(flujo, lector, udp.UpdateActivity(temporizador))
	}

	if err := task.Run(ctx, subida, task.OnSuccess(bajada, task.Close(enlace.Writer))); err != nil {
		return errors.New("mx: conexion finalizada").Base(err)
	}
	return nil
}

func init() {
	common.Must(common.RegisterConfig((*ConfiguracionCliente)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NuevoCliente(ctx, config.(*ConfiguracionCliente))
	}))
}
