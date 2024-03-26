#!/bin/sh
MW_KUBE_AGENT_HOME_GO=/usr/local/bin/mw-agent-kube-go
export MW_KUBE_AGENT_HOME_GO

sudo su << EOSUDO
mkdir -p $MW_KUBE_AGENT_HOME_GO
touch -p $MW_KUBE_AGENT_HOME_GO/agent.yaml
wget -O $MW_KUBE_AGENT_HOME_GO/agent.yaml https://kube.middleware.io/scripts/mw-kube-agent.yaml
EOSUDO

if [ -z "${MW_KUBECONFIG}" ]; then
    sed -e 's|MW_API_KEY_VALUE|'${MW_API_KEY}'|g' -e 's|TARGET_VALUE|'${TARGET}'|g' -e 's|NAMESPACE_VALUE|mw-agent-ns-'${MW_API_KEY:0:5}'|g' $MW_KUBE_AGENT_HOME_GO/agent.yaml | sudo tee $MW_KUBE_AGENT_HOME_GO/agent.yaml
    kubectl apply --kubeconfig=${MW_KUBECONFIG}  -f $MW_KUBE_AGENT_HOME_GO/agent.yaml
else
    sed -e 's|MW_API_KEY_VALUE|'${MW_API_KEY}'|g' -e 's|TARGET_VALUE|'${TARGET}'|g' -e 's|NAMESPACE_VALUE|mw-agent-ns-'${MW_API_KEY:0:5}'|g' $MW_KUBE_AGENT_HOME_GO/agent.yaml | sudo tee $MW_KUBE_AGENT_HOME_GO/agent.yaml
    kubectl apply -f $MW_KUBE_AGENT_HOME_GO/agent.yaml
fi


echo '
  MW Kube Agent Installed Successfully !
  --------------------------------------------------
  /usr/local/bin 
    └───mw-kube-agent-go
            └───agent.yaml: Contains definitions for all required kubernetes components for Agent
'
