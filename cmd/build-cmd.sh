#!/usr/bin/env bash
# Elastic Beanstalk provides these environment variables at this build stage and
# directly to webapp, so we build them into the binary
LDFLAGS="-X gitern/stripehelper.STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY} -X gitern/db.RDS_DB_NAME=${RDS_DB_NAME} -X gitern/db.RDS_USERNAME=${RDS_USERNAME} -X gitern/db.RDS_PASSWORD=${RDS_PASSWORD} -X gitern/db.RDS_HOSTNAME=${RDS_HOSTNAME} -X gitern/db.RDS_PORT=${RDS_PORT}"
PREFIX=gitern
# min and med security
CMDS_MIN=/jail/git-shell-commands
CMDS_MED=/jail/git-shell-commands-MEDSEC

## remove old commands from commissary just in case the names changed
# rm -rf ${CMDS_MIN}/${PREFIX}-* ${CMDS_MED}/${PREFIX}-*

# don't allow serfs to execute modifying commands
lordsOnly () {
    chmod 750 $1
    chown root:git $1
}

## authorized-keys
go build -ldflags "${LDFLAGS}" -o /usr/bin/${PREFIX}-authorized-keys authorized-keys/*

## dump-args
go build -ldflags "${LDFLAGS}" -o ${CMDS_MED}/${PREFIX}-dump-args dump-args/*
lordsOnly ${CMDS_MED}/${PREFIX}-dump-args

## create
go build -ldflags "${LDFLAGS}" -o ${CMDS_MED}/${PREFIX}-create create/*
lordsOnly ${CMDS_MED}/${PREFIX}-create

## list is in both LVL1 (listing all accounts) and LVL2
go build -ldflags "${LDFLAGS}" -o ${CMDS_MED}/${PREFIX}-list list/*
go build -ldflags "${LDFLAGS}" -o ${CMDS_MIN}/${PREFIX}-list list/*

## account is in both LVL1 (one unspecified account) and LVL2 (a specified account)
go build -ldflags "${LDFLAGS}" -o ${CMDS_MED}/${PREFIX}-account account/*
go build -ldflags "${LDFLAGS}" -o ${CMDS_MIN}/${PREFIX}-account account/*
lordsOnly ${CMDS_MIN}/${PREFIX}-account
lordsOnly ${CMDS_MED}/${PREFIX}-account

## delete
go build -ldflags "${LDFLAGS}" -o ${CMDS_MED}/${PREFIX}-delete delete/*
lordsOnly ${CMDS_MED}/${PREFIX}-delete

## no-interactive-login
go build -o ${CMDS_MIN}/no-interactive-login no-interactive-login/*

## intake-command
INTAKE_COMMAND=${CMDS_MIN}/${PREFIX}-intake
go build -ldflags "${LDFLAGS}" -o ${INTAKE_COMMAND} intake/*
# setuid: intake command sets up an account cell and jails them if needed
chmod 4755 ${INTAKE_COMMAND}

## pubkey add
go build -ldflags "${LDFLAGS}" -o ${CMDS_MIN}/${PREFIX}-pubkey-add pubkey/add/*
lordsOnly ${CMDS_MIN}/${PREFIX}-pubkey-add

## pubkey list
go build -ldflags "${LDFLAGS}" -o ${CMDS_MIN}/${PREFIX}-pubkey-list pubkey/list/*
lordsOnly ${CMDS_MIN}/${PREFIX}-pubkey-list

## pubkey remove
go build -ldflags "${LDFLAGS}" -o ${CMDS_MIN}/${PREFIX}-pubkey-remove pubkey/remove/*
lordsOnly ${CMDS_MIN}/${PREFIX}-pubkey-remove