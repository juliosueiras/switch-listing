package utils

type NSCollectorSheetItem struct {
	GameTitle string `csv:"gametitle"`
	ReleaseId string `csv:"release"`
	USADate string `csv:"usadate"`
	JPNDate string `csv:"jpndate"`
	EUDate string `csv:"eurdate"`
	AUSDate string `csv:"ausdate"`
	USACartID string `csv:"usacart"`
	JPNCartID string `csv:"jpncart"`
	EUCartID string `csv:"eurcart"`
	AUSCartID string `csv:"auscart"`
	EnglishOnCart string `csv:"english"`
	Notes string `csv:"notes"`
}
