apiVersion: batch/v1
kind: Job
metadata:
  generateName: wrk2-__WRK2_NAME__-
spec:
  template:
    spec:
      containers:
      - name: wrk2
        image: bootjp/wrk2
        command:
        - wrk
        - --latency
        - --duration
        - "300"
        - --threads
        - "40"
        - --connections
        - "100"
        - --rate
        - "1000"
        - --header
        - "Host: test-ingressroute0"
        - __WRK2_URL__
      restartPolicy: Never
      nodeSelector:
        workload: wrk2
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            - labelSelector:
                matchExpressions:
                  - key: "app"
                    operator: In
                    values:
                      - wrk2
              topologyKey: kubernetes.io/hostname


