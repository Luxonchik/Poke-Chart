package models

// Card описывает структуру покемон-карточки.
// Теги `json:"..."` указывают, как именно эти поля будут называться 
// когда мы превратим их в JSON для фронтенда.
type Card struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
	CardNo   string `json:"card_no"`
	Rarity   string `json:"rarity"`
	Series   string `json:"series"`
}

type CardWithListingCount struct {
	Card
	ListingCount int  `json:"listing_count"`
	MinPrice     *int `json:"min_price"`     // lowest active listing price in KRW, null when none
	VariantCount int  `json:"variant_count"` // master_cards rows for the same printing (e.g. regular + mirror)
}