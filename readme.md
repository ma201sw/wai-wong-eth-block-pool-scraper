### Quick start
This is a CLI program that efficiently gets the highest earning liquidity pool over n days if you invest $1

you can run it by going into the folder and running go run . or use the wai-wong.exe included

You can use go run . by running this in command line for example:
go run . scraper highestearningpool --days 1 --daysslidingwindowminutes 180 --noofblocksslidingwindow 10 --minlockedusd 1.0

You can use the exe provided by running this in command line for example:
.\wai-wong.exe scraper highestearningpool --days 1 --daysslidingwindowminutes 180 --noofblocksslidingwindow 10 --minlockedusd 1.0

Here are some commands to run for 1,3,10 days:
go run . scraper highestearningpool --days 1 --daysslidingwindowminutes 180 --noofblocksslidingwindow 8 --minlockedusd 1.0

go run . scraper highestearningpool --days 3 --daysslidingwindowminutes 180 --noofblocksslidingwindow 24 --minlockedusd 1.0

go run . scraper highestearningpool --days 10 --daysslidingwindowminutes 180 --noofblocksslidingwindow 90 --minlockedusd 1.0

And for the exe:
.\wai-wong.exe scraper highestearningpool --days 1 --daysslidingwindowminutes 180 --noofblocksslidingwindow 8 --minlockedusd 1.0

.\wai-wong.exe scraper highestearningpool --days 3 --daysslidingwindowminutes 180 --noofblocksslidingwindow 24 --minlockedusd 1.0

.\wai-wong.exe scraper highestearningpool --days 10 --daysslidingwindowminutes 180 --noofblocksslidingwindow 90 --minlockedusd 1.0


### Arguments explained:
--days are the number of days you want to search prior to today

--daysslidingwindowminutes is the sliding window for getting eth blocks, each window spawns a go routine. It needs to be able to divide into 1440 as a round number, 1440 is the number of minutes in a day. The lower this number the more go routines and connections spawn, if you want to lower connections you can use 360 if you are searching through many days ie 20 days

--noofblocksslidingwindow is the sliding window when searching the blocks for pools, the lower this number the more connections and go routines spawn. 1 is the fastest but you can get connection errors if the search is too large. The example commands above all download at about 60Mbps. If you have too many connection problem increase this number

--minlockedusd finds locked volume pools that are more than 1 dollar. If you set this to 0.0 it finds pools which are less than 1 dollar and because earnings are reported per USD, I report the earnings per USD as whole earnings pot (fee X volumeUSD)


### formula:
I use the Ethereum Blocks API schema item blocks to get the eth blocks and Uniswap V3 Subgraph schema item pools to get the pools from the blocks. I use the formula to calculate earnings per USD for a pool ((feeTier / 1000000) * volumeUSD) * (1/totalValueLockedUSD) I assume totalValueLockedUSD is the liquidity pool provided in dollars amount

(feeTier / 1000000) * volumeUSD) calculates the total share of earnings to be shared across liquidty providers. (1/totalValueLockedUSD) is your proportionate share if you invest 1 dollar

note if the totalValueLockedUSD is less than 1 dollars I make the earnings per USD the whole shared liquidity provider earnings pot (feeTier / 1000000) * volumeUSD), this also only happens if you put the argument --minlockedusd 0.0


### packages, structs and interfaces:
package cmdhighestearningpool contains the entrypoint when running the program, I use a bunch of go routines to call the APIs. This uses the highestearningpool.New(myHTTPClient MyHTTPClientSrv) which takes in a httpclient for dependency injection, this httpclient can use a real one or a mock one. You can have a struct with more dependencies and inject that to be able to mock more libraries and methods

package common contains type errors, I avoid sentinel errors

package constant contains constants and config

package domain contains the structs for marshaling and unmarshaling eth blocks and pools

package highestearningpool contains a service for the API calls, this is called by cmdhighestearningpool/cmdhighestearningpool.go

highestearningpool/mock.go contains mocks I have written(I prefer to write my own mocks than use mockgen as it gives more flexibility and you learn more). With my flexible mocks I can potentially achieve 100% coverage. Mocks are dependency injected. Another way to dependency inject is to use key value in context

highestearningpool/highestearningpool_test.go uses the mock MockHTTPClientImpl I have written. Note I have only tested the method GetUniswapV3HighestEarningPoolAddressAPIAs to save time. 


### other notes:
Spent time finding which APIs to use, what the formulas are and understanding how it all works

Uses UTC time

if you have connection error you may need to throttle, set a higher blocks sliding window e.g. --noofblocksslidingwindow 100. Try to tinker with the setting so you download at your broadband bandwidth limit, mine is about 70Mbps

did some optimisations by sorting by volumeUSD desc and multiplying the volumeUSD by the highest fee(0.01) and if this is less than the leading earnings, move onto the next block

I use wait groups and atomic vars to help aggregrate the results in a thread safe manner

I store the ethblocks in memory, there maybe some further optimizations here

avoid sentinel errors, used type errors. If I spent more time I probably would use error AS/IS error matching to improve errors. Errors should also be propagated up in a format like service1: service2: token error: the error

should be using decimal for currency instead of float. Go does not have decimal so [Eric Lagergren's decimal](https://github.com/ericlagergren/decimal) should work

use golangci-lint to enforce go standards

There is possible a possible race condition with the load and store for earningsPerUSDLeader, to fix this race condition I can use a mutex but it will slow things down a lot. Because of how big the queries are, I have chosen to prefer speed and accept the low risk of the race condition for the time being


Please let me know if something doesn't work thanks!
