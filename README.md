# OpenGate Go Client

OpenGate Go Client is a lightweight and idiomatic Go library that provides a convenient interface for interacting with the OpenGate IoT platform’s REST API. It abstracts away raw HTTP calls, offering strongly-typed methods and data structures to manage devices, data models, rules, ingestions, and other platform resources.

Status: Early preview, only south data ingestion supported — the API may evolve.

## Features

- Simple and idiomatic Go API for OpenGate
- Built-in JWT authentication and automatic token refresh
- Strongly-typed request/response models
- Support for synchronous and asynchronous calls
- Easy integration into CLI tools, backend services, or automation scripts

---

## Installation

```bash
go get github.com/<your-github-username>/opengate-go
```

Then import it in your code:

```go
import "github.com/<your-github-username>/opengate-go"
```

Requirements: Go 1.20+ is recommended.

---

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/<your-github-username>/opengate-go"
)

func main() {
    client, err := opengate.NewClient(opengate.Config{
        BaseURL:  "https://<your-opengate-instance>/api",
        Username: "your-username",
        Password: "your-password",
    })
    if err != nil {
        log.Fatalf("failed to create client: %v", err)
    }

    // Example: Get all devices
    devices, err := client.Devices.List()
    if err != nil {
        log.Fatalf("failed to get devices: %v", err)
    }

    for _, d := range devices {
        fmt.Printf("Device: %s (%s)\n", d.Name, d.ID)
    }
}
```

---

## Authentication

This library handles JWT-based authentication automatically:

- Logs in with username and password
- Stores the JWT token
- Automatically refreshes the token when it expires

---

## API Coverage

| Resource      | Supported Methods                           |
| ------------- | ------------------------------------------- |
| Devices       | `List`, `Get`, `Create`, `Update`, `Delete` |
| Data Models   | `List`, `Get`, `Create`, `Update`, `Delete` |
| Rules         | `List`, `Get`, `Create`, `Update`, `Delete` |
| Ingestions    | `List`, `Get`, `Create`, `Update`, `Delete` |
| Organizations | `List`, `Get`, `Create`, `Update`, `Delete` |

Note: This list will grow as the library evolves.

---

## Configuration

You can configure the client in two ways:

1) Pass a `Config` struct directly in Go:

```go
client, err := opengate.NewClient(opengate.Config{
    BaseURL:  "https://<your-opengate-instance>/api",
    Username: os.Getenv("OPENGATE_USERNAME"),
    Password: os.Getenv("OPENGATE_PASSWORD"),
})
```

2) Use environment variables:

```bash
export OPENGATE_BASE_URL=https://<your-opengate-instance>/api
export OPENGATE_USERNAME=your-username
export OPENGATE_PASSWORD=your-password
```

---

## Error Handling

All API methods return `(result, error)`. When possible, errors include the HTTP status code and the OpenGate API error message for easier diagnostics. Always check and handle the returned `error`.

---

## Roadmap

- [ ] Complete coverage of OpenGate API endpoints
- [ ] Add WebSocket support for real-time data
- [ ] Unit tests and integration tests
- [ ] Example CLI tool

---

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub:

```text
https://github.com/<your-github-username>/opengate-go
```

Before submitting, please run linters/formatters as applicable and include tests when adding behavior.

---

## License

This project is licensed under the MIT License — see the `LICENSE` file for details.

---

## Disclaimer

This project is not officially affiliated with Amplía Soluciones S.L. or the OpenGate platform. It is an independent open-source client library.
