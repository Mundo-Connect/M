package mx

import (
	"strings"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common/errors"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
)

type Validador struct {
	acceso    sync.RWMutex
	porId     map[string]*protocol.MemoryUser
	porCorreo map[string]*protocol.MemoryUser
}

func (v *Validador) Add(usuario *protocol.MemoryUser) error {
	cuenta, ok := usuario.Account.(*CuentaMemoria)
	if !ok {
		return errors.New("mx: cuenta invalida")
	}
	id := limpiarId(cuenta.Id)
	if id == "" {
		return errors.New("mx: id vacio")
	}
	correo := strings.ToLower(usuario.Email)

	v.acceso.Lock()
	defer v.acceso.Unlock()

	if v.porId == nil {
		v.porId = make(map[string]*protocol.MemoryUser)
		v.porCorreo = make(map[string]*protocol.MemoryUser)
	}
	if _, encontrado := v.porId[id]; encontrado {
		return errors.New("mx: id duplicado")
	}
	if correo != "" {
		if _, encontrado := v.porCorreo[correo]; encontrado {
			return errors.New("mx: correo duplicado")
		}
	}

	cuenta.Id = id
	v.porId[id] = usuario
	if correo != "" {
		v.porCorreo[correo] = usuario
	}
	return nil
}

func (v *Validador) Del(correo string) error {
	correo = strings.ToLower(strings.TrimSpace(correo))
	if correo == "" {
		return errors.New("mx: correo vacio")
	}

	v.acceso.Lock()
	defer v.acceso.Unlock()

	usuario, encontrado := v.porCorreo[correo]
	if !encontrado {
		return errors.New("mx: usuario no encontrado")
	}

	delete(v.porCorreo, correo)
	delete(v.porId, usuario.Account.(*CuentaMemoria).Id)
	return nil
}

func (v *Validador) Buscar(id string) *protocol.MemoryUser {
	v.acceso.RLock()
	defer v.acceso.RUnlock()
	return v.porId[limpiarId(id)]
}
