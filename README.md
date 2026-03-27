# GoChess

GoChess is a compact chess engine written in Go with two front ends: a local browser UI and a UCI mode for chess GUIs. It uses a bitboard-based engine, iterative deepening alpha-beta search, and built-in perft support for engine validation.

## Features

- Go 1.21 project with no third-party runtime dependencies
- Bitboard board representation with incremental make/unmake
- UCI protocol support for external GUI integration
- Built-in web UI served by the engine over HTTP
- Alpha-beta search with iterative deepening and aspiration windows
- Quiescence search, null-move pruning, killer/history move ordering, and transposition tables
- Zobrist hashing, repetition detection, FEN parsing, and perft support

## Project Layout

- `main.go` selects `web` or `uci` mode
- `engine/` contains board logic, move generation, evaluation, search, and UCI handling
- `web/` contains the HTTP server and embedded single-page interface

## Requirements

- Go 1.21 or newer

## Build

```bash
go build -o gochess .
```

On Windows the output binary will be `gochess.exe`.

## Run

Start the browser UI on the default port:

```bash
go run . -mode web
```

Start the browser UI on a custom port:

```bash
go run . -mode web -port 9090
```

Then open `http://localhost:8080` or your chosen port.

Start UCI mode:

```bash
go run . -mode uci
```

Build once, then run the binary:

```bash
./gochess -mode web
./gochess -mode uci
```

## UCI Usage

Minimal manual session:

```text
uci
isready
ucinewgame
position startpos
go depth 8
quit
```

Supported commands in the current implementation include:

- `uci`
- `isready`
- `ucinewgame`
- `position startpos`
- `position fen <fen> moves <...>`
- `go depth <n>`
- `go movetime <ms>`
- `go wtime <ms> btime <ms> [winc <ms>] [binc <ms>] [movestogo <n>]`
- `stop`
- `perft <depth>`
- `d`
- `eval`

## Web UI

The web mode exposes:

- `GET /` for the UI
- `GET /api/state` for current board state
- `POST /api/move` to play a move
- `POST /api/newgame` to start a game with selected color and depth
- `POST /api/undo` to undo the last full move pair
- `GET /api/legalmoves?square=<sq>` for square-specific legal moves

The UI lets you:

- play as White or Black
- set engine depth from 1 to 20
- view move list, evaluation, nodes, depth, and NPS
- undo the last move pair and flip the board

## Verify

```bash
go build ./...
go test ./...
```

`go test ./...` is still useful as a package check, but the current workspace does not include any `_test.go` files.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
