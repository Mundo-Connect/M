package mx

import (
	"encoding/binary"
	"io"
	gonet "net"

	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
)

const (
	versionMX  byte = 1
	comandoTCP byte = 1
	comandoUDP byte = 2
)

var analizadorDireccion = protocol.NewAddressParser(
	protocol.AddressFamilyByte(0x01, net.AddressFamilyIPv4),
	protocol.AddressFamilyByte(0x04, net.AddressFamilyIPv6),
	protocol.AddressFamilyByte(0x03, net.AddressFamilyDomain),
	protocol.PortThenAddress(),
)

type Solicitud struct {
	Id      string
	Destino net.Destination
}

type escritorPaqueteTamano struct {
	io.Writer
}

func (e *escritorPaqueteTamano) WriteMultiBuffer(mb buf.MultiBuffer) error {
	defer buf.ReleaseMulti(mb)
	for _, b := range mb {
		if b == nil {
			continue
		}
		if err := escribirPaqueteTamano(e.Writer, b.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

type lectorPaqueteTamano struct {
	io.Reader
}

func (l *lectorPaqueteTamano) ReadMultiBuffer() (buf.MultiBuffer, error) {
	return leerPaqueteTamano(l.Reader)
}

type escritorPaqueteDireccionado struct {
	io.Writer
}

func (e *escritorPaqueteDireccionado) WriteTo(datos []byte, direccion gonet.Addr) (int, error) {
	return e.escribir(datos, net.DestinationFromAddr(direccion))
}

func (e *escritorPaqueteDireccionado) escribir(datos []byte, destino net.Destination) (int, error) {
	cabecera := buf.StackNew()
	defer cabecera.Release()

	if err := analizadorDireccion.WriteAddressPort(&cabecera, destino.Address, destino.Port); err != nil {
		return 0, err
	}

	var longitud [2]byte
	binary.BigEndian.PutUint16(longitud[:], uint16(len(datos)))
	if _, err := cabecera.Write(longitud[:]); err != nil {
		return 0, err
	}

	if err := escribirTodo(e.Writer, cabecera.Bytes()); err != nil {
		return 0, err
	}
	if err := escribirTodo(e.Writer, datos); err != nil {
		return 0, err
	}
	return len(datos), nil
}

type paqueteDireccionado struct {
	destino net.Destination
	buffer  buf.MultiBuffer
}

type lectorPaqueteDireccionado struct {
	io.Reader
}

func (l *lectorPaqueteDireccionado) leer() (*paqueteDireccionado, error) {
	direccion, puerto, err := analizadorDireccion.ReadAddressPort(nil, l.Reader)
	if err != nil {
		return nil, errors.New("mx: no se pudo leer direccion").Base(err)
	}

	var longitud [2]byte
	if _, err := io.ReadFull(l.Reader, longitud[:]); err != nil {
		return nil, errors.New("mx: no se pudo leer longitud").Base(err)
	}

	tamano := int(binary.BigEndian.Uint16(longitud[:]))
	buffer := buf.NewWithSize(int32(tamano))
	if _, err := buffer.ReadFullFrom(l.Reader, int32(tamano)); err != nil {
		buffer.Release()
		return nil, errors.New("mx: no se pudo leer paquete").Base(err)
	}

	return &paqueteDireccionado{
		destino: net.UDPDestination(direccion, puerto),
		buffer:  buf.MultiBuffer{buffer},
	}, nil
}

type lectorConexionPaquete struct {
	lector  *lectorPaqueteDireccionado
	paquete *paqueteDireccionado
}

func (l *lectorConexionPaquete) ReadFrom(datos []byte) (int, gonet.Addr, error) {
	var err error
	if l.paquete == nil || l.paquete.buffer.IsEmpty() {
		l.paquete, err = l.lector.leer()
		if err != nil {
			return 0, nil, err
		}
	}

	direccion := &gonet.UDPAddr{
		IP:   l.paquete.destino.Address.IP(),
		Port: int(l.paquete.destino.Port),
	}
	var n int
	l.paquete.buffer, n = buf.SplitFirstBytes(l.paquete.buffer, datos)
	return n, direccion, nil
}

func escribirCabecera(escritor io.Writer, id string, destino net.Destination) error {
	if len(id) > 255 {
		return errors.New("mx: id demasiado largo")
	}

	buffer := buf.StackNew()
	defer buffer.Release()

	if err := buffer.WriteByte(versionMX); err != nil {
		return err
	}
	comando := comandoTCP
	if destino.Network == net.Network_UDP {
		comando = comandoUDP
	}
	if err := buffer.WriteByte(comando); err != nil {
		return err
	}
	if err := buffer.WriteByte(byte(len(id))); err != nil {
		return err
	}
	if _, err := buffer.WriteString(id); err != nil {
		return err
	}
	if err := analizadorDireccion.WriteAddressPort(&buffer, destino.Address, destino.Port); err != nil {
		return err
	}
	return escribirTodo(escritor, buffer.Bytes())
}

func leerCabecera(lector io.Reader) (*Solicitud, error) {
	var cabecera [3]byte
	if _, err := io.ReadFull(lector, cabecera[:]); err != nil {
		return nil, errors.New("mx: no se pudo leer cabecera").Base(err)
	}
	if cabecera[0] != versionMX {
		return nil, errors.New("mx: version invalida")
	}

	id := make([]byte, int(cabecera[2]))
	if _, err := io.ReadFull(lector, id); err != nil {
		return nil, errors.New("mx: no se pudo leer id").Base(err)
	}

	direccion, puerto, err := analizadorDireccion.ReadAddressPort(nil, lector)
	if err != nil {
		return nil, errors.New("mx: no se pudo leer destino").Base(err)
	}

	red := net.Network_TCP
	if cabecera[1] == comandoUDP {
		red = net.Network_UDP
	} else if cabecera[1] != comandoTCP {
		return nil, errors.New("mx: comando invalido")
	}

	return &Solicitud{
		Id:      limpiarId(string(id)),
		Destino: net.Destination{Network: red, Address: direccion, Port: puerto},
	}, nil
}

func escribirPaqueteTamano(escritor io.Writer, datos []byte) error {
	var longitud [2]byte
	binary.BigEndian.PutUint16(longitud[:], uint16(len(datos)))
	if err := escribirTodo(escritor, longitud[:]); err != nil {
		return err
	}
	return escribirTodo(escritor, datos)
}

func leerPaqueteTamano(lector io.Reader) (buf.MultiBuffer, error) {
	var longitud [2]byte
	if _, err := io.ReadFull(lector, longitud[:]); err != nil {
		return nil, errors.New("mx: no se pudo leer longitud").Base(err)
	}

	restante := int(binary.BigEndian.Uint16(longitud[:]))
	if restante == 0 {
		return buf.MultiBuffer{buf.New()}, nil
	}

	mb := make(buf.MultiBuffer, 0, 1+(restante/buf.Size))
	for restante > 0 {
		tamano := buf.Size
		if restante < tamano {
			tamano = restante
		}
		b := buf.New()
		mb = append(mb, b)
		if _, err := b.ReadFullFrom(lector, int32(tamano)); err != nil {
			buf.ReleaseMulti(mb)
			return nil, errors.New("mx: no se pudo leer paquete").Base(err)
		}
		restante -= tamano
	}
	return mb, nil
}

func escribirTodo(escritor io.Writer, datos []byte) error {
	for len(datos) > 0 {
		n, err := escritor.Write(datos)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		datos = datos[n:]
	}
	return nil
}
