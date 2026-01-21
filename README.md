# WePick

**WePick** is a location-based group decision mobile app that helps people quickly decide where to go together. It turns the common “where should we go?” problem into a simple, swipe-based experience powered by real-time collaboration and location-aware discovery.

The project is built as a **cross-platform React Native app** backed by a **scalable Go service architecture**, with a strong emphasis on clean separation of concerns, concurrency safety, and real-time state management.

---

## Overview

Users join a shared group session and swipe to like or skip nearby places. When all participants like the same option, a match is found and the decision is made. The app uses the user’s current GPS location to surface relevant restaurants and activities in their area.

---

## Key Features

- Location-aware discovery using real-time GPS data
- Swipe-based group voting with match detection
- Shared group sessions with live state synchronization
- Real-time UI updates with clear loading and error states
- Cleanly decoupled mobile client and backend services

---

## Technical Architecture

### Mobile Client
- Built with **React Native** for iOS and Android
- Handles location permissions, swipe interactions, and session UI
- Consumes backend services via JSON-based REST APIs

### Backend Services
- Implemented in **Go**
- Stateless RESTful APIs for:
  - Place suggestions
  - Swipe and vote processing
  - Group session lifecycle management

### Data & Coordination
- **Redis** used as a shared, in-memory coordination layer
- Enables safe concurrent updates across multiple Go service instances
- Supports real-time group state and vote aggregation

---

## Tech Stack

- Frontend: React Native
- Backend: Go (Golang)
- Coordination Layer: Redis
- APIs: REST + JSON
- Location: GPS-based discovery

---

## Project Focus

This project emphasizes:
- Scalable backend design
- Concurrency-safe real-time systems
- Clear frontend–backend boundaries
- Production-style architecture suitable for multi-instance deployments

---

## Status

Work in progress. Core group matching, session handling, and real-time voting are actively being developed.

---

## License

MIT
