package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	serv_http   string
	serv_dir    StringFlags
	handle_path []string
)

type StringFlags []string

type neuteredFileSystem struct {
	fs http.FileSystem
}

//所有对目录的请求（没有index.html文件）都将返回404 Not Found响应，而不是目录列表或重定向

func (sfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := sfs.fs.Open(path)

	if err != nil {
		return nil, err
	}

	s, err := f.Stat()

	if s.IsDir() {
		index := strings.TrimSuffix(path, "/") + "/index.html"
		if _, err := sfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}
			return nil, err
		}
	}

	return f, nil
}

func (opt *StringFlags) String() string {
	return fmt.Sprint(*opt)
}

func (opt *StringFlags) Set(value string) error {
	*opt = append(*opt, value)
	return nil
}

func main() {
	flag.Var(&serv_dir, "dir",
		"Dirs for http file server. (e.g. 'list1:/dir1')")
	flag.StringVar(&serv_http, "http", "8880",
		"HTTP service address.")
	flag.Parse()
	//手动加冒号
	if !strings.HasPrefix(serv_http, ":") {
		serv_http = fmt.Sprintf(":%v", serv_http)
	}

	if len(serv_dir) == 0 {
		serv_dir = append(serv_dir, "/:.")
	}
	for _, dir := range serv_dir {
		handle_path = strings.SplitN(dir, ":", 2)

		//是否已 '/' 开头
		if !strings.HasPrefix(handle_path[0], "/") {
			handle_path[0] = "/" + handle_path[0]
		}
		//是否已 '/' 结尾
		if !strings.HasSuffix(handle_path[0], "/") {
			handle_path[0] = handle_path[0] + "/"
		}

		fs := http.FileServer(neuteredFileSystem{http.Dir(handle_path[1])})
		http.Handle(handle_path[0], http.StripPrefix(handle_path[0], fs))

	}
	log.Println("Http server is started at", serv_http)
	log.Println(serv_dir)
	err := http.ListenAndServe(serv_http, nil)
	if err != nil {
		log.Println("ListenAndServe: ", err)

		os.Exit(1)
	}
	return
}
