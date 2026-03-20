package shared

// StandardProject represents the different types of standard projects
type StandardProject string

const (
	StandardProjectSellPatents              StandardProject = "sell-patents"
	StandardProjectPowerPlant               StandardProject = "power-plant"
	StandardProjectAsteroid                 StandardProject = "asteroid"
	StandardProjectAquifer                  StandardProject = "aquifer"
	StandardProjectGreenery                 StandardProject = "greenery"
	StandardProjectCity                     StandardProject = "city"
	StandardProjectAirScrapping             StandardProject = "air-scrapping"
	StandardProjectConvertPlantsToGreenery  StandardProject = "convert-plants-to-greenery"
	StandardProjectConvertHeatToTemperature StandardProject = "convert-heat-to-temperature"
)

// StandardProjectCost represents the credit cost for each standard project
var StandardProjectCost = map[StandardProject]int{
	StandardProjectSellPatents:  0,
	StandardProjectPowerPlant:   11,
	StandardProjectAsteroid:     14,
	StandardProjectAquifer:      18,
	StandardProjectGreenery:     23,
	StandardProjectCity:         25,
	StandardProjectAirScrapping: 15,
}
