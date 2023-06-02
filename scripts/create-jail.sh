#!/usr/bin/env bash
jail=/jail
accounts=/gitern/accounts

# create the jail's commissary
mkdir -v -p -m 755 $jail
mkdir -v -p -m 755 $jail/{accounts,dev,git-shell-commands,git-shell-commands-MEDSEC,/usr/share/git-core/templates,/etc/git/hooks}

## populate commissary
# binaries
cp -v --parents /usr/bin/git /usr/bin/git-* $jail
# libraries
libs=$(ldd /usr/bin/git* | egrep -o '/lib.*?\.[0-9]*' | sort --unique)
for i in $libs; do
  sudo cp -v --parents $i $jail
done
# dev
mknod -m 666 $jail/dev/null c 1 3
mknod -m 666 $jail/dev/zero c 1 5
mknod -m 666 $jail/dev/tty  c 5 0
mknod -m 666 $jail/dev/random c 1 8
mknod -m 666 $jail/dev/urandom c 1 9
chown root:tty $jail/dev/tty

# etc/resolv.conf for resolving rds domain name
# ssl certs to call stripe api
cp -r -v --parents /etc/pki /etc/ssl /etc/resolv.conf $jail

# git configure --global (ie HOME) to use /etc/git/hooks
git config --file $jail/etc/gitconfig core.hooksPath /etc/git/hooks

## mount accounts in cell
if ! findmnt --mtab --target $jail/accounts; then
  mount -v -o bind,nosuid $accounts $jail/accounts
fi