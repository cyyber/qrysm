// Prometheus exporter for Zond address balances.
// Forked from https://github.com/hunterlong/ethexporter
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/go-zond/params"
	"github.com/theQRL/go-zond/zondclient"
	_ "github.com/theQRL/qrysm/runtime/maxprocs"
)

var (
	allWatching []*Watching
	loadSeconds float64
	totalLoaded int64
	eth         *zondclient.Client
)

var (
	port = flag.Int("port", 9090, "Port to serve /metrics")
	// TODO(now.youtrack.cloud/issue/TQ-1)
	web3URL         = flag.String("web3-provider", "https://goerli.prylabs.net", "Web3 URL to access information about ETH1")
	prefix          = flag.String("prefix", "", "Metrics prefix.")
	addressFilePath = flag.String("addresses", "", "File path to addresses text file.")
)

func main() {
	flag.Parse()

	if *addressFilePath == "" {
		log.Println("--addresses is required")
		return
	}

	err := OpenAddresses(*addressFilePath)
	if err != nil {
		panic(err)
	}

	err = ConnectionToGzond(*web3URL)
	if err != nil {
		panic(err)
	}

	// check address balances
	go func() {
		for {
			totalLoaded = 0
			t1 := time.Now()
			fmt.Printf("Checking %v wallets...\n", len(allWatching))
			for _, v := range allWatching {
				v.Balance = EthBalance(v.Address).String()
				totalLoaded++
			}
			t2 := time.Now()
			loadSeconds = t2.Sub(t1).Seconds()
			fmt.Printf("Finished checking %v wallets in %0.0f seconds, sleeping for %v seconds.\n", len(allWatching), loadSeconds, 15)
			time.Sleep(15 * time.Second)
		}
	}()

	block := CurrentBlock()

	fmt.Printf("ETHexporter has started on port %v using web3 server: %v at block #%v\n", *port, *web3URL, block)

	http.HandleFunc("/metrics", MetricsHTTP)
	http.HandleFunc("/reload", ReloadHTTP)
	srv := &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", *port),
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

// Watching address wrapper
type Watching struct {
	Name    string
	Address string
	Balance string
}

// ConnectionToGzond - Connect to remote server.
func ConnectionToGzond(url string) error {
	var err error
	eth, err = zondclient.Dial(url)
	return err
}

// EthBalance from remote server.
func EthBalance(address string) *big.Float {
	addr, err := common.NewAddressFromString(address)
	if err != nil {
		fmt.Printf("Error fetching Zond Balance for address: %v\n", address)
		return nil
	}
	balance, err := eth.BalanceAt(context.TODO(), addr, nil)
	if err != nil {
		fmt.Printf("Error fetching Zond Balance for address: %v\n", address)
	}
	return ToZond(balance)
}

// CurrentBlock in ETH1.
func CurrentBlock() uint64 {
	block, err := eth.BlockByNumber(context.TODO(), nil)
	if err != nil {
		fmt.Printf("Error fetching current block height: %v\n", err)
		return 0
	}
	return block.NumberU64()
}

// ToZond from Planck.
func ToZond(o *big.Int) *big.Float {
	planck := big.NewFloat(0)
	planck.SetInt(o)
	return new(big.Float).Quo(planck, big.NewFloat(params.Zond))
}

// MetricsHTTP - HTTP response handler for /metrics.
func MetricsHTTP(w http.ResponseWriter, _ *http.Request) {
	allOut := make([]string, 0, len(allWatching))
	total := big.NewFloat(0)
	for _, v := range allWatching {
		if v.Balance == "" {
			v.Balance = "0"
		}
		bal := big.NewFloat(0)
		bal.SetString(v.Balance)
		total.Add(total, bal)
		allOut = append(allOut, fmt.Sprintf("%veth_balance{name=\"%v\",address=\"%v\"} %v", *prefix, v.Name, v.Address, v.Balance))
	}
	allOut = append(allOut,
		fmt.Sprintf("%veth_balance_total %0.18f", *prefix, total),
		fmt.Sprintf("%veth_load_seconds %0.2f", *prefix, loadSeconds),
		fmt.Sprintf("%veth_loaded_addresses %v", *prefix, totalLoaded),
		fmt.Sprintf("%veth_total_addresses %v", *prefix, len(allWatching)))

	if _, err := fmt.Fprintln(w, strings.Join(allOut, "\n")); err != nil {
		logrus.WithError(err).Error("Failed to write metrics")
	}
}

// ReloadHTTP reloads the addresses from disk.
func ReloadHTTP(w http.ResponseWriter, _ *http.Request) {
	if err := OpenAddresses(*addressFilePath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	log.Println("Reloaded addresses")
}

// OpenAddresses from text file (name:address)
func OpenAddresses(filename string) error {
	file, err := os.Open(filename) // #nosec G304
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()
	scanner := bufio.NewScanner(file)
	allWatching = []*Watching{}
	for scanner.Scan() {
		object := strings.Split(scanner.Text(), ":")
		if common.IsAddress(object[1]) {
			w := &Watching{
				Name:    object[0],
				Address: object[1],
			}
			allWatching = append(allWatching, w)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return err
}
