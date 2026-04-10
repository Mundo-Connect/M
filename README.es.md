```text
    MM       MM
    MMM     MMM
    MM M   M MM
    MM  M M  MM
    MM   M   MM
    MM       MM
    MM       MM
```

# Project M

`MConnectOSS` forma parte de **Project M**, el ecosistema abierto de Mundo Connect.

Project M es una nueva solución de proxy para Internet desarrollada para la familia Mundo Connect. Está diseñada como un núcleo moderno, modular y extensible para proxying de Internet, adaptación de transporte y evolución de protocolos.

Este repositorio ofrece una implementación abierta del núcleo de Project M y continúa una línea de código que se origina en V2Ray, en particular en [`v2fly/v2ray-core`](https://github.com/v2fly/v2ray-core).

Sitio web: [668993.xyz](https://668993.xyz)

Idioma: [English](README.md) | **Español**

## Descripción general

Project M está dirigido a operadores, desarrolladores, investigadores e integradores que necesitan un núcleo de proxy flexible, con una arquitectura limpia y una superficie de extensión estable.

Este repositorio se posiciona como:

- una implementación del núcleo de Project M dentro del ecosistema Mundo Connect;
- un motor modular de proxy para composición de entrada, salida, enrutamiento y transporte;
- una base práctica para integración de protocolos, experimentación de transporte y despliegue;
- una implementación de código abierto distribuida bajo la licencia MIT.

## Project M y Mundo Connect

Project M es una iniciativa central dentro del ecosistema Mundo Connect.

El objetivo de Project M es ofrecer una nueva solución de proxy para Internet con:

- una arquitectura modular;
- extensibilidad para protocolos y transportes;
- capacidad de despliegue orientada a producción;
- un núcleo abierto y mantenible;
- una ruta clara para futuras familias de protocolos bajo el marco de Mundo Connect.

Mundo Connect no se limita a un solo protocolo ni a un solo transporte. Project M está pensado como base de implementación para un ecosistema más amplio de tecnologías de proxy interoperables.

## Origen

Este repositorio se origina en la línea de código de V2Ray y se basa en la arquitectura y en la base de código de [`v2fly/v2ray-core`](https://github.com/v2fly/v2ray-core).

Project M continúa esa base técnica mientras establece su propia dirección de producto, estilo documental, identidad de ecosistema y hoja de ruta a largo plazo dentro de Mundo Connect.

## Documentación

- Sitio web oficial: [668993.xyz](https://668993.xyz)
- Documentación en inglés: [README.md](README.md)
- Origen técnico ascendente: [`v2fly/v2ray-core`](https://github.com/v2fly/v2ray-core)

## Alcance

La base de código está orientada a soportar:

- implementaciones modulares de proxy de entrada y salida;
- composición de transporte y evolución de la capa de transporte;
- control de enrutamiento, despacho y políticas;
- desarrollo de protocolos dentro del ecosistema Project M y Mundo Connect;
- integración en productos mayores de cliente, servidor y plataforma.

## Compilación

Este repositorio sigue el modelo de compilación en Go heredado de la arquitectura ascendente.

Flujo típico de desarrollo:

### Windows (PowerShell)

```powershell
go build -o mproxy.exe -trimpath -ldflags="-s -w" -v ./main
```

### Linux / macOS

```bash
CGO_ENABLED=0 go build -o mproxy -trimpath -ldflags="-s -w" -v ./main
```

El empaquetado específico de plataforma, la integración con productos y el uso embebido pueden requerir pasos adicionales según el destino.

## Licencia

Este repositorio se distribuye bajo la [licencia MIT](LICENSE).

La implementación abierta y el trabajo de protocolos contenidos en este repositorio se ofrecen bajo MIT.

## Reconocimiento

Project M reconoce la base técnica establecida por la comunidad V2Ray y el trabajo ascendente mantenido en [`v2fly/v2ray-core`](https://github.com/v2fly/v2ray-core).

Project M se desarrolla como parte del ecosistema más amplio de Mundo Connect y representa una dirección diferenciada para una nueva generación de soluciones de proxy para Internet.
