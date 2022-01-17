package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	orangeSDK "github.com/orange-protocol/orange-provider-go-sdk"
	orangeOnt "github.com/orange-protocol/orange-provider-go-sdk/ont"
)

type BalanceReq struct {
	ProviderDID string  `json:"provider_did"`
	Data        ReqData `json:"data"`
	Encrypted   string  `json:"encrypted"`
}

type BalanceData struct {
	Balance string `json:"balance"`
}

type ReqData struct {
	Data BalanceData `json:"data"`
	Sig  string      `json:"sig"`
}

var didsdk *orangeSDK.OrangeProviderSdk

func main() {
	didsdk, err := orangeOnt.NewOrangeProviderOntSdk("./wallet.dat", "123456", "TESTNET")
	if err != nil {
		panic(err)
	}
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/calcScore", func(c *gin.Context) {
		requestJson := &BalanceReq{}
		if err := c.ShouldBindJSON(requestJson); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		encryptMsg, err := hex.DecodeString(requestJson.Encrypted)
		if err != nil {
			fmt.Printf("DecodeString encryptMsg failed:%s\n", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		decryptedbts, err := didsdk.DecryptData(encryptMsg)
		if err != nil {
			fmt.Printf("DecryptMsg failed:%s\n", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		dataWithSig := &ReqData{}
		err = json.Unmarshal(decryptedbts, dataWithSig)
		if err != nil {
			fmt.Printf("Unmarshal dataWithSig failed:%s\n", err.Error())

			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		msgbytes, err := json.Marshal(dataWithSig.Data)
		if err != nil {
			fmt.Printf("Marshal msgbytes failed:%s\n", err.Error())

			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		sigbytes, err := hex.DecodeString(dataWithSig.Sig)
		if err != nil {
			fmt.Printf("DecodeString sigbytes failed:%s\n", err.Error())

			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		f, err := didsdk.VerifySig(requestJson.ProviderDID, msgbytes, sigbytes)
		if err != nil || !f {
			fmt.Printf("VerifySig  failed:%s\n", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("invalid signature")})
			return
		}

		balance := dataWithSig.Data.Balance
		fmt.Printf("balance is %d\n", balance)

		c.JSON(200, gin.H{
			"score": 500,
		})

	})
	r.Run(":3001") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
