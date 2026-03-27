# Flow

the following is the expected user flow for cowork

`cowork` is the cli command. The root commands are:


1. `init` - intiailize the current repos as a cowork repos. If it is not a git repos or if it is already a cowork repos throw a soft error. have a flag to avoid error if its already init'ed.

2. `config` - to be described below

3. `task` - to be described below

4. `go` - to be described below


Command explainers:

## `config`
`
config handles the configuration of global and local cw setups. acts as an organization spot for a lot of the existing commands.

All config attributes, settings, keys, etc follow the order of:

1. flags passed to the param
2. local .cw folder configs
3. global .cw configs.

### `show`

shows the current configuration settings in a human readable manner. Never show keys

### `provider`

handles all settings for the providers (github, gitlab, bitbucket)

### `git`

handles all git settings

### `container`

Handles all docker/podman settings.

### `agent`

alows the user to create settings for the agent (docker image, command structure, keys, etc)

### `env`

Allows the user to save and store .env vars that are passed to the container at run time. These are stored encrypted. Can only be local, not global

### `save` and `load

Save/load yaml files for the configuration settings. Can only be done to local settings, not global.


## `task`

Does everyting we'd expect tasks to do now. Move workspace commands to be a sub part of tasks, since we won't allow workspaces to be created without tasks anymore. Workspaces and tasks must always be synced up (and thus always have the same id, etc)


### `list`

lists all tasks

### `sync`

Sync down tasks and statuses from associated git provider

### `describe`

### `priority`

a set of commands to change a tasks priority. `freeze` is a special priority that prevents it from being ran until the user runs `unfreeze`.

### `start`

starts the task, with the agent, if it can. reports status back

### `stop`

stops the task (pause if its possible, otherwise full stop)

### `kill`

forcibly kill the agent container if its working

### `logs`

Show the log output of the agent as it works. --tail or -t will go ahead and contnually output it.
