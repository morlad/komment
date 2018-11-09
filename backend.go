package main

import (
  "fmt"
  "os"
  "syscall"
  "io/ioutil"
  "net/http"
  "net/http/cgi"
  "encoding/json" // for config file
)

func emit_status_500(msg string) {

  fmt.Printf("Status: 500 Script Error\r\n")
  fmt.Printf("Content-Type: text/plain\r\n")
  fmt.Printf("\r\n")
  fmt.Printf("%s\r\n", msg)
}

func handler(w http.ResponseWriter, r *http.Request) {

  w.Header().Set("Content-Type", "text/json")
  w.WriteHeader(200)

  r.ParseForm()

  data := make(map[string]string)
  data["komment_id"] = r.Form["komment_id"][0]
  data["comment"] = r.Form["comment"][0]
  data["name"] = r.Form["name"][0]

  b, err := json.Marshal(data)
  if err == nil {
    ioutil.WriteFile("comments_"+data["komment_id"]+".json", b, 0644)
  } else {
    b = []byte(err.Error())
    ioutil.WriteFile("comments_"+data["komment_id"]+".json", b, 0644)
  }
}

func main() {

  // redirect <stderr> to logfile
  logFile, err := os.OpenFile("komment.log", os.O_WRONLY | os.O_CREATE | os.O_SYNC, 0664)
  if err != nil {
    emit_status_500(err.Error())
    return
  }
  syscall.Dup2(int(logFile.Fd()), 2)

  err = cgi.Serve(http.HandlerFunc(handler))
  if err != nil {
    emit_status_500(err.Error())
  }
}
