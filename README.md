# barber-shared

Shared libraries and contracts for the barber-booking system.

## Layout

```
barber-shared/
├── go/
│   ├── authmiddleware/   JWT verification middleware (Chi/Gin compatible)
│   ├── natsschemas/      NATS envelope + event payload types
│   └── paymentpb/        Generated protobuf for payment-svc gRPC
├── rust/
│   └── auth-middleware/  Tower layer + Axum extractor for JWT verification
├── proto/
│   └── payment.proto     Single source of truth for payment gRPC contract
└── buf.yaml
```

Each Go subdirectory has its own `go.mod` so consumers import only what they need and bump independently.

## Versioning

Tags use prefix per module:

```bash
git tag go/authmiddleware/v1.0.0
git tag go/natsschemas/v0.3.1
git tag go/paymentpb/v0.2.0
git tag rust-auth-middleware-v1.0.0
```

## Consuming from a Go service

```bash
go get github.com/napakornsk/barber-shared/go/authmiddleware@v1.0.0
go get github.com/napakornsk/barber-shared/go/natsschemas@v0.3.1
go get github.com/napakornsk/barber-shared/go/paymentpb@v0.2.0
```

## Private repo setup (per dev machine + CI runner)

```bash
go env -w GOPRIVATE=github.com/napakornsk/*
git config --global url."git@github.com:napakornsk/".insteadOf "https://github.com/napakornsk/"
```

## Consuming from Rust (auth-svc)

```toml
[dependencies]
barber-auth-middleware = { git = "ssh://git@github.com/napakornsk/barber-shared.git", tag = "rust-auth-middleware-v1.0.0" }
```

## Regenerating protobuf

```bash
buf generate
# or:
protoc --go_out=. --go-grpc_out=. proto/payment.proto
```
