__PASTE_APEX__ {
  proxy / localhost:__PASTE_PORT__ {
    transparent
  }
  tls chilts@appsattic.com
  log stdout
  errors stderr
}

www.__PASTE_APEX__ {
  redir http://__PASTE_APEX__{uri} 302
}
