package v4_test

import (
	"testing"

	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon"
	"github.com/v2fly/v2ray-core/v5/infra/conf/cfgcommon/testassist"
	v4 "github.com/v2fly/v2ray-core/v5/infra/conf/v4"
	"github.com/v2fly/v2ray-core/v5/proxy/mx"
)

func TestMXSalida(t *testing.T) {
	creator := func() cfgcommon.Buildable {
		return new(v4.MXConfiguracionSalida)
	}

	testassist.RunMultiTestCase(t, []testassist.TestCase{
		{
			Input: `{
				"vnext": [{
					"address": "example.com",
					"port": 443,
					"users": [{
						"id": "27848739-7e62-4138-9fd3-098a63964b6b",
						"level": 0
					}]
				}]
			}`,
			Parser: testassist.LoadJSON(creator),
			Output: &mx.ConfiguracionCliente{
				Vnext: []*protocol.ServerEndpoint{
					{
						Address: &net.IPOrDomain{
							Address: &net.IPOrDomain_Domain{Domain: "example.com"},
						},
						Port: 443,
						User: []*protocol.User{
							{
								Account: serial.ToTypedMessage(&mx.Cuenta{Id: "27848739-7e62-4138-9fd3-098a63964b6b"}),
								Level:   0,
							},
						},
					},
				},
			},
		},
	})
}

func TestMXEntrada(t *testing.T) {
	creator := func() cfgcommon.Buildable {
		return new(v4.MXConfiguracionEntrada)
	}

	testassist.RunMultiTestCase(t, []testassist.TestCase{
		{
			Input: `{
				"users": [{
					"id": "27848739-7e62-4138-9fd3-098a63964b6b",
					"level": 0,
					"email": "mx@v2fly.org"
				}]
			}`,
			Parser: testassist.LoadJSON(creator),
			Output: &mx.ConfiguracionServidor{
				Users: []*protocol.User{
					{
						Account: serial.ToTypedMessage(&mx.Cuenta{Id: "27848739-7e62-4138-9fd3-098a63964b6b"}),
						Level:   0,
						Email:   "mx@v2fly.org",
					},
				},
			},
		},
	})
}
