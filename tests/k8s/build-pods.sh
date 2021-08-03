#!/usr/bin/env bash

set -e

POD_TOTAL=${1:-5}
IS_HEALTH=${2:-"notok"}
POD_TEST_FILE="pods-test.yaml"

# The threshold for notok Pods
NOTOK_THRESHOLD=$((POD_TOTAL/9))

# debug settings
cat << END
  SLEEP_TIME: $SLEEP_TIME
  POD_TOTAL: $POD_TOTAL
  IS_HEALTH: $IS_HEALTH
  POD_TEST_FILE: $POD_TEST_FILE
  NOTOK_THRESHOLD: $NOTOK_THRESHOLD
END

# Add k8s Volume to the Pod spec
function addVolume() {
cat << END
    volumeMounts:
        - name: config-vol
          mountPath: /etc/nginx/conf.d/

  volumes:
    - name: config-vol
      configMap:
        name: nginx
        items:
          - key: default.conf
            path: default.conf
END
}

function addPod() {
cat << END
apiVersion: v1
kind: Pod
metadata:
  name: nginx-${PODNUM}
  labels:
    app: nginx-${PODNUM}
    group: nginx-healthcat
    health: $IS_HEALTH
spec:
  containers:
  - name: nginx-${PODNUM}
    image: nginx:1.14.0
    ports:
    - name: http
      containerPort: 80

END
}

function addService() {
cat << END
apiVersion: v1
kind: Service
metadata:
  name: nginx-${PODNUM}
  labels:
    app: nginx-${PODNUM}
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: http
      protocol: TCP
      name: http
  selector:
    app: nginx-${PODNUM}

END
}

if [[ -f $POD_TEST_FILE  ]]; then
  echo "Removing file $POD_TEST_FILE ..."
  rm -f $POD_TEST_FILE
fi

function addLine() {
  echo
  echo "---"
  echo
}

counter=1
while [[ $counter -le $POD_TOTAL ]];
do
  RANDOM_NUM=$(shuf -i 1-100 -n 1)
  # echo "RANDOM_NUM: $RANDOM_NUM"
  if [[ $RANDOM_NUM -lt $NOTOK_THRESHOLD ]]; then
    IS_HEALTH="notok"
    echo "  Adding Volume"
  else
    IS_HEALTH="ok"
  fi

  PODNUM=$counter
  echo "Creating Pod# $PODNUM"
  {
    addPod
    if [[ "${IS_HEALTH}" == "ok" ]]; then
      addVolume
    fi

    addLine
    addService

    # end of this Pod

    # if not last Pod add a line
    if [[ $counter -lt $POD_TOTAL ]]; then
      addLine
    fi

  } >> $POD_TEST_FILE

  ((counter++))
done

# TODO automate k8s deployment
# check if nginx namespace exists
# {
#   kubectl get ns nginx
# } > /dev/null
# if [[ $? -gt 0 ]]; then
#   echo "Creating missing nginx namespace"
#   kubectl create ns nginx
# fi

# TODO automate k8s deployment of all components
# echo "creating ConfigMap "
# kubectl apply -f configmap.yaml
# sleep 2

# kubectl -n nginx \
# apply -f $POD_TEST_FILE
