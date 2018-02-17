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

const resultChanTtl = 60

var resultChannels = make(map[int]chan *Result)

func Request(c *gin.Context) {
	var _, ridSent = c.GetQuery("rid")
	if ridSent {
		return
	}
	var _, isASync = c.GetQuery("async")
	rid := rand.Int()
	fmt.Println(rid)
	resultChannels[rid] = make(chan *Result,1)

	c.Set("rid", rid)
	c.Abort()
	go c.Handler()(c)
	if isASync {
		c.JSON(200, map[string]interface{}{
			"rid": fmt.Sprint(rid),
		})
		c.Abort()
		return
	} else {
		response(c, rid)
	}
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
		fmt.Printf("Result channel with id '%d' ditroid\n", rid)
	}
}

func response(c *gin.Context, rid int) {

	result := <-resultChannels[rid]
	c.JSON(result.Code, result.Data)
	delete(resultChannels, rid)
}

func Worker(c *gin.Context) {
	time.Sleep(time.Second * 5)
	fmt.Println("Worker end")

	ApiResponse(c, 201, map[string]interface{}{
		"message": "success!",
	})
}


/*
<?php


class async
{
    /** @var  Generator * /
private $g;
public function request()
{
$this->g = $this->send();
if($this->g->current())
return true;
else
throw new Exception("ERROR");
}

public function wait()
{
$this->g->next();
return $this->g->current();
}

private function send()
{
$url = 'http://127.0.0.1:8080/test/t2';
$res = json_decode(file_get_contents($url . '?async'));
yield true;
echo "inja";
$res = json_decode(file_get_contents($url . '?rid='.$res->rid));
echo "inja";
yield $res;
}
}


$api = new async();

$api->request();

echo "yekar!\n";
//sleep(1);
echo "dokar!\n";

$result =  $api->wait();

die(json_encode($result, JSON_PRETTY_PRINT));



*/