# Redis Replica
To run redis with replica, we will use 1 master and 1 replica container. master will use port 6379 and replica will use port 6380. 

We will use custom config for redis. The config lies on redis.conf on master folder for master container and replica folder for replica container.

master config : 
```
bind 0.0.0.0
port 6379
masterauth risktechacademy
requirepass risktechacademy
```

### running the redis
go to this directory
To run master node, run this docker command : 
```
docker run --rm -it --name redis0 -p 6379:6379 -v ${PWD}/master:/usr/local/etc/redis/ redis:5.0-alpine redis-server /usr/local/etc/redis/redis.conf
```
${PWD} is current directory

replica config :
```
bind 0.0.0.0
port 6380
masterauth risktechacademy
requirepass risktechacademy
replicaof <master_container's ip> 6379
```

bind 0.0.0.0 is needed to expose it to all interface ( not for production ).

this config replicaof <master_container's ip> 6379 will tell the node to replicate from master
you can get master container's ip by using this command : 
```
docker inspect redis0
```
To run replica node : 
```
docker run --rm -it --name redis1 -p 6380:6380 -v ${PWD}/replica1:/usr/local/etc/redis/ redis:5.0-alpine redis-server /usr/local/etc/redis/redis.conf
```