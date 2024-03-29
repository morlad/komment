package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/cgi"
	"net/smtp"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	LOG_PATH = "komment.log"

	LOG_MODE = 0664
	LOG_FLAG = os.O_WRONLY | os.O_CREATE | os.O_SYNC | os.O_APPEND

	COMMENT_FLAG = os.O_CREATE | os.O_WRONLY | os.O_EXCL
	COMMENT_MODE = 0664

	LIMIT_COMMENTS = 500

	COOKIE_PREFIX = "komment_ownership_"
)

type Configuration struct {
	CgiPath       string `json:"CgiPath"`
	MessagesPath  string `json:"MessagesPath"`
	TemplatePath  string `json:"TemplatePath"`
	EditWindow    int    `json:"EditWindow"`
	DateFormat    string `json:"DateFormat"`
	SmtpHostname  string `json:"SmtpHostname"`
	SmtpPort      int    `json:"SmtpPort"`
	SmtpUser      string `json:"SmtpUser"`
	SmtpPassword  string `json:"SmtpPassword"`
	SmtpFrom      string `json:"SmtpFrom"`
	SmtpTo        string `json:"SmtpTo"`
	ListenOn      string `json:"ListenOn"`
	MaxLength     int    `json:"MaxLength"`
	MaxNameLength int    `json:"MaxNameLength"`
	Whitelist     string `json:"Whitelist"`
	IdValidator   string `json:"IdValidator"`
	BucketSize    int64  `json:"BucketSize"`
	TokenRate     int64  `json:"TokenRate"`
}

var g_config Configuration

type TokenBucket struct {
	Tokens  int64
	LastAdd int64
}

var g_token TokenBucket

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Fprintf(os.Stderr, "%s\t=\t%v\n", what, time.Since(start))
	}
}

func emit_status(w http.ResponseWriter, status_code int, msg string) {
	w.WriteHeader(status_code)
	fmt.Fprintf(w, "%v", msg)
}

type Comment struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Date    string `json:"date"`
	Stamp   string `json:"stamp"`
	Deleted bool   `json:"deleted"`
}

func uid_gen(r *http.Request, komment_id string) string {

	buf := bytes.NewBufferString(komment_id)
	buf.WriteString(strconv.FormatInt(time.Now().UnixNano(), 16))
	buf.WriteString(r.FormValue("message"))
	buf.WriteString(r.RemoteAddr)
	buf.WriteString(strconv.FormatUint(rand.Uint64(), 16))

	h := sha256.New()
	h.Write(buf.Bytes())
	return fmt.Sprintf("%x", h.Sum(nil))
}

func sanitize_komment_id(in string) string {
	rex := regexp.MustCompile("(^\\.|[/\r\n\t])")
	out := rex.ReplaceAllLiteralString(in, "_")
	return strings.ToLower(out)
}

func sanitize_message(in string) string {
	in = strings.Replace(in, "\r", "", -1)
	rex := regexp.MustCompile("\n{3,}")
	out := rex.ReplaceAllLiteralString(in, "\n\n")
	if g_config.MaxLength > 0 && len(out) > g_config.MaxLength {
		out = out[:g_config.MaxLength] + " ..."
	}
	return out
}

func sanitize_name(in string) string {
	if g_config.MaxNameLength > 0 && len(in) > g_config.MaxNameLength {
		return in[:g_config.MaxNameLength] + " ..."
	} else {
		return in
	}
}

func min(a, b int64) int64 {
	if a <= b {
		return a
	} else {
		return b
	}
}

