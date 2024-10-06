docker build -t redis-operator/redis-worker .
docker tag redis-operator/redis-worker:latest public.ecr.aws/f1r9h5l7/redis-operator/redis-worker:latest
docker push public.ecr.aws/f1r9h5l7/redis-operator/redis-worker:latest