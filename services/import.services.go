package services

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

type ImportCandidate struct {
	Index       int             `json:"index"`
	Type        string          `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Details     string          `json:"details"`
	Geojson     json.RawMessage `json:"geojson"`
	IsPublic    bool            `json:"is_public"`
}

func ParseImportFile(path, fileName string) ([]ImportCandidate, error) {
	name := strings.ToLower(fileName)

	isSQLite, isGPX, err := detectFormat(path, name)
	if err != nil {
		return nil, err
	}

	switch {
	case isSQLite:
		return parseSQLite(path)
	case isGPX:
		return parseGPX(path)
	}

	return nil, errors.New("unsupported file format")
}

func detectFormat(path, name string) (isSQLite, isGPX bool, err error) {
	switch {
	case strings.HasSuffix(name, ".gpx"):
		return false, true, nil
	case strings.HasSuffix(name, ".db"),
		strings.HasSuffix(name, ".sqlite"),
		strings.HasSuffix(name, ".sqlite3"):
		return true, false, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return false, false, err
	}
	defer f.Close()

	head := make([]byte, 512)
	n, _ := io.ReadFull(f, head)
	prefix := string(head[:n])

	if strings.HasPrefix(prefix, "SQLite format 3") {
		return true, false, nil
	}

	if strings.Contains(strings.ToLower(prefix), "<gpx") {
		return false, true, nil
	}

	return false, false, errors.New("unrecognized file format (expected a GPX file or SQLite database)")
}

func parseSQLite(path string) ([]ImportCandidate, error) {
	dbt, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	defer dbt.Close()

	if err := dbt.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	rows, err := dbt.Query(`SELECT p.title, p.description, p.latitude, p.longitude, COALESCE(g.title, '')
		FROM place p
		LEFT JOIN item_group g ON p.group_id = g._id
		WHERE p.deleted = 0 AND p.latitude IS NOT NULL AND p.longitude IS NOT NULL
		ORDER BY p.title`)
	if err != nil {
		return nil, fmt.Errorf("query place: %w", err)
	}
	defer rows.Close()

	var candidates []ImportCandidate
	for rows.Next() {
		var title, description, groupTitle string
		var lat, lon float64
		if err := rows.Scan(&title, &description, &lat, &lon, &groupTitle); err != nil {
			return nil, fmt.Errorf("scan place: %w", err)
		}

		geojson, err := pointGeoJSON(lon, lat)
		if err != nil {
			return nil, err
		}

		details := "Point"
		if groupTitle != "" {
			details += " · " + groupTitle
		}

		candidates = append(candidates, ImportCandidate{
			Index:       len(candidates),
			Type:        "point",
			Title:       title,
			Description: description,
			Details:     details,
			Geojson:     geojson,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return nil, errors.New("no importable places found in the SQLite database")
	}

	return candidates, nil
}

type gpxXML struct {
	XMLName xml.Name   `xml:"gpx"`
	Tracks  []gpxTrack `xml:"trk"`
}

type gpxTrack struct {
	Name     string            `xml:"name"`
	Desc     string            `xml:"desc"`
	Segments []gpxTrackSegment `xml:"trkseg"`
}

type gpxTrackSegment struct {
	Points []gpxPoint `xml:"trkpt"`
}

type gpxPoint struct {
	Lat float64 `xml:"lat,attr"`
	Lon float64 `xml:"lon,attr"`
}

func parseGPX(path string) ([]ImportCandidate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read gpx: %w", err)
	}

	var gpx gpxXML
	if err := xml.Unmarshal(data, &gpx); err != nil {
		return nil, fmt.Errorf("parse gpx: %w", err)
	}

	var candidates []ImportCandidate

	for i, trk := range gpx.Tracks {
		var coords [][]float64
		for _, seg := range trk.Segments {
			for _, pt := range seg.Points {
				coords = append(coords, []float64{pt.Lon, pt.Lat})
			}
		}
		if len(coords) == 0 {
			continue
		}

		geojson, err := lineGeoJSON(coords)
		if err != nil {
			return nil, err
		}

		title := trk.Name
		if title == "" {
			title = fmt.Sprintf("GPX track %d", i+1)
		}

		candidates = append(candidates, ImportCandidate{
			Index:       len(candidates),
			Type:        "line",
			Title:       title,
			Description: trk.Desc,
			Details:     fmt.Sprintf("Line · %d point(s)", len(coords)),
			Geojson:     geojson,
		})
	}

	if len(candidates) == 0 {
		return nil, errors.New("no tracks found in the GPX file")
	}

	return candidates, nil
}

func pointGeoJSON(lon, lat float64) (json.RawMessage, error) {
	return json.Marshal(struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	}{
		Type:        "Point",
		Coordinates: []float64{lon, lat},
	})
}

func lineGeoJSON(coords [][]float64) (json.RawMessage, error) {
	return json.Marshal(struct {
		Type        string      `json:"type"`
		Coordinates [][]float64 `json:"coordinates"`
	}{
		Type:        "LineString",
		Coordinates: coords,
	})
}
