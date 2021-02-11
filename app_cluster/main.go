package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

type HTTPHandler struct {
	rd          *redis.ClusterClient
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

		_, err := h.rd.SetNX("rl:"+reqData.Phone, 0, 0).Result()
		if err != nil {
			fmt.Println(err.Error())
		}

		ctr, err := h.rd.Incr("rl:" + reqData.Phone).Result()
		if err != nil {
			fmt.Println(err.Error())
			ctr, err = h.rd.Get("rl:" + reqData.Phone).Int64() // modify line to read cache
			if err != nil {
				fmt.Println(err.Error())
			}
		}
		if ctr > int64(h.limitConfig[path]) { // rate limit threshold
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
	ctr, err := h.rd.Get("rl:" + phone).Int()
	if err != nil {
		fmt.Println(err.Error())
	}

	w.Write([]byte(`{"counter":` + strconv.Itoa(ctr) + `}`))
}

func main() {
	clusterSlots := func() ([]redis.ClusterSlot, error) {
		slots := []redis.ClusterSlot{
			// First node with 1 master and 1 replica.
			{
				Nodes: []redis.ClusterNode{
					{
						Addr: "172.22.0.3:6379",
					},
					{
						Addr: "172.22.0.5:6379",
					},
				},
			},
		}
		return slots, nil
	}

	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		ClusterSlots:  clusterSlots,
		RouteRandomly: true,
		Password:      "risktechacademy",
	})

	defer rdb.Close()

	rdb.Ping()

	// ReloadState reloads cluster state. It calls ClusterSlots func
	// to get cluster slots information.
	err := rdb.ReloadState()
	if err != nil {
		fmt.Println(err.Error())
	}

	hdl := HTTPHandler{
		rd: rdb,
		limitConfig: map[string]int{
			"/otp": 5,
		},
	}

	http.HandleFunc("/otp", hdl.RateLimit(hdl.GenerateOTP))
	http.HandleFunc("/otp/counter", hdl.GetOTPCounter)
	fmt.Println("start listening on port 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("error listening on port 80 for http")
	}
}
