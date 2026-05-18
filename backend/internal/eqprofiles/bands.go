package eqprofiles

type Bands struct {
	SubBass  int `json:"subBass" validate:"min=-12,max=12"`
	Bass     int `json:"bass" validate:"min=-12,max=12"`
	Mid      int `json:"mid" validate:"min=-12,max=12"`
	Treble   int `json:"treble" validate:"min=-12,max=12"`
	Presence int `json:"presence" validate:"min=-12,max=12"`
}
