package models

// VisionResult - это то, что мы заставим вернуть нейросеть
type VisionResult struct {
	Name     string `json:"name"`      // Например: "피카츄"
	Rarity   string `json:"rarity"`    // Например: "SAR", "ex", "UR" (если есть)
	CardNo   string `json:"card_no"`   // Например: "017/086" (если видно на фото)
	IsPokemonCard bool `json:"is_pokemon_card"` // Защита от фото кроссовок или мебели на 당근마켓
}