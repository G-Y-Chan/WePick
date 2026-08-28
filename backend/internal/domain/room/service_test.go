package room

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"testing"
	"time"

	"backend/internal/apperr"
)

// ---------------------------------------------------------------------------
// Fakes — hand-written implementations of the three domain ports.
// ---------------------------------------------------------------------------

type fakeRepository struct {
	mu sync.Mutex

	rooms       map[Code]bool
	started     map[Code]bool
	cards       map[Code][]Card
	pageTokens  map[Code]string
	clientCount map[Code]int64
	votes       map[Code]map[string]int

	published []Event

	createCollisions int
	majorityResult   bool

	createErr      error
	existsErr      error
	isStartedErr   error
	markStartedErr error
	saveCardsErr   error
	loadCardsErr   error
	savePageErr    error
	loadPageErr    error
	incrErr        error
	decrErr        error
	clientCountErr error
	voteErr        error
	publishErr     error
	subscribeErr   error

	subscriptions  []*fakeEventSubscription
	subscribeCalls int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		rooms:       make(map[Code]bool),
		started:     make(map[Code]bool),
		cards:       make(map[Code][]Card),
		pageTokens:  make(map[Code]string),
		clientCount: make(map[Code]int64),
		votes:       make(map[Code]map[string]int),
	}
}

func (f *fakeRepository) CreateRoom(_ context.Context, code Code, _ time.Duration) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.createErr != nil {
		return false, f.createErr
	}
	if f.createCollisions > 0 {
		f.createCollisions--
		return false, nil
	}
	if f.rooms[code] {
		return false, nil
	}
	f.rooms[code] = true
	f.started[code] = false
	f.clientCount[code] = 0
	return true, nil
}

func (f *fakeRepository) Exists(_ context.Context, code Code) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.existsErr != nil {
		return false, f.existsErr
	}
	return f.rooms[code], nil
}

func (f *fakeRepository) IsStarted(_ context.Context, code Code) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.isStartedErr != nil {
		return false, f.isStartedErr
	}
	return f.started[code], nil
}

func (f *fakeRepository) MarkStarted(_ context.Context, code Code) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.markStartedErr != nil {
		return f.markStartedErr
	}
	f.started[code] = true
	return nil
}

func (f *fakeRepository) SaveCards(_ context.Context, code Code, cards []Card) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.saveCardsErr != nil {
		return f.saveCardsErr
	}
	f.cards[code] = append([]Card(nil), cards...)
	return nil
}

func (f *fakeRepository) LoadCards(_ context.Context, code Code) ([]Card, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.loadCardsErr != nil {
		return nil, f.loadCardsErr
	}
	cards, ok := f.cards[code]
	if !ok {
		return nil, nil
	}
	return append([]Card(nil), cards...), nil
}

func (f *fakeRepository) SavePageToken(_ context.Context, code Code, token string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.savePageErr != nil {
		return f.savePageErr
	}
	f.pageTokens[code] = token
	return nil
}

func (f *fakeRepository) LoadPageToken(_ context.Context, code Code) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.loadPageErr != nil {
		return "", f.loadPageErr
	}
	return f.pageTokens[code], nil
}

func (f *fakeRepository) IncrClientCount(_ context.Context, code Code) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.incrErr != nil {
		return 0, f.incrErr
	}
	f.clientCount[code]++
	return f.clientCount[code], nil
}

func (f *fakeRepository) DecrClientCount(_ context.Context, code Code) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.decrErr != nil {
		return 0, f.decrErr
	}
	f.clientCount[code]--
	return f.clientCount[code], nil
}

func (f *fakeRepository) ClientCount(_ context.Context, code Code) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.clientCountErr != nil {
		return 0, f.clientCountErr
	}
	return f.clientCount[code], nil
}

func (f *fakeRepository) RegisterAcceptVote(_ context.Context, code Code, cardID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.voteErr != nil {
		return false, f.voteErr
	}
	if f.votes[code] == nil {
		f.votes[code] = make(map[string]int)
	}
	f.votes[code][cardID]++
	return f.majorityResult, nil
}

