package swapi

import (
	"context"
	"strings"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

func init() { kit.Register(Domain{}) }

// Domain is the swapi kit driver. It carries no state; the per-run client is
// built by the factory Register hands to kit.
type Domain struct{}

// Info describes the scheme, hostnames, and the identity used in help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "swapi",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "swapi",
			Short:  "A command line for the Star Wars API.",
			Long: `A command line for the Star Wars API.

swapi reads public Star Wars data from swapi.py4e.com over plain HTTPS,
shapes it into clean records, and prints output that pipes into the rest
of your tools. No API key, nothing to run alongside it.

Browse people, films, planets, starships, vehicles, and species from the
Star Wars universe.`,
			Site: Host,
			Repo: "https://github.com/tamnd/swapi-cli",
		},
	}
}

// Register installs the client factory and all six resource operations onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{Name: "people", Group: "read", List: true,
		Summary: "List Star Wars characters"}, peopleOp)

	kit.Handle(app, kit.OpMeta{Name: "films", Group: "read", List: true,
		Summary: "List all canonical Star Wars films"}, filmsOp)

	kit.Handle(app, kit.OpMeta{Name: "planets", Group: "read", List: true,
		Summary: "List Star Wars planets"}, planetsOp)

	kit.Handle(app, kit.OpMeta{Name: "starships", Group: "read", List: true,
		Summary: "List starships from the Star Wars films"}, starshipsOp)

	kit.Handle(app, kit.OpMeta{Name: "vehicles", Group: "read", List: true,
		Summary: "List vehicles from the Star Wars films"}, vehiclesOp)

	kit.Handle(app, kit.OpMeta{Name: "species", Group: "read", List: true,
		Summary: "List species and races from Star Wars"}, speciesOp)
}

// newClient builds the client from the host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- input structs ---

type peopleInput struct {
	Search string  `kit:"flag" help:"filter by character name"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type filmsInput struct {
	Client *Client `kit:"inject"`
}

type planetsInput struct {
	Search string  `kit:"flag" help:"filter by planet name"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type starshipsInput struct {
	Search string  `kit:"flag" help:"filter by name or model"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type vehiclesInput struct {
	Search string  `kit:"flag" help:"filter by name or model"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type speciesInput struct {
	Search string  `kit:"flag" help:"filter by species name"`
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func peopleOp(ctx context.Context, in peopleInput, emit func(Person) error) error {
	items, err := in.Client.People(ctx, in.Search, in.Limit)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func filmsOp(ctx context.Context, in filmsInput, emit func(Film) error) error {
	items, err := in.Client.Films(ctx)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func planetsOp(ctx context.Context, in planetsInput, emit func(Planet) error) error {
	items, err := in.Client.Planets(ctx, in.Search, in.Limit)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func starshipsOp(ctx context.Context, in starshipsInput, emit func(Starship) error) error {
	items, err := in.Client.Starships(ctx, in.Search, in.Limit)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func vehiclesOp(ctx context.Context, in vehiclesInput, emit func(Vehicle) error) error {
	items, err := in.Client.Vehicles(ctx, in.Search, in.Limit)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func speciesOp(ctx context.Context, in speciesInput, emit func(Species) error) error {
	items, err := in.Client.Species(ctx, in.Search, in.Limit)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

// Classify turns a SWAPI URL or resource path into (type, id).
func (Domain) Classify(input string) (string, string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", errs.Usage("empty swapi reference")
	}
	return "person", input, nil
}

// Locate returns the live API URL for a (type, id).
func (Domain) Locate(t, id string) (string, error) {
	return BaseURL + "/api/" + t + "/" + id, nil
}
