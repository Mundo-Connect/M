package mx

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/signal"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/policy"
)

type ladoCopia struct {
	lector       buf.Reader
	escritor     buf.Writer
	cerrarSalida bool
}

func relevarConexion(ctx context.Context, sesion policy.Session, subida ladoCopia, bajada ladoCopia) error {
	ctx, cancelar := context.WithCancel(ctx)
	defer cancelar()

	temporizador := signal.CancelAfterInactivity(ctx, cancelar, sesion.Timeouts.ConnectionIdle)

	tareaSubida := func() error {
		defer temporizador.SetTimeout(sesion.Timeouts.DownlinkOnly)
		if err := buf.Copy(subida.lector, subida.escritor, buf.UpdateActivity(temporizador)); err != nil {
			return errors.New("mx: error en subida").Base(err)
		}
		return nil
	}
	tareaBajada := func() error {
		defer temporizador.SetTimeout(sesion.Timeouts.UplinkOnly)
		if err := buf.Copy(bajada.lector, bajada.escritor, buf.UpdateActivity(temporizador)); err != nil {
			return errors.New("mx: error en bajada").Base(err)
		}
		return nil
	}

	if subida.cerrarSalida {
		tareaSubida = task.OnSuccess(tareaSubida, task.Close(subida.escritor))
	}
	if bajada.cerrarSalida {
		tareaBajada = task.OnSuccess(tareaBajada, task.Close(bajada.escritor))
	}

	if err := task.Run(ctx, tareaSubida, tareaBajada); err != nil {
		_ = common.Interrupt(subida.lector)
		_ = common.Interrupt(subida.escritor)
		_ = common.Interrupt(bajada.lector)
		_ = common.Interrupt(bajada.escritor)
		return errors.New("mx: conexion finalizada").Base(err)
	}
	return nil
}
