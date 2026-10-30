package models

// SearchRouteKey is the loaderData key of the daangn.com buy-sell search page.
const SearchRouteKey = "routes/kr.search.buy-sell._index"

// RemixContext mirrors the part of window.__remixContext the scraper needs.
type RemixContext struct {
	State struct {
		LoaderData map[string]SearchRoute `json:"loaderData"`
	} `json:"state"`
}

// SearchRoute is the loader data of the search page; articles are server-rendered.
type SearchRoute struct {
	Articles []Article `json:"buySellArticles"`
}
