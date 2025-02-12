package postgres

import (
	"fmt"
	"regexp"
	"strconv"
)

// Point represents a 2D point with X and Y coordinates.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// UnmarshalJSON custom unmarshals a POINT from Supabase ("(x,y)") to Point struct.
func (p *Point) UnmarshalJSON(data []byte) error {
	// Convert bytes to string and remove quotes
	raw := string(data)
	if len(raw) < 5 { // Minimal valid string: "(0,0)"
		return fmt.Errorf("invalid POINT format")
	}

	// Use regex to extract x and y values
	re := regexp.MustCompile(`\(([-\d.]+),([-\d.]+)\)`)
	matches := re.FindStringSubmatch(raw)

	if len(matches) != 3 {
		return fmt.Errorf("invalid POINT format: %s", raw)
	}

	// Convert extracted values to float
	x, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return err
	}

	y, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return err
	}

	// Assign values to struct
	p.X = x
	p.Y = y
	return nil
}

// MarshalJSON converts a Point struct into Supabase's POINT format "(x,y)"
func (p Point) MarshalJSON() ([]byte, error) {
	// Format as "(x,y)" string
	pointStr := fmt.Sprintf(`"(%f,%f)"`, p.X, p.Y)
	return []byte(pointStr), nil
}
