package trade_bot

import (
	"fmt"
	"kasegu/external/algorithms"
	"kasegu/internal/kraken"
	"log"
)

const (
	baseCurrency = "ZUSD"
	quoteCoin    = "XXBT"
	tradePair    = "BTC/USD"
)

type Client interface {
	Action()
}

type client struct {
	kClient *kraken.Kraken
}

func New() (Client, error) {
	c, err := kraken.NewClient()
	if err != nil {
		return nil, fmt.Errorf("could not create kraken client: %w", err)
	}
	return &client{&c}, nil
}

func (c *client) Action() {
	data, err := (*c.kClient).GetOHCLData("BTC/USD", 1440)
	if err != nil {
		//TODO: Make it keep trying, probably
		log.Printf("could not get data from kraken: %v", err)
		return
	}
	p, err := kraken.ParseOHCLData(data)
	if err != nil {
		log.Printf("could not parse data from kraken: %v", err)
		return
	}
	fa := make([]float64, len(*p))
	for i, v := range *p {
		fa[i] = v.Close
	}
	masei, err := algorithms.CalculateMasei(fa)
	if err != nil {
		log.Printf("could not calculate masei: %v", err)
		return
	}
	mi := uint32(len(*masei) - 1)
	index := (*masei)[mi].Index
	pi := len(*p) - 1
	if index == uint32(pi) {
		if (*masei)[mi].IsLongCond {
			c.Buy()
		} else {
			c.Sell()
		}
	}
}

func (c *client) addOrder(pair string, asset string, transactionType string) error {
	fmt.Printf("%sing trade ...", transactionType)
	bal, err := (*c.kClient).GetAccountBalance()
	if err != nil {
		return fmt.Errorf("could not get account balance: %w", err)
	}
	fmt.Println(bal)
	amt, ok := (*bal)[asset]
	if !ok {
		return fmt.Errorf("could not get account to invest")
	}
	err = (*c.kClient).AddOrder(pair, amt, "buy")
	if err != nil {
		log.Printf("order had an error: %v, retrying...", err)
		for i := 1; i <= 3; i++ {
			err = (*c.kClient).AddOrder(pair, amt, transactionType)
			if err == nil {
				return fmt.Errorf("could not make the trade")
			}
		}
	}
	log.Println("successfully made the trade")
	return nil
}

func (c *client) Buy() {
	err := c.addOrder(tradePair, baseCurrency, "buy")
	if err != nil {
		log.Printf("could not make the trade: %v", err)
	}
}

func (c *client) Sell() {
	err := c.addOrder(tradePair, quoteCoin, "sell")
	if err != nil {
		log.Printf("could not make the trade: %v", err)
	}
}
