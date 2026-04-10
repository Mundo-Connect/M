package mx

import (
	"strings"

	"github.com/v2fly/v2ray-core/v5/common/protocol"
)

func (c *Cuenta) AsAccount() (protocol.Account, error) {
	return &CuentaMemoria{Id: limpiarId(c.Id)}, nil
}

type CuentaMemoria struct {
	Id string
}

func (c *CuentaMemoria) Equals(otra protocol.Account) bool {
	rival, ok := otra.(*CuentaMemoria)
	return ok && c.Id == rival.Id
}

func limpiarId(id string) string {
	return strings.TrimSpace(id)
}
