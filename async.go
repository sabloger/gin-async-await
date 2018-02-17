package gin_async_await

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"strconv"
	"time"
)

// Request and Response funcs should be the latest middleware!

type Result struct {
	Code int
	Data interface{}
}

const (
	resultChanTtl = 60
)

var resultChannels = make(map[int]chan *Result)
var deadChannels = make(map[int]chan bool)

func Request(c *gin.Context) {
	var _, ridSent = c.GetQuery("rid")
	if ridSent {
		return
	}
	var _, isASync = c.GetQuery("async")
	rid := rand.Int()
	fmt.Println(rid)
	resultChannels[rid] = make(chan *Result, 10)
	deadChannels[rid] = make(chan bool, 1)

	c.Set("rid", rid)
	c.Abort()
	go funcHandler(c, isASync, rid)
	if isASync {
		c.JSON(200, map[string]interface{}{// todo:: => 204 and rid in header
			"rid": fmt.Sprint(rid),
		})
		c.Abort()
		return
	} else {
		response(c, rid)
	}
}

func funcHandler(c *gin.Context, isASync bool, rid int) {
	c.Handler()(c)
	fmt.Println("mikham bokoshamesh")
	deadChannels[rid] <- true
	fmt.Println("koshtamesh!!")
}

func Response(c *gin.Context) {
	fmt.Println("Response")
	var ridStr, ridSent = c.GetQuery("rid")
	if !ridSent {
		return
	}
	c.Abort()

	var rid, err = strconv.Atoi(ridStr)
	if err != nil {
		c.JSON(400, map[string]interface{}{
			"message": "Invalid request ID.",
		})
		return
	}
	if _, ok := resultChannels[rid]; ok {
		response(c, rid)
	} else {
		c.JSON(404, map[string]interface{}{
			"message": "request not found!",
		})
	}
}

func ApiResponse(c *gin.Context, code int, data interface{}) {
	var rid = c.GetInt("rid")
	go deleteChan(rid)
	resultChannels[rid] <- &Result{
		Code: code,
		Data: data,
	}
}

func deleteChan(rid int) {
	<-time.After(time.Second * resultChanTtl)
	if _, ok := resultChannels[rid]; ok {
		delete(resultChannels, rid)
		fmt.Printf("Result channel with id '%d' destroid\n", rid)
	}
}

func response(c *gin.Context, rid int) {
	fmt.Println("umade inja")
	select {
	case result := <-resultChannels[rid]:
		fmt.Println("umad:", result)
		c.JSON(result.Code, result.Data)
	case <-deadChannels[rid]:
		select {
		case result := <-resultChannels[rid]:
			fmt.Println("umad:", result)
			c.JSON(result.Code, result.Data)
		default:
			fmt.Println("naaaaa!!")
		}
		fmt.Println("naumad!!")
	}
	delete(resultChannels, rid)
	delete(deadChannels, rid)
}

// Sample api controller:
//func Worker(c *gin.Context) {
//	time.Sleep(time.Second * 5)
//	fmt.Println("Worker end")
//
//	ApiResponse(c, 201, map[string]interface{}{
//		"message": "success!",
//	})
//}