func handler(w http.ResponseWriter, r *http.Request) {

	request := r.FormValue("r")
	raw_komment_id := r.FormValue("komment_id")
	komment_id := sanitize_komment_id(r.FormValue("komment_id"))

	// append new message
	if request == "a" {

		defer elapsed("append: " + komment_id)()

		// rate limit - if enabled
		if g_config.TokenRate > 0 && g_config.BucketSize > 0 {
			now := time.Now().Unix()
			// initialize bucket, if necessary
			new_tokens := (now - g_token.LastAdd) / g_config.TokenRate
			g_token.Tokens = min(g_token.Tokens+new_tokens, g_config.BucketSize)
			g_token.LastAdd += new_tokens * g_config.TokenRate
			if g_token.Tokens > 0 {
				g_token.Tokens -= 1
			} else {
				fmt.Fprintf(os.Stderr, "%v - (%v - %v)", g_config.TokenRate, now, g_token.LastAdd)
				timeout := g_config.TokenRate - (now - g_token.LastAdd)
				emit_status(w, 429, "Rate Limited: wait "+strconv.Itoa(int(timeout))+" seconds and retry")
				return
			}
		}

		// validate komment_id
		var is_valid_id bool = g_config.Whitelist == "" && g_config.IdValidator == ""
		if !is_valid_id && g_config.Whitelist != "" {
			file, err := os.Open(g_config.Whitelist)
			if err != nil {
				panic(err.Error())
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				matched, err := regexp.MatchString(scanner.Text(), komment_id)
				if err == nil && matched {
					is_valid_id = true
					break
				}
			}
		}
		if !is_valid_id && g_config.IdValidator != "" {
			cmd := exec.Command(g_config.IdValidator, raw_komment_id, komment_id)
			err := cmd.Run()
			if err == nil {
				is_valid_id = true
			} else {
				w.WriteHeader(403)
				fmt.Fprintf(os.Stderr, "IdValidator: %v\n", err)
				return
			}
		}
		if !is_valid_id {
			w.WriteHeader(403)
			return
		}

		var comment Comment
		comment.Comment = sanitize_message(r.FormValue("message"))
		comment.Name = sanitize_name(r.FormValue("name"))
		comment.Date = time.Now().UTC().Format(time.RFC3339)
		comment.Stamp = uid_gen(r, komment_id)

		b, err := json.Marshal(comment)
		if err != nil {
			panic(err.Error())
		}
		number := 1
		var f *os.File
		// make sure the requested directory exists
		os.MkdirAll(fmt.Sprintf("%v/%v", g_config.MessagesPath, komment_id), 0755)
		for ; number <= LIMIT_COMMENTS; number += 1 {
			f, err = os.OpenFile(
				fmt.Sprintf("%v/%v/%v.json", g_config.MessagesPath, komment_id, number),
				COMMENT_FLAG,
				COMMENT_MODE)
			if err == nil {
				break
			} else if !os.IsExist(err) {
				emit_status(w, 500, err.Error())
				return
			}
		}
		f.Write(b)
		f.Close()

		w.Header().Set("Content-Type", "text/json")

		var cookie http.Cookie
		cookie.Name = COOKIE_PREFIX + comment.Stamp
		cookie.Value = "true"
		cookie.MaxAge = g_config.EditWindow
		http.SetCookie(w, &cookie)

		// send mail
		if g_config.SmtpTo != "" {
			mail_msg := "Id: " + raw_komment_id + "\r\n" + comment.Comment
			mail_to := []string{g_config.SmtpTo}
			msg := []byte("To: " + g_config.SmtpTo + "\r\n" +
				"From: " + g_config.SmtpFrom + "\r\n" +
				"Subject: New Comment!\r\n" +
				"\r\n" +
				mail_msg + "\r\n")
			auth := smtp.PlainAuth("", g_config.SmtpUser, g_config.SmtpPassword, g_config.SmtpHostname)
			err := smtp.SendMail(g_config.SmtpHostname+":"+strconv.Itoa(g_config.SmtpPort), auth, g_config.SmtpFrom, mail_to, msg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "SendMail: %v\n", err)
			}
		}

		w.WriteHeader(200)
		fmt.Fprintf(w, "{ \"result\": %v }\n", number)

		// count
	} else if request == "c" {

		defer elapsed("count: " + komment_id)()

		// load template
		templ, err := template.ParseFiles(g_config.TemplatePath + "/count.html.tmpl")
		if err != nil {
			emit_status(w, 500, err.Error())
			return
		}
		type CountTemplateData struct {
			Count int
		}
		var tdata CountTemplateData

		// open file
		path := g_config.MessagesPath + "/" + komment_id
		jsonpath, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(200)
				tdata.Count = 0
				err = templ.Execute(w, tdata)
				if err != nil {
					emit_status(w, 500, err.Error())
					return
				}
				return
			} else {
				emit_status(w, 500, err.Error())
				return
			}
		}
		defer jsonpath.Close()

		names, err := jsonpath.Readdirnames(0)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		tdata.Count = len(names)
		err = templ.Execute(w, tdata)
		if err != nil {
			emit_status(w, 500, err.Error())
			return
		}

		// form
	} else if request == "form" {

		defer elapsed("form: " + komment_id)()

		// load template
		templ, err := template.ParseFiles(g_config.TemplatePath + "/form.html.tmpl")
		if err != nil {
			emit_status(w, 500, err.Error())
			return
		}
		type FormTemplateData struct {
			CgiPath string
		}
		var tdata FormTemplateData

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		tdata.CgiPath = g_config.CgiPath
		err = templ.Execute(w, tdata)
		if err != nil {
			emit_status(w, 500, err.Error())
			return
		}

		// script
	} else if request == "script" {

		defer elapsed("script: " + komment_id)()

		// load template
		templ, err := template.ParseFiles(g_config.TemplatePath + "/frontend.js.tmpl")
		if err != nil {
			emit_status(w, 500, err.Error())
			return
		}
		type FormTemplateData struct {
			CgiPath string
		}
		var tdata FormTemplateData

		w.Header().Set("Content-Type", "application/javascript")
		w.WriteHeader(200)
		tdata.CgiPath = g_config.CgiPath
		err = templ.Execute(w, tdata)
		if err != nil {
			emit_status(w, 500, err.Error())
			return
		}

		// list all messages
	} else if request == "l" {

		defer elapsed("list: " + komment_id)()

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)

		templ, err := template.ParseFiles(g_config.TemplatePath + "/message.html.tmpl")
		if err != nil {
			emit_status(w, 500, err.Error())
			return
		}

		for number := 1; number <= LIMIT_COMMENTS; number += 1 {
			content, err := ioutil.ReadFile(fmt.Sprintf("%v/%v/%v.json", g_config.MessagesPath, komment_id, number))
			if err != nil {
				if os.IsNotExist(err) {
					break // reached end of files
				} else {
					emit_status(w, 500, err.Error())
					return
				}
			}
			var comment Comment
			err = json.Unmarshal(content, &comment)
			if err != nil {
				emit_status(w, 500, err.Error())
				return
			}
			cookie, err := r.Cookie(COOKIE_PREFIX + comment.Stamp)

			type CommentTemplateData struct {
				Name       template.HTML
				Comment    template.HTML
				RawComment string
				CanEdit    bool
				MessageId  string
				KommentId  string
				Deleted    bool
				CgiPath    string
				Date       string
			}

			var tdata CommentTemplateData
			tdata.Name = template.HTML(template.HTMLEscapeString(comment.Name))
			html_comment := template.HTMLEscapeString(comment.Comment)
			html_comment = strings.Replace(html_comment, "\n", "<br/>", -1)
			tdata.Comment = template.HTML(html_comment)
			tdata.RawComment = comment.Comment
			tdata.KommentId = raw_komment_id
			tdata.Deleted = comment.Deleted
			tdata.CgiPath = g_config.CgiPath
			tdata.MessageId = fmt.Sprintf("%v", number)
			date, err := time.Parse(time.RFC3339, comment.Date)
			tdata.Date = date.Format(g_config.DateFormat)
			if cookie != nil {
				tdata.CanEdit = true
			}
			err = templ.Execute(w, tdata)
			if err != nil {
				emit_status(w, 500, err.Error())
				return
			}
		}

		// edit messages
	} else if request == "e" {

		message_id := r.FormValue("message_id")
		path := fmt.Sprintf("%s/%v/%v.json", g_config.MessagesPath, komment_id, message_id)

		new_comment := sanitize_message(r.FormValue("message"))

		content, err := ioutil.ReadFile(path)
		if err != nil {
			emit_status(w, 500, err.Error())
			return
		}
		var comment Comment
		err = json.Unmarshal(content, &comment)
		if err != nil {
			emit_status(w, 500, err.Error())
			return
		}
		_, err = r.Cookie(COOKIE_PREFIX + comment.Stamp)
		if err != nil {
			emit_status(w, 403, "Not authorized to edit message.")
			return
		}
		comment.Comment = new_comment
		b, err := json.Marshal(comment)
		if err != nil {
			emit_status(w, 500, err.Error())
			return
		}
		file, err := os.Create(path)
		file.Write(b)

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)

		// no request type -> error
	} else {
		emit_status(w, 500, "Invalid Request")
		return
	}
}

func main() {

	// redirect <stderr> to logfile
	logFile, err := os.OpenFile(LOG_PATH, LOG_FLAG, LOG_MODE)
	if err != nil {
		log.Fatalln(err)
	}
	syscall.Dup2(int(logFile.Fd()), 2)

	content, err := ioutil.ReadFile("config/komment.json")
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal(content, &g_config)
	if err != nil {
		log.Fatalln(err)
	}

	// serve
	if g_config.ListenOn != "" {
		http.HandleFunc("/", handler)
		err := http.ListenAndServe(g_config.ListenOn, nil)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		err = cgi.Serve(http.HandlerFunc(handler))
		if err != nil {
			log.Fatalln(err)
		}
	}
}
