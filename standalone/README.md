# Redis Standalone

Using docker, we can install redis this way : 
```
docker run -it --rm --name redis -p 6379:6379 redis:5.0-alpine
```

This will run a container named redis using redis 5 image based on linux alpine. We will use -it to make command interactive. After the container stoped, using -rm we will remove the container.

After running the command, we will bind container's port 6379 to host port 6379.