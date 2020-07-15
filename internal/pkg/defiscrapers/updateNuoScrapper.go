package defiscrapers

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"

	"github.com/diadata-org/diadata/pkg/dia"
	"github.com/diadata-org/diadata/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type NuoMarket struct {
	Data struct {
		Users []User `json:"users"`
	} `json:"data"`
}

type NuoProtocol struct {
	scraper  *DefiScraper
	protocol dia.DefiProtocol
}

type User struct {
	ID                  string `json:"id"`
	LastUpdateTimestamp int    `json:"lastUpdateTimestamp"`
	LiquidityRate       string `json:"liquidityRate"`
	Name                string `json:"name"`
	Price               struct {
		ID string `json:"id"`
	} `json:"price"`
	TotalOrdersLiquidated string `json:"totalOrdersLiquidated"`
	TotalOrdersSettled    string `json:"totalOrdersSettled"`
}

func NewNuo(scraper *DefiScraper, protocol dia.DefiProtocol) *NuoProtocol {
	return &NuoProtocol{scraper: scraper, protocol: protocol}
}

func fetchNuoMarkets() (nuorate NuoMarket, err error) {
	jsonData := map[string]string{
		"query": `
		{
			makerToCompounds {
			  cdpNumber
			  owner
			  daiAmount
			  ethAmount
			  fees
			  transactionHash
			  timestamp
			}
		  }
		  
`,
	}
	jsonValue, _ := json.Marshal(jsonData)
	jsondata, err := utils.PostRequest("https://api.thegraph.com/subgraphs/name/sudeepb02/nuonetwork", bytes.NewBuffer(jsonValue))
	err = json.Unmarshal(jsondata, &nuorate)
	log.Println(nuorate)
	return
}

func getNuoAssetByAddress(address string) (user User, err error) {
	markets, err := fetchNuoMarkets()
	if err != nil {
		return
	}
	for _, market := range markets.Data.Users {
		if market.ID == address {
			user = market
		}
	}
	return
}

func (proto *NuoProtocol) UpdateRate() error {
	log.Printf("Updating DEFI Rate for %+v \n ", proto.protocol.Name)

	markets, err := fetchNuoMarkets()
	if err != nil {
		return err
	}

	for _, market := range markets.Data.Users {

		totalSupplyAPR, err := strconv.ParseFloat(market.TotalOrdersLiquidated, 64)
		if err != nil {
			return err
		}
		totalBorrowAPR, err := strconv.ParseFloat(market.TotalOrdersSettled, 64)
		if err != nil {
			return err
		}
		asset := &dia.DefiRate{
			Timestamp:     time.Now(),
			Asset:         "Nuo",
			Protocol:      proto.protocol,
			LendingRate:   totalSupplyAPR,
			BorrowingRate: totalBorrowAPR,
		}
		log.Printf("writing DEFI rate for  %#v in %v\n", asset, proto.scraper.RateChannel())
		proto.scraper.RateChannel() <- asset

	}

	log.Info("Update complete")
	return nil
	// log.Printf("NuoScraper update")

	// markets, err := fetchNuoMarkets()
	// if err != nil {
	// 	return err
	// }

	// for _, market := range markets.Data.Users {

	// 	totalSupplyAPR, err := strconv.ParseFloat(market.TotalOrdersLiquidated, 64)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	totalBorrowAPR, err := strconv.ParseFloat(market.TotalOrdersSettled, 64)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	asset := &dia.DefiRate{
	// 		Timestamp:     time.Now(),
	// 		Asset:         "Nuo",
	// 		Protocol:      protocol,
	// 		LendingRate:   totalSupplyAPR,
	// 		BorrowingRate: totalBorrowAPR,
	// 	}
	// 	log.Printf("writing DEFI rate for  %#v in %v\n", asset, s.chanDefiRate)
	// 	s.chanDefiRate <- asset

	// }

	// log.Info("Update complete")
	// return nil
}

func (proto *NuoProtocol) UpdateState() error {
	log.Print("Updating DEFI state for %+v\\n ", proto.protocol)
	usdcMarket, err := getNuoAssetByAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")
	if err != nil {
		return err
	}
	ethMarket, err := getNuoAssetByAddress("0x0000000000000000000000000000000000000000")
	if err != nil {
		return err
	}
	totalUSDCSupplyPAR, err := strconv.ParseFloat(usdcMarket.TotalOrdersLiquidated, 64)
	if err != nil {
		return err
	}
	totalETHSupplyPAR, err := strconv.ParseFloat(ethMarket.TotalOrdersSettled, 64)
	if err != nil {
		return err
	}

	defistate := &dia.DefiProtocolState{
		TotalUSD:  totalUSDCSupplyPAR,
		TotalETH:  totalETHSupplyPAR,
		Protocol:  proto.protocol.Name,
		Timestamp: time.Now(),
	}
	proto.scraper.StateChannel() <- defistate
	log.Printf("writing DEFI state for  %#v in %v\n", defistate, proto.scraper.StateChannel())

	log.Info("Update State complete")

	return nil

}
