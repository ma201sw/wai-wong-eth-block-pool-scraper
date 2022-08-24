package highestearningpool

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"go.uber.org/atomic"

	"wai-wong/internal/domain"
)

func Test_highestEarningPoolImpl_GetUniswapV3HighestEarningPoolAddressAPIAs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		myHTTPClientMock       func(t *testing.T) *MockHTTPClientImpl
		expectedPoolAddress    string
		expectedEarningsPerUSD float64
		expectedNextVolume     string
		wantErr                bool
	}{
		{
			name: "error response",
			myHTTPClientMock: func(t *testing.T) *MockHTTPClientImpl {
				t.Helper()

				return &MockHTTPClientImpl{
					DoFn: func(req *http.Request) (*http.Response, error) {
						return nil, fmt.Errorf("test error")
					},
				}
			},
			wantErr: true,
		},
		{
			name:                   "successful response",
			expectedPoolAddress:    "0x11b815efb8f581194ae79006d24e0d814b7697f6",
			expectedEarningsPerUSD: 0.604147301673943,
			expectedNextVolume:     "44954936422.71104360912169302257554",
			myHTTPClientMock: func(t *testing.T) *MockHTTPClientImpl {
				t.Helper()

				return &MockHTTPClientImpl{
					DoFn: func(req *http.Request) (*http.Response, error) {
						json := `{
							"data": {
							  "pools": [
								{
								  "id": "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640",
								  "feeTier": "500",
								  "totalValueLockedUSD": "326468570.9903251612828761943181873",
								  "volumeUSD": "258527716184.7313543475828786371853"
								},
								{
								  "id": "0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8",
								  "feeTier": "3000",
								  "totalValueLockedUSD": "343850955.5084472013093565690716945",
								  "volumeUSD": "60182915976.70729418898882712041335"
								},
								{
								  "id": "0x11b815efb8f581194ae79006d24e0d814b7697f6",
								  "feeTier": "500",
								  "totalValueLockedUSD": "37205277.83385939796263244717340579",
								  "volumeUSD": "44954936422.71104360912169302257554"
								}
							  ]
							}
						  }`

						jsonRC := io.NopCloser(strings.NewReader(json))

						return &http.Response{
							StatusCode: 200,
							Body:       jsonRC,
						}, nil
					},
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var poolAddressResult atomic.String
			var earningsPerUSDResult atomic.Float64

			poolAddressResult.Store("")
			earningsPerUSDResult.Store(0)

			highestEarningsSrv := New(tt.myHTTPClientMock(t))

			got, err := highestEarningsSrv.GetUniswapV3HighestEarningPoolAddressAPIAs(domain.Block{}, &poolAddressResult, &earningsPerUSDResult, "", 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("highestEarningPoolImpl.GetUniswapV3HighestEarningPoolAddressAPIAs() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if err == nil && got != tt.expectedNextVolume {
				t.Errorf("highestEarningPoolImpl.GetUniswapV3HighestEarningPoolAddressAPIAs() = %v, want %v", got, tt.expectedNextVolume)

				return
			}

			if err == nil && poolAddressResult.Load() != tt.expectedPoolAddress {
				t.Errorf("highestEarningPoolImpl.GetUniswapV3HighestEarningPoolAddressAPIAs() = %v, want %v", poolAddressResult.Load(), tt.expectedPoolAddress)

				return
			}

			if err == nil && earningsPerUSDResult.Load() != tt.expectedEarningsPerUSD {
				t.Errorf("highestEarningPoolImpl.GetUniswapV3HighestEarningPoolAddressAPIAs() = %v, want %v", earningsPerUSDResult.Load(), tt.expectedEarningsPerUSD)

				return
			}
		})
	}
}
