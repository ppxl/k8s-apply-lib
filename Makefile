ARTIFACT_ID=k8s-apply-lib
VERSION=0.1.0
GOTAG?=1.18
MAKEFILES_VERSION=6.0.1
.DEFAULT_GOAL:=default

include build/make/variables.mk
include build/make/self-update.mk
include build/make/dependencies-gomod.mk
include build/make/build.mk
include build/make/test-common.mk
include build/make/test-unit.mk
include build/make/static-analysis.mk
include build/make/clean.mk

K8S_RUN_PRE_TARGETS=install setup-etcd-port-forward
PRE_COMPILE=vet

.PHONY: default
default: unit-test vet