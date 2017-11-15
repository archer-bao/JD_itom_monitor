package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type item struct {
	Op float64 `json:"op,string"`
	M  float64 `json:"m,string"`
	Id string  `json:"id"`
	P  float64 `json:"p,string"`
}

var logger = Log()

func Log() *log.Logger {
	fileName := "jd_Monitor.log"
	logFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalln("open file error !")
		os.Exit(-1)
	}
	defer logFile.Close()

	// 创建一个日志对象
	debugLog := log.New(logFile, "[Debug]", log.LstdFlags)
	//配置一个日志格式的前缀
	debugLog.SetPrefix("[Info]")
	return debugLog
}

var f = retry()

func retry() func() bool {
	var retry int = 3
	return func() bool {
		if retry--; retry == 0 {
			return false
		} else {
			return true
		}
	}
}

func getItem() {
	res, er := http.Get(`https://p.3.cn/prices/mgets?skuIds=J_1470147`)

	if er != nil {
		logger.Println("network error")
		os.Exit(-1)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		// handle error
		if f() {
			logger.Println("response:", string(body), " error: ", err)
			sendSms("JD API CANNOT ACCESS (API接口无法访问)")
			os.Exit(-1)
		}
	} else {
		result := string(body)
		result = result[1 : len(result)-2]
		var goods item
		// print(result, "\n")

		err := json.Unmarshal([]byte(result), &goods)
		if err != nil {
			logger.Println("response:", result, " error: ", err)
			sendSms("JD API ERROR (API接口访问返回错误，可能需要验证码)")
			os.Exit(-1)
		} else {
			if goods.P < 699 {
				// print("promotion")
				// print(strconv.FormatFloat(goods.P, 'f', 2, 32))
				logger.Println("price:", goods.P)
				sendSms("飞利浦（PHILIPS）行车记录仪ADR810已降价，当前价格:" + strconv.FormatFloat(goods.P, 'f', 2, 32))
			}
		}
	}
}

func sendSms(content string) {
	client := &http.Client{}

	req, _ := http.NewRequest("POST", "http://sms-api.luosimao.com/v1/send.json", strings.NewReader("mobile=13512345678&message="+content+"【sign】"))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic YXBpOmtleSr240stuek2budewtMjp9txt1c1hf3lxy3sjRlNw")

	resp, err := client.Do(req)
	if err != nil {
		logger.Println("error:", err)
		return
	}
	defer resp.Body.Close()

	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		logger.Println("response:", string(body), " error: ", err2)
	}
}

func GenerateRangeNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max-min) + min
	return randNum
}

func ServiceCheck() {
	t1 := time.Now().Day()
	t2 := time.Now().Hour()
	t3 := time.Now().Minute()
	if t1%5 == 0 && t2 == 10 && t3 == 10 {
		sendSms("服务存活检测")
	}
}

func main() {
	for {
		ServiceCheck()
		getItem()
		time.Sleep(time.Duration(GenerateRangeNum(15, 25)) * time.Minute)
	}
}
