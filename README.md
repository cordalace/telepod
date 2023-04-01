# telepod

**Telepod** adds Telegram notifications feature to
[podman auto-update](https://docs.podman.io/en/latest/markdown/podman-auto-update.1.html).

Telepod looks up for running containers and checks if an image has changed since last run.
If the image has been updated Telepod sends a telegram message.

## Environment variables

- `TELEGRAM_CHAT_ID`
- `TELEGRAM_BOT_TOKEN`

## Use with cron

```
*/5 *  * * *   TELEGRAM_CHAT_ID=-1003920707150 TELEGRAM_BOT_TOKEN=7917343448:906-MIX0YZRc6cHETY5KEmrDBA5fHW4Ye26 /usr/local/bin/telepod
```

## Use with systemd

File `~/.config/systemd/user/telepod.timer`:

```
[Unit]
Description=Telepod image update notifications

[Timer]
OnCalendar=*:0/5
Unit=telepod.target

[Install]
WantedBy=timers.target
```

File `~/.config/systemd/user/telepod.service`:

```
[Unit]
Description=Telepod image update notifications
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/telepod
```

## User

Telepod tracks container versions of the current user. It could be
started as root or as regular user.

## Versions file

Telepod tracks image versions in json-file:

- `/var/lib/telepod/db.json` if started as root
- `$XDG_STATE_HOME/telepod/db.json` if started as regular user

