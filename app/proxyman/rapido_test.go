package proxyman

import (
	"context"
	"testing"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

func TestSesionAutenticacionRapida(t *testing.T) {
	sesion := NuevaSesionAutenticacionRapida("mx", nil, net.TCPDestination(net.DomainAddress("example.com"), net.Port(443)))
	if sesion.OmiteCabecera() {
		t.Fatal("unexpected active fast auth session")
	}
	sesion.Activar()
	if !sesion.OmiteCabecera() {
		t.Fatal("fast auth session should skip classic header")
	}

	ctx := ContextWithSesionAutenticacionRapida(context.Background(), sesion)
	if SesionAutenticacionRapidaFromContext(ctx) != sesion {
		t.Fatal("failed to retrieve fast auth session from context")
	}
}

func TestResultadoAutenticacionRapida(t *testing.T) {
	resultado := &ResultadoAutenticacionRapida{
		Protocolo: "mx",
		Destino:   net.TCPDestination(net.DomainAddress("example.com"), net.Port(443)),
	}
	ctx := ContextWithResultadoAutenticacionRapida(context.Background(), resultado)
	if ResultadoAutenticacionRapidaFromContext(ctx) != resultado {
		t.Fatal("failed to retrieve fast auth result from context")
	}
}
