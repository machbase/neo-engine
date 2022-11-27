
- required dependencies

```
    sudo apt-get install libjemalloc-dev
    sudo apt-get install libssl-dev
```

- enable c/c++ plugin of visual studio code

- required docker plugins for cross building

> https://hub.docker.com/r/docker/binfmt/tags

```
docker run --privileged --rm tonistiigi/binfmt --install all
docker pull docker/binfmt:a7996909642ee92942dcd6cff44b9b95f08dad64
```