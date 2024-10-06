docker build -t redis-operator/container .
docker tag redis-operator/container:latest public.ecr.aws/f1r9h5l7/redis-operator/container:latest
docker push public.ecr.aws/f1r9h5l7/redis-operator/container:latest
#make deploy IMG=public.ecr.aws/f1r9h5l7/redis-operator/container:latest