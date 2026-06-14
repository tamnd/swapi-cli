package swapi

// Person is a character from the Star Wars films.
type Person struct {
	Name      string `json:"name"`
	Height    string `json:"height"`
	Mass      string `json:"mass"`
	HairColor string `json:"hair_color"`
	BirthYear string `json:"birth_year"`
	Gender    string `json:"gender"`
	URL       string `json:"url"`
}

// Film is one of the 7 canonical Star Wars films.
type Film struct {
	Title       string `json:"title"`
	EpisodeID   int    `json:"episode_id"`
	Director    string `json:"director"`
	Producer    string `json:"producer"`
	ReleaseDate string `json:"release_date"`
	URL         string `json:"url"`
}

// Planet is a planet in the Star Wars universe.
type Planet struct {
	Name       string `json:"name"`
	Climate    string `json:"climate"`
	Terrain    string `json:"terrain"`
	Population string `json:"population"`
	URL        string `json:"url"`
}

// Starship is a starship used in the Star Wars films.
type Starship struct {
	Name             string `json:"name"`
	Model            string `json:"model"`
	Manufacturer     string `json:"manufacturer"`
	StarshipClass    string `json:"starship_class"`
	HyperdriveRating string `json:"hyperdrive_rating"`
	URL              string `json:"url"`
}

// Vehicle is a ground or air vehicle in the Star Wars universe.
type Vehicle struct {
	Name         string `json:"name"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	VehicleClass string `json:"vehicle_class"`
	URL          string `json:"url"`
}

// Species is a species or race in the Star Wars universe.
type Species struct {
	Name            string `json:"name"`
	Classification  string `json:"classification"`
	Language        string `json:"language"`
	AverageLifespan string `json:"average_lifespan"`
	URL             string `json:"url"`
}
