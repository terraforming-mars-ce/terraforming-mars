package dto

import "terraforming-mars-backend/internal/maps"

// MapPreviews projects the registered boards for setup and lobby selection.
func MapPreviews(mapRegistry *maps.MapRegistry) []MapInfoDto {
	availableMaps := make([]MapInfoDto, 0)
	for _, m := range mapRegistry.ListMaps() {
		mapDef, _ := mapRegistry.GetMap(m.ID)
		tiles := maps.GenerateBoardFromMap(mapDef, false)
		previewTiles := make([]MapPreviewTile, 0)
		for _, t := range tiles {
			if t.Location != "mars" || string(t.Type) == "empty" {
				continue
			}
			volcanic := false
			for _, tag := range t.Tags {
				if tag == "volcanic" {
					volcanic = true
					break
				}
			}
			previewTiles = append(previewTiles, MapPreviewTile{
				Q:        t.Coordinates.Q,
				R:        t.Coordinates.R,
				S:        t.Coordinates.S,
				Type:     string(t.Type),
				Volcanic: volcanic,
				Tags:     t.Tags,
			})
		}
		availableMaps = append(availableMaps, MapInfoDto{ID: m.ID, Name: m.Name, Description: mapDef.Description, Tiles: previewTiles})
	}
	return availableMaps
}
