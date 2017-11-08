package main

import (
	"encoding/json"
	"fmt"
	"github.com/xjdrew/xxc"
	"log"
	"net/http"
	"sync"
)

type UserManager struct {
	lock  sync.Mutex
	hosts map[string](map[string]*User)
}

var g_user_mgr UserManager

type UserError struct {
	host string
	user string
	msg  string
}

func (e UserError) Error() string {
	return fmt.Sprintf("host: %s, user:%s, msg:%s", e.host, e.user, e.msg)
}

func __get_user(mgr UserManager, host string, user string) *User {
	host_info := mgr.hosts[host]
	if host_info == nil {
		return nil
	}
	return host_info[user]
}

func __add_user(mgr UserManager, host string, user string, user_obj *User) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	host_info := mgr.hosts[host]
	if host_info == nil {
		mgr.hosts[host] = make(map[string]*User)
		host_info = mgr.hosts[host]
	}
	host_info[user] = user_obj
	log.Printf("add user succ, host:%s, user:%s", host, user)
}

func __del_user(mgr UserManager, host string, user string) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	host_info := mgr.hosts[host]
	if host_info != nil {
		_, ok := host_info[user]
		if ok {
			delete(host_info, user)
			log.Printf("del user succ, host:%s, user:%s", host, user)
		}
	}
}

func _get_user(host string, user string, password string) (*User, error) {
	// read from cache
	user_obj := __get_user(g_user_mgr, host, user)
	if user_obj != nil {
		if user_obj.config.Password == password {
			return user_obj, nil
		} else {
			return nil, UserError{host, user, "password unmatch"}
		}
	}

	config := &xxc.ClientConfig{}
	config.Host = host
	config.User = user
	config.Password = password
	user_obj, err := CreateUser(config)

	if err != nil {
		msg := fmt.Sprintf("create user failed: %s", err)
		return user_obj, UserError{host, user, msg}
	}

	// defer user_obj.Fini()
	// TODO:客户端退出时，需要删除

	// write to cache
	__add_user(g_user_mgr, host, user, user_obj)
	return user_obj, err
}

func _retpack_json_response(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	s, err := json.Marshal(data)
	if err != nil {
		return
	}
	fmt.Fprintf(w, string(s))
}

func logic_init() {
	g_user_mgr.hosts = make(map[string](map[string]*User))
	http.HandleFunc("/help", handle_help)
	http.HandleFunc("/sayto", handle_sayto)
}

func handle_help(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "help")
}

func handle_sayto(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	host := req.Form["host"][0]
	user := req.Form["user"][0]
	pwd := req.Form["password"][0]
	user_obj, err := _get_user(host, user, pwd)
	if err != nil {
		ret := map[string]interface{}{
			"code": 10000,
			"msg":  fmt.Sprintf("%s", err),
		}
		log.Printf("host:%s, user:%s, get user failed: %s", host, user, err)
		_retpack_json_response(w, ret)
		return
	}

	account := req.Form["account"]
	groupid := req.Form["groupid"]
	content := req.Form["content"]

	if account != nil && content != nil {
		err := user_obj.SayToUser(account[0], content[0])
		if err != nil {
			log.Printf("host:%s, user:%s, sayto account:%s failed: %s", host, user, account[0], err)
			ret := map[string]interface{}{
				"code": 500,
				"msg":  fmt.Sprintf("%s", err),
			}
			_retpack_json_response(w, ret)
			__del_user(g_user_mgr, host, user)
			return
		}
	}

	if groupid != nil && content != nil {
		err := user_obj.SayToGroup(groupid[0], content[0])
		if err != nil {
			log.Printf("host:%s, user:%s, sayto group:%s failed: %s", host, user, groupid[0], err)
			ret := map[string]interface{}{
				"code": 500,
				"msg":  fmt.Sprintf("%s", err),
			}
			_retpack_json_response(w, ret)
			__del_user(g_user_mgr, host, user)
			return
		}
	}

	ret := map[string]interface{}{
		"code": 200,
		"msg":  "ok",
	}
	_retpack_json_response(w, ret)
	return
}
