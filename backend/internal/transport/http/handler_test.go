package http

import (
	"context"
	"errors"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/domain/room"
)

// stubService is a hand-written fake implementing room.Service. Each test can
// override only the methods it needs; default methods return safe zero values.
type stubService struct {
	createRoom func(ctx context.Context) (room.Code, error)
	joinRoom   func(ctx context.Context, code room.Code) (bool, error)
	startRoom  func(ctx context.Context, code room.Code, filters room.SearchFilters) error
	cards      func(ctx context.Context, code room.Code) ([]room.Card, error)
	photoURL   func(ctx context.Context, photoRef string) (string, error)
	submitVote func(ctx context.Context, vote room.Vote) error
	connect    func(ctx context.Context, code room.Code, client room.Client) error
	disconnect func(ctx context.Context, code room.Code, client room.Client)
}

func (s *stubService) CreateRoom(ctx context.Context) (room.Code, error) {
	if s.createRoom != nil {
		return s.createRoom(ctx)
	}
	return "", nil
}

func (s *stubService) JoinRoom(ctx context.Context, code room.Code) (bool, error) {
	if s.joinRoom != nil {
		return s.joinRoom(ctx, code)
	}
	return false, nil
}

func (s *stubService) StartRoom(ctx context.Context, code room.Code, filters room.SearchFilters) error {
	if s.startRoom != nil {
		return s.startRoom(ctx, code, filters)
	}
	return nil
}

func (s *stubService) Cards(ctx context.Context, code room.Code) ([]room.Card, error) {
	if s.cards != nil {
		return s.cards(ctx, code)
	}
	return nil, nil
}

func (s *stubService) PhotoURL(ctx context.Context, photoRef string) (string, error) {
	if s.photoURL != nil {
		return s.photoURL(ctx, photoRef)
	}
	return "", nil
}

func (s *stubService) SubmitVote(ctx context.Context, vote room.Vote) error {
	if s.submitVote != nil {
		return s.submitVote(ctx, vote)
	}
	return nil
}

func (s *stubService) Connect(ctx context.Context, code room.Code, client room.Client) error {
	if s.connect != nil {
		return s.connect(ctx, code, client)
	}
	return nil
}

func (s *stubService) Disconnect(ctx context.Context, code room.Code, client room.Client) {
	if s.disconnect != nil {
		s.disconnect(ctx, code, client)
	}
}

func newTestHandler(svc room.Service) *Handler {
	return NewHandler(svc, nil)
}

func doRequest(t *testing.T, h *Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	h.Routes().ServeHTTP(rr, req)
	return rr
}

func bodyOf(rr *httptest.ResponseRecorder) string {
	return rr.Body.String()
}

func TestCreateRoom_ByteForByteJSONParity(t *testing.T) {
	svc := &stubService{
		createRoom: func(ctx context.Context) (room.Code, error) {
			return room.Code("123456"), nil
		},
	}
	h := newTestHandler(svc)

	rr := doRequest(t, h, nethttp.MethodGet, "/rooms", "")
	if rr.Code != nethttp.StatusOK {
		t.Fatalf("status: want %d, got %d", nethttp.StatusOK, rr.Code)
	}

	want := "{\"Header\":\"Room Code\",\"Body\":\"123456\"}\n"
	if got := bodyOf(rr); got != want {
		t.Fatalf("body mismatch\nwant: %q\n got: %q", want, got)
	}
}

func TestCreateRoom_ErrorEnvelope(t *testing.T) {
	svc := &stubService{
		createRoom: func(ctx context.Context) (room.Code, error) {
			return "", errors.New("redis down")
		},
	}
	h := newTestHandler(svc)

	rr := doRequest(t, h, nethttp.MethodGet, "/rooms", "")
	if rr.Code != nethttp.StatusInternalServerError {
		t.Fatalf("status: want %d, got %d", nethttp.StatusInternalServerError, rr.Code)
	}

	want := "{\"Header\":\"Room Code Error\",\"Body\":\"internal server error\",\"Message\":\"\"}\n"
	if got := bodyOf(rr); got != want {
		t.Fatalf("body mismatch\nwant: %q\n got: %q", want, got)
	}
}

