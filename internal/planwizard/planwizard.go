package planwizard

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type PlanWizard struct {
	APIEndpoint string
	PlanID      int64
}

func (pw PlanWizard) FetchPlanNodes(limit int, offset int) (*[]Node, error) {
	type Response struct {
		Error  string  `json:"error"`
		Reason string  `json:"reason"`
		Data   *[]Node `json:"data"`
	}

	args := fmt.Sprintf(
		"?limit=%d&offset=%d",
		limit,
		offset,
	)

	url := pw.APIEndpoint + "/plans/" + fmt.Sprintf("%d", pw.PlanID) + "/nodes" + args

	req, _ := http.NewRequest("GET", url, nil)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response *Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if response.Error != "" {
		return nil, errors.New("success `false` returned from Plan Wizard API when fetching nodes — " + response.Error + ": " + response.Reason)
	}

	return response.Data, nil
}
