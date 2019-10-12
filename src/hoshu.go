package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	_ "strconv"
	"strings"
	"time"
)

type RequestParameter struct {
	cgi     string
	referer string
	bbs     string
	key     string
	mail    string
	message string
	dat     string
}

var cookiePath string
var cookiePathTemp string
var messagePath string

func init() {
	log.SetPrefix("[LOG] ")
}

func initialize() error {
	// env
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dir := filepath.Dir(exe)
	cookiePath = filepath.Join(dir, "cookie.txt")
	cookiePathTemp = cookiePath + ".temp"
	messagePath = filepath.Join(dir, "message.txt")
	return nil
}

func readCookieFile() ([]string, error) {
	_, err := os.Stat(cookiePath)
	if err != nil {
		// file not exists
		return make([]string, 0), nil
	}
	f, err := os.OpenFile(cookiePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ret := make([]string, 0, 8)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		ret = append(ret, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

func writeCookieFile(reqCookies []string, resCookies []*http.Cookie) error {
	f, err := os.Create(cookiePathTemp)
	if err != nil {
		return err
	}
	defer f.Close()
	wr := bufio.NewWriter(f)

	cookies := make(map[string]string)
	for _, c := range reqCookies {
		cookies[parseCookieNameAndDomain(c)] = c
	}
	for _, c := range resCookies {
		cookies[c.Name+";"+c.Domain] = c.Raw
	}
	for _, cookie := range cookies {
		fmt.Fprintln(wr, cookie)
	}
	wr.Flush()
	return nil
}

func parseCookieNameAndDomain(cookieString string) string {
	re := regexp.MustCompile(`^([^=]+)=`)
	match := re.FindStringSubmatch(cookieString)
	name := match[1]
	domain := ""
	split := strings.Split(cookieString, ";")
	for _, s := range split {
		opt := strings.Split(strings.TrimSpace(s), "=")
		if len(opt) == 2 {
			if opt[0] == "domain" {
				domain = opt[1]
			}
		}
	}
	return name + ";" + domain
}

func readMessageFile() (*RequestParameter, error) {
	content, err := ioutil.ReadFile(messagePath)
	if err != nil {
		return &RequestParameter{}, err
	}
	split := strings.Split(string(content), "\n")

	// http://egg.2ch.net/test/read.cgi/software/1568944812
	re := regexp.MustCompile(`(http://[a-zA-Z0-9\.]+)/test/read\.cgi/([a-zA-Z0-9]+)/([0-9]+)/`)
	match := re.FindStringSubmatch(split[0])

	param := &RequestParameter{
		cgi:     match[1] + "/test/bbs.cgi",
		bbs:     match[2],
		key:     match[3],
		referer: split[0],
		mail:    split[1],
		message: split[2],
		dat:     match[1] + "/" + match[2] + "/dat/" + match[3] + ".dat",
	}
	return param, nil
}

func makeHeadRequest(param *RequestParameter) (*http.Request, error) {
	// request
	req, err := http.NewRequest("HEAD", param.dat, nil)
	if err != nil {
		return nil, err
	}

	// header
	req.Header.Add("User-Agent", "Monazilla/1.00")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Connection", "close")

	return req, nil
}

func makePostRequest(param *RequestParameter, cookies []string) (*http.Request, error) {
	// parameter
	values := url.Values{}
	values.Add("bbs", param.bbs)
	values.Add("key", param.key)
	values.Add("time", "1")
	values.Add("submit", "書き込む")
	values.Add("FROM", "")
	values.Add("mail", param.mail)
	values.Add("MESSAGE", param.message)
	values.Add("time", "1")

	// request
	req, err := http.NewRequest("POST", param.cgi, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	// header
	req.Header.Add("User-Agent", "Monazilla/1.00")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Referer", param.referer)
	req.Header.Add("Connection", "close")

	// cookies
	for _, cookie := range cookies {
		req.Header.Add("Cookie", cookie)
	}

	return req, nil
}

func main() {

	// initialize
	if err := initialize(); err != nil {
		log.Fatalln(err)
	}

	// message
	param, err := readMessageFile()
	if err != nil {
		log.Fatalln(err)
	}

	// cookie
	cookies, err := readCookieFile()
	if err != nil {
		log.Fatalln(err)
	}

	// parse args
	// sendTime, err := time.ParseInLocation("2006-01-02 15:04:05.000", os.Args[1], time.Local)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// delay, err := strconv.Atoi(os.Args[2])
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	var contentLength int64 = 0
	for {

		time.Sleep(5 * time.Minute)

		// request(head)
		req, err := makeHeadRequest(param)
		if err != nil {
			log.Fatalln(err)
		}
		// client
		proxy, _ := url.Parse("http://localhost:8085/")
		client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}
		// send request
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalln(err)
		}
		if contentLength != resp.ContentLength {
			contentLength = resp.ContentLength
			fmt.Println("need not hoshu.", time.Now(), contentLength)
			continue
		}

		// request(post)
		req, err = makePostRequest(param, cookies)
		if err != nil {
			log.Fatalln(err)
		}
		// client
		client = &http.Client{}

		// send request
		resp, err = client.Do(req)
		fmt.Println("Response Time:", time.Now())

		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()

		// parse response
		fmt.Println("Response Status:", resp.Status)
		fmt.Println("Response Header:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Response Body:", string(body))
	}

	// write cookie
	//	writeCookieFile(cookies, resp.Cookies())

	// end
	//	terminate()
}

func terminate() {
	os.Rename(cookiePathTemp, cookiePath)
}
