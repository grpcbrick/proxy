package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	grpc "google.golang.org/grpc"
)

var clientPool map[string]*grpc.ClientConn

func init() {
	clientPool = make(map[string]*grpc.ClientConn, 10)
}

func main() {
	http.HandleFunc("/", ProxyHandler)
	http.ListenAndServe("0.0.0.0:8000", nil)
}

// ProxyHandler ProxyHandler
func ProxyHandler(response http.ResponseWriter, request *http.Request) {
	var err error
	var args interface{}
	var reply interface{}
	serviceURL := request.Header.Get("service")
	functionURL := request.Header.Get("function")

	buf, err := ioutil.ReadAll(request.Body)
	err = json.Unmarshal(buf, args)
	defer request.Body.Close()

	if err != nil {
		response.WriteHeader(400)
		return
	}

	client, err := CreateGrpcClientConn(serviceURL)
	if err != nil {
		response.WriteHeader(500)
		return
	}

	if err = client.Invoke(context.TODO(), functionURL, args, reply); err != nil {
		response.WriteHeader(500)
		return
	}

	replyJSOM, err := json.Marshal(reply)
	if err != nil {
		response.WriteHeader(500)
		return
	}

	response.Write(replyJSOM)
	response.WriteHeader(200)
}

// CreateGrpcClientConn CreateGrpcClientConn
func CreateGrpcClientConn(url string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	symbol, err := MakeMD5(opts)
	if err != nil {
		return nil, err
	}

	client := clientPool[url+symbol]
	if client != nil {
		return client, nil
	}

	conn, err := grpc.Dial(url, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// MakeMD5 is make md5.
func MakeMD5(target interface{}) (string, error) {
	data, err := json.Marshal(target)
	if err != nil {
		return "", err
	}

	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}
