package main

import ("io" ; "net/http")

type Hello struct{}

func hello(res http.ResponseWriter, req *http.Request){
    res.Header().Set("Content-Type", "text/html",)
    io.WriteString(res, "hello")
}

// func landing(res http.ResponseWriter, req *http.Request){
//     io.WriteString(res, "landing")
// }

func main() {
    http.HandleFunc("/hello", hello)
    http.ListenAndServe("localhost:8080", nil)
}