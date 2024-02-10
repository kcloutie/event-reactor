#!/usr/bin/env bash

set -euf
cd $(dirname $(readlink -f ${0}))

export KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-kind}
export KUBECONFIG=${HOME}/.kube/config.${KIND_CLUSTER_NAME}
export TARGET=kubernetes
export DOMAIN_NAME=paac-127-0-0-1.nip.io

if ! builtin type -p kind &>/dev/null; then
    echo "Install kind. https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
    exit 1
fi
kind=$(type -p kind)
if ! builtin type -p ko &>/dev/null; then
    echo "Install ko. https://ko.build/install/"
    exit 1
fi
ko=$(type -p ko)
if ! builtin type -p gosmee &>/dev/null; then
    echo "Install gosmee. https://github.com/chmouel/gosmee?tab=readme-ov-file#install"
    exit 1
fi

TMPD=$(mktemp -d /tmp/.GITXXXX)
REG_PORT='5000'
REG_NAME='kind-registry'
INSTALL_FROM_RELEASE=
SUDO=sudo
EVENT_REACTOR_DIR=${EVENT_REACTOR_DIR:-""}


[[ $(uname -s) == "Darwin" ]] && {
    SUDO=
}

# cleanup on exit (useful for running locally)
cleanup() { rm -rf ${TMPD} ;}
trap cleanup EXIT

function start_registry() {
    running="$(docker inspect -f '{{.State.Running}}' ${REG_NAME} 2>/dev/null || echo false)"

    if [[ ${running} != "true" ]];then
        docker rm -f kind-registry || true
        docker run \
               -d --restart=always -p "127.0.0.1:${REG_PORT}:5000" \
               -e REGISTRY_HTTP_SECRET=secret \
               --name "${REG_NAME}" \
               registry:2
    fi
}

function reinstall_kind() {
	${SUDO} $kind delete cluster --name ${KIND_CLUSTER_NAME} || true
	sed "s,%DOCKERCFG%,${HOME}/.docker/config.json,"  kind.yaml > ${TMPD}/kconfig.yaml

       cat <<EOF >> ${TMPD}/kconfig.yaml
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${REG_PORT}"]
    endpoint = ["http://${REG_NAME}:5000"]
EOF

	${SUDO} ${kind} create cluster --name ${KIND_CLUSTER_NAME} --config  ${TMPD}/kconfig.yaml
	mkdir -p $(dirname ${KUBECONFIG})
	${SUDO} ${kind} --name ${KIND_CLUSTER_NAME} get kubeconfig > ${KUBECONFIG}


    docker network connect "kind" "${REG_NAME}" 2>/dev/null || true
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${REG_PORT}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

}

function install_nginx() {
    echo "Installing nginx"
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml >/dev/null
    i=0
    echo -n "Waiting for nginx to come up: "
	while true;do
		[[ ${i} == 120 ]] && exit 1
		ep=$(kubectl wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=180s 2>/dev/null || true)
		[[ -n ${ep} ]] && break
		sleep 5
		i=$((i+1))
	done
    echo "done."
}

function install_event-reactor() {
    [[ -z ${EVENT_REACTOR_DIR} && $(git rev-parse --show-toplevel 2>/dev/null) != "" ]] && \
        EVENT_REACTOR_DIR=$(git rev-parse --show-toplevel)

    [[ -z ${EVENT_REACTOR_DIR} && $(git rev-parse --show-toplevel 2>/dev/null) == "" ]] &&  \
        EVENT_REACTOR_DIR=$GOPATH/src/github.com/kcloutie/event-reactor

	if [[ -n ${INSTALL_FROM_RELEASE} ]];then
		kubectl apply -f ${PAC_RELEASE:-https://github.com/kcloutie/event-reactor/raw/stable/release.k8s.yaml}
	else
        [[ -d ${EVENT_REACTOR_DIR} ]] || {
            echo "I cannot find the EVENT_REACTOR installation directory, set the variable \$EVENT_REACTOR_DIR to define it. or launch this script from inside where the pac code is checkout"
            exit 1
        }
        oldPwd=${PWD}
        cd ${EVENT_REACTOR_DIR}
        echo "Deploying EVENT_REACTOR from ${EVENT_REACTOR_DIR}"
        [[ -n ${PAC_DEPLOY_SCRIPT:-""} ]] && ${PAC_DEPLOY_SCRIPT} || env KO_DOCKER_REPO=localhost:5000 $ko apply -f config --sbom=none -B >/dev/null
        cd ${oldPwd}
    fi
    configure_event-reactor
    echo "application: http://app.${DOMAIN_NAME}"
}

function remove_event-reactor() {
    [[ -z ${EVENT_REACTOR_DIR} && $(git rev-parse --show-toplevel 2>/dev/null) != "" ]] && \
        EVENT_REACTOR_DIR=$(git rev-parse --show-toplevel)

    [[ -z ${EVENT_REACTOR_DIR} && $(git rev-parse --show-toplevel 2>/dev/null) == "" ]] &&  \
        EVENT_REACTOR_DIR=$GOPATH/src/github.com/kcloutie/event-reactor

	if [[ -n ${INSTALL_FROM_RELEASE} ]];then
		kubectl delete -f ${PAC_RELEASE:-https://github.com/kcloutie/event-reactor/raw/stable/release.k8s.yaml}
	else
        [[ -d ${EVENT_REACTOR_DIR} ]] || {
            echo "I cannot find the EVENT_REACTOR installation directory, set the variable \$EVENT_REACTOR_DIR to define it. or launch this script from inside where the pac code is checkout"
            exit 1
        }
        oldPwd=${PWD}
        cd ${EVENT_REACTOR_DIR}
        echo "Removing EVENT_REACTOR from ${EVENT_REACTOR_DIR}"
        [[ -n ${PAC_DEPLOY_SCRIPT:-""} ]] && ${PAC_DEPLOY_SCRIPT} || env KO_DOCKER_REPO=localhost:5000 $ko delete -f config >/dev/null
        cd ${oldPwd}
    fi
 
    echo "application has been removed"
}

function configure_event-reactor() {
    echo "Configuring EVENT_REACTOR"
}



main() {
    start_registry
	reinstall_kind
	install_nginx
	install_event-reactor
    echo "And we are done :): "
}

function usage() {
    cat <<EOF
Usage: $0 [OPTIONS]

Options:
  -h          Show this message
  -b          Only install the registry/kind/nginx and don't install EVENT_REACTOR
  -c          Configure EVENT_REACTOR
  -p          Install only EVENT_REACTOR
  -r          Install from release instead of local checkout with ko
  -R          Restart the EVENT_REACTOR pods
  -d          Delete EVENT_REACTOR deployment
EOF
}

while getopts "hbcpdRr" o; do
    case "${o}" in
        h)
            usage
            exit
            ;;
        b)
            start_registry
            reinstall_kind
            install_nginx
            exit
            ;;
        c)
            configure_event-reactor
            exit
            ;;
        p)
            install_event-reactor
            exit
            ;;
        d)
            remove_event-reactor
            exit
            ;;
        R)

            echo "Restarting event-reactor pods"
            kubectl delete pod -l app.kubernetes.io/part-of=event-reactor -n event-reactor || true
            ;;
	    r)
		    INSTALL_FROM_RELEASE=yes
            ;;

        *)
            echo "Invalid option"; exit 1;
            ;;
    esac
done
shift $((OPTIND-1))

main
