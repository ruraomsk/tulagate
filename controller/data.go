package controller

type Controller struct {
	ID             string
	AddressRu      string
	AddressLatin   string
	Geom           Geom
	LastProgrammId string
}
type Geom struct {
	Latitude  float64 //			Широта
	Longitude float64 // Долгота
}
