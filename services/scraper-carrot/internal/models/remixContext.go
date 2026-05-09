package models

type RemixContext struct {
	State struct {
		LoaderData struct {
			Route struct {
				AllPage struct {
					Articles []Article `json:"fleamarketArticles"`
				} `json:"allPage"`
				PoW struct {
					Challenge  string `json:"challenge"`
					Difficulty int    `json:"difficulty"`
					ExpiresAt  int64  `json:"expiresAt"`
					URI        string `json:"uri"`
				} `json:"pow"`
				CurrentFilters struct {
					RegionID   string `json:"regionId"`
					Search     string `json:"search"`
					CategoryID string `json:"categoryId"`
					Price      string `json:"price"`
					OnlyOnSale bool   `json:"onlyOnSale"`
				} `json:"currentFilters"`
			} `json:"routes/kr.buy-sell.s"`
		} `json:"loaderData"`
	} `json:"state"`
}

type FleamarketSearchResponse struct {
	FleamarketArticles []Article `json:"fleamarketArticles"`
}
