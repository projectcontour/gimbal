build:
	go install ./cmd/gimbalbench

simplerun:
	gimbalbench run \
      --gimbal-kubecfg-file ${GIMBAL_KUBECONFIG} \
      --backend-kubecfg-file ${BACKEND_KUBECONFIG} \
      --loadgen-kubecfg-file ${LOADGEN_KUBECONFIG} \
      --gimbal-url ${GIMBAL_URL} \
      --test-concurrent-connections 250,500 \
      --test-backend-services 10,25 \
      --test-backend-endpoints 10,25 \
      --test-kubernetes-discovery-time 10,25 \
      --test-gimbal-ingresses 10,25 \
      --test-gimbal-ingressroutes 10,25

fullrun:
	gimbalbench run \
      --gimbal-kubecfg-file ${GIMBAL_KUBECONFIG} \
      --backend-kubecfg-file ${BACKEND_KUBECONFIG} \
      --loadgen-kubecfg-file ${LOADGEN_KUBECONFIG} \
      --gimbal-url ${GIMBAL_URL} \
      --test-concurrent-connections 1000,2500,5000,10000 \
      --test-backend-services 25,50,100,250,500,1000,2500 \
      --test-backend-endpoints 10,25,50,100 \
      --test-kubernetes-discovery-time 100,250,500,1000 \
      --test-gimbal-ingresses 100,250,500,2500,5000 \
      --test-gimbal-ingressroutes 100,250,500,2500,5000

latestreport:
	gimbalbench report logs/$(shell ls -t logs | head -n 1)