func (f *fakeRepository) PublishEvent(_ context.Context, evt Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.publishErr != nil {
		return f.publishErr
	}
	f.published = append(f.published, evt)
	return nil
}

func (f *fakeRepository) SubscribeEvents(_ context.Context) (EventSubscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.subscribeCalls++
	if f.subscribeErr != nil && f.subscribeCalls == 1 {
		return nil, f.subscribeErr
	}

	sub := newFakeEventSubscription()
	f.subscriptions = append(f.subscriptions, sub)
	return sub, nil
}

func (f *fakeRepository) subscriptionCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.subscriptions)
}

func (f *fakeRepository) subscription(i int) *fakeEventSubscription {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.subscriptions[i]
}

type fakeEventSubscription struct {
	events    chan Event
	closeOnce sync.Once
}

func newFakeEventSubscription() *fakeEventSubscription {
	return &fakeEventSubscription{events: make(chan Event, 8)}
}

func (f *fakeEventSubscription) Events() <-chan Event { return f.events }

func (f *fakeEventSubscription) Close() error {
	f.closeOnce.Do(func() { close(f.events) })
	return nil
}

type fakePlaces struct {
	mu            sync.Mutex
	cards         []Card
	nextPageToken string
	photoURL      string
	searchErr     error
	photoErr      error
}

func newFakePlaces() *fakePlaces { return &fakePlaces{} }

func (f *fakePlaces) Search(_ context.Context, _ SearchFilters, _ string) ([]Card, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.searchErr != nil {
		return nil, "", f.searchErr
	}
	return append([]Card(nil), f.cards...), f.nextPageToken, nil
}

func (f *fakePlaces) PhotoURL(_ context.Context, _ string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.photoErr != nil {
		return "", f.photoErr
	}
	return f.photoURL, nil
}

type majorityBroadcast struct {
	code   Code
	cardID string
}

type fakeBroadcaster struct {
	mu          sync.Mutex
	clients     map[Code][]Client
	started     []Code
	majority    []majorityBroadcast
	registerErr error
}

func newFakeBroadcaster() *fakeBroadcaster {
	return &fakeBroadcaster{clients: make(map[Code][]Client)}
}

func (b *fakeBroadcaster) Register(code Code, client Client) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.registerErr != nil {
		return b.registerErr
	}
	b.clients[code] = append(b.clients[code], client)
	return nil
}

func (b *fakeBroadcaster) Unregister(code Code, client Client) {
	b.mu.Lock()
	defer b.mu.Unlock()
	clients := b.clients[code]
	for i, c := range clients {
		if c == client {
			b.clients[code] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
}

func (b *fakeBroadcaster) BroadcastRoomStarted(code Code) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.started = append(b.started, code)
}

func (b *fakeBroadcaster) BroadcastMajorityFound(code Code, cardID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.majority = append(b.majority, majorityBroadcast{code: code, cardID: cardID})
}

func (b *fakeBroadcaster) registeredCount(code Code) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.clients[code])
}

func (b *fakeBroadcaster) startedCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.started)
}

func (b *fakeBroadcaster) majorityCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.majority)
}

type fakeClient struct {
	id string
}

func (c *fakeClient) ID() string                 { return c.id }
func (c *fakeClient) Send(_ OutboundEvent) error { return nil }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestService(repo *fakeRepository, places *fakePlaces, broadcaster *fakeBroadcaster) *roomService {
	svc := NewService(repo, places, broadcaster)
	rs, ok := svc.(*roomService)
	if !ok {
		panic("NewService did not return *roomService")
	}
	// Keep event-loop tests fast and deterministic.
	rs.reconnectDelay = time.Millisecond
	return rs
}

func assertErrorCode(t *testing.T, err error, want apperr.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected an error with code %q, got nil", want)
	}
	var aerr *apperr.Error
	if !errors.As(err, &aerr) {
		t.Fatalf("expected *apperr.Error, got %T: %v", err, err)
	}
	if aerr.Code != want {
		t.Fatalf("error code = %q, want %q (message: %s)", aerr.Code, want, aerr.Message)
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}

