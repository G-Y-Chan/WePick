# Project Memory: WePick - Group Decision App

## Project Overview
WePick is a location-based group decision mobile app that helps users choose a venue together through swipe-based voting, room-based collaboration, and real-time updates.

## System Architecture
- `/backend`: Go service with HTTP endpoints and WebSocket room sessions
  - Entry point is `backend/cmd/server/main.go`
  - HTTP handlers live in `backend/internal/transport/http/`
  - WebSocket handler and connection hub live in `backend/internal/transport/ws/`
  - Room lifecycle, voting, and fan-out logic live in `backend/internal/room/`
  - Redis client setup lives in `backend/internal/infra/redis.go`
  - Environment loading lives in `backend/internal/config/config.go`
  - Places API client lives in `backend/internal/places/`
  - CORS middleware lives in `backend/internal/transport/http/middleware/`
- `/frontend`: Expo Router React Native app written in TypeScript
  - Root navigation lives in `frontend/app/_layout.tsx`
  - Room-scoped navigation lives in `frontend/app/(room)/_layout.tsx`
  - Route screens live in `frontend/app/` and `frontend/app/(room)/`
  - Reusable swipe UI logic lives in `frontend/src/swipe/`
  - App state currently uses React Context in `frontend/src/context/`, not Zustand
  - HTTP and WebSocket integrations live in `frontend/services/`
- `/nginx` and `docker-compose.yml`: Multi-instance backend setup
  - Nginx load-balances multiple Go app instances
  - Redis is the shared coordination layer across instances

## Distributed System Reality
The backend is designed for horizontally distributed rooms. Users in the same room may be connected to different Go server instances behind Nginx. Because of that:
- Local in-memory connection maps in `RoomManager` and `RoomConnections` are instance-local only
- Shared room state, counters, and cross-instance events must go through Redis
- Pub-sub on the `room_events` channel is how one instance notifies other instances about room-wide events such as `room_started` and `majority_found`
- Any new room, vote, presence, or match feature must be designed to work correctly when participants are split across multiple servers

## Backend Design Notes
- `RoomManager` owns process-local orchestration and relays broadcast events to the `Hub` (connection registry in `internal/transport/ws/`)
- `RoomRepository` (concrete implementation in `internal/room/repository.go`) owns Redis-backed persistence, counters, scripts, and pub-sub
- The domain layer (`internal/room/`) uses the `Repository` interface — no direct Redis dependency in the manager
- `HandleVote` uses Redis-backed client counts and vote increments with SADD-based dedup via clientID, so majority detection is cross-instance and replay-proof
- `StartEventListener` subscribes every server instance to room events so each instance can rebroadcast to its own local WebSocket clients
- HTTP DTOs live in `backend/internal/transport/http/dto.go`, separate from domain types in `internal/room/room.go`

## Frontend Design Notes
- Expo Router route groups are used to scope room screens under `frontend/app/(room)/`
- `RoomProvider` manages the room WebSocket lifecycle and exposes `status`, `lastMessage`, and `send`
- `PlacesProvider` stores card data for result lookup after navigation
- Swipe interactions are split into focused modules: `SwipeCard`, `SwipeDeck`, and `useSwipeDeck`
- Frontend message types live in `frontend/services/types.ts` and swipe domain types live in `frontend/src/swipe/types.ts`

## Core Immutable Rules
1. Never design room features as if all participants are connected to the same backend instance.
2. Use `RoomRepository` for shared room state, counters, Redis scripts, and pub-sub; use `RoomManager` for instance-local orchestration only.
3. For cross-instance room events, publish through Redis and make every server instance handle the event via `StartEventListener`.
4. Do not store authoritative room state only in process memory; in-memory maps are for local WebSocket connections, not global truth.
5. Preserve message contract compatibility between `backend/internal/transport/http/dto.go`, `frontend/services/types.ts`, and any route logic consuming those messages.
6. Follow the existing Expo Router structure and keep room-scoped state inside providers rather than passing duplicated params through many screens.
7. Do not introduce new third-party dependencies without explicit permission.
