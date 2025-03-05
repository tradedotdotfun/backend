package types

type GetAccountResponse struct {
	Name                 string  `json:"name"`
	Rank                 uint64  `json:"rank"`
	USDAmount            float64 `json:"usd_amount"`
	EstimatedTotalAmount float64 `json:"estimated_total_amount"`
}

type AddNameRequest struct {
	Name string `json:"name"`
}

type AddNameResponse struct {
	Status string `json:"status"`
}
