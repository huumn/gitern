#!/usr/bin/env bash
LDFLAGS="-X gitern/stripehelper.STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY} -X gitern/db.RDS_DB_NAME=${RDS_DB_NAME} -X gitern/db.RDS_USERNAME=${RDS_USERNAME} -X gitern/db.RDS_PASSWORD=${RDS_PASSWORD} -X gitern/db.RDS_HOSTNAME=${RDS_HOSTNAME} -X gitern/db.RDS_PORT=${RDS_PORT}"
HOOKS_PATH=/jail/etc/git/hooks

## pre-receive
go build -ldflags "${LDFLAGS}" -o $HOOKS_PATH/pre-receive pre-receive/*