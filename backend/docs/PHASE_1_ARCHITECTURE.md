# Phase 1 Architecture

`cmd/api` starts the application.

`internal/app` composes dependencies.

`internal/server` owns HTTP routing.

`internal/database` owns MySQL connectivity.

`internal/middleware` owns cross-cutting HTTP protections.

Future business modules follow handler -> service -> repository -> database.

The public PAYE calculator will remain usable without authentication.
