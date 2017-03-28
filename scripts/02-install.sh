#!/bin/bash
## --------------------------------------------------------------------------------------------------------------------

set -e

echo "Checking ask.sh is installed ..."
if [ ! /home/chilts/bin/ask.sh ]; then
    echo "Please put ask.sh into ~/bin (should already be in your path from ~/.profile):"
    echo ""
    echo "    mkdir ~/bin"
    echo "    wget -O ~/bin/ask.sh https://gist.githubusercontent.com/chilts/6b547307a6717d53e14f7403d58849dd/raw/ecead4db87ad4e7674efac5ab0e7a04845be642c/ask.sh"
    echo "    chmod +x ~/bin/ask.sh"
    echo ""
    exit 2
fi
echo

# General
WHO=`whoami`
PASTE_PORT=`ask.sh paste PASTE_PORT 'Which local port should the server listen on (e.g. 8420):'`
PASTE_APEX=`ask.sh paste PASTE_APEX 'What is the apex (e.g. localhost:8420 or paste.gd) :'`
PASTE_BASE_URL=`ask.sh paste PASTE_BASE_URL 'What is the base URL (e.g. http://localhost:1234 or https://paste.gd) :'`
PASTE_DIR=`ask.sh paste PASTE_DIR 'What is the storage dir (e.g. /var/lib/paste/raw) :'`
PASTE_DUMP_DIR=`ask.sh paste PASTE_DUMP_DIR 'What is the dump dir (e.g. /var/lib/paste/dump) :'`
PASTE_GOOGLE_ANALYTICS=`ask.sh paste PASTE_GOOGLE_ANALYTICS 'What is the Google Analytics code (e.g. UA-123-4) :'`

echo "Building code ..."
gb build
echo

echo "Minifying assets ..."
make minify
echo

echo "Creating storage dir ..."
sudo mkdir -p $PASTE_DIR
sudo chown ${WHO}.${WHO} $PASTE_DIR
sudo mkdir -p $PASTE_DUMP_DIR
sudo chown ${WHO}.${WHO} $PASTE_DUMP_DIR
echo

# copy the supervisor script into place
echo "Copying supervisor config ..."
m4 \
    -D __PASTE_PORT__=$PASTE_PORT \
    -D __PASTE_APEX__=$PASTE_APEX \
    -D __PASTE_BASE_URL__=$PASTE_BASE_URL \
    -D __PASTE_DIR__=$PASTE_DIR \
    -D __PASTE_GOOGLE_ANALYTICS__=$PASTE_GOOGLE_ANALYTICS \
    etc/supervisor/conf.d/paste.conf.m4 | sudo tee /etc/supervisor/conf.d/paste.conf
echo

# restart supervisor
echo "Restarting supervisor ..."
sudo systemctl restart supervisor.service
echo

# copy the caddy conf
echo "Copying Caddy config config ..."
m4 \
    -D __PASTE_PORT__=$PASTE_PORT \
    -D __PASTE_APEX__=$PASTE_APEX \
    etc/caddy/vhosts/paste.conf.m4 | sudo tee /etc/caddy/vhosts/paste.conf
echo

# restarting Caddy
echo "Restarting caddy ..."
sudo systemctl restart caddy.service
echo

## --------------------------------------------------------------------------------------------------------------------