func TestJoinRoom_Success(t *testing.T) {
	svc := &stubService{
		joinRoom: func(ctx context.Context, code room.Code) (bool, error) {
			if code != "123456" {
				t.Fatalf("unexpected code: %q", code)
			}
			return true, nil
		},
	}
	h := newTestHandler(svc)

	rr := doRequest(t, h, nethttp.MethodPost, "/rooms/123456/join", "")
	if rr.Code != nethttp.StatusOK {
		t.Fatalf("status: want %d, got %d", nethttp.StatusOK, rr.Code)
	}

	want := "{\"Header\":\"Join Status\",\"Body\":\"true\"}\n"
	if got := bodyOf(rr); got != want {
		t.Fatalf("body mismatch\nwant: %q\n got: %q", want, got)
	}
}

func TestJoinRoom_AlreadyStarted_Conflict409(t *testing.T) {
	svc := &stubService{
		joinRoom: func(ctx context.Context, code room.Code) (bool, error) {
			return false, room.ErrRoomAlreadyStarted
		},
	}
	h := newTestHandler(svc)

	rr := doRequest(t, h, nethttp.MethodPost, "/rooms/123456/join", "")
	if rr.Code != nethttp.StatusConflict {
		t.Fatalf("status: want %d, got %d", nethttp.StatusConflict, rr.Code)
	}

	want := "{\"Header\":\"Join Room Error\",\"Body\":\"room already started\",\"Message\":\"\"}\n"
	if got := bodyOf(rr); got != want {
		t.Fatalf("body mismatch\nwant: %q\n got: %q", want, got)
	}
}

func TestJoinRoom_MalformedCode(t *testing.T) {
	h := newTestHandler(&stubService{})

	rr := doRequest(t, h, nethttp.MethodPost, "/rooms/not-a-code/join", "")
	if rr.Code != nethttp.StatusBadRequest {
		t.Fatalf("status: want %d, got %d", nethttp.StatusBadRequest, rr.Code)
	}

	want := "{\"Header\":\"Join Room Error\",\"Body\":\"room code must be exactly 6 digits\",\"Message\":\"\"}\n"
	if got := bodyOf(rr); got != want {
		t.Fatalf("body mismatch\nwant: %q\n got: %q", want, got)
	}
}

func TestStartRoom_Success(t *testing.T) {
	svc := &stubService{
		startRoom: func(ctx context.Context, code room.Code, filters room.SearchFilters) error {
			if code != "123456" {
				t.Fatalf("unexpected code: %q", code)
			}
			if filters.RadiusM != 5000 {
				t.Fatalf("unexpected radius: %d", filters.RadiusM)
			}
			return nil
		},
	}
	h := newTestHandler(svc)

	payload := `{"filters":{"latitude":1.3521,"longitude":103.8198,"radius":5000,"maxPrice":4,"category":"restaurant","openNow":true}}`
	rr := doRequest(t, h, nethttp.MethodPost, "/rooms/123456/start", payload)
	if rr.Code != nethttp.StatusOK {
		t.Fatalf("status: want %d, got %d", nethttp.StatusOK, rr.Code)
	}

	want := "{\"Header\":\"Start Status\",\"Body\":\"true\"}\n"
	if got := bodyOf(rr); got != want {
		t.Fatalf("body mismatch\nwant: %q\n got: %q", want, got)
	}
}

func TestStartRoom_NoPlacesFound_NotFound404(t *testing.T) {
	svc := &stubService{
		startRoom: func(ctx context.Context, code room.Code, filters room.SearchFilters) error {
			return room.ErrNoPlacesFound
		},
	}
	h := newTestHandler(svc)

	payload := `{"filters":{"latitude":0,"longitude":0,"radius":1000,"maxPrice":4,"category":"restaurant","openNow":false}}`
	rr := doRequest(t, h, nethttp.MethodPost, "/rooms/123456/start", payload)
	if rr.Code != nethttp.StatusNotFound {
		t.Fatalf("status: want %d, got %d", nethttp.StatusNotFound, rr.Code)
	}

	want := "{\"Header\":\"Start Room Error\",\"Body\":\"no places found within the specified area\",\"Message\":\"\"}\n"
	if got := bodyOf(rr); got != want {
		t.Fatalf("body mismatch\nwant: %q\n got: %q", want, got)
	}
}

