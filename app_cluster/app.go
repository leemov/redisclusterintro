package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-redis/redis"
)

type HTTPHandler struct {
	rd *redis.ClusterClient
}

func (h *HTTPHandler) Count(w http.ResponseWriter, r *http.Request) {
	// ctx := context.Background()
	_, err := h.rd.SetNX("hitcounter2021", 0, 0).Result()
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte("there is an error, " + err.Error()))
		return
	}

	ctr, err := h.rd.Incr("hitcounter2021").Result()
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte("there is an error, " + err.Error()))
		return
	}

	w.Write([]byte(strconv.FormatInt(ctr, 10)))
}

func (h *HTTPHandler) CheckCounter(w http.ResponseWriter, r *http.Request) {
	exist, err := h.rd.Exists("hitcounter2021").Result()
	if exist == 0 {
		w.Write([]byte("0"))
		return
	}

	ctr, err := h.rd.Get("hitcounter2021").Result()
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte("there is an error, " + err.Error()))
		return
	}

	w.Write([]byte(ctr))
}

func main() {
	clusterSlots := func() ([]redis.ClusterSlot, error) {
		slots := []redis.ClusterSlot{
			// First node with 1 master and 1 replica.
			{
				Nodes: []redis.ClusterNode{
					{
						Addr: ":6379", // master
					},
					{
						Addr: ":6379", // replica
					},
				},
			},
		}
		return slots, nil
	}

	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		ClusterSlots:  clusterSlots,
		RouteRandomly: true,
		// Password:      "risktechacademy",
	})

	defer rdb.Close()

	rdb.Ping()

	// ReloadState reloads cluster state. It calls ClusterSlots func
	// to get cluster slots information.
	rdb.ReloadState()

	hdl := HTTPHandler{
		rd: rdb,
	}

	http.HandleFunc("/check", hdl.CheckCounter)
	http.HandleFunc("/test", hdl.Count)
	fmt.Println("start listening on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("error listening on port 80 for http")
	}
}
