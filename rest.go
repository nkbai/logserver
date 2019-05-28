package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/karlseguin/ccache"

	"github.com/ant0ine/go-json-rest/rest"
)

func Start(port int) {
	api := rest.NewApi()

	api.Use(rest.DefaultCommonStack...)

	router, err := rest.MakeRouter(
		//peer 提交Partner的BalanceProof,更新Partner的余额
		rest.Get("/logsrv/1/assignid", AssignID),
		rest.Post("/logsrv/1/log/:address/:id", Log),
		rest.Post("/logsrv/1/upload",upload),
	)
	if err != nil {
		log.Fatalf("maker router :%s", err)
	}
	api.SetApp(router)
	listen := fmt.Sprintf("0.0.0.0:%d", port)
	log.Fatalf("http listen and serve :%s", http.ListenAndServe(listen, api.MakeHandler()))
}

func AssignID(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(NextID())
}
func upload(w rest.ResponseWriter, r *rest.Request) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	//解压缩
	zr, err := gzip.NewReader(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer zr.Close()
	handler.Filename=path.Base(handler.Filename)
	f, err := os.OpenFile(path.Join(logdir,"upload", handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, zr)
	fmt.Fprintln(w.(http.ResponseWriter), "upload ok!")
}
var cache = ccache.New(ccache.Configure().MaxSize(50).ItemsToPrune(5).OnDelete(func(item *ccache.Item) {
	item.Value().(*os.File).Close()
}))

func Log(w rest.ResponseWriter, r *rest.Request) {
	address := r.PathParam("address")
	id := r.PathParam("id")
	log.Printf("address=%s,id=%s\n", address, id)
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rest.Error(w, "body err", http.StatusConflict)
		return
	}
	doLog(address, id, w, msg)
}

func doLog(address, id string, w rest.ResponseWriter, msg []byte) {
	if len(address) <= 0 || len(id) <= 0 {
		rest.Error(w, "arg error ", http.StatusBadRequest)
		return
	}
	key := fmt.Sprintf("%s-%s", address, id)
	it := cache.Get(key)
	if it == nil {
		filename := fmt.Sprintf("%s-%s.log", address, id)
		idFile, err := os.OpenFile(filepath.Join(logdir, filename), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			rest.Error(w, fmt.Sprintf("OpenFile for file %s err %s", filename, err), http.StatusInternalServerError)
			return
		}
		cache.Set(key, idFile, time.Second)
		it = cache.Get(key)
	}
	it.Value().(*os.File).Write(msg)
}
