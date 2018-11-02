package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/karlseguin/ccache"

	"github.com/ant0ine/go-json-rest/rest"
)

func Start(port int) {
	api := rest.NewApi()

	api.Use(rest.DefaultProdStack...)

	router, err := rest.MakeRouter(
		//peer 提交Partner的BalanceProof,更新Partner的余额
		rest.Get("/logsrv/1/assignid", AssignID),
		rest.Put("/logsrv/1/log/:address/:id", Log),
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

var cache = ccache.New(ccache.Configure().MaxSize(50).ItemsToPrune(5).OnDelete(func(item *ccache.Item) {
	item.Value().(*os.File).Close()
}))

func Log(w rest.ResponseWriter, r *rest.Request) {
	address := r.PathParam("address")
	id := r.PathParam("id")
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
