# Redis Cluster Setup
To setup redis cluster, simply run this bash command.
Basically the command will run this 3 steps : 

1. Setup 6 redis node with cluste-enabled config set to yes
```
start_cmd='redis-server --port 6379 --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000 --appendonly yes'
redis_image='redis:5.0-rc'
network_name='risktechacademy'

docker network create $network_name
echo $network_name " created"

for port in `seq 6379 6384`; do \
 docker run -d --name "redis-"$port -p $port:6379 --net $network_name $redis_image $start_cmd;
 echo "created redis cluster node redis-"$port
done
```
2. Obtaining node's ips 
```
for port in `seq 6379 6384`; do \
 hostip=`docker inspect -f '{{(index .NetworkSettings.Networks "redis_cluster_net").IPAddress}}' "redis-"$port`;
 echo "IP for cluster node redis-"$port "is" $hostip
 cluster_hosts="$cluster_hosts$hostip:6379 ";
done
```

3. Setup the cluster 
```
echo "cluster hosts "$cluster_hosts
echo "creating cluster...."
echo 'yes' | docker run -i --rm --net $network_name $redis_image redis-cli --cluster create $cluster_hosts --cluster-replicas 1;
```

This way, we can use cluster supported redis client to access the redis. If you use redis-cli, you can run the command using 
```
redis-cli -c -h 127.0.0.1 -p 6379
```