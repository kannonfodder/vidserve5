docker buildx build --platform "linux/arm64,linux/amd64" --tag docker.kannonfoundry.dev/api-go --push .
kubectl rollout restart deployment --namespace vids api-go