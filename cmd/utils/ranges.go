/**
 * BTCGO
 *
 * Modulo : Ranges
 */

package utils

import (
	"encoding/json"
	"errors"
	"os"
)

// Range struct to hold the minimum, maximum, and status
type ranges struct {
	Min    string `json:"min"`
	Max    string `json:"max"`
	Status int    `json:"status"`
}

// Ranges struct to hold an array of ranges
type Range struct {
	FileName   string
	DataRanges map[int]ranges
}

// Cria uma instancia
func NewRanges(filename string) *Range {
	return &Range{
		FileName:   filename,
		DataRanges: make(map[int]ranges),
	}
}

// Ler os ranges para memoria
func (r *Range) Load() error {
	data, err := os.ReadFile(r.FileName)
	if err != nil {
		return err
	}

	// Temp Struct
	var xTmpRanges struct {
		Rang []ranges `json:"ranges"`
	}
	if err := json.Unmarshal(data, &xTmpRanges); err != nil {
		return err
	}
	for i, rgd := range xTmpRanges.Rang {
		r.DataRanges[i+1] = rgd
	}

	return nil
}

// Get range data
func (r Range) Get(rangeid int) (ranges, error) {
	if _, ok := r.DataRanges[rangeid]; ok {
		return r.DataRanges[rangeid], nil
	}
	return ranges{}, errors.New("range not found")
}

// Get Min
func (r Range) GetMin(rangeid int) (string, error) {
	if _, ok := r.DataRanges[rangeid]; ok {
		return r.DataRanges[rangeid].Min, nil
	}
	return "", errors.New("range not found")
}

// Get Max
func (r Range) GetMax(rangeid int) (string, error) {
	if _, ok := r.DataRanges[rangeid]; ok {
		return r.DataRanges[rangeid].Max, nil
	}
	return "", errors.New("range not found")
}

// Get Status
func (r Range) GetStatus(rangeid int) (int, error) {
	if _, ok := r.DataRanges[rangeid]; ok {
		return r.DataRanges[rangeid].Status, nil
	}
	return -1, errors.New("range not found")
}

// Get Status
func (r Range) Count() int {
	return len(r.DataRanges)
}
