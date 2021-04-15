load('ext://restart_process', 'docker_build_with_restart')

# Records the current time, then kicks off a server update.
# Normally, you would let Tilt do deploys automatically, but this
# shows you how to set up a custom workflow that measures it.
local_resource(
    'deploy',
    './tilt/record-start-time.py',
)              

local_resource(
  'healthcat-compile',
  'BINARY_NAME=healthcat CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build',
  deps=['./server', './main.go', './start.go'],
  resource_deps = ['deploy'])

docker_build_with_restart(
  '715824223074.dkr.ecr.us-east-1.amazonaws.com/healthcat',
  '.',
  entrypoint='/healthcat --cluster-id=catdev --port=80 --namespaces=nginx,dimitar --log-preset=prod',
  dockerfile='tilt/Dockerfile',
  only=[
    './build'
  ],
  live_update=[
    sync('./build/bin/linux/', '/'),
  ])


allow_k8s_contexts('edpub-us-east-1.nonprod.edpub.wiley.com')

yaml = helm(
  'helm/',
  # The release name, equivalent to helm --name
  name='healthcat',
  # The namespace to install in, equivalent to helm --namespace
  namespace='healthcat',
  # The values file to substitute into the chart.
  values=['./helm_vars/edpub/nonprod/values.yaml'],
)

k8s_yaml(yaml)
k8s_resource(workload='healthcat', 
              port_forwards='8080:80',
              resource_deps=['deploy', 'healthcat-compile'])

