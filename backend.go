package main

import (
  "fmt"
  "os"
  "syscall"
  "net/http"
  "net/http/cgi"
  "encoding/json"
  "math/rand"
)


const LOG_PATH = "komment.log"
const LOG_MODE = 0664
const LOG_FLAG = os.O_WRONLY | os.O_CREATE | os.O_SYNC

const COMMENT_EXT = "_comments.json"
const COMMENT_FLAG = os.O_APPEND | os.O_CREATE | os.O_WRONLY
const COMMENT_MODE = 0664

// in seconds
const COOKIE_EDIT_WINDOW = 300
const COOKIE_PREFIX = "komment_ownership_"

func emit_status_500(msg string) {

  fmt.Printf("Status: 500 Script Error\r\n")
  fmt.Printf("Content-Type: text/plain\r\n")
  fmt.Printf("\r\n")
  fmt.Printf("%s\r\n", msg)
}


func handler(w http.ResponseWriter, r *http.Request) {

  request := r.FormValue("r")
  komment_id := r.FormValue("komment_id")

  // append new comment
  if request == "a" {

    data := make(map[string]string)
    data["comment"] = r.FormValue("comment")
    data["name"] = r.FormValue("name")
    data["stamp"] = fmt.Sprintf("%x", rand.Uint64())

    b, err := json.Marshal(data)
    if err != nil {
      panic(err.Error())
    }
    f, err := os.OpenFile(komment_id + COMMENT_EXT, COMMENT_FLAG, COMMENT_MODE)
    if err != nil {
      panic(err.Error())
    }
    f.Write(b)

    w.Header().Set("Content-Type", "text/json")

    var cookie http.Cookie
    cookie.Name = COOKIE_PREFIX + data["stamp"]
    cookie.Value = "true"
    cookie.MaxAge = COOKIE_EDIT_WINDOW
    http.SetCookie(w, &cookie)

    w.WriteHeader(200)
    fmt.Fprintln(w, "{ \"result\": \"ok\" }")

  // count
  } else if request == "c" {
  // list all comments
  } else if request == "l" {
  // edit comment
  } else if request == "e" {
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
