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
)

const LOG_PATH = "komment.log"
const LOG_MODE = 0664
const LOG_FLAG = os.O_WRONLY | os.O_CREATE | os.O_SYNC

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
    fmt.Fprintf(os.Stderr, "%s took %v\n", what, time.Since(start))
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
}

func handler(w http.ResponseWriter, r *http.Request) {

  request := r.FormValue("r")
  komment_id := r.FormValue("komment_id")

  // append new comment
  if request == "a" {

    var comment Comment
    comment.Comment = r.FormValue("comment")
    comment.Name = r.FormValue("name")
    comment.Stamp = fmt.Sprintf("%x", rand.Uint64())

    b, err := json.Marshal(comment)
    if err != nil {
      panic(err.Error())
    }
    number := 1
    var f *os.File
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

    elapsed("count")()

    // open file
    path := COMMENT_PATH + "/" + komment_id
    jsonpath, err := os.Open(path)
    if err != nil {
      if os.IsNotExist(err) {
        w.WriteHeader(200)
        fmt.Fprint(w, "{ \"count\": 0 }")
        return
      } else {
        emit_status_500(err.Error())
        return
      }
    }
    defer jsonpath.Close()

    names, err := jsonpath.Readdirnames(0)

    w.WriteHeader(200)
    fmt.Fprintf(w, "{ \"count\": %v }", len(names))


  // list all comments
  } else if request == "l" {

    elapsed("list")()

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
      err = template.Execute(w, comment)
      if err != nil {
        emit_status_500(err.Error())
      }
    }

  // edit comment
  } else if request == "e" {
  // no request type -> error
  } else {
    emit_status_500("Invalid Request")
  }
}


func main() {

  elapsed("main")()

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
