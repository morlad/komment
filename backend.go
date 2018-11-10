package main

import (
  "fmt"
  "os"
  "syscall"
  "net/http"
  "net/http/cgi"
  "encoding/json"
  "math/rand"
  "io/ioutil"
  "time"
  "html/template"
  "strconv"
  "crypto/sha256"
  "bytes"
)

const LOG_PATH = "komment.log"
const LOG_MODE = 0664
const LOG_FLAG = os.O_WRONLY | os.O_CREATE | os.O_SYNC | os.O_APPEND

const COMMENT_PATH = "comments"
const COMMENT_FLAG = os.O_CREATE | os.O_WRONLY | os.O_EXCL
const COMMENT_MODE = 0664

const LIMIT_COMMENTS = 500

// in seconds
const COOKIE_EDIT_WINDOW = 300
const COOKIE_PREFIX = "komment_ownership_"


func elapsed(what string) func() {
  start := time.Now()
  return func() {
    fmt.Fprintf(os.Stderr, "%s -> %v\n", what, time.Since(start))
  }
}


func emit_status_500(msg string) {

  fmt.Printf("Status: 500 Script Error\r\n")
  fmt.Printf("Content-Type: text/plain\r\n")
  fmt.Printf("\r\n")
  fmt.Printf("%s\r\n", msg)
  os.Exit(500)
}


type Comment struct {
  Name string `json:"name"`
  Comment string `json:"comment"`
  Stamp string `json:"stamp"`
  Deleted bool `json:"deleted"`
}

type CommentTemplateData struct {
  Name string
  Comment string
  CanEdit bool
  MessageId string
  KommentId string
  Deleted bool
}

func uid_gen(r *http.Request) string {

  buf := bytes.NewBufferString(r.FormValue("komment_id"))
  buf.WriteString(strconv.FormatInt(time.Now().UnixNano(), 16))
  buf.WriteString(r.FormValue("comment"))
  buf.WriteString(r.RemoteAddr)
  buf.WriteString(strconv.FormatUint(rand.Uint64(), 16))

  h := sha256.New()
  h.Write(buf.Bytes())
  return fmt.Sprintf("%x", h.Sum(nil))
}

func handler(w http.ResponseWriter, r *http.Request) {

  request := r.FormValue("r")
  komment_id := r.FormValue("komment_id")

  // append new comment
  if request == "a" {

    defer elapsed("append:"+komment_id)()

    var comment Comment
    comment.Comment = r.FormValue("comment")
    comment.Name = r.FormValue("name")
    comment.Stamp = uid_gen(r)

    b, err := json.Marshal(comment)
    if err != nil {
      panic(err.Error())
    }
    number := 1
    var f *os.File
    // make sure the requested directory exists
    os.MkdirAll(fmt.Sprintf("%v/%v", COMMENT_PATH, komment_id), 0755)
    for ; number <= LIMIT_COMMENTS; number += 1 {
      f, err = os.OpenFile(
        fmt.Sprintf("%v/%v/%v.json", COMMENT_PATH, komment_id, number),
        COMMENT_FLAG,
        COMMENT_MODE)
      if err == nil {
        break
      } else if !os.IsExist(err) {
        emit_status_500(err.Error())
      }
    }
    f.Write(b)
    f.Close()

    w.Header().Set("Content-Type", "text/json")

    var cookie http.Cookie
    cookie.Name = COOKIE_PREFIX + comment.Stamp
    cookie.Value = "true"
    cookie.MaxAge = COOKIE_EDIT_WINDOW
    http.SetCookie(w, &cookie)

    w.WriteHeader(200)
    fmt.Fprintf(w, "{ \"result\": %v }\n", number)

  // count
  } else if request == "c" {

    defer elapsed("count:"+komment_id)()

    // load template
    template, err := template.ParseFiles("count.html.tmpl")
    if err != nil {
      emit_status_500(err.Error())
    }
    type CountTemplateData struct {
      Count int
    }
    var tdata CountTemplateData

    // open file
    path := COMMENT_PATH + "/" + komment_id
    jsonpath, err := os.Open(path)
    if err != nil {
      if os.IsNotExist(err) {
        w.Header().Set("Content-Type", "text/html")
        w.WriteHeader(200)
        tdata.Count = 0
        err = template.Execute(w, tdata)
        if err != nil {
          emit_status_500(err.Error())
        }
        return
      } else {
        emit_status_500(err.Error())
        return
      }
    }
    defer jsonpath.Close()

    names, err := jsonpath.Readdirnames(0)
    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(200)
    tdata.Count = len(names)
    err = template.Execute(w, tdata)
    if err != nil {
      emit_status_500(err.Error())
    }

  // list all comments
  } else if request == "l" {

    defer elapsed("list:"+komment_id)()

    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(200)

    template, err := template.ParseFiles("comment.html.tmpl")
    if err != nil {
      emit_status_500(err.Error())
    }

    for number := 1; number <= LIMIT_COMMENTS; number += 1 {
      content, err := ioutil.ReadFile(fmt.Sprintf("%v/%v/%v.json", COMMENT_PATH, komment_id, number))
      if err != nil {
        if os.IsNotExist(err) {
          break // reached end of files
        } else {
          emit_status_500(err.Error())
        }
      }
      var comment Comment
      err = json.Unmarshal(content, &comment)
      if err != nil {
        emit_status_500(err.Error())
      }
      cookie, err := r.Cookie(COOKIE_PREFIX + comment.Stamp)

      var tdata CommentTemplateData
      tdata.Comment = comment.Comment
      tdata.Name = comment.Name
      tdata.KommentId = komment_id
      tdata.Deleted = comment.Deleted
      tdata.MessageId = fmt.Sprintf("%v", number)
      if cookie != nil {
        tdata.CanEdit = true
      }
      err = template.Execute(w, tdata)
      if err != nil {
        emit_status_500(err.Error())
      }
    }

  // edit comment
  } else if request == "e" {

    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(200)

    message_id := r.FormValue("message_id")
    path := fmt.Sprintf("%s/%v/%v.json", COMMENT_PATH, komment_id, message_id)

    new_comment := r.FormValue("comment")

    content, err := ioutil.ReadFile(path)
    if err != nil {
      emit_status_500(err.Error())
    }
    var comment Comment
    err = json.Unmarshal(content, &comment)
    if err != nil {
      emit_status_500(err.Error())
    }
    comment.Comment = new_comment
    b, err := json.Marshal(comment)
    if err != nil {
      panic(err.Error())
    }
    file, err := os.Create(path)
    file.Write(b)

  // no request type -> error
  } else {
    emit_status_500("Invalid Request")
  }
}


func main() {

  // redirect <stderr> to logfile
  logFile, err := os.OpenFile(LOG_PATH, LOG_FLAG, LOG_MODE)
  if err != nil {
    emit_status_500(err.Error())
    return
  }
  syscall.Dup2(int(logFile.Fd()), 2)

  // serve
  err = cgi.Serve(http.HandlerFunc(handler))
  if err != nil {
    emit_status_500(err.Error())
  }
}
