// Copyright 2021 J. Thomas Eck. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/tidwall/gjson"
)

// Solana testnet WebSocket address
var addr = flag.String("addr", "testnet.solana.com:8900", "http service address")

// Address on Solana chain for merchant to accept USDC payments
var gMerchantSolanaAddress = "FUoAafzWRYp8dsshzKqadN7QXGZQAJ6M5dc95jN1d9GJ"

var API_KEY = "QVBJX0tFWTpmNTE0ZDU5MWM5YTE4MjI4NGViZGMxNmYwNmQ4ZGVhMjpiOWFlZmEwODU2ZTA4ZDVhODgxNjY2MzQ3NGQ4ODA5Nw"

var gMerchantWalletAddress = "1000072207"

// Tracks merchant balance
var gMerchantBalance = 0.0

// Extract the amount of USDC sent to the merchant in this update
func getAmount(paymentString string) string {

	// TODO JTE this is an unholy way to extract the amount of the payment just received
	idx1 := strings.Index(paymentString, "uiAmount")
	tails := strings.Split(paymentString[idx1:], ":")
	uiAmount := strings.Split(tails[1], "}")[0]

	return uiAmount
}

func getMerchantAccountBalance() float64 {

	// Create an http client
	client := &http.Client{
		CheckRedirect: nil,
	}

	// Setup http request
	req, err := http.NewRequest("GET", "https://api-sandbox.circle.com/v1/wallets/"+gMerchantWalletAddress, nil)
	req.Header.Add("Authorization", "Bearer "+API_KEY)

	// Make http call
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		// handle error
		log.Printf("Error obtaining merchant account balance: %s", resp)
		return -1.0
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	// log.Printf(string(body))

	// Extract the account balance
	strBalance := gjson.Get(string(body), "data.balances.0.amount")
	fAmt, _ := strconv.ParseFloat(strBalance.String(), 32)

	if fAmt != gMerchantBalance {
		gMerchantBalance = fAmt
		log.Printf("Merchant wallet balance is %s USD", strBalance)
	}
	return fAmt
}

func printBanner() {
	log.Printf("====================================================================")
	log.Printf("==>>                                                            <<==")
	log.Printf("==============   CIRCLE+SOLANA  Point Of Sale Demo   ===============")
	log.Printf("==>>                                                            <<==")
	log.Printf("====================================================================")
}

func main() {

	printBanner()

	// Get merchant account balance
	getMerchantAccountBalance()

	u := url.URL{Scheme: "ws", Host: *addr, Path: ""}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})
	defer close(done)

	// Handle the receipt messages from server
	go func() {

		// Loop forever, handling messages received from server
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			// Convert bytes to string and echo what we received from server
			strResp := string(message)
			//log.Printf("recv: %s", strResp)

			// Looks like merchant received a payment or update to his account
			if strings.Contains(strResp, "tokenAmount") {
				amountStr := getAmount(strResp)

				if amountStr != "0.0" {
					log.Printf(">>> Merchant received an updated amount at his USDC address: %s", amountStr)
				}
			}
		}
	}()

	// Send an accountSubscribe message to the Solana chain
	//  indicating we want to receive updates for this
	go func() {

		log.Printf("Registering for updates to merchant address %s", gMerchantSolanaAddress)
		subscripJSON := []byte(`{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "accountSubscribe",
			"params": [
			  "FUoAafzWRYp8dsshzKqadN7QXGZQAJ6M5dc95jN1d9GJ",   
			  {
				"encoding": "jsonParsed",
				"commitment" : "complete"
			  }
			]
		}`)

		err := c.WriteMessage(websocket.TextMessage, subscripJSON)
		if err != nil {
			log.Println("write:", err)
			return
		}
	}()

	// Check for updated Circle wallet balance every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				getMerchantAccountBalance()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		}
	}
}
