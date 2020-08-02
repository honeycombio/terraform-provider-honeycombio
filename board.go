package honeycombio

// Compile-time proof of interface implementation.
var _ Boards = (*boards)(nil)

// Boards describes all the boards related methods that Honeycomb supports.
type Boards interface {
	// List all boards present in this dataset.
	List() ([]Board, error)

	// Get a board by its ID. Returns nil, ErrNotFound if there is no board
	// with the given ID.
	Get(id string) (*Board, error)

	// Create a new board.
	Create(data Board) (*Board, error)
}

// boards implements Boards.
type boards struct {
	client *Client
}

// Board represents a Honeycomb board, as described by https://docs.honeycomb.io/api/boards-api/#fields-on-a-board
type Board struct {
	ID          string       `json:"id,omitempty"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Style       string       `json:"style"`
	Queries     []BoardQuery `json:"queries"`
}

type BoardQuery struct {
	Caption string `json:"caption,omitempty"`
	Dataset string `json:"dataset"`
	Query   QuerySpec  `json:"query"`
}

func (s *boards) List() ([]Board, error) {
	req, err := s.client.newRequest("GET", "/1/boards", nil)
	if err != nil {
		return nil, err
	}

	var b []Board
	err = s.client.do(req, &b)
	return b, err
}

func (s *boards) Get(ID string) (*Board, error) {
	req, err := s.client.newRequest("GET", "/1/boards/"+ID, nil)
	if err != nil {
		return nil, err
	}

	var b Board
	err = s.client.do(req, &b)
	return &b, err
}

func (s *boards) Create(b Board) (*Board, error) {
	req, err := s.client.newRequest("POST", "/1/boards", b)
	if err != nil {
		return nil, err
	}

	err = s.client.do(req, &b)
	return &b, err
}
