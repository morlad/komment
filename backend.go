package main

import (
  "fmt"
  //"os"
  "io/ioutil"
  "net/http"
  "net/http/cgi"
  //"encoding/json" // for config file
)

func errorResponse(code int, msg string) {

  fmt.Printf("Status: %d Script Error\r\n", code)
  fmt.Printf("Content-Type: text/plain\r\n")
  fmt.Printf("Connection: close\r\n")
  fmt.Printf("\r\n")
  fmt.Printf("%s\r\n", msg)
  return
}

func myhandler(w http.ResponseWriter, r *http.Request) {

  r.Header.Set("Content-Type:", "text/json")

  r.ParseForm()
  komment := r.Form["komment"][0]
  komment_id := r.Form["komment_id"][0]
  komment2 := []byte(komment)

  //b, err := json.Marshal(komment)
  //if err != nil {
    ioutil.WriteFile("comment"+komment_id+".txt", komment2, 0644)
  //}
}

func main() {

  if err := cgi.Serve(http.HandlerFunc(myhandler)); err != nil {
    d1 := []byte("script error\n")
    ioutil.WriteFile("logfile.txt", d1, 0644)
    errorResponse(500, "Script Error " + err.Error())
  }
}
