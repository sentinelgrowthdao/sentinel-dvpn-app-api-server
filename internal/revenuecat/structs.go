package revenuecat

type Event struct {
	Id          string `json:"id"`
	Environment string `json:"environment"`
	ProductId   string `json:"product_id"`
	AppUserId   string `json:"app_user_id"`
}

type Payload struct {
	Event Event `json:"event"`
}
