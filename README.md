# PMPOS
### Poor man's Point of Sale terminal
Submitted for the Solana DeFi Hackathon March 1 2021

# Installation
    go get "github.com/gorilla/websocket"

    go get "github.com/tidwall/gjson"

# Running
    go run pmpos.go

# What it does
This application simulates the back-end of a merchant's Point of Sale (PoS) system. 
It performs two basic functions:

1. Keeps a tab on the merchant's USDC balance of a soft wallet held on Circle.com.  It accomplishes this by polling the Circle Wallet API periodically
2. Connects to the Solana testnet WebSocket and submits a request to be notified of all changes to an address on the Solana chain to which a customer pays the merchant in USDC

This server should be running while the other component of this hackathon (tomeck/solana-hackathon-cust-app) simulates a checkout flow at the merchant's PoS.

Due to the rapid settlement time of Solana, the merchant is assured that the transfer transaction has completed and it is safe to provide the customer with the goods or services he just paid for.
