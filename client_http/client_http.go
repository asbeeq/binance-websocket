package client_http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	base_url = "https://api.binance.com/api/v3/depth"
	limit    = 20
)

type OrderBook struct {
	LastUpdateId    int        `json:"lastUpdateId"`
	Bids            [][]string `json:"bids"`
	Asks            [][]string `json:"asks"`
	SumBidsQuantity float64
	SumAsksQuantity float64
}

func MakeRequest(symbol string) (*OrderBook, error) {
	url := fmt.Sprintf("%s?limit=%d&symbol=%s", base_url, limit, symbol)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// преобразовать json в струтуру
	orderBooks := OrderBook{}
	err = json.Unmarshal(body, &orderBooks)
	if err != nil {
		return nil, err
	}

	// проверка если элементы найдены
	if len(orderBooks.Bids) == 20 && len(orderBooks.Asks) == 20 {
		// оставить только по 15 элементов внутри
		orderBooks.Bids = orderBooks.Bids[:15]
		orderBooks.Asks = orderBooks.Asks[:15]

		for i := range orderBooks.Bids {
			BidsQuantity, err := strconv.ParseFloat(orderBooks.Bids[i][1], 64)
			if err != nil {
				return nil, err
			}
			orderBooks.SumBidsQuantity += BidsQuantity

			AsksQuantity, err := strconv.ParseFloat(orderBooks.Asks[i][1], 64)
			if err != nil {
				return nil, err
			}
			orderBooks.SumAsksQuantity += AsksQuantity
		}
	}

	return &orderBooks, nil
}
