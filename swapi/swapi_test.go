package swapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tamnd/swapi-cli/swapi"
)

func newTestClient(ts *httptest.Server) *swapi.Client {
	cfg := swapi.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return swapi.NewClient(cfg)
}

func TestPeopleSendsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua == "" {
			t.Error("request carried no User-Agent")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"count":0,"next":null,"previous":null,"results":[]}`)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.People(context.Background(), "", 10)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPeopleParsesItems(t *testing.T) {
	payload := `{
		"count": 2,
		"next": null,
		"previous": null,
		"results": [
			{"name":"Luke Skywalker","height":"172","mass":"77","hair_color":"blond","birth_year":"19BBY","gender":"male","url":"https://swapi.py4e.com/api/people/1/"},
			{"name":"Darth Vader","height":"202","mass":"136","hair_color":"none","birth_year":"41.9BBY","gender":"male","url":"https://swapi.py4e.com/api/people/4/"}
		]
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, payload)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	items, err := c.People(context.Background(), "", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Name != "Luke Skywalker" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "Luke Skywalker")
	}
	if items[0].Height != "172" {
		t.Errorf("items[0].Height = %q, want %q", items[0].Height, "172")
	}
	if items[0].Gender != "male" {
		t.Errorf("items[0].Gender = %q, want %q", items[0].Gender, "male")
	}
	if items[1].Name != "Darth Vader" {
		t.Errorf("items[1].Name = %q, want %q", items[1].Name, "Darth Vader")
	}
}

func TestPeopleLimit(t *testing.T) {
	payload := `{
		"count": 2,
		"next": null,
		"previous": null,
		"results": [
			{"name":"Luke Skywalker","height":"172","mass":"77","hair_color":"blond","birth_year":"19BBY","gender":"male","url":"https://swapi.py4e.com/api/people/1/"},
			{"name":"Darth Vader","height":"202","mass":"136","hair_color":"none","birth_year":"41.9BBY","gender":"male","url":"https://swapi.py4e.com/api/people/4/"}
		]
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, payload)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	items, err := c.People(context.Background(), "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d, want 1 (limit respected)", len(items))
	}
}

func TestPeopleRetriesOn503(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"count":0,"next":null,"previous":null,"results":[]}`)
	}))
	defer srv.Close()

	cfg := swapi.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := swapi.NewClient(cfg)

	start := time.Now()
	_, err := c.People(context.Background(), "", 10)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestFilmsParsesItems(t *testing.T) {
	payload := `{
		"count": 2,
		"next": null,
		"previous": null,
		"results": [
			{"title":"A New Hope","episode_id":4,"director":"George Lucas","producer":"Gary Kurtz, Rick McCallum","release_date":"1977-05-25","url":"https://swapi.py4e.com/api/films/1/"},
			{"title":"The Empire Strikes Back","episode_id":5,"director":"Irvin Kershner","producer":"Gary Kurtz, Rick McCallum","release_date":"1980-05-17","url":"https://swapi.py4e.com/api/films/2/"}
		]
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, payload)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	items, err := c.Films(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Title != "A New Hope" {
		t.Errorf("items[0].Title = %q, want %q", items[0].Title, "A New Hope")
	}
	if items[0].EpisodeID != 4 {
		t.Errorf("items[0].EpisodeID = %d, want 4", items[0].EpisodeID)
	}
	if items[0].Director != "George Lucas" {
		t.Errorf("items[0].Director = %q, want %q", items[0].Director, "George Lucas")
	}
}

func TestPlanetsParseItems(t *testing.T) {
	payload := `{
		"count": 1,
		"next": null,
		"previous": null,
		"results": [
			{"name":"Tatooine","climate":"arid","terrain":"desert","population":"200000","url":"https://swapi.py4e.com/api/planets/1/"}
		]
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, payload)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	items, err := c.Planets(context.Background(), "tatooine", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].Name != "Tatooine" {
		t.Errorf("items[0].Name = %q, want %q", items[0].Name, "Tatooine")
	}
	if items[0].Climate != "arid" {
		t.Errorf("items[0].Climate = %q, want %q", items[0].Climate, "arid")
	}
}

// Verify JSON round-trip for Starship type.
func TestStarshipJSONRoundtrip(t *testing.T) {
	raw := `{"name":"Millennium Falcon","model":"YT-1300 light freighter","manufacturer":"Corellian Engineering Corporation","starship_class":"Light freighter","hyperdrive_rating":"0.5","url":"https://swapi.py4e.com/api/starships/10/"}`
	var s swapi.Starship
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if s.Name != "Millennium Falcon" {
		t.Errorf("Name = %q", s.Name)
	}
	if s.HyperdriveRating != "0.5" {
		t.Errorf("HyperdriveRating = %q", s.HyperdriveRating)
	}
}
