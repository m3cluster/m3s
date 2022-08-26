#!/bin/bash

export KUBECONFIG=$MESOS_SANDBOX/kubeconfig.yaml
export BRANCH=update-k3s

if [ -n $1 ]
then
  kill -9 $1 && curl https://raw.githubusercontent.com/AVENTER-UG/mesos-m3s/${BRANCH}/bootstrap/server > $MESOS_SANDBOX/server
  exec ./server &
fi
