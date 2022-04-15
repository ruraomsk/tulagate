package controller

type ExternalController struct {
	IDExternal     string //Их ID
	AddressRu      string
	AddressLatin   string
	Geom           Geom
	LastProgrammId string
}
type Geom struct {
	Latitude  float64 // Широта
	Longitude float64 // Долгота
}
type DKSet struct {
	DKSets []OneSet `json:"dkset"`
}
type OneSet struct {
	IDExternal string `json:"idext"` //Их ID
	Area       int    `json:"area"`  //Наш Перекресток
	ID         int    `json:"id"`
	MGRs       []MGR  `json:"mgrs"`
}
type MGR struct {
	Phase int `json:"phase"` // Номер фаза после которой вставлять MГР
	Tlen  int `json:"tlen"`  // Минимальная длительность фазы после которой вставлять МГР
	TMGR  int `json:"tmgr"`  // Длительность МГР
}
