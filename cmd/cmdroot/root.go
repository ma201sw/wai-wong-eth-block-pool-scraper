package cmdroot

import (
	"wai-wong/cmd/cmdhighestearningpool"
	"wai-wong/cmd/cmdscraper"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "root",
	Short: "scraper cli tools",
}

func init() {
	RootCmd.AddCommand(cmdscraper.ScraperCmd)
	cmdscraper.ScraperCmd.AddCommand(cmdhighestearningpool.HighestEarningPoolCmd)
	cmdhighestearningpool.HighestEarningPoolCmd.Flags().Int("days", 1, "default: 1, days pio to today")
	cmdhighestearningpool.HighestEarningPoolCmd.Flags().Int("daysslidingwindowminutes", 180, "default: 180,  sliding window minutes of days to process(has to be a round number that can divide into 1440 which is 1 day in minutes, the lower this number the more connectins are made)")
	cmdhighestearningpool.HighestEarningPoolCmd.Flags().Int("noofblocksslidingwindow", 10, "default: 10, number of blocks sliding window to process(the lower this number the more connection are made)")
	cmdhighestearningpool.HighestEarningPoolCmd.Flags().Float64("minlockedusd", 1.0, "default: 1.0, minimum locked USD liquidiy in a pool")
}
