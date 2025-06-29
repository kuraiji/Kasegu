package trade_bot

import (
	"fmt"
	"kasegu/external/algorithms"
	"kasegu/internal/kraken"
	"log"
	"strconv"
)

const (
	baseCurrency = "ZUSD"
	quoteCoin    = "PENGU"
	tradePair    = "PENGU/USD"
)

type Client interface {
	Action()
	Buy()
	Sell()
}

type client struct {
	kClient *kraken.Kraken
}

func New(apiKeyEnv string, privateKeyEnv string) (Client, error) {
	c, err := kraken.NewClient(apiKeyEnv, privateKeyEnv)
	if err != nil {
		return nil, fmt.Errorf("could not create kraken client: %w", err)
	}
	return &client{&c}, nil
}

func (c *client) Action() {
	log.Printf("Commening Action, BaseCurrency: %s | QuoteCoin: %s | TradePair: %s", baseCurrency, quoteCoin, tradePair)
	data, err := (*c.kClient).GetOHCLData(tradePair, 1440)
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
	log.Printf("MaseiIndex: %d | PIndex: %d", index, uint32(pi))
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
	if transactionType == "buy" {
		ti, err := (*c.kClient).GetTickerInformation(pair)
		if err != nil {
			return fmt.Errorf("could not get ticker information: %w", err)
		}
		fmt.Println(ti)
		a, ok := (*ti)[pair]
		if !ok {
			return fmt.Errorf("ticker info didn't have ask values")
		}
		p, err := strconv.ParseFloat(a.A[0], 64)
		if err != nil {
			return fmt.Errorf("could not parse price: %w", err)
		}
		af, err := strconv.ParseFloat(amt, 64)
		if err != nil {
			return fmt.Errorf("could not parse amount: %w", err)
		}
		amt = fmt.Sprint(af / p)
		fmt.Println(amt)
	}
	err = (*c.kClient).AddOrder(pair, amt, transactionType)
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
