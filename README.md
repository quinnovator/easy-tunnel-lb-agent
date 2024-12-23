# Easy Tunnel Load Balancer Agent

A lightweight load balancer and tunnel manager that accepts tunnel connections from a Kubernetes cluster and exposes cluster services to the public internet through the tunnel, similar to Nginx.

## Features

- HTTP and TCP load balancing
- WireGuard tunnel support
- Host-based and port-based routing
- RESTful API for tunnel management
- TLS support for secure connections
- Graceful shutdown handling
- Structured logging

## Prerequisites

- Go 1.19 or later
- WireGuard tools installed on the system
- (Optional) TLS certificates for HTTPS

## Installation

```bash
go get github.com/quinnovator/easy-tunnel-lb-agent
```

Or clone the repository and build:

```bash
git clone https://github.com/quinnovator/easy-tunnel-lb-agent.git
cd easy-tunnel-lb-agent
go build -o easy-tunnel-lb-agent cmd/main.go
```

## Configuration

The agent can be configured using environment variables:

```bash
# API Server settings
export API_PORT=8080
export API_HOST=0.0.0.0
export API_BASE_PATH=/api

# Public Load Balancer settings
export PUBLIC_PORT=443
export PUBLIC_HOST=0.0.0.0

# TLS Configuration (optional)
export TLS_CERT_PATH=/path/to/cert.pem
export TLS_KEY_PATH=/path/to/key.pem

# Tunnel settings
export MAX_TUNNELS=100

# Logging
export LOG_LEVEL=info
```

## Usage

### Starting the Agent

```bash
./easy-tunnel-lb-agent --log-level=info
```

### API Endpoints

1. Create a new tunnel:

```bash
curl -X POST http://localhost:8080/api/new-tunnel \
  -H "Content-Type: application/json" \
  -d '{
    "tunnel_id": "my-service",
    "hostname": "service.example.com",
    "target_port": 8000,
    "wireguard_public_key": "your-wireguard-public-key",
    "metadata": {
      "environment": "production",
      "service": "web"
    }
  }'
```

2. Remove a tunnel:

```bash
curl -X POST http://localhost:8080/api/remove-tunnel \
  -H "Content-Type: application/json" \
  -d '{
    "tunnel_id": "my-service"
  }'
```

3. Get agent status:

```bash
curl http://localhost:8080/api/status
```

## Architecture

The agent consists of several components:

1. **API Server**: Handles tunnel management requests
2. **Load Balancer**: Routes incoming traffic to the appropriate tunnel
3. **Tunnel Manager**: Manages tunnel lifecycle and configuration
4. **Router**: Maintains routing tables for hostname and port-based routing
5. **WireGuard Manager**: Handles WireGuard tunnel setup and configuration

## Development

### Project Structure

```
easy-tunnel-lb-agent/
├── cmd/
│   └── main.go                 # Entry point
├── internal/
│   ├── api/                    # API handlers and models
│   ├── loadbalancer/          # Load balancing logic
│   ├── tunnel/                # Tunnel management
│   ├── config/                # Configuration handling
│   └── utils/                 # Utilities (logging, etc.)
└── README.md
```

### Building and Testing

```bash
# Build the binary
go build -o easy-tunnel-lb-agent cmd/main.go

# Run tests
go test ./...

# Run with debug logging
./easy-tunnel-lb-agent --log-level=debug
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