func TestStartRoom_InvalidJSON(t *testing.T) {
	h := newTestHandler(&stubService{})

	rr := doRequest(t, h, nethttp.MethodPost, "/rooms/123456/start", `{"filters":`)
	if rr.Code != nethttp.StatusBadRequest {
		t.Fatalf("status: want %d, got %d", nethttp.StatusBadRequest, rr.Code)
	}

	want := "{\"Header\":\"Start Room Error\",\"Body\":\"invalid JSON payload\",\"Message\":\"\"}\n"
	if got := bodyOf(rr); got != want {
		t.Fatalf("body mismatch\nwant: %q\n got: %q", want, got)
	}
}

func TestGetCards_EmptyPreservesLegacyShape(t *testing.T) {
	svc := &stubService{
		cards: func(ctx context.Context, code room.Code) ([]room.Card, error) {
			return nil, nil
		},
	}
	h := newTestHandler(svc)

	rr := doRequest(t, h, nethttp.MethodGet, "/rooms/123456/cards", "")
	if rr.Code != nethttp.StatusOK {
		t.Fatalf("status: want %d, got %d", nethttp.StatusOK, rr.Code)
	}

	// util.Message has `json:"Cards,omitempty"` and an empty slice is omitted
	// by encoding/json, so the legacy empty-card shape has no Cards key.
	want := "{\"Header\":\"CARD_DATA\"}\n"
	if got := bodyOf(rr); got != want {
		t.Fatalf("body mismatch\nwant: %q\n got: %q", want, got)
	}
}

func TestGetCards_NonEmpty(t *testing.T) {
	svc := &stubService{
		cards: func(ctx context.Context, code room.Code) ([]room.Card, error) {
			return []room.Card{
				{
					ID:          "place_1",
					Title:       "Test Place",
					Category:    "restaurant",
					PriceLevel:  "$$",
					Rating:      4.3,
					ReviewCount: 120,
					OpenNow:     true,
					Summary:     "A nice place",
					Address:     "123 Test St",
					PhotoRef:    "photo_ref_1",
				},
			}, nil
		},
	}
	h := newTestHandler(svc)

	rr := doRequest(t, h, nethttp.MethodGet, "/rooms/123456/cards", "")
	if rr.Code != nethttp.StatusOK {
		t.Fatalf("status: want %d, got %d", nethttp.StatusOK, rr.Code)
	}

	want := "{\"Header\":\"CARD_DATA\",\"Cards\":[{\"id\":\"place_1\",\"title\":\"Test Place\",\"category\":\"restaurant\",\"priceLevel\":\"$$\",\"rating\":4.3,\"reviewCount\":120,\"openNow\":true,\"summary\":\"A nice place\",\"address\":\"123 Test St\",\"photoName\":\"photo_ref_1\"}]}\n"
	if got := bodyOf(rr); got != want {
		t.Fatalf("body mismatch\nwant: %q\n got: %q", want, got)
	}
}

func TestGetPhoto_Redirect(t *testing.T) {
	svc := &stubService{
		photoURL: func(ctx context.Context, photoRef string) (string, error) {
			if photoRef != "photo_ref_1" {
				t.Fatalf("unexpected photo ref: %q", photoRef)
			}
			return "https://example.com/photo.jpg", nil
		},
	}
	h := newTestHandler(svc)

	rr := doRequest(t, h, nethttp.MethodGet, "/image?photoName=photo_ref_1", "")
	if rr.Code != nethttp.StatusFound {
		t.Fatalf("status: want %d, got %d", nethttp.StatusFound, rr.Code)
	}
	if got := rr.Header().Get("Location"); got != "https://example.com/photo.jpg" {
		t.Fatalf("Location: want %q, got %q", "https://example.com/photo.jpg", got)
	}
}

func TestGetPhoto_MissingParam(t *testing.T) {
	h := newTestHandler(&stubService{})

	rr := doRequest(t, h, nethttp.MethodGet, "/image", "")
	if rr.Code != nethttp.StatusBadRequest {
		t.Fatalf("status: want %d, got %d", nethttp.StatusBadRequest, rr.Code)
	}

	want := "{\"Header\":\"Image Error\",\"Body\":\"missing photo name parameter\",\"Message\":\"\"}\n"
	if got := bodyOf(rr); got != want {
		t.Fatalf("body mismatch\nwant: %q\n got: %q", want, got)
	}
}
