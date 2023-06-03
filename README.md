This is the source for https://gitern.com ... in summary it's
1. a multitenant git host
2. that uses ssh keys for auth exclusively
3. and has a command-line ui

I'm sharing it hoping it might be able to help someone build a git host on nostr. This isn't particularly relevant to be honest but does
1. have novel auth and
2. is very minimal

I will one day describe more about how it works (especially if someone bothers to ask), but this gist is:
- lots of SSHD hacks
- users are progressively chroot/jailed into isolated parts of the filesytem upon authentication with different sets of commands available

As a starting point, this is where auth begins (this is from the sshd config):

```txt
    Match User git
        AuthorizedKeysCommand /usr/bin/gitern-authorized-keys %f %t %k
        AuthorizedKeysCommandUser authorized-keys-command
        ForceCommand gitern-intake
        AcceptEnv none
        ChrootDirectory /jail
```

They then get forced through `gitern-intake` whose source is `/cmd/intake/intake.go`.

That should be enough for someone curious to start. Huzzah!

Extra credit
---------------
I also wrote a remote helper for gitern that end to end encrypts git repos: https://github.com/huumn/git-remote-gitern ... It's pretty inefficient and naive but it kind of shows how you might do something ***really freaking weird*** with git that runs exclusively on the client (no special git server required).


