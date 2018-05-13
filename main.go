package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"math/rand"

	"github.com/gorilla/mux"
	"github.com/icza/session"
)

const (
	htmlDocPath = "html/"
	//TulingRequestURL :
	TulingRequestURL = "http://www.tuling123.com/openapi/api"
	//TulingRequestKey :
	TulingRequestKey = "577173e62a2ff6627b62e94d663b449c"
	//TulingCodeText :
	TulingCodeText = 100000
	//TulingCodeURL :
	TulingCodeURL = 200000
	//TulingCodeNews :
	TulingCodeNews = 302000
	//TulingCodeErrorKey :
	TulingCodeErrorKey = 40001
	//TulingCodeErrorInfo :
	TulingCodeErrorInfo = 40002
	//TulingCodeErrorDeplete :
	TulingCodeErrorDeplete = 40004
	//TulingCodeErrorFormat :
	TulingCodeErrorFormat = 40007
)

//TulingResponseCode :
type TulingResponseCode struct {
	Code int
}

//TulingResponseText :
type TulingResponseText struct {
	Code int
	Text string
}

//TulingResponseURL :
type TulingResponseURL struct {
	Code int
	Text string
	URL  string
}

//TulingResponseNews :
type TulingResponseNews struct {
	Code int
	Text string
	List []struct {
		Article   string
		Source    string
		Icon      string
		DetailURL string
	}
}

func (trc *TulingResponseCode) getCode(body []byte) (int, error) {
	err := json.Unmarshal(body, &trc)
	if err != nil {
		return -1, err
	}
	return trc.Code, nil
}

func getData(trc TulingResponseCode, body []byte) (interface{}, error) {
	switch trc.Code {
	case TulingCodeText:
		var trt TulingResponseText
		err := json.Unmarshal(body, &trt)
		if err != nil {
			return nil, err
		}
		return trt, nil
	case TulingCodeURL:
		var tru TulingResponseURL
		err := json.Unmarshal(body, &tru)
		if err != nil {
			return nil, err
		}
		return tru, nil
	case TulingCodeNews:
		var trn TulingResponseNews
		err := json.Unmarshal(body, &trn)
		if err != nil {
			return nil, err
		}
		return trn, nil
	case TulingCodeErrorKey:
		return nil, errors.New("参数key错误")
	case TulingCodeErrorInfo:
		return nil, errors.New("请求内容info为空")
	case TulingCodeErrorDeplete:
		return nil, errors.New("当天请求次数已使用完")
	case TulingCodeErrorFormat:
		return nil, errors.New("数据格式异常")
	}
	return nil, errors.New("未知标识码")
}

func main() {
	session.Global.Close()
	session.Global = session.NewCookieManagerOptions(
		session.NewInMemStore(),
		&session.CookieMngrOptions{AllowHTTP: true},
	)
	run(":8080")
}

func run(port string) {
	setRouter()
	log.Println("start listening on", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func setRouter() {
	r := mux.NewRouter()
	r.HandleFunc("/login", login).Methods("GET")
	r.HandleFunc("/login/submit", ajaxLogin).Methods("POST")
	r.HandleFunc("/logout", ajaxLogout).Methods("GET")
	r.HandleFunc("/choose", choose).Methods("GET")
	r.HandleFunc("/chat/boy1", chat).Methods("GET")
	r.HandleFunc("/chat/boy2", chat).Methods("GET")
	r.HandleFunc("/chat/boy3", chat).Methods("GET")
	r.HandleFunc("/chat/girl1", chat).Methods("GET")
	r.HandleFunc("/chat/girl2", chat).Methods("GET")
	r.HandleFunc("/chat/girl3", chat).Methods("GET")
	r.HandleFunc("/chat/new", ajaxChatNew).Methods("POST")
	r.HandleFunc("/chat/like", ajaxChatLike).Methods("POST")
	r.HandleFunc("/show", show).Methods("GET")
	r.HandleFunc("/report", report).Methods("GET")
	// r.HandleFunc("/skill", skill).Methods("GET")
	// r.HandleFunc("/knowledge", knowledge).Methods("GET")
	r.HandleFunc("/map", mapPage).Methods("GET")
	http.Handle("/", r)
	http.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(
				http.Dir("static"),
			),
		),
	)
}

func verifyID(id string) bool {
	ok, _ := regexp.MatchString(`^20[01][0-9]{7}$`, id)
	return ok
}

func verifyPassword(pw string) bool {
	ok, _ := regexp.MatchString(`^[[:graph:]]{6,20}$`, pw)
	return ok
}

func verifyLoginForm(id, pw string) bool {
	return verifyID(id) && verifyPassword(pw)
}

func getIDFromSession(r *http.Request) string {
	sess := session.Get(r)
	if sess == nil {
		return ""
	}
	studentID := sess.CAttr("studentID").(string)
	if !verifyID(studentID) {
		return ""
	}
	return studentID
}

