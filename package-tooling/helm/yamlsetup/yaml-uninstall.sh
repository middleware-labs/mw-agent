#!/bin/sh
MW_KUBE_AGENT_HOME_GO=/usr/local/bin/mw-agent-kube-go
export MW_KUBE_AGENT_HOME_GO

if [ -z "${MW_KUBECONFIG}" ]; then
    kubectl delete --kubeconfig=${MW_KUBECONFIG}  -f /usr/local/bin/mw-kube-agent/agent.yaml
else
    kubectl delete -f /usr/local/bin/mw-kube-agent/agent.yaml
fi

sudo su << EOSUDO
rm -rf $MW_KUBE_AGENT_HOME_GO
EOSUDO