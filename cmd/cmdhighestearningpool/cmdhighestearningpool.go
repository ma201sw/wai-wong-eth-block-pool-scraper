package cmdhighestearningpool

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"wai-wong/internal/constant"
	"wai-wong/internal/domain"
	"wai-wong/internal/highestearningpool"

	"github.com/spf13/cobra"
	"go.uber.org/atomic"
)

var HighestEarningPoolCmd = &cobra.Command{
	Use:   "highestearningpool",
	Short: "highest earning pool",
	Run:   getHighestEarningPool,
}

func getHighestEarningPool(cmd *cobra.Command, args []string) {
	days, err := cmd.Flags().GetInt(constant.Days)
	if err != nil {
		panic(err)
	}

	if days == 0 {
		log.Fatal(`missing or incorrect "` + constant.Days + `" param`)
	}

	daysSlidingWindowmins, err := cmd.Flags().GetInt(constant.DaysSlidingWindowmins)
	if err != nil {
		panic(err)
	}

	if daysSlidingWindowmins == 0 {
		log.Fatal(`missing or incorrect "` + constant.DaysSlidingWindowmins + `" param`)
	}

	noOfblocksSlidingWindow, err := cmd.Flags().GetInt(constant.NoOfblocksSlidingWindow)
	if err != nil {
		panic(err)
	}

	if noOfblocksSlidingWindow == 0 {
		log.Fatal(`missing or incorrect "` + constant.NoOfblocksSlidingWindow + `" param`)
	}

	minLockedUsd, err := cmd.Flags().GetFloat64(constant.MinLockedUsd)
	if err != nil {
		panic(err)
	}

	highestEarningPoolSrv := highestearningpool.New(&http.Client{Timeout: time.Second * constant.HTTPClientTimeout})

	var poolAddress atomic.String

	var earningsPerUSDLeader atomic.Float64

	fmt.Println("running...")

	ethBlocks, err := getEthBlocks(days, daysSlidingWindowmins, highestEarningPoolSrv)
	if err != nil {
		log.Fatalf("getEthblocks error: %v\n", err)
	}

	getHighestEarningPoolAddressAs(ethBlocks, &poolAddress, &earningsPerUSDLeader, noOfblocksSlidingWindow, minLockedUsd, highestEarningPoolSrv)

	fmt.Println("final result:")
	fmt.Printf("pool address: %v, earnings per usd: %v\n", poolAddress.Load(), fmt.Sprintf("%f", earningsPerUSDLeader.Load()))
}

func getEthBlocks(noOfDaysToCheck int, minsInSlidingWindow int, highestEarningPoolSrv highestearningpool.Service) ([]domain.EthBlocks, error) {
	now := time.Now()
	timezone := time.UTC

	noOfWindows := (noOfDaysToCheck * constant.MinsInADay) / minsInSlidingWindow
	minsInSlidingWindowTime := time.Minute * time.Duration(minsInSlidingWindow)

	beginningOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, timezone)

	var waitGroup sync.WaitGroup

	var ethBlocks []domain.EthBlocks

	for i := 0; i < noOfWindows; i++ {
		start := beginningOfDay.Add(-time.Duration(i+1) * minsInSlidingWindowTime)
		end := beginningOfDay.Add(-time.Duration(i) * minsInSlidingWindowTime)

		waitGroup.Add(1)

		go func(goStart, goEnd time.Time) {
			defer waitGroup.Done()

			startID := ""

			for {
				ethBlocksSubset, nextID, err := highestEarningPoolSrv.GetEthBlocksAPI(goStart.Unix(), goEnd.Unix(), startID)
				if err != nil {
					log.Printf("get eth blocks api error: %v\n", err)

					continue
				}

				if nextID == "-1" { // no more blocks
					break
				}

				startID = nextID

				ethBlocks = append(ethBlocks, *ethBlocksSubset)
			}
		}(start, end)
	}

	waitGroup.Wait()

	return ethBlocks, nil
}

func getHighestEarningPoolAddressAs(ethBlocks []domain.EthBlocks, poolAddress *atomic.String, earningsPerUSDLeader *atomic.Float64, window int, lockedUSDGte float64, highestEarningPoolSrv highestearningpool.Service) {
	var waitGroup sync.WaitGroup

	endOfEthBlocks := len(ethBlocks)
	windowEnd := window

	if windowEnd > endOfEthBlocks {
		windowEnd = endOfEthBlocks
	}

	windowStart := 0

	for {
		waitGroup.Add(1)

		go func(ethBlocks []domain.EthBlocks) {
			defer waitGroup.Done()

			for _, ethBlocksData := range ethBlocks {
				for _, block := range ethBlocksData.Data.Blocks {
					startVolume := "999999999999999999"

					for {
						nextVolume, uniErr := highestEarningPoolSrv.GetUniswapV3HighestEarningPoolAddressAPIAs(block, poolAddress, earningsPerUSDLeader, startVolume, lockedUSDGte)
						if uniErr != nil {
							log.Printf("uniswap v3 error: %v\n", uniErr)

							continue
						}

						if nextVolume == "-1" {
							break
						}

						nextVolumeFloat, err := strconv.ParseFloat(nextVolume, 64)
						if err != nil {
							log.Printf("parse float error: %v\n", err)

							continue
						}

						totalLiqPrvdrEarningsUSD := constant.HighestFee * nextVolumeFloat

						if totalLiqPrvdrEarningsUSD < earningsPerUSDLeader.Load() {
							break
						}

						startVolume = nextVolume
					}
				}
			}
		}(ethBlocks[windowStart:windowEnd])

		if windowEnd == endOfEthBlocks {
			break
		}

		windowStart += window
		windowEnd += window

		if windowEnd > endOfEthBlocks {
			windowEnd = endOfEthBlocks
		}
	}
	waitGroup.Wait()
}