var roomCodePattern = regexp.MustCompile(`^\d{6}$`)

// ---------------------------------------------------------------------------
// CreateRoom
// ---------------------------------------------------------------------------

func TestCreateRoom(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*fakeRepository)
		wantErrCode apperr.Code
	}{
		{
			name:  "success",
			setup: func(_ *fakeRepository) {},
		},
		{
			name: "collision then success",
			setup: func(r *fakeRepository) {
				r.createCollisions = 5
			},
		},
		{
			name: "repository failure",
			setup: func(r *fakeRepository) {
				r.createErr = errors.New("redis unavailable")
			},
			wantErrCode: apperr.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			tt.setup(repo)
			svc := NewService(repo, newFakePlaces(), newFakeBroadcaster())

			code, err := svc.CreateRoom(context.Background())

			if tt.wantErrCode != "" {
				assertErrorCode(t, err, tt.wantErrCode)
				if code != "" {
					t.Fatalf("expected empty code on error, got %q", code)
				}
				return
			}

			if err != nil {
				t.Fatalf("CreateRoom returned unexpected error: %v", err)
			}
			if !roomCodePattern.MatchString(code.String()) {
				t.Fatalf("room code %q is not a 6-digit string", code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// JoinRoom
// ---------------------------------------------------------------------------

func TestJoinRoom(t *testing.T) {
	tests := []struct {
		name         string
		started      bool
		isStartedErr error
		wantJoined   bool
		wantSentinel error
		wantErrCode  apperr.Code
	}{
		{
			name:       "success",
			started:    false,
			wantJoined: true,
		},
		{
			name:         "already started",
			started:      true,
			wantJoined:   false,
			wantSentinel: ErrRoomAlreadyStarted,
		},
		{
			name:         "room not found",
			isStartedErr: ErrRoomNotFound,
			wantJoined:   false,
			wantSentinel: ErrRoomNotFound,
		},
		{
			name:         "repository failure",
			isStartedErr: errors.New("redis unavailable"),
			wantJoined:   false,
			wantErrCode:  apperr.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.started["123456"] = tt.started
			repo.isStartedErr = tt.isStartedErr

			svc := NewService(repo, newFakePlaces(), newFakeBroadcaster())
			joined, err := svc.JoinRoom(context.Background(), "123456")

			if joined != tt.wantJoined {
				t.Fatalf("joined = %v, want %v", joined, tt.wantJoined)
			}

			if tt.wantSentinel != nil {
				if !errors.Is(err, tt.wantSentinel) {
					t.Fatalf("error = %v, want sentinel %v", err, tt.wantSentinel)
				}
				return
			}
			if tt.wantErrCode != "" {
				assertErrorCode(t, err, tt.wantErrCode)
				return
			}
			if err != nil {
				t.Fatalf("JoinRoom returned unexpected error: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// StartRoom
// ---------------------------------------------------------------------------

func TestStartRoom(t *testing.T) {
	validFilters := SearchFilters{
		Latitude:  37.7749,
		Longitude: -122.4194,
		RadiusM:   500,
		MaxPrice:  2,
		Category:  CategoryRestaurant,
	}
	sampleCards := []Card{{ID: "p1", Title: "Place One"}}

	tests := []struct {
		name           string
		filters        SearchFilters
		cards          []Card
		nextToken      string
		searchErr      error
		saveCardsErr   error
		savePageErr    error
		markStartedErr error
		wantSentinel   error
		wantErrCode    apperr.Code
	}{
		{
			name:      "success",
			filters:   validFilters,
			cards:     sampleCards,
			nextToken: "next-token",
		},
		{
			name:         "invalid filters",
			filters:      SearchFilters{Latitude: 999},
			wantSentinel: ErrInvalidFilters,
		},
		{
			name:        "places upstream failure",
			filters:     validFilters,
			searchErr:   errors.New("places api network error"),
			wantErrCode: apperr.CodeUpstream,
		},
		{
			name:         "no places found",
			filters:      validFilters,
			cards:        nil,
			wantSentinel: ErrNoPlacesFound,
		},
		{
			name:         "save cards failure",
			filters:      validFilters,
			cards:        sampleCards,
			saveCardsErr: errors.New("redis unavailable"),
			wantErrCode:  apperr.CodeInternal,
		},
		{
			name:        "save page token failure",
			filters:     validFilters,
			cards:       sampleCards,
			savePageErr: errors.New("redis unavailable"),
			wantErrCode: apperr.CodeInternal,
		},
		{
			name:           "mark started failure",
			filters:        validFilters,
			cards:          sampleCards,
			markStartedErr: errors.New("redis unavailable"),
			wantErrCode:    apperr.CodeInternal,
		},
		{
			name:           "mark started room not found",
			filters:        validFilters,
			cards:          sampleCards,
			markStartedErr: ErrRoomNotFound,
			wantSentinel:   ErrRoomNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			places := newFakePlaces()
			places.cards = tt.cards
			places.nextPageToken = tt.nextToken
			places.searchErr = tt.searchErr

			repo.saveCardsErr = tt.saveCardsErr
			repo.savePageErr = tt.savePageErr
			repo.markStartedErr = tt.markStartedErr

			svc := NewService(repo, places, newFakeBroadcaster())
			err := svc.StartRoom(context.Background(), "123456", tt.filters)

			if tt.wantSentinel != nil {
				if !errors.Is(err, tt.wantSentinel) {
					t.Fatalf("error = %v, want sentinel %v", err, tt.wantSentinel)
				}
				return
			}
			if tt.wantErrCode != "" {
				assertErrorCode(t, err, tt.wantErrCode)
				return
			}
			if err != nil {
				t.Fatalf("StartRoom returned unexpected error: %v", err)
			}

			cards, _ := repo.LoadCards(context.Background(), "123456")
			if len(cards) != len(tt.cards) {
				t.Fatalf("saved cards count = %d, want %d", len(cards), len(tt.cards))
			}

			token, _ := repo.LoadPageToken(context.Background(), "123456")
			if token != tt.nextToken {
				t.Fatalf("saved page token = %q, want %q", token, tt.nextToken)
			}

			started, _ := repo.IsStarted(context.Background(), "123456")
			if !started {
				t.Fatal("room should be started after successful StartRoom")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Cards
// ---------------------------------------------------------------------------

func TestCards(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := newFakeRepository()
		repo.cards["123456"] = []Card{{ID: "p1"}, {ID: "p2"}}

		svc := NewService(repo, newFakePlaces(), newFakeBroadcaster())
		cards, err := svc.Cards(context.Background(), "123456")
		if err != nil {
			t.Fatalf("Cards returned unexpected error: %v", err)
		}
		if len(cards) != 2 {
			t.Fatalf("cards count = %d, want 2", len(cards))
		}
	})

	t.Run("repository failure", func(t *testing.T) {
		repo := newFakeRepository()
		repo.loadCardsErr = errors.New("redis unavailable")

		svc := NewService(repo, newFakePlaces(), newFakeBroadcaster())
		_, err := svc.Cards(context.Background(), "123456")
		assertErrorCode(t, err, apperr.CodeInternal)
	})
}

// ---------------------------------------------------------------------------
// PhotoURL
// ---------------------------------------------------------------------------

func TestPhotoURL(t *testing.T) {
	tests := []struct {
		name        string
		photoRef    string
		photoURL    string
		photoErr    error
		wantErrCode apperr.Code
	}{
		{
			name:     "success",
			photoRef: "photos/abc123",
			photoURL: "https://cdn.example/photo.jpg",
		},
		{
			name:        "missing photo reference",
			photoRef:    "",
			wantErrCode: apperr.CodeInvalid,
		},
		{
			name:        "places upstream failure",
			photoRef:    "photos/abc123",
			photoErr:    errors.New("places api network error"),
			wantErrCode: apperr.CodeUpstream,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			places := newFakePlaces()
			places.photoURL = tt.photoURL
			places.photoErr = tt.photoErr

			svc := NewService(newFakeRepository(), places, newFakeBroadcaster())
			url, err := svc.PhotoURL(context.Background(), tt.photoRef)

			if tt.wantErrCode != "" {
				assertErrorCode(t, err, tt.wantErrCode)
				return
			}
			if err != nil {
				t.Fatalf("PhotoURL returned unexpected error: %v", err)
			}
			if url != tt.photoURL {
				t.Fatalf("url = %q, want %q", url, tt.photoURL)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SubmitVote
// ---------------------------------------------------------------------------

func TestSubmitVote(t *testing.T) {
	t.Run("reject is silently ignored", func(t *testing.T) {
		repo := newFakeRepository()
		svc := NewService(repo, newFakePlaces(), newFakeBroadcaster())

		err := svc.SubmitVote(context.Background(), Vote{
			Room:   "123456",
			CardID: "card-1",
			Result: VoteReject,
		})
		if err != nil {
			t.Fatalf("SubmitVote returned unexpected error: %v", err)
		}
		if got := len(repo.votes["123456"]); got != 0 {
			t.Fatalf("reject vote should not be registered, got %d registrations", got)
		}
		if len(repo.published) != 0 {
			t.Fatalf("reject vote should not publish, got %d events", len(repo.published))
		}
	})

	t.Run("missing room is invalid", func(t *testing.T) {
		svc := NewService(newFakeRepository(), newFakePlaces(), newFakeBroadcaster())
		err := svc.SubmitVote(context.Background(), Vote{CardID: "card-1", Result: VoteAccept})
		if !errors.Is(err, ErrInvalidVote) {
			t.Fatalf("error = %v, want ErrInvalidVote", err)
		}
	})

	t.Run("missing card id is invalid", func(t *testing.T) {
		svc := NewService(newFakeRepository(), newFakePlaces(), newFakeBroadcaster())
		err := svc.SubmitVote(context.Background(), Vote{Room: "123456", Result: VoteAccept})
		if !errors.Is(err, ErrInvalidVote) {
			t.Fatalf("error = %v, want ErrInvalidVote", err)
		}
	})

	t.Run("accept without majority does not publish", func(t *testing.T) {
		repo := newFakeRepository()
		repo.majorityResult = false

		svc := NewService(repo, newFakePlaces(), newFakeBroadcaster())
		err := svc.SubmitVote(context.Background(), Vote{
			Room:   "123456",
			CardID: "card-1",
			Result: VoteAccept,
		})
		if err != nil {
			t.Fatalf("SubmitVote returned unexpected error: %v", err)
		}
		if len(repo.published) != 0 {
			t.Fatalf("no-majority vote should not publish, got %d events", len(repo.published))
		}
	})

	t.Run("accept with majority publishes majority event", func(t *testing.T) {
		repo := newFakeRepository()
		repo.majorityResult = true

		svc := NewService(repo, newFakePlaces(), newFakeBroadcaster())
		err := svc.SubmitVote(context.Background(), Vote{
			Room:   "123456",
			CardID: "card-1",
			Result: VoteAccept,
		})
		if err != nil {
			t.Fatalf("SubmitVote returned unexpected error: %v", err)
		}
		if len(repo.published) != 1 {
			t.Fatalf("published events = %d, want 1", len(repo.published))
		}
		evt := repo.published[0]
		if evt.Type != EventMajorityFound || evt.Room != "123456" || evt.CardID != "card-1" {
			t.Fatalf("unexpected published event: %+v", evt)
		}
	})

	t.Run("vote repository failure", func(t *testing.T) {
		repo := newFakeRepository()
		repo.voteErr = errors.New("redis unavailable")

		svc := NewService(repo, newFakePlaces(), newFakeBroadcaster())
		err := svc.SubmitVote(context.Background(), Vote{
			Room:   "123456",
			CardID: "card-1",
			Result: VoteAccept,
		})
		assertErrorCode(t, err, apperr.CodeInternal)
	})

	t.Run("publish failure is logged and swallowed", func(t *testing.T) {
		repo := newFakeRepository()
		repo.majorityResult = true
		repo.publishErr = errors.New("pubsub unavailable")

		svc := NewService(repo, newFakePlaces(), newFakeBroadcaster())
		err := svc.SubmitVote(context.Background(), Vote{
			Room:   "123456",
			CardID: "card-1",
			Result: VoteAccept,
		})
		if err != nil {
			t.Fatalf("vote was already recorded; publish failure should be swallowed, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// Connect
// ---------------------------------------------------------------------------

func TestConnect(t *testing.T) {
	tests := []struct {
		name           string
		roomExists     bool
		existsErr      error
		registerErr    error
		incrErr        error
		wantSentinel   error
		wantErrCode    apperr.Code
		wantRegistered bool
		wantCount      int64
	}{
		{
			name:           "success",
			roomExists:     true,
			wantRegistered: true,
			wantCount:      1,
		},
		{
			name:         "room not found",
			roomExists:   false,
			wantSentinel: ErrRoomNotFound,
		},
		{
			name:        "exists failure",
			roomExists:  true,
			existsErr:   errors.New("redis unavailable"),
			wantErrCode: apperr.CodeInternal,
		},
		{
			name:        "register failure",
			roomExists:  true,
			registerErr: errors.New("nil client"),
			wantErrCode: apperr.CodeInternal,
		},
		{
			name:        "increment failure unregisters client",
			roomExists:  true,
			incrErr:     errors.New("redis unavailable"),
			wantErrCode: apperr.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepository()
			repo.rooms["123456"] = tt.roomExists
			repo.existsErr = tt.existsErr
			repo.incrErr = tt.incrErr

			broadcaster := newFakeBroadcaster()
			broadcaster.registerErr = tt.registerErr

			client := &fakeClient{id: "conn-1"}
			svc := NewService(repo, newFakePlaces(), broadcaster)

			err := svc.Connect(context.Background(), "123456", client)

			if tt.wantSentinel != nil {
				if !errors.Is(err, tt.wantSentinel) {
					t.Fatalf("error = %v, want sentinel %v", err, tt.wantSentinel)
				}
				return
			}
			if tt.wantErrCode != "" {
				assertErrorCode(t, err, tt.wantErrCode)
				if tt.incrErr != nil && broadcaster.registeredCount("123456") != 0 {
					t.Fatal("client should be unregistered after failed client-count increment")
				}
				return
			}
			if err != nil {
				t.Fatalf("Connect returned unexpected error: %v", err)
			}

			if got := broadcaster.registeredCount("123456"); got != 1 {
				t.Fatalf("registered clients = %d, want 1", got)
			}
			if got := repo.clientCount["123456"]; got != tt.wantCount {
				t.Fatalf("client count = %d, want %d", got, tt.wantCount)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Disconnect
// ---------------------------------------------------------------------------

func TestDisconnect(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := newFakeRepository()
		repo.rooms["123456"] = true
		repo.clientCount["123456"] = 1

		broadcaster := newFakeBroadcaster()
		client := &fakeClient{id: "conn-1"}
		if err := broadcaster.Register("123456", client); err != nil {
			t.Fatalf("fake Register: %v", err)
		}

		svc := NewService(repo, newFakePlaces(), broadcaster)
		svc.Disconnect(context.Background(), "123456", client)

		if got := broadcaster.registeredCount("123456"); got != 0 {
			t.Fatalf("registered clients after disconnect = %d, want 0", got)
		}
		if got := repo.clientCount["123456"]; got != 0 {
			t.Fatalf("client count after disconnect = %d, want 0", got)
		}
	})

	t.Run("decrement failure is swallowed", func(t *testing.T) {
		repo := newFakeRepository()
		repo.decrErr = errors.New("redis unavailable")

		broadcaster := newFakeBroadcaster()
		client := &fakeClient{id: "conn-1"}
		if err := broadcaster.Register("123456", client); err != nil {
			t.Fatalf("fake Register: %v", err)
		}

		svc := NewService(repo, newFakePlaces(), broadcaster)
		svc.Disconnect(context.Background(), "123456", client)

		if got := broadcaster.registeredCount("123456"); got != 0 {
			t.Fatalf("registered clients after disconnect = %d, want 0", got)
		}
	})
}

// ---------------------------------------------------------------------------
// StartEventListener
// ---------------------------------------------------------------------------

func TestStartEventListener(t *testing.T) {
	t.Run("broadcasts room started event", func(t *testing.T) {
		repo := newFakeRepository()
		broadcaster := newFakeBroadcaster()
		svc := newTestService(repo, newFakePlaces(), broadcaster)

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			defer close(done)
			svc.StartEventListener(ctx)
		}()
		t.Cleanup(func() {
			cancel()
			<-done
		})

		waitFor(t, func() bool { return repo.subscriptionCount() == 1 })
		repo.subscription(0).events <- Event{Type: EventRoomStarted, Room: "123456"}

		waitFor(t, func() bool { return broadcaster.startedCount() == 1 })

		cancel()
		<-done
	})

	t.Run("broadcasts majority found event", func(t *testing.T) {
		repo := newFakeRepository()
		broadcaster := newFakeBroadcaster()
		svc := newTestService(repo, newFakePlaces(), broadcaster)

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			defer close(done)
			svc.StartEventListener(ctx)
		}()
		t.Cleanup(func() {
			cancel()
			<-done
		})

		waitFor(t, func() bool { return repo.subscriptionCount() == 1 })
		repo.subscription(0).events <- Event{Type: EventMajorityFound, Room: "123456", CardID: "card-9"}

		waitFor(t, func() bool { return broadcaster.majorityCount() == 1 })

		cancel()
		<-done
	})

	t.Run("ignores unknown events", func(t *testing.T) {
		repo := newFakeRepository()
		broadcaster := newFakeBroadcaster()
		svc := newTestService(repo, newFakePlaces(), broadcaster)

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			defer close(done)
			svc.StartEventListener(ctx)
		}()
		t.Cleanup(func() {
			cancel()
			<-done
		})

		waitFor(t, func() bool { return repo.subscriptionCount() == 1 })
		sub := repo.subscription(0)

		// Channel is FIFO: the unknown event is processed before the known one.
		sub.events <- Event{Type: EventType("bogus"), Room: "123456"}
		sub.events <- Event{Type: EventRoomStarted, Room: "123456"}

		waitFor(t, func() bool { return broadcaster.startedCount() == 1 })

		if broadcaster.majorityCount() != 0 {
			t.Fatal("unknown event should not trigger any broadcast")
		}

		cancel()
		<-done
	})

	t.Run("exits when context is canceled before subscribing", func(t *testing.T) {
		repo := newFakeRepository()
		repo.subscribeErr = errors.New("redis unavailable")
		broadcaster := newFakeBroadcaster()
		svc := newTestService(repo, newFakePlaces(), broadcaster)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel before the listener starts

		done := make(chan struct{})
		go func() {
			defer close(done)
			svc.StartEventListener(ctx)
		}()

		waitFor(t, func() bool {
			select {
			case <-done:
				return true
			default:
				return false
			}
		})
	})

	t.Run("retries after subscribe failure", func(t *testing.T) {
		repo := newFakeRepository()
		repo.subscribeErr = errors.New("redis unavailable")
		broadcaster := newFakeBroadcaster()
		svc := newTestService(repo, newFakePlaces(), broadcaster)

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			defer close(done)
			svc.StartEventListener(ctx)
		}()
		t.Cleanup(func() {
			cancel()
			<-done
		})

		// First SubscribeEvents call fails; after reconnectDelay the second
		// call succeeds and an event should be delivered.
		waitFor(t, func() bool { return repo.subscriptionCount() == 1 })
		repo.subscription(0).events <- Event{Type: EventRoomStarted, Room: "123456"}

		waitFor(t, func() bool { return broadcaster.startedCount() == 1 })

		cancel()
		<-done
	})
}
