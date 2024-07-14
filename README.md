# Raiden Framework
[![Build Status](https://github.com/sev-2/raiden/actions/workflows/raiden.yaml/badge.svg?branch=main)](https://github.com/sev-2/raiden/actions?query=branch%3Amain)
[![codecov](https://codecov.io/gh/sev-2/raiden/branch/main/graph/badge.svg)](https://codecov.io/gh/sev-2/raiden)
[![Go Report Card](https://goreportcard.com/badge/github.com/sev-2/raiden)](https://goreportcard.com/report/github.com/sev-2/raiden)

## Introduction
Raiden is a cutting-edge framework designed for seamless integration with Supabase, focusing on enhancing security, streamlining backend processes, and providing consistent schema management. It specifically addresses the need to avoid direct client-side calls to the database, ensuring more secure and efficient data handling.

## Key Objectives
- Enhanced Security: Prevent direct client-side database calls to bolster security.
- Unified Backend Management: Introduce a unified layer for managing Remote Procedure Calls (RPC), Edge Functions, and standard APIs, simplifying backend complexity.
- Consistent Schema Management: Provide tools for consistent and efficient management and persistence of database schemas.

## Features
- Secure Database Interaction: Ensures secure communication between client and database, mitigating risks associated with direct database access.
- Unified Backend Layer: Streamlines the creation and management of RPCs, Edge Functions, and APIs, offering a centralized way to handle backend logic.
- Schema Consistency: Tools to manage database schemas with ease, ensuring consistency across different stages of development.

## Getting Started
### Prerequisites
- Go (version 1.21.6 or higher)
- Supabase account and project setup

### Installation
Download our binary, or build from your local.
```
Usage:
  raiden [command]

Available Commands:
  apply       Apply resource to supabase
  build       Build app binary
  completion  Generate the autocompletion script for the specified shell
  configure   Configure project
  generate    Generate application resource
  help        Help about any command
  imports     Import supabase resource
  init        Init golang app
  run         Run app server
  serve       Serve app binary
  start       Start new app
  version     Show application information

Flags:
      --debug   enable log with debug mode
  -h, --help    help for raiden
      --trace   enable log with trace mode

Use "raiden [command] --help" for more information about a command.
```

## Documentation
For detailed documentation, including security practices and schema management, visit [raiden.sev-2.com](https://raiden.sev-2.com).

## Contributing
Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md) for more information.

## License
Raiden is open source and is licensed under the [MIT License](LICENSE).

## Contact
For support or queries, please contact us at admin@refactory.id.
