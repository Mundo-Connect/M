package mx

import (
	"bytes"
	"testing"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
)

func TestCabecera(t *testing.T) {
	destino := net.TCPDestination(net.DomainAddress("example.com"), net.Port(443))
	var b bytes.Buffer
	if err := escribirCabecera(&b, "abc", destino); err != nil {
		t.Fatal(err)
	}

	solicitud, err := leerCabecera(&b)
	if err != nil {
		t.Fatal(err)
	}
	if solicitud.Id != "abc" {
		t.Fatalf("id inesperado: %q", solicitud.Id)
	}
	if solicitud.Destino.Network != destino.Network || solicitud.Destino.Port != destino.Port || solicitud.Destino.Address.String() != destino.Address.String() {
		t.Fatalf("destino inesperado: %#v", solicitud.Destino)
	}
}

func TestPaqueteTamano(t *testing.T) {
	var b bytes.Buffer
	escritor := &escritorPaqueteTamano{Writer: &b}
	if err := escritor.WriteMultiBuffer(buf.MultiBuffer{buf.FromBytes([]byte("hola"))}); err != nil {
		t.Fatal(err)
	}

	lector := &lectorPaqueteTamano{Reader: &b}
	mb, err := lector.ReadMultiBuffer()
	if err != nil {
		t.Fatal(err)
	}
	defer buf.ReleaseMulti(mb)
	if len(mb) != 1 || string(mb[0].Bytes()) != "hola" {
		t.Fatalf("paquete inesperado: %q", string(mb[0].Bytes()))
	}
}

func TestPaqueteDireccionado(t *testing.T) {
	var b bytes.Buffer
	escritor := &escritorPaqueteDireccionado{Writer: &b}
	direccion := &net.UDPAddr{IP: []byte{1, 1, 1, 1}, Port: 53}
	if _, err := escritor.WriteTo([]byte("dns"), direccion); err != nil {
		t.Fatal(err)
	}

	lector := &lectorConexionPaquete{lector: &lectorPaqueteDireccionado{Reader: &b}}
	datos := make([]byte, 16)
	n, addr, err := lector.ReadFrom(datos)
	if err != nil {
		t.Fatal(err)
	}
	udpAddr := addr.(*net.UDPAddr)
	if n != 3 || string(datos[:n]) != "dns" || udpAddr.Port != 53 || !udpAddr.IP.Equal([]byte{1, 1, 1, 1}) {
		t.Fatalf("paquete direccionado inesperado: %d %q %v", n, string(datos[:n]), udpAddr)
	}
}
