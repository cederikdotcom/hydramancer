# Gallo-Romeins Perforce: getting started for developers

This guide helps a new Cyborn developer connect to the Gallo-Romeins Museum
Perforce server and deliver an Unreal build. Follow the steps in order.

You do not need SSH. You connect only with the Perforce client on port 1666.

## What you need

- Your Perforce user name and a temporary password. The admin sends these to you.
- The Helix Visual Client (P4V), or the `p4` command-line client.

Server details:

| Field | Value |
|-------|-------|
| Server (P4PORT) | `ssl:perforce.galloromeins.experiencenet.com:1666` |
| Depot | `//galloromeinsmuseum/main` (a stream depot) |
| Your access | write, in `//galloromeinsmuseum/` only |

## Step 1: install P4V

Download P4V from https://www.perforce.com/downloads/helix-visual-client-p4v
and install it. P4V is the easiest way to start.

## Step 2: connect and trust the server

1. Open P4V. Open **Connection > Set Up Connection**.
2. Set **Server** to `ssl:perforce.galloromeins.experiencenet.com:1666`.
3. Set **User** to your user name.
4. Accept the SSL fingerprint when P4V asks. This is normal on the first connect.

Command line:

```bash
p4 set P4PORT=ssl:perforce.galloromeins.experiencenet.com:1666
p4 set P4USER=<your-user>
p4 trust -y
```

## Step 3: log in and set your own password

1. Log in with the temporary password.
2. The server asks you to set a new password at once. Set a strong one.

Command line:

```bash
p4 login          # type the temporary password
p4 passwd         # set your own password
```

The server uses Security Level 3. Your password must be strong.

## Step 4: make a stream workspace

You must bind your workspace to the stream. A normal workspace cannot submit to a
stream depot.

In P4V:

1. Open **View > Workspaces**, then **New Workspace**.
2. Select **Stream** and pick `//galloromeinsmuseum/main`.
3. Set a local root folder for your files.

Command line:

```bash
p4 client -S //galloromeinsmuseum/main <your-workspace-name>
```

## Step 5: get the files and submit your work

```bash
p4 sync                      # get the latest files
# add or change files, then:
p4 reconcile                 # detect adds, edits, and deletes
p4 submit -d "your message"  # send your change to the server
```

## Rules for builds

- Put packaged builds under `//galloromeinsmuseum/main/Builds/`. The build watcher
  reads this folder. It publishes new builds there to the render fleet on its own.
- If you also submit the Unreal project source, turn on Source Control in the
  editor and set it to Perforce.
- Keep `Binaries/`, `Build/`, and `Intermediate/` out of Perforce.
- Do not exclude `Binaries/ThirdParty/`. It holds NVIDIA and DLSS files that you
  cannot rebuild.

## Get help

If you cannot connect or submit, contact the admin (cederik@experiencenet.com).
Give the exact error text.

---

## Appendix: how an admin adds a new Cyborn developer

The server uses the free license. It allows 5 users. Three seats are in use
(`hydra_admin`, `hydraperforce_svc`, `koen`). Two seats are free.

To add a developer:

```bash
ssh root@195.201.88.170
export P4PORT=ssl:localhost:1666 P4USER=hydra_admin
p4 login                                   # hydra_admin password

p4 user -f -i <<'U'                        # create the account
User: <newuser>
Email: <email>
FullName: <name> (Cyborn)
U

printf 'TEMP\nTEMP\n' | p4 passwd <newuser>  # set a temporary password

p4 group -o CybornContrib | \
  sed "/^Users:/a\\\t<newuser>" | p4 group -i   # add to the write group
```

The new user joins group `CybornContrib`. That group already has write access to
`//galloromeinsmuseum/...`. You do not need to change the protections table.

Security Level 3 marks an admin-set password as "must reset on first login". This
is what you want for a temporary password. Send the temporary password to the new
developer. Ask them to set their own on first login.

If all 5 seats are full, free one first. Delete a dormant user, or buy a license.
See `galloromeins-perforce.md` in `docs/runbooks/`.
