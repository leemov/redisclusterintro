#------------ bootstrap the cluster nodes --------------------

start_cmd='redis-server --port 6379 --cluster-enabled yes --requirepass risktechacademy --cluster-config-file nodes.conf --cluster-node-timeout 5000 --appendonly yes'
redis_image='redis:5.0-rc'
network_name='risktechacademy'

docker network create $network_name
echo $network_name " created"

#---------- create the cluster ------------------------

for port in `seq 6379 6384`; do \
 docker run -d --rm --name "redis-"$port -p $port:6379 --net $network_name $redis_image $start_cmd;
 echo "created redis cluster node redis-"$port
done

cluster_hosts=''

for port in `seq 6379 6384`; do \
 hostip=`docker inspect -f '{{(index .NetworkSettings.Networks "risktechacademy").IPAddress}}' "redis-"$port`;
 echo "IP for cluster node redis-"$port "is" $hostip
 cluster_hosts="$cluster_hosts$hostip:6379 ";
done

echo "cluster hosts "$cluster_hosts
echo "creating cluster...."
echo 'yes' | docker run -i --rm --net $network_name $redis_image redis-cli -a risktechacademy --cluster  create $cluster_hosts --cluster-replicas 1;