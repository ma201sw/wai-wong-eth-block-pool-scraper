package highestearningpool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"wai-wong/internal/common"
	"wai-wong/internal/constant"
	"wai-wong/internal/domain"

	"go.uber.org/atomic"
)

type highestEarningPoolImpl struct {
	MyHTTPClient MyHTTPClientSrv
}

// verify interface compliance
var _ Service = (*highestEarningPoolImpl)(nil)

type Service interface {
	GetEthBlocksAPI(timestampGt, timestampLt int64, idGt string) (*domain.EthBlocks, string, error)
	GetUniswapV3HighestEarningPoolAddressAPIAs(block domain.Block, poolAddress *atomic.String, earningsPerUSD *atomic.Float64, volumeUSDLt string, lockedUSDGte float64) (string, error)
}

func New(myHTTPClient MyHTTPClientSrv) highestEarningPoolImpl {
	return highestEarningPoolImpl{
		MyHTTPClient: myHTTPClient,
	}
}

type MyHTTPClientSrv interface {
	Do(req *http.Request) (*http.Response, error)
}

func (h highestEarningPoolImpl) GetEthBlocksAPI(timestampGt, timestampLt int64, idGt string) (*domain.EthBlocks, string, error) {
	jsonData := map[string]string{
		"query": `
			{
				blocks(first:10,where: { timestamp_gt:` + strconv.FormatInt(timestampGt, 10) + `, timestamp_lt:` + strconv.FormatInt(timestampLt, 10) + `, id_gt: "` + idGt + `"}) {
				id
				}
		  }
        `,
	}

	jsonValue, jsonErr := json.Marshal(jsonData)
	if jsonErr != nil {
		return nil, "", fmt.Errorf("get eth blocks api error: %w", jsonErr)
	}

	request, err := http.NewRequest("POST", constant.EthBlocksEndpoint, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, "", fmt.Errorf("the http request failed with error %w", err)
	}

	httpClient := &http.Client{Timeout: time.Second * constant.HTTPClientTimeout}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, "", fmt.Errorf("the http response failed with error %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, "", common.HTTPStatusError{StatusCode: response.StatusCode, Status: response.Status}
	}

	data, ioErr := io.ReadAll(response.Body)
	if ioErr != nil {
		return nil, "", fmt.Errorf("io error: %v", ioErr)
	}

	ethBlocks := &domain.EthBlocks{}

	if unMarshalErr := json.Unmarshal(data, &ethBlocks); unMarshalErr != nil {
		return nil, "", fmt.Errorf("response body unmarshal error: %w", unMarshalErr)
	}

	if len(ethBlocks.Data.Blocks) == 0 {
		return nil, "-1", nil
	}

	return ethBlocks, ethBlocks.Data.Blocks[len(ethBlocks.Data.Blocks)-1].ID, nil
}

func (h highestEarningPoolImpl) GetUniswapV3HighestEarningPoolAddressAPIAs(block domain.Block, poolAddress *atomic.String, earningsPerUSD *atomic.Float64, volumeUSDLt string, lockedUSDGte float64) (string, error) {
	jsonData := map[string]string{
		"query": `
				{
					pools(first:1000, block: { hash:"` + block.ID + `" }, orderBy: volumeUSD, orderDirection:desc, where:{volumeUSD_gt:"0", volumeUSD_lt: ` + volumeUSDLt + `}) {
						id
						feeTier
						totalValueLockedUSD
						volumeUSD
					} 
			  }
			`,
	}

	jsonValue, jsonErr := json.Marshal(jsonData)
	if jsonErr != nil {
		return "", fmt.Errorf("get uniswap v3 api marshal error: %w", jsonErr)
	}

	request, err := http.NewRequest("POST", constant.UniswapV3SubgphEndpoint, bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", fmt.Errorf("the http request failed with error %w", err)
	}

	client := h.MyHTTPClient

	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("the http response failed with error: %w", err)
	}

	if response.StatusCode != 200 {
		return "", common.HTTPStatusError{StatusCode: response.StatusCode, Status: response.Status}
	}

	defer response.Body.Close()

	data, ioReadErr := io.ReadAll(response.Body)
	if ioReadErr != nil {
		return "", fmt.Errorf("io read error: %w", ioReadErr)
	}

	pools := domain.Pools{}

	if unMarshalErr := json.Unmarshal(data, &pools); unMarshalErr != nil {
		return "", fmt.Errorf("failed to unmarshal err: %w", unMarshalErr)
	}

	if len(pools.Data.Pools) == 0 { // reached the end
		return "-1", nil
	}

	for _, pool := range pools.Data.Pools {
		feeTier, err := strconv.ParseFloat(pool.FeeTier, 64)
		if err != nil {
			log.Printf("parse float err: %v\n", err)

			continue
		}

		volumeUSDFloat, vErr := strconv.ParseFloat(pool.VolumeUSD, 64)
		if vErr != nil {
			log.Printf("parse float err: %v\n", vErr)

			continue
		}

		totalLiqPrvdrEarningsUSD := (feeTier / 1000000) * volumeUSDFloat

		totalLockedUSD, tErr := strconv.ParseFloat(pool.TotalValueLockedUSD, 64)
		if tErr != nil {
			log.Printf("parse float err: %v\n", tErr)

			continue
		}

		if totalLockedUSD < lockedUSDGte {
			continue
		}

		oneDollarPercentage := 1 / totalLockedUSD

		earningShare := oneDollarPercentage * totalLiqPrvdrEarningsUSD

		if oneDollarPercentage >= 1 {
			earningShare = totalLiqPrvdrEarningsUSD
		}

		earningsPerUSDLoaded := earningsPerUSD.Load()
		if earningShare > earningsPerUSDLoaded {
			earningsPerUSD.Store(earningShare)
			poolAddress.Store(pool.ID)
			fmt.Printf("pool address: %v, earnings per usd: %v\n", poolAddress.Load(), fmt.Sprintf("%f", earningsPerUSDLoaded))
		}
	}

	return pools.Data.Pools[len(pools.Data.Pools)-1].VolumeUSD, nil
}
