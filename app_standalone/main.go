package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

type HTTPHandler struct {
	rd          *redis.Pool
	limitConfig map[string]int
}

func (h *HTTPHandler) RateLimit(f http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// read path and phone number
		path := r.URL.Path

		bodyByte, _ := ioutil.ReadAll(r.Body)
		type req struct {
			Phone string `json:"phone"`
		}

		reqData := req{}
		json.Unmarshal(bodyByte, &reqData)

		// assume that format always valid
		if reqData.Phone == "" {
			f(w, r)
			return
		}

		conn := h.rd.Get()
		defer conn.Close()

		_, err := conn.Do("SETNX", "rl:"+reqData.Phone, 0)
		if err == nil {
			conn.Do("EXPIRE", "rl:"+reqData.Phone, 60) // rate limit for 60 secs
		} else {
			fmt.Println(err.Error())
		}

		ctr, err := redis.Int(conn.Do("INCR", "rl:"+reqData.Phone))
		if err != nil {
			fmt.Println(err.Error())
		}
		if ctr > h.limitConfig[path] { // rate limit threshold
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("OOOPS PELAN-PELAN"))
			return
		}

		f(w, r)
	})
}

func (h *HTTPHandler) GenerateOTP(w http.ResponseWriter, r *http.Request) {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	w.Write([]byte(`{"otpcode":` + strconv.Itoa(r1.Intn(1000000)) + `}`)) // randomize otp code

	return
}

func (h *HTTPHandler) GetOTPCounter(w http.ResponseWriter, r *http.Request) {
	phone := r.URL.Query().Get("phone")
	conn := h.rd.Get()
	defer conn.Close()
	ctr, err := redis.Int(conn.Do("GET", "rl:"+phone))
	if err != nil {
		fmt.Println(err.Error())
	}

	w.Write([]byte(`{"counter":` + strconv.Itoa(ctr) + `}`))
}

func main() {
	pool := &redis.Pool{
		MaxIdle:     1000,
		IdleTimeout: 2 * time.Second,
		MaxActive:   1000,
		// Dial or DialContext must be set. When both are set, DialContext takes precedence over Dial.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				return nil, err
			}

			// PASSWORD SETUP
			// if _, err := c.Do("AUTH", "risktechacademy"); err != nil {
			// 	c.Close()
			// 	return nil, err
			// }

			return c, nil
		},
	}

	defer pool.Close()

	hdl := HTTPHandler{
		rd: pool,
		limitConfig: map[string]int{
			"/otp": 5,
		},
	}

	http.HandleFunc("/otp", hdl.RateLimit(hdl.GenerateOTP))
	fmt.Println("start listening on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("error listening on port 80 for http")
	}
}