func newPage(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	verify bool,
) {
	log.Printf("[%s]   %s   %s", r.Method, r.URL.RequestURI(), r.RemoteAddr)
	m := map[string]interface{}{}
	if verify {
		studentID := getIDFromSession(r)
		if studentID == "" {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		m["StudentID"] = studentID
	}
	if name == "chat" {
		if strings.Index(r.URL.RequestURI(), "boy3") > -1 {
			m["Name"] = "社团达人"
			m["Path"] = "boy3.png"
			m["Question1"] = "无人机协会。http://sua.ccnu.edu.cn/info/1015/2062.htm"
			m["Question2"] = `社团招新时间在10月份“百团大战”。`
			m["Question3"] = "TEDxCCNU团队、华中师范大学创行团队、春晖社、全纳服务队、心心火义教之家、华心自强社、社工协会、心理协会、圣兵爱心社、绿丝带爱心驿站、春野环保协会、国旗护卫队、心语爱心社、党团先锋队天域爱心社。"
			m["Question4"] = ""
			m["Question5"] = ""
		}
		if strings.Index(r.URL.RequestURI(), "boy2") > -1 {
			m["Name"] = "运动男神"
			m["Path"] = "boy2.png"
			m["Question1"] = "佑铭体育场（400米标准跑道），高职体育馆（300米）标准跑道。"
			m["Question2"] = "佑铭体育场"
			m["Question3"] = "华中师范大学室外游泳馆，电话：15007107895，营业时间：09:00-21:00 周一至周日"
			m["Question4"] = ""
			m["Question5"] = ""
		}
		if strings.Index(r.URL.RequestURI(), "boy1") > -1 {
			m["Name"] = "主席学长"
			m["Path"] = "boy1.png"
			m["Question1"] = "学生会办公室、学术与创新中心、宣传中心、维权与服务中心、文体中心、财务中心以及调研与评估中心。"
			m["Question2"] = "新生军训期间交报名表，军训之后面试。"
			m["Question3"] = "华大电视台，华大青年，华大桂生、华大在线、I华大。"
			m["Question4"] = "不用谢，学生会的宗旨就是“真诚沟通，服务全体”。"
			m["Question5"] = ""
		}
		if strings.Index(r.URL.RequestURI(), "girl2") > -1 {
			m["Name"] = "生活能手"
			m["Path"] = "girl2.png"
			m["Question1"] = "东一食堂、学子餐厅、东二食堂、桂香园、沁园春、博雅园、北区食堂、南湖食堂"
			m["Question2"] = "东区学子超市，东区购物广告，沁园春超市，图书馆天翼爱心超市"
			m["Question3"] = "东十六学生宿舍为男生4人间标准素质，上床下桌内设独立洗澡间，宿舍分布如图所示："
			m["Question4"] = "圆通快递，申通快递，中通快递，韵达快递，顺丰快递，百世汇通，EMS快递。"
			m["Question5"] = ""
		}
		if strings.Index(r.URL.RequestURI(), "girl1") > -1 {
			m["Name"] = "学霸学姐"
			m["Path"] = "girl1.png"
			m["Question1"] = "转专业有两次机会。每年的11月组织报名，大一大二均可报名参加。"
			m["Question2"] = "转专业是在每年11月组织报名，跨学院转专业学生统一网上报名，统一参加考试。具体要等待教务处通知啊"
			m["Question3"] = "有志者，事竟成，努力学习就不难的。"
			m["Question4"] = "每年5月1日到10月1日之间不断电，其余时间十一点半以后断电"
			m["Question5"] = ""
		}
		if strings.Index(r.URL.RequestURI(), "girl3") > -1 {
			m["Name"] = "知心姐姐"
			m["Path"] = "girl3.png"
			m["Question1"] = ""
			m["Question2"] = ""
			m["Question3"] = ""
			m["Question4"] = ""
			m["Question5"] = ""
		}
		m["Color"] = "#" + fmt.Sprintf("%x", rand.Int63())[:6]
	}
	t, _ := template.ParseFiles(htmlDocPath + name + ".html")
	t.Execute(w, m)
}

func login(w http.ResponseWriter, r *http.Request) {
	newPage(w, r, "login", false)
}

func show(w http.ResponseWriter, r *http.Request) {
	newPage(w, r, "show", false)
}

func report(w http.ResponseWriter, r *http.Request) {
	newPage(w, r, "report", false)
}

func choose(w http.ResponseWriter, r *http.Request) {
	newPage(w, r, "choose", true)
}

func chat(w http.ResponseWriter, r *http.Request) {
	newPage(w, r, "chat", true)
}

func skill(w http.ResponseWriter, r *http.Request) {
	newPage(w, r, "skill", true)
}

func knowledge(w http.ResponseWriter, r *http.Request) {
	newPage(w, r, "knowledge", true)
}

func mapPage(w http.ResponseWriter, r *http.Request) {
	newPage(w, r, "map", true)
}

func ajaxLogin(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s]   %s   %s", r.Method, r.URL.RequestURI(), r.RemoteAddr)
	r.ParseForm()
	studentID := r.FormValue("studentID")
	password := r.FormValue("password")
	if !verifyLoginForm(studentID, password) {
		io.WriteString(w, "login:illegal")
		return
	}
	if password != "123456" {
		io.WriteString(w, "login:invalid")
		return
	}
	sess := session.Get(r)
	sess = session.NewSessionOptions(&session.SessOptions{
		CAttrs:  map[string]interface{}{"studentID": studentID},
		Timeout: time.Hour * 2,
	})
	session.Add(sess, w)
	io.WriteString(w, "login:success")
}

func ajaxLogout(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s]   %s   %s", r.Method, r.URL.RequestURI(), r.RemoteAddr)
	sess := session.Get(r)
	if sess != nil {
		session.Remove(sess, w)
	}
}

func ajaxChatNew(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s]   %s   %s", r.Method, r.URL.RequestURI(), r.RemoteAddr)
	studentID := getIDFromSession(r)
	if studentID == "" {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	r.ParseForm()
	info := r.FormValue("info")
	requestData := url.Values{
		"key":  {TulingRequestKey},
		"info": {info},
	}
	resp, err := http.PostForm(TulingRequestURL, requestData)
	if err != nil {
		log.Println(err)
		io.WriteString(w, "Unknown Error")
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	io.WriteString(w, string(body))
}

func ajaxChatLike(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s]   %s   %s", r.Method, r.URL.RequestURI(), r.RemoteAddr)

}
