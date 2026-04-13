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

`MConnectOSS` is part of **Project M**, the open Mundo Connect ecosystem.

Project M is a new Internet proxy solution developed for the Mundo Connect family. It is designed as a modern, modular, and extensible connectivity core for Internet proxying, transport adaptation, and protocol evolution.

This repository provides an open-source Project M core implementation and continues a code lineage that originates from V2Ray, specifically [`v2fly/v2ray-core`](https://github.com/v2fly/v2ray-core).

Website: [668993.xyz](https://668993.xyz)

Language: **English** | [Español](README.es.md)

## Overview

Project M is intended for operators, developers, researchers, and integrators who require a flexible proxy core with a clean architecture and a stable extension surface.

This repository is positioned as:

- a Project M core implementation within the Mundo Connect ecosystem;
- a modular proxy engine for inbound, outbound, routing, and transport composition;
- a practical base for protocol integration, transport experimentation, and deployment work;
- an open-source implementation distributed under the MIT License.

## Project M and Mundo Connect

Project M is a core initiative within the Mundo Connect ecosystem.

The objective of Project M is to provide a new Internet proxy solution with:

- a modular architecture;
- protocol and transport extensibility;
- production-oriented deployment capability;
- a maintainable open-source core;
- a clear path for future protocol families under the Mundo Connect umbrella.

Mundo Connect is not limited to a single protocol or a single transport. Project M is intended to serve as an implementation foundation for a broader ecosystem of interoperable proxy technologies.

## Origin

This repository originates from the V2Ray code lineage and is based on the architecture and source foundation of [`v2fly/v2ray-core`](https://github.com/v2fly/v2ray-core).

Project M continues that technical base while establishing its own product direction, documentation style, ecosystem identity, and long-term roadmap within Mundo Connect.

## Documentation

- Official website: [668993.xyz](https://668993.xyz)
- Spanish documentation: [README.es.md](README.es.md)
- Upstream technical origin: [`v2fly/v2ray-core`](https://github.com/v2fly/v2ray-core)

## Scope

The codebase is intended to support:

- modular inbound and outbound proxy implementations;
- transport composition and transport-layer evolution;
- routing, dispatch, and policy control;
- protocol development within the Project M and Mundo Connect ecosystem;
- integration into larger client, server, and platform-specific products.

## Build

This repository follows the Go-based build model inherited from the upstream core architecture.

Typical development workflow:

### Windows (PowerShell)

```powershell
go build -o mproxy.exe -trimpath -ldflags="-s -w" -v ./main
```

### Linux / macOS

```bash
go build -o mproxy -trimpath -ldflags="-s -w" -v ./main
```

Environment-specific packaging, platform integration, and embedded use may apply additional build requirements depending on the target product.

## License

This repository is distributed under the [MIT License](LICENSE).

The open-source implementation and protocol work contained in this repository are provided under MIT.

## Acknowledgement

Project M acknowledges the technical foundation established by the V2Ray community and the upstream work maintained in [`v2fly/v2ray-core`](https://github.com/v2fly/v2ray-core).

Project M is developed as part of the broader Mundo Connect ecosystem and represents a distinct direction for a new generation of Internet proxy solutions.
