[program:paste]
directory = /home/chilts/src/appsattic-paste.gd
command = /home/chilts/src/appsattic-paste.gd/bin/paste
user = chilts
autostart = true
autorestart = true
stdout_logfile = /var/log/chilts/paste-stdout.log
stderr_logfile = /var/log/chilts/paste-stderr.log
environment =
    PASTE_PORT="__PASTE_PORT__",
    PASTE_APEX="__PASTE_APEX__",
    PASTE_BASE_URL="__PASTE_BASE_URL__",
    PASTE_DIR="__PASTE_DIR__",
    PASTE_GOOGLE_ANALYTICS="__PASTE_GOOGLE_ANALYTICS__"
