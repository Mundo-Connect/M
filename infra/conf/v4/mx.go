package v4

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v5/proxy/mx"
)

type MXConfiguracionEntrada struct {
	Users []json.RawMessage `json:"users"`
}

func (c *MXConfiguracionEntrada) Build() (proto.Message, error) {
	configuracion := new(mx.ConfiguracionServidor)
	configuracion.Users = make([]*protocol.User, len(c.Users))

	for i, bruto := range c.Users {
		usuario := new(protocol.User)
		if err := json.Unmarshal(bruto, usuario); err != nil {
			return nil, newError("MX users: invalid user").Base(err)
		}
		cuenta := new(mx.Cuenta)
		if err := json.Unmarshal(bruto, cuenta); err != nil {
			return nil, newError("MX users: invalid user").Base(err)
		}
		usuario.Account = serial.ToTypedMessage(cuenta)
		configuracion.Users[i] = usuario
	}

	return configuracion, nil
}

type MXDestino struct {
	Address *cfgcommon.Address `json:"address"`
	Port    uint16             `json:"port"`
	Users   []json.RawMessage  `json:"users"`
}

type MXConfiguracionSalida struct {
	Vnext []*MXDestino `json:"vnext"`
}

func (c *MXConfiguracionSalida) Build() (proto.Message, error) {
	if len(c.Vnext) == 0 {
		return nil, newError(`MX settings: "vnext" is empty`)
	}

	configuracion := new(mx.ConfiguracionCliente)
	configuracion.Vnext = make([]*protocol.ServerEndpoint, len(c.Vnext))

	for i, destino := range c.Vnext {
		if destino.Address == nil {
			return nil, newError(`MX vnext: "address" is not set`)
		}
		if len(destino.Users) == 0 {
			return nil, newError(`MX vnext: "users" is empty`)
		}

		servidor := &protocol.ServerEndpoint{
			Address: destino.Address.Build(),
			Port:    uint32(destino.Port),
			User:    make([]*protocol.User, len(destino.Users)),
		}

		for j, bruto := range destino.Users {
			usuario := new(protocol.User)
			if err := json.Unmarshal(bruto, usuario); err != nil {
				return nil, newError("MX users: invalid user").Base(err)
			}
			cuenta := new(mx.Cuenta)
			if err := json.Unmarshal(bruto, cuenta); err != nil {
				return nil, newError("MX users: invalid user").Base(err)
			}
			usuario.Account = serial.ToTypedMessage(cuenta)
			servidor.User[j] = usuario
		}

		configuracion.Vnext[i] = servidor
	}

	return configuracion, nil
}
